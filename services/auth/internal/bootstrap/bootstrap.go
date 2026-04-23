// Package bootstrap creates the initial administrator on first startup.
//
// The auth service, when deployed through the Helm chart, comes with a Secret
// holding the bootstrap credentials. On startup, before the manager begins
// reconciling, this package:
//
//  1. Reads the Secret (username/password keys).
//  2. Ensures a matching User object exists, carrying the bcrypt hash on its
//     Status.
//  3. Scrubs the plaintext password from the Secret.
//
// Subsequent runs are idempotent: if the User already exists with a
// populated password hash the bootstrap is a no-op. This allows admins to
// change the admin password later by rotating the Secret's `password` key
// — the next controller loop will rehash and project into Dex.
package bootstrap

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	authv1alpha1 "github.com/bzzdevlabs/crossplane-ui/pkg/apis/auth/v1alpha1"
	"github.com/bzzdevlabs/crossplane-ui/services/auth/internal/password"
)

// BootstrapAdminGroup is added to .spec.groups of the auto-created admin.
const BootstrapAdminGroup = "crossplane-ui:admins"

// Config drives the one-shot bootstrap.
type Config struct {
	// Namespace is where the Secret lives.
	Namespace string
	// SecretName is the name of the Secret carrying the "username" and
	// "password" keys.
	SecretName string
	// DefaultUsername is used when the Secret does not carry a "username"
	// key (it should, but the chart default is nonetheless surfaced here).
	DefaultUsername string
}

// Run executes the bootstrap. It is safe to call repeatedly.
//
// Success semantics:
//   - the User exists, is not disabled, and carries a non-empty
//     Status.PasswordHash;
//   - the Secret no longer carries the plaintext password.
//
// If the Secret is missing and no User exists yet, Run returns an error so
// operators notice the misconfiguration. If a User already exists the
// missing-Secret case is ignored.
func Run(ctx context.Context, c client.Client, cfg Config) error {
	if err := validate(cfg); err != nil {
		return err
	}

	sec, missing, err := loadSecret(ctx, c, cfg)
	if err != nil {
		return err
	}
	username := resolveUsername(sec, cfg, missing)
	if username == "" {
		return errors.New("bootstrap: empty admin username")
	}

	existing, exists, err := loadUser(ctx, c, username)
	if err != nil {
		return err
	}

	switch {
	case missing && !exists:
		return fmt.Errorf("bootstrap: secret %q not found and no %q User exists yet", cfg.SecretName, username)
	case missing:
		// Secret gone, User already there. Controller will carry on.
		return nil
	}

	plaintext := string(sec.Data["password"])
	if !exists {
		if err := createAdminUser(ctx, c, username, plaintext); err != nil {
			return err
		}
	} else if plaintext != "" && existing.Status.PasswordHash == "" {
		if err := rehashExistingUser(ctx, c, existing, plaintext); err != nil {
			return err
		}
	}

	if plaintext != "" {
		if err := scrubSecret(ctx, c, sec); err != nil {
			return err
		}
	}
	return nil
}

func validate(cfg Config) error {
	if cfg.Namespace == "" {
		return errors.New("bootstrap: namespace is empty")
	}
	if cfg.SecretName == "" {
		return errors.New("bootstrap: secret name is empty")
	}
	return nil
}

func loadSecret(ctx context.Context, c client.Client, cfg Config) (*corev1.Secret, bool, error) {
	var sec corev1.Secret
	err := c.Get(ctx, types.NamespacedName{Namespace: cfg.Namespace, Name: cfg.SecretName}, &sec)
	if apierrors.IsNotFound(err) {
		return nil, true, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("bootstrap: get secret %q: %w", cfg.SecretName, err)
	}
	return &sec, false, nil
}

func loadUser(ctx context.Context, c client.Client, username string) (*authv1alpha1.User, bool, error) {
	var u authv1alpha1.User
	err := c.Get(ctx, types.NamespacedName{Name: username}, &u)
	if apierrors.IsNotFound(err) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("bootstrap: get user %q: %w", username, err)
	}
	return &u, true, nil
}

func resolveUsername(sec *corev1.Secret, cfg Config, missing bool) string {
	username := cfg.DefaultUsername
	if missing || sec == nil {
		return username
	}
	if u := string(sec.Data["username"]); u != "" {
		return u
	}
	return username
}

func createAdminUser(ctx context.Context, c client.Client, username, plaintext string) error {
	hash := ""
	if plaintext != "" {
		h, err := password.Hash(plaintext)
		if err != nil {
			return fmt.Errorf("bootstrap: hash password: %w", err)
		}
		hash = h
	}
	user := authv1alpha1.User{
		ObjectMeta: metav1.ObjectMeta{
			Name: username,
			Labels: map[string]string{
				"app.kubernetes.io/managed-by":    "crossplane-ui-auth",
				"auth.crossplane-ui.io/bootstrap": "true",
			},
		},
		Spec: authv1alpha1.UserSpec{
			Username: username,
			Email:    username + "@crossplane-ui.local",
			Groups:   []string{BootstrapAdminGroup},
		},
	}
	if err := c.Create(ctx, &user); err != nil {
		return fmt.Errorf("bootstrap: create user: %w", err)
	}
	if hash == "" {
		return nil
	}
	patch := client.MergeFrom(user.DeepCopy())
	user.Status.PasswordHash = hash
	user.Status.UserID = uuid.NewString()
	if err := c.Status().Patch(ctx, &user, patch); err != nil {
		return fmt.Errorf("bootstrap: patch user status: %w", err)
	}
	return nil
}

func rehashExistingUser(ctx context.Context, c client.Client, existing *authv1alpha1.User, plaintext string) error {
	hash, err := password.Hash(plaintext)
	if err != nil {
		return fmt.Errorf("bootstrap: hash password: %w", err)
	}
	patch := client.MergeFrom(existing.DeepCopy())
	existing.Status.PasswordHash = hash
	if existing.Status.UserID == "" {
		existing.Status.UserID = uuid.NewString()
	}
	if err := c.Status().Patch(ctx, existing, patch); err != nil {
		return fmt.Errorf("bootstrap: patch user status: %w", err)
	}
	return nil
}

func scrubSecret(ctx context.Context, c client.Client, sec *corev1.Secret) error {
	secPatch := client.MergeFrom(sec.DeepCopy())
	delete(sec.Data, "password")
	if sec.Annotations == nil {
		sec.Annotations = map[string]string{}
	}
	sec.Annotations["auth.crossplane-ui.io/password-consumed"] = metav1.Now().
		UTC().
		Format("2006-01-02T15:04:05Z07:00")
	if err := c.Patch(ctx, sec, secPatch); err != nil {
		return fmt.Errorf("bootstrap: scrub secret: %w", err)
	}
	return nil
}
