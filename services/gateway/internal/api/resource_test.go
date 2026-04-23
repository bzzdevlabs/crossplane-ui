package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	fakedynamic "k8s.io/client-go/dynamic/fake"

	"github.com/bzzdevlabs/crossplane-ui/services/gateway/internal/api"
	"github.com/bzzdevlabs/crossplane-ui/services/gateway/internal/oidc"
)

func authCtx(req *http.Request) *http.Request {
	ctx := oidc.WithUser(req.Context(), oidc.User{
		Subject:           "abc",
		PreferredUsername: "alice",
		Groups:            []string{"admins"},
	})
	return req.WithContext(ctx)
}

func TestResourceGetReturnsObject(t *testing.T) {
	t.Parallel()

	gvr := schema.GroupVersionResource{Group: "pkg.crossplane.io", Version: "v1", Resource: "providers"}
	u := &unstructured.Unstructured{}
	u.SetAPIVersion("pkg.crossplane.io/v1")
	u.SetKind("Provider")
	u.SetName("provider-aws")
	scheme := runtime.NewScheme()
	dyn := fakedynamic.NewSimpleDynamicClientWithCustomListKinds(
		scheme,
		map[schema.GroupVersionResource]string{gvr: "ProviderList"},
		u,
	)
	factory := &fakeCrossplaneFactory{dyn: dyn}

	h := api.ResourceHandler(newLogger(), factory)
	req := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/crossplane/resource?group=pkg.crossplane.io&version=v1&resource=providers&name=provider-aws",
		nil,
	)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, authCtx(req))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body=%s", rec.Code, rec.Body.String())
	}
	var got map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	meta, _ := got["metadata"].(map[string]any)
	if meta["name"] != "provider-aws" {
		t.Errorf("metadata.name: got %v", meta["name"])
	}
}

func TestResourceGetRejectsMissingQuery(t *testing.T) {
	t.Parallel()
	h := api.ResourceHandler(newLogger(), &fakeCrossplaneFactory{})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/crossplane/resource?group=x", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, authCtx(req))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestResourceDeleteRemovesObject(t *testing.T) {
	t.Parallel()

	gvr := schema.GroupVersionResource{Group: "pkg.crossplane.io", Version: "v1", Resource: "providers"}
	u := &unstructured.Unstructured{}
	u.SetAPIVersion("pkg.crossplane.io/v1")
	u.SetKind("Provider")
	u.SetName("provider-aws")
	scheme := runtime.NewScheme()
	dyn := fakedynamic.NewSimpleDynamicClientWithCustomListKinds(
		scheme,
		map[schema.GroupVersionResource]string{gvr: "ProviderList"},
		u,
	)
	factory := &fakeCrossplaneFactory{dyn: dyn}

	h := api.ResourceHandler(newLogger(), factory)
	req := httptest.NewRequest(
		http.MethodDelete,
		"/api/v1/crossplane/resource?group=pkg.crossplane.io&version=v1&resource=providers&name=provider-aws",
		nil,
	)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, authCtx(req))

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, body=%s", rec.Code, rec.Body.String())
	}
}

func TestResourcePutRejectsEmptyBody(t *testing.T) {
	t.Parallel()
	h := api.ResourceHandler(newLogger(), &fakeCrossplaneFactory{})
	req := httptest.NewRequest(http.MethodPut,
		"/api/v1/crossplane/resource?group=pkg.crossplane.io&version=v1&resource=providers",
		bytes.NewReader([]byte(`{}`)))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, authCtx(req))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestResourcePutRejectsInvalidJSON(t *testing.T) {
	t.Parallel()
	h := api.ResourceHandler(newLogger(), &fakeCrossplaneFactory{})
	req := httptest.NewRequest(http.MethodPut,
		"/api/v1/crossplane/resource?group=pkg.crossplane.io&version=v1&resource=providers",
		bytes.NewReader([]byte(`not json`)))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, authCtx(req))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}
