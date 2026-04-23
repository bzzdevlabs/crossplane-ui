package kube_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"k8s.io/client-go/rest"

	"github.com/bzzdevlabs/crossplane-ui/services/gateway/internal/kube"
)

// TestFactoryForwardsImpersonationHeaders spins a tiny httptest apiserver
// that records the headers it received and asserts the impersonation
// headers are forwarded on a Namespaces().List call.
func TestFactoryForwardsImpersonationHeaders(t *testing.T) {
	t.Parallel()

	var (
		gotUser   string
		gotGroups []string
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUser = r.Header.Get("Impersonate-User")
		gotGroups = r.Header.Values("Impersonate-Group")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"kind":"NamespaceList","apiVersion":"v1","items":[]}`))
	}))
	t.Cleanup(srv.Close)

	factory := kube.NewClientFactory(&rest.Config{Host: srv.URL})
	client, err := factory.For("alice", []string{"admins", "devs"})
	if err != nil {
		t.Fatalf("factory.For: %v", err)
	}

	if _, err := client.CoreV1().Namespaces().List(context.Background(), metav1ListOptions()); err != nil {
		t.Fatalf("List: %v", err)
	}

	if gotUser != "alice" {
		t.Errorf("Impersonate-User = %q, want alice", gotUser)
	}
	if strings.Join(gotGroups, ",") != "admins,devs" {
		t.Errorf("Impersonate-Group = %v, want [admins devs]", gotGroups)
	}
}

func TestFactoryRejectsEmptyUser(t *testing.T) {
	t.Parallel()

	factory := kube.NewClientFactory(&rest.Config{Host: "http://unused"})
	if _, err := factory.For("", nil); err == nil {
		t.Fatal("expected error on empty user")
	}
}
