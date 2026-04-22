package dex_test

import (
	"context"
	"encoding/base64"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	authv1alpha1 "gitlab.telespazio-digital-factory.fr/icdo/tpzf/crossplane-ui/pkg/apis/auth/v1alpha1"
	"gitlab.telespazio-digital-factory.fr/icdo/tpzf/crossplane-ui/services/auth/internal/dex"
	"gitlab.telespazio-digital-factory.fr/icdo/tpzf/crossplane-ui/services/auth/internal/kube"
)

const testNS = "crossplane-ui"

func passwordList() *unstructured.UnstructuredList {
	l := &unstructured.UnstructuredList{}
	l.SetGroupVersionKind(dex.PasswordGVK.GroupVersion().WithKind(dex.KindPassword + "List"))
	return l
}

func TestSyncCreatesPasswordsForEnabledUsers(t *testing.T) {
	t.Parallel()

	c := fake.NewClientBuilder().WithScheme(kube.Scheme).Build()
	users := []authv1alpha1.User{
		{
			ObjectMeta: metav1.ObjectMeta{Name: "alice"},
			Spec:       authv1alpha1.UserSpec{Username: "alice", Email: "Alice@Example.com"},
			Status:     authv1alpha1.UserStatus{PasswordHash: "$2a$10$HASH-A", UserID: "uid-a"},
		},
		// No hash yet → skipped.
		{
			ObjectMeta: metav1.ObjectMeta{Name: "bob"},
			Spec:       authv1alpha1.UserSpec{Username: "bob", Email: "bob@example.com"},
		},
		// Disabled → skipped.
		{
			ObjectMeta: metav1.ObjectMeta{Name: "charlie"},
			Spec:       authv1alpha1.UserSpec{Username: "charlie", Email: "charlie@example.com", Disabled: true},
			Status:     authv1alpha1.UserStatus{PasswordHash: "$2a$10$HASH-C", UserID: "uid-c"},
		},
	}

	changed, err := dex.Sync(context.Background(), c, testNS, users)
	if err != nil {
		t.Fatalf("Sync: %v", err)
	}
	if !changed {
		t.Fatal("Sync reported no change on creation")
	}

	list := passwordList()
	if err := c.List(context.Background(), list); err != nil {
		t.Fatalf("List: %v", err)
	}
	if got, want := len(list.Items), 1; got != want {
		t.Fatalf("expected %d Password objects, got %d", want, got)
	}
	item := list.Items[0]
	if got, want := item.Object["email"], "alice@example.com"; got != want {
		t.Errorf("email: got %q, want %q (expected lowercase)", got, want)
	}
	if got, want := item.Object["userID"], "uid-a"; got != want {
		t.Errorf("userID: got %q, want %q", got, want)
	}
	wantHash := base64.StdEncoding.EncodeToString([]byte("$2a$10$HASH-A"))
	if got := item.Object["hash"]; got != wantHash {
		t.Errorf("hash: got %q, want %q", got, wantHash)
	}
}

func TestSyncIsIdempotent(t *testing.T) {
	t.Parallel()

	c := fake.NewClientBuilder().WithScheme(kube.Scheme).Build()
	users := []authv1alpha1.User{
		{
			Spec:   authv1alpha1.UserSpec{Username: "alice", Email: "a@b.com"},
			Status: authv1alpha1.UserStatus{PasswordHash: "$2a$10$HASH", UserID: "uid"},
		},
	}
	if _, err := dex.Sync(context.Background(), c, testNS, users); err != nil {
		t.Fatalf("first Sync: %v", err)
	}
	changed, err := dex.Sync(context.Background(), c, testNS, users)
	if err != nil {
		t.Fatalf("second Sync: %v", err)
	}
	if changed {
		t.Fatal("second Sync reported change on unchanged input")
	}
}

func TestSyncPrunesStalePasswords(t *testing.T) {
	t.Parallel()

	c := fake.NewClientBuilder().WithScheme(kube.Scheme).Build()
	users := []authv1alpha1.User{
		{
			Spec:   authv1alpha1.UserSpec{Username: "alice", Email: "a@b.com"},
			Status: authv1alpha1.UserStatus{PasswordHash: "$2a$10$A", UserID: "uid-a"},
		},
		{
			Spec:   authv1alpha1.UserSpec{Username: "zed", Email: "z@b.com"},
			Status: authv1alpha1.UserStatus{PasswordHash: "$2a$10$Z", UserID: "uid-z"},
		},
	}
	if _, err := dex.Sync(context.Background(), c, testNS, users); err != nil {
		t.Fatalf("first Sync: %v", err)
	}

	// Drop zed → the second reconcile must prune her Password.
	if _, err := dex.Sync(context.Background(), c, testNS, users[:1]); err != nil {
		t.Fatalf("second Sync: %v", err)
	}

	list := passwordList()
	if err := c.List(context.Background(), list); err != nil {
		t.Fatalf("List: %v", err)
	}
	if got, want := len(list.Items), 1; got != want {
		t.Fatalf("expected %d Password after prune, got %d", want, got)
	}
	if got, want := list.Items[0].Object["email"], "a@b.com"; got != want {
		t.Errorf("survivor email: got %q, want %q", got, want)
	}
}

func TestSyncUpdatesChangedHash(t *testing.T) {
	t.Parallel()

	c := fake.NewClientBuilder().WithScheme(kube.Scheme).Build()
	users := []authv1alpha1.User{
		{
			Spec:   authv1alpha1.UserSpec{Username: "alice", Email: "a@b.com"},
			Status: authv1alpha1.UserStatus{PasswordHash: "$2a$10$OLD", UserID: "uid"},
		},
	}
	if _, err := dex.Sync(context.Background(), c, testNS, users); err != nil {
		t.Fatalf("first Sync: %v", err)
	}
	users[0].Status.PasswordHash = "$2a$10$NEW"
	changed, err := dex.Sync(context.Background(), c, testNS, users)
	if err != nil {
		t.Fatalf("second Sync: %v", err)
	}
	if !changed {
		t.Fatal("Sync reported no change after hash rotation")
	}

	list := passwordList()
	if err := c.List(context.Background(), list); err != nil {
		t.Fatalf("List: %v", err)
	}
	wantHash := base64.StdEncoding.EncodeToString([]byte("$2a$10$NEW"))
	if got := list.Items[0].Object["hash"]; got != wantHash {
		t.Errorf("hash not updated: got %q, want %q", got, wantHash)
	}
}

func TestEnsureOAuth2ClientCreatesAndIsIdempotent(t *testing.T) {
	t.Parallel()

	c := fake.NewClientBuilder().WithScheme(kube.Scheme).Build()
	cfg := &dex.OAuth2ClientConfig{
		Namespace:    testNS,
		ID:           "crossplane-ui",
		Secret:       "s3cret",
		Name:         "Crossplane UI",
		RedirectURIs: []string{"https://example.com/auth/callback"},
	}

	changed, err := dex.EnsureOAuth2Client(context.Background(), c, cfg)
	if err != nil {
		t.Fatalf("EnsureOAuth2Client: %v", err)
	}
	if !changed {
		t.Fatal("first EnsureOAuth2Client reported no change")
	}

	changed, err = dex.EnsureOAuth2Client(context.Background(), c, cfg)
	if err != nil {
		t.Fatalf("second EnsureOAuth2Client: %v", err)
	}
	if changed {
		t.Fatal("second EnsureOAuth2Client reported a change on unchanged input")
	}

	list := &unstructured.UnstructuredList{}
	list.SetGroupVersionKind(dex.OAuth2ClientGVK.GroupVersion().WithKind(dex.KindOAuth2Client + "List"))
	if err := c.List(context.Background(), list); err != nil {
		t.Fatalf("List oauth2clients: %v", err)
	}
	if got, want := len(list.Items), 1; got != want {
		t.Fatalf("expected %d OAuth2Client, got %d", want, got)
	}
	got := list.Items[0].Object
	if got["id"] != "crossplane-ui" {
		t.Errorf("id: got %v, want %q", got["id"], "crossplane-ui")
	}
	if got["secret"] != "s3cret" {
		t.Errorf("secret not persisted")
	}
}

func TestEnsureOAuth2ClientUpdatesSecretRotation(t *testing.T) {
	t.Parallel()

	c := fake.NewClientBuilder().WithScheme(kube.Scheme).Build()
	cfg := &dex.OAuth2ClientConfig{
		Namespace:    testNS,
		ID:           "crossplane-ui",
		Secret:       "old",
		RedirectURIs: []string{"https://example.com/auth/callback"},
	}
	if _, err := dex.EnsureOAuth2Client(context.Background(), c, cfg); err != nil {
		t.Fatalf("first EnsureOAuth2Client: %v", err)
	}
	cfg.Secret = "new"
	changed, err := dex.EnsureOAuth2Client(context.Background(), c, cfg)
	if err != nil {
		t.Fatalf("second EnsureOAuth2Client: %v", err)
	}
	if !changed {
		t.Fatal("secret rotation did not trigger an update")
	}
}

func TestSplitRedirectURIs(t *testing.T) {
	t.Parallel()

	cases := []struct {
		in   string
		want []string
	}{
		{"", nil},
		{"https://a/cb", []string{"https://a/cb"}},
		{"https://a/cb,https://b/cb", []string{"https://a/cb", "https://b/cb"}},
		{"https://a/cb  https://b/cb\n\thttps://c/cb", []string{"https://a/cb", "https://b/cb", "https://c/cb"}},
	}
	for _, tc := range cases {
		got := dex.SplitRedirectURIs(tc.in)
		if len(got) != len(tc.want) {
			t.Fatalf("%q: got %v, want %v", tc.in, got, tc.want)
		}
		for i := range got {
			if got[i] != tc.want[i] {
				t.Errorf("%q[%d]: got %q, want %q", tc.in, i, got[i], tc.want[i])
			}
		}
	}
}
