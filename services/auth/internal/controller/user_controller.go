// Package controller holds the controller-runtime reconcilers for the auth
// service (User and Group custom resources).
package controller

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	authv1alpha1 "gitlab.telespazio-digital-factory.fr/icdo/tpzf/crossplane-ui/pkg/apis/auth/v1alpha1"
	"gitlab.telespazio-digital-factory.fr/icdo/tpzf/crossplane-ui/services/auth/internal/dex"
	"gitlab.telespazio-digital-factory.fr/icdo/tpzf/crossplane-ui/services/auth/internal/password"
)

const (
	passwordSecretKey = "password"
	// consumedAnnotation marks a Secret whose password has already been
	// consumed by the controller. It avoids rehashing the same plaintext over
	// and over when we leave the "password" key untouched.
	consumedAnnotation = "auth.crossplane-ui.io/password-consumed"

	reasonMissingSecret = "PasswordSecretMissing"
	reasonNoPassword    = "NoPassword"
	reasonProjected     = "ProjectedToDex"
)

// UserReconciler reconciles User custom resources.
//
// Its loop is small:
//  1. If Spec.PasswordSecretRef points at a Secret carrying a password key,
//     consume it (bcrypt → Status.PasswordHash; blank out the key to avoid
//     leaving plaintext at rest).
//  2. Ensure Status.UserID is set.
//  3. Rewrite the Dex Password objects from the current User set.
type UserReconciler struct {
	Client    client.Client
	Scheme    *runtime.Scheme
	Namespace string
}

// SetupWithManager wires the reconciler up with the manager and asks for a
// requeue on Dex Password events so that an external edit of a managed
// Password object is detected and overwritten.
func (r *UserReconciler) SetupWithManager(mgr ctrl.Manager) error {
	passwordList := &unstructured.Unstructured{}
	passwordList.SetGroupVersionKind(dex.PasswordGVK)

	return ctrl.NewControllerManagedBy(mgr).
		Named("user").
		For(&authv1alpha1.User{}).
		WatchesRawSource(
			source.Kind(
				mgr.GetCache(),
				client.Object(passwordList),
				handler.EnqueueRequestsFromMapFunc(r.mapPasswordToUsers),
				managedByPredicate{},
			),
		).
		Complete(r)
}

// Reconcile implements reconcile.Reconciler.
func (r *UserReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	logger := log.FromContext(ctx).WithValues("user", req.Name)

	var u authv1alpha1.User
	if err := r.Client.Get(ctx, req.NamespacedName, &u); err != nil {
		if apierrors.IsNotFound(err) {
			// Deletion — resync so the user disappears from Dex.
			return reconcile.Result{}, r.syncDex(ctx, logger)
		}
		return reconcile.Result{}, fmt.Errorf("get user: %w", err)
	}

	patch := client.MergeFrom(u.DeepCopy())

	if u.Status.UserID == "" {
		u.Status.UserID = uuid.NewString()
	}

	if err := r.consumePasswordSecret(ctx, &u); err != nil {
		setCondition(&u.Status.Conditions, metav1.ConditionFalse, reasonMissingSecret, err.Error(), u.Generation)
		_ = r.Client.Status().Patch(ctx, &u, patch)
		return reconcile.Result{}, err
	}

	u.Status.ObservedGeneration = u.Generation

	switch u.Status.PasswordHash {
	case "":
		setCondition(&u.Status.Conditions, metav1.ConditionFalse, reasonNoPassword,
			"no password hash set; populate spec.passwordSecretRef or status.passwordHash", u.Generation)
	default:
		setCondition(&u.Status.Conditions, metav1.ConditionTrue, reasonProjected,
			"user projected into Dex password storage", u.Generation)
	}

	if err := r.Client.Status().Patch(ctx, &u, patch); err != nil {
		return reconcile.Result{}, fmt.Errorf("patch user status: %w", err)
	}

	if err := r.syncDex(ctx, logger); err != nil {
		return reconcile.Result{}, err
	}
	return reconcile.Result{}, nil
}

// consumePasswordSecret turns spec.passwordSecretRef into status.passwordHash.
// Both arms of the function are safe to call repeatedly; only a non-empty
// password in the referenced Secret produces a new hash.
func (r *UserReconciler) consumePasswordSecret(ctx context.Context, u *authv1alpha1.User) error {
	if u.Spec.PasswordSecretRef == nil {
		return nil
	}
	name := u.Spec.PasswordSecretRef.Name

	var sec corev1.Secret
	err := r.Client.Get(ctx, types.NamespacedName{Namespace: r.Namespace, Name: name}, &sec)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return fmt.Errorf("secret %q not found in %q", name, r.Namespace)
		}
		return fmt.Errorf("get secret %q: %w", name, err)
	}

	raw, ok := sec.Data[passwordSecretKey]
	if !ok || len(raw) == 0 {
		// Already consumed (or never populated). Nothing to do.
		return nil
	}

	// If the caller already gave us a bcrypt hash inside the Secret, trust it.
	// Otherwise hash the plaintext.
	plaintext := string(raw)
	var hash string
	if password.IsBcrypt(plaintext) {
		hash = plaintext
	} else {
		h, err := password.Hash(plaintext)
		if err != nil {
			return fmt.Errorf("hash password: %w", err)
		}
		hash = h
	}
	u.Status.PasswordHash = hash

	// Blank out the plaintext key and record consumption so we do not rehash
	// the same value on every reconcile.
	secPatch := client.MergeFrom(sec.DeepCopy())
	delete(sec.Data, passwordSecretKey)
	if sec.Annotations == nil {
		sec.Annotations = map[string]string{}
	}
	sec.Annotations[consumedAnnotation] = metav1.Now().UTC().Format("2006-01-02T15:04:05Z07:00")
	if err := r.Client.Patch(ctx, &sec, secPatch); err != nil {
		return fmt.Errorf("scrub password from secret %q: %w", name, err)
	}
	return nil
}

func (r *UserReconciler) syncDex(ctx context.Context, logger logr.Logger) error {
	var list authv1alpha1.UserList
	if err := r.Client.List(ctx, &list); err != nil {
		return fmt.Errorf("list users: %w", err)
	}
	changed, err := dex.Sync(ctx, r.Client, r.Namespace, list.Items)
	if err != nil {
		logger.Error(err, "dex sync failed", "namespace", r.Namespace)
		return err
	}
	if changed {
		logger.Info("dex password objects reconciled",
			"namespace", r.Namespace,
			"users", len(list.Items))
	}
	return nil
}

// mapPasswordToUsers maps a Dex Password event to a list of User reconcile
// requests so the sync loop reconstructs the collection whenever something
// upstream diverges from the desired state.
func (r *UserReconciler) mapPasswordToUsers(ctx context.Context, _ client.Object) []reconcile.Request {
	var list authv1alpha1.UserList
	if err := r.Client.List(ctx, &list); err != nil {
		return nil
	}
	out := make([]reconcile.Request, 0, len(list.Items))
	for i := range list.Items {
		out = append(out, reconcile.Request{NamespacedName: types.NamespacedName{Name: list.Items[i].Name}})
	}
	if len(out) == 0 {
		// Trigger at least one reconcile so the empty-list path prunes any
		// leftover Password that no longer belongs to a live User.
		out = append(out, reconcile.Request{NamespacedName: types.NamespacedName{Name: "__rebuild__"}})
	}
	return out
}
