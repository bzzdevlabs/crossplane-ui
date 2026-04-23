package bootstrap_test

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	authv1alpha1 "github.com/bzzdevlabs/crossplane-ui/pkg/apis/auth/v1alpha1"
	"github.com/bzzdevlabs/crossplane-ui/services/auth/internal/bootstrap"
	"github.com/bzzdevlabs/crossplane-ui/services/auth/internal/kube"
	"github.com/bzzdevlabs/crossplane-ui/services/auth/internal/password"
)

func adminSecret(ns, name, user, pwd string) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: name},
		Data: map[string][]byte{
			"username": []byte(user),
			"password": []byte(pwd),
		},
	}
}

func TestRunCreatesAdminAndScrubsSecret(t *testing.T) {
	t.Parallel()

	ns := "crossplane-ui"
	sec := adminSecret(ns, "bootstrap", "admin", "hunter2")

	c := fake.NewClientBuilder().
		WithScheme(kube.Scheme).
		WithObjects(sec).
		WithStatusSubresource(&authv1alpha1.User{}).
		Build()

	err := bootstrap.Run(context.Background(), c, bootstrap.Config{
		Namespace:       ns,
		SecretName:      "bootstrap",
		DefaultUsername: "admin",
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	var got authv1alpha1.User
	if err := c.Get(context.Background(), types.NamespacedName{Name: "admin"}, &got); err != nil {
		t.Fatalf("User not created: %v", err)
	}
	if got.Status.PasswordHash == "" {
		t.Fatal("PasswordHash empty after bootstrap")
	}
	ok, verr := password.Verify(got.Status.PasswordHash, "hunter2")
	if verr != nil || !ok {
		t.Fatalf("password did not verify: ok=%v err=%v", ok, verr)
	}
	if got.Status.UserID == "" {
		t.Fatal("UserID not generated")
	}

	var after corev1.Secret
	if err := c.Get(context.Background(), types.NamespacedName{Namespace: ns, Name: "bootstrap"}, &after); err != nil {
		t.Fatalf("Get secret: %v", err)
	}
	if _, still := after.Data["password"]; still {
		t.Fatal("plaintext password was not scrubbed from the Secret")
	}
	if after.Annotations["auth.crossplane-ui.io/password-consumed"] == "" {
		t.Fatal("consumed annotation missing")
	}
}

func TestRunIsIdempotent(t *testing.T) {
	t.Parallel()

	ns := "crossplane-ui"
	sec := adminSecret(ns, "bootstrap", "admin", "hunter2")
	c := fake.NewClientBuilder().
		WithScheme(kube.Scheme).
		WithObjects(sec).
		WithStatusSubresource(&authv1alpha1.User{}).
		Build()

	cfg := bootstrap.Config{Namespace: ns, SecretName: "bootstrap", DefaultUsername: "admin"}
	if err := bootstrap.Run(context.Background(), c, cfg); err != nil {
		t.Fatalf("first Run: %v", err)
	}
	// A second call must not error out.
	if err := bootstrap.Run(context.Background(), c, cfg); err != nil {
		t.Fatalf("second Run: %v", err)
	}
}

func TestRunErrorsWhenSecretAndUserMissing(t *testing.T) {
	t.Parallel()

	c := fake.NewClientBuilder().
		WithScheme(kube.Scheme).
		WithStatusSubresource(&authv1alpha1.User{}).
		Build()

	err := bootstrap.Run(context.Background(), c, bootstrap.Config{
		Namespace: "x", SecretName: "missing", DefaultUsername: "admin",
	})
	if err == nil {
		t.Fatal("Run returned nil when Secret and User both absent")
	}
}

func TestRunSkipsWhenSecretMissingButUserExists(t *testing.T) {
	t.Parallel()

	existing := &authv1alpha1.User{
		ObjectMeta: metav1.ObjectMeta{Name: "admin"},
		Spec:       authv1alpha1.UserSpec{Username: "admin", Email: "admin@crossplane-ui.local"},
		Status:     authv1alpha1.UserStatus{PasswordHash: "$2a$10$alreadyhashed", UserID: "u"},
	}
	c := fake.NewClientBuilder().
		WithScheme(kube.Scheme).
		WithObjects(existing).
		WithStatusSubresource(&authv1alpha1.User{}).
		Build()

	if err := bootstrap.Run(context.Background(), c, bootstrap.Config{
		Namespace: "x", SecretName: "missing", DefaultUsername: "admin",
	}); err != nil {
		t.Fatalf("Run should tolerate missing Secret when User exists: %v", err)
	}

	// Sanity: no accidental mutation of the Status.
	var after authv1alpha1.User
	if err := c.Get(context.Background(), types.NamespacedName{Name: "admin"}, &after); err != nil {
		t.Fatalf("Get: %v", err)
	}
	if after.Status.PasswordHash != existing.Status.PasswordHash {
		t.Fatal("PasswordHash clobbered")
	}
}
