package controller_test

import (
	"context"
	"encoding/base64"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	authv1alpha1 "github.com/bzzdevlabs/crossplane-ui/pkg/apis/auth/v1alpha1"
	"github.com/bzzdevlabs/crossplane-ui/services/auth/internal/controller"
	"github.com/bzzdevlabs/crossplane-ui/services/auth/internal/dex"
	"github.com/bzzdevlabs/crossplane-ui/services/auth/internal/kube"
	"github.com/bzzdevlabs/crossplane-ui/services/auth/internal/password"
)

func passwordList() *unstructured.UnstructuredList {
	l := &unstructured.UnstructuredList{}
	l.SetGroupVersionKind(dex.PasswordGVK.GroupVersion().WithKind(dex.KindPassword + "List"))
	return l
}

func TestUserReconcileHashesSecretAndProjectsToDex(t *testing.T) {
	t.Parallel()

	const ns = "crossplane-ui"

	user := &authv1alpha1.User{
		ObjectMeta: metav1.ObjectMeta{Name: "alice", Generation: 1},
		Spec: authv1alpha1.UserSpec{
			Username:          "alice",
			Email:             "alice@example.com",
			PasswordSecretRef: &authv1alpha1.SecretReference{Name: "alice-pw"},
		},
	}
	sec := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: "alice-pw"},
		Data:       map[string][]byte{"password": []byte("letmein")},
	}

	c := fake.NewClientBuilder().
		WithScheme(kube.Scheme).
		WithObjects(user, sec).
		WithStatusSubresource(&authv1alpha1.User{}).
		Build()

	r := &controller.UserReconciler{
		Client:    c,
		Scheme:    kube.Scheme,
		Namespace: ns,
	}

	if _, err := r.Reconcile(context.Background(), reconcile.Request{
		NamespacedName: types.NamespacedName{Name: "alice"},
	}); err != nil {
		t.Fatalf("Reconcile: %v", err)
	}

	// User status carries a hash that verifies.
	var got authv1alpha1.User
	if err := c.Get(context.Background(), types.NamespacedName{Name: "alice"}, &got); err != nil {
		t.Fatalf("Get user: %v", err)
	}
	if got.Status.PasswordHash == "" {
		t.Fatal("PasswordHash empty after reconcile")
	}
	if got.Status.UserID == "" {
		t.Fatal("UserID empty after reconcile")
	}
	ok, err := password.Verify(got.Status.PasswordHash, "letmein")
	if err != nil || !ok {
		t.Fatalf("hash doesn't verify: ok=%v err=%v", ok, err)
	}

	// Secret is scrubbed.
	var after corev1.Secret
	if err := c.Get(context.Background(), types.NamespacedName{Namespace: ns, Name: "alice-pw"}, &after); err != nil {
		t.Fatalf("Get secret: %v", err)
	}
	if _, ok := after.Data["password"]; ok {
		t.Fatal("password key not scrubbed")
	}

	// Dex Password CR is written with our base64-wrapped hash.
	list := passwordList()
	if err := c.List(context.Background(), list); err != nil {
		t.Fatalf("List passwords: %v", err)
	}
	if len(list.Items) != 1 {
		t.Fatalf("expected 1 Password, got %d", len(list.Items))
	}
	item := list.Items[0]
	if email, _ := item.Object["email"].(string); email != "alice@example.com" {
		t.Errorf("email: got %q", email)
	}
	wantHash := base64.StdEncoding.EncodeToString([]byte(got.Status.PasswordHash))
	if h, _ := item.Object["hash"].(string); h != wantHash {
		t.Errorf("hash: got %q, want %q", h, wantHash)
	}
}

func TestUserReconcileDeletionPrunesPassword(t *testing.T) {
	t.Parallel()

	const ns = "crossplane-ui"

	// Seed a stale Password that belongs to no live User.
	stale := &unstructured.Unstructured{}
	stale.SetGroupVersionKind(dex.PasswordGVK)
	stale.SetNamespace(ns)
	stale.SetName("stale-password")
	stale.SetLabels(map[string]string{dex.ManagedByLabel: dex.ManagedByValue})
	stale.Object["email"] = "ghost@example.com"
	stale.Object["hash"] = "Zm9v"
	stale.Object["username"] = "ghost"
	stale.Object["userID"] = "uid-ghost"

	c := fake.NewClientBuilder().
		WithScheme(kube.Scheme).
		WithObjects(stale).
		WithStatusSubresource(&authv1alpha1.User{}).
		Build()

	r := &controller.UserReconciler{
		Client:    c,
		Scheme:    kube.Scheme,
		Namespace: ns,
	}

	// Reconcile for a non-existent User: the stale Password must be pruned.
	if _, err := r.Reconcile(context.Background(), reconcile.Request{
		NamespacedName: types.NamespacedName{Name: "ghost"},
	}); err != nil {
		t.Fatalf("Reconcile: %v", err)
	}

	list := passwordList()
	if err := c.List(context.Background(), list); err != nil {
		t.Fatalf("List passwords: %v", err)
	}
	if got := len(list.Items); got != 0 {
		t.Fatalf("expected 0 Passwords after prune, got %d", got)
	}
}
