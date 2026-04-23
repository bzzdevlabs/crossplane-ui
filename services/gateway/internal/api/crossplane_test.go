package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	fakediscovery "k8s.io/client-go/discovery/fake"
	"k8s.io/client-go/dynamic"
	fakedynamic "k8s.io/client-go/dynamic/fake"
	clienttesting "k8s.io/client-go/testing"

	"gitlab.telespazio-digital-factory.fr/icdo/tpzf/crossplane-ui/services/gateway/internal/api"
	"gitlab.telespazio-digital-factory.fr/icdo/tpzf/crossplane-ui/services/gateway/internal/oidc"
)

type fakeCrossplaneFactory struct {
	dyn  dynamic.Interface
	disc discovery.DiscoveryInterface
}

func (f *fakeCrossplaneFactory) Dynamic(string, []string) (dynamic.Interface, error) {
	return f.dyn, nil
}

func (f *fakeCrossplaneFactory) Discovery(string, []string) (discovery.DiscoveryInterface, error) {
	return f.disc, nil
}

func newUnstructured(apiVersion, kind, name, ready, synced string) *unstructured.Unstructured {
	u := &unstructured.Unstructured{}
	u.SetAPIVersion(apiVersion)
	u.SetKind(kind)
	u.SetName(name)
	u.Object["status"] = map[string]any{
		"conditions": []any{
			map[string]any{"type": "Ready", "status": ready},
			map[string]any{"type": "Synced", "status": synced},
		},
	}
	return u
}

func TestCrossplaneResourcesAggregatesCategories(t *testing.T) {
	t.Parallel()

	scheme := runtime.NewScheme()

	compositionGVR := schema.GroupVersionResource{
		Group: "apiextensions.crossplane.io", Version: "v1", Resource: "compositions",
	}
	providerGVR := schema.GroupVersionResource{
		Group: "pkg.crossplane.io", Version: "v1", Resource: "providers",
	}
	functionGVR := schema.GroupVersionResource{
		Group: "pkg.crossplane.io", Version: "v1", Resource: "functions",
	}
	bucketGVR := schema.GroupVersionResource{
		Group: "s3.aws.upbound.io", Version: "v1beta1", Resource: "buckets",
	}

	gvrToListKind := map[schema.GroupVersionResource]string{
		compositionGVR: "CompositionList",
		providerGVR:    "ProviderList",
		functionGVR:    "FunctionList",
		bucketGVR:      "BucketList",
	}

	initial := []runtime.Object{
		newUnstructured("apiextensions.crossplane.io/v1", "Composition", "aws-vpc", "True", "True"),
		newUnstructured("pkg.crossplane.io/v1", "Provider", "provider-aws", "True", "True"),
		newUnstructured("pkg.crossplane.io/v1", "Function", "fn-go", "False", "Unknown"),
		newUnstructured("s3.aws.upbound.io/v1beta1", "Bucket", "my-bucket", "True", "True"),
	}

	dyn := fakedynamic.NewSimpleDynamicClientWithCustomListKinds(scheme, gvrToListKind, initial...)

	fakeDisco := &fakediscovery.FakeDiscovery{Fake: &clienttesting.Fake{}}
	fakeDisco.Resources = []*metav1.APIResourceList{
		{
			GroupVersion: "s3.aws.upbound.io/v1beta1",
			APIResources: []metav1.APIResource{
				{Name: "buckets", Kind: "Bucket", Categories: []string{"managed", "crossplane"}},
			},
		},
	}

	factory := &fakeCrossplaneFactory{dyn: dyn, disc: fakeDisco}
	h := api.CrossplaneResourcesHandler(newLogger(), factory)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/crossplane/resources", nil)
	ctx := oidc.WithUser(req.Context(), oidc.User{
		Subject:           "abc",
		PreferredUsername: "alice",
		Groups:            []string{"admins"},
	})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req.WithContext(ctx))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}

	var out api.CrossplaneSummary
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	byCat := map[string]api.CrossplaneGroup{}
	for _, g := range out.Groups {
		byCat[g.Category] = g
	}
	if g := byCat["composition"]; len(g.Items) != 1 || g.Items[0].Name != "aws-vpc" {
		t.Errorf("composition items = %+v", g.Items)
	}
	if g := byCat["provider"]; len(g.Items) != 1 || g.Items[0].Ready != "True" {
		t.Errorf("provider items = %+v", g.Items)
	}
	if g := byCat["function"]; len(g.Items) != 1 || g.Items[0].Ready != "False" {
		t.Errorf("function items = %+v", g.Items)
	}
	if g := byCat["managed"]; len(g.Items) != 1 || g.Items[0].Kind != "Bucket" {
		t.Errorf("managed items = %+v", g.Items)
	}
	if g := byCat["composite"]; len(g.Items) != 0 {
		t.Errorf("composite items should be empty, got %+v", g.Items)
	}
}

func TestCrossplaneResourcesRejectsWithoutUser(t *testing.T) {
	t.Parallel()
	factory := &fakeCrossplaneFactory{}
	h := api.CrossplaneResourcesHandler(newLogger(), factory)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/crossplane/resources", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}
