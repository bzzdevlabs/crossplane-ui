package controller

import (
	"context"
	"fmt"
	"sort"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	authv1alpha1 "gitlab.telespazio-digital-factory.fr/icdo/tpzf/crossplane-ui/pkg/apis/auth/v1alpha1"
)

// GroupReconciler keeps Group.status.members in sync with User.spec.groups.
// It is deliberately read-heavy and write-light: every group change rescans
// the full User list, and every User change triggers the group(s) it points
// at.
type GroupReconciler struct {
	Client client.Client
	Scheme *runtime.Scheme
}

// SetupWithManager wires the reconciler up with the manager.
func (r *GroupReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named("group").
		For(&authv1alpha1.Group{}).
		Watches(&authv1alpha1.User{}, handler.EnqueueRequestsFromMapFunc(r.mapUserToGroups)).
		Complete(r)
}

// Reconcile implements reconcile.Reconciler.
func (r *GroupReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	var g authv1alpha1.Group
	if err := r.Client.Get(ctx, req.NamespacedName, &g); err != nil {
		if apierrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, fmt.Errorf("get group: %w", err)
	}

	var users authv1alpha1.UserList
	if err := r.Client.List(ctx, &users); err != nil {
		return reconcile.Result{}, fmt.Errorf("list users: %w", err)
	}

	members := make([]string, 0)
	for i := range users.Items {
		u := &users.Items[i]
		for _, gr := range u.Spec.Groups {
			if gr == g.Name {
				members = append(members, u.Spec.Username)
				break
			}
		}
	}
	sort.Strings(members)

	patch := client.MergeFrom(g.DeepCopy())
	g.Status.Members = members
	g.Status.ObservedGeneration = g.Generation
	setCondition(&g.Status.Conditions, metav1.ConditionTrue, "Reconciled",
		fmt.Sprintf("%d member(s)", len(members)), g.Generation)
	if err := r.Client.Status().Patch(ctx, &g, patch); err != nil {
		return reconcile.Result{}, fmt.Errorf("patch group status: %w", err)
	}
	return reconcile.Result{}, nil
}

// mapUserToGroups maps a User event to a Reconcile request per declared group.
func (r *GroupReconciler) mapUserToGroups(_ context.Context, obj client.Object) []reconcile.Request {
	u, ok := obj.(*authv1alpha1.User)
	if !ok {
		return nil
	}
	out := make([]reconcile.Request, 0, len(u.Spec.Groups))
	for _, gr := range u.Spec.Groups {
		out = append(out, reconcile.Request{NamespacedName: types.NamespacedName{Name: gr}})
	}
	return out
}
