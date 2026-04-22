package api_test

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"

	"gitlab.telespazio-digital-factory.fr/icdo/tpzf/crossplane-ui/services/gateway/internal/api"
	"gitlab.telespazio-digital-factory.fr/icdo/tpzf/crossplane-ui/services/gateway/internal/oidc"
)

// fakeFactory returns the same pre-baked clientset on every call and
// records the (user, groups) it was invoked with.
type fakeFactory struct {
	client    kubernetes.Interface
	err       error
	lastUser  string
	lastGroup []string
}

func (f *fakeFactory) For(user string, groups []string) (kubernetes.Interface, error) {
	f.lastUser = user
	f.lastGroup = groups
	return f.client, f.err
}

func newLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestNamespacesHandlerRejectsWithoutUser(t *testing.T) {
	t.Parallel()

	h := api.NamespacesHandler(newLogger(), &fakeFactory{})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/namespaces", nil))

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}

func TestNamespacesHandlerReturnsProjectedList(t *testing.T) {
	t.Parallel()

	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "kube-system",
			Labels:            map[string]string{"foo": "bar"},
			CreationTimestamp: metav1.Time{Time: time.Date(2026, 4, 22, 10, 0, 0, 0, time.UTC)},
		},
		Status: corev1.NamespaceStatus{Phase: corev1.NamespaceActive},
	}
	client := fake.NewClientset(ns)
	factory := &fakeFactory{client: client}

	h := api.NamespacesHandler(newLogger(), factory)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/namespaces", nil)
	ctx := oidc.WithUser(req.Context(), oidc.User{
		Subject:           "abc",
		PreferredUsername: "alice",
		Groups:            []string{"admins"},
	})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req.WithContext(ctx))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var out api.NamespacesList
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if len(out.Items) != 1 || out.Items[0].Name != "kube-system" {
		t.Fatalf("items = %+v", out.Items)
	}
	if out.Items[0].Phase != "Active" {
		t.Errorf("phase = %q, want Active", out.Items[0].Phase)
	}
	if factory.lastUser != "alice" || factory.lastGroup[0] != "admins" {
		t.Errorf("factory called with %q %v, want alice [admins]", factory.lastUser, factory.lastGroup)
	}
}

func TestNamespacesHandlerFactoryFailure(t *testing.T) {
	t.Parallel()

	factory := &fakeFactory{err: errors.New("boom")}
	h := api.NamespacesHandler(newLogger(), factory)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/namespaces", nil)
	ctx := oidc.WithUser(req.Context(), oidc.User{Subject: "abc", PreferredUsername: "alice"})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req.WithContext(ctx))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
}

func TestNamespacesHandlerMethodNotAllowed(t *testing.T) {
	t.Parallel()

	h := api.NamespacesHandler(newLogger(), &fakeFactory{})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/namespaces", nil)
	ctx := oidc.WithUser(req.Context(), oidc.User{Subject: "abc", PreferredUsername: "alice"})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req.WithContext(ctx))

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want 405", rec.Code)
	}
}
