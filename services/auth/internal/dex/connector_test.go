package dex_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	authv1alpha1 "github.com/bzzdevlabs/crossplane-ui/pkg/apis/auth/v1alpha1"
	"github.com/bzzdevlabs/crossplane-ui/services/auth/internal/dex"
	"github.com/bzzdevlabs/crossplane-ui/services/auth/internal/kube"
)

func connectorList() *unstructured.UnstructuredList {
	l := &unstructured.UnstructuredList{}
	l.SetGroupVersionKind(dex.ConnectorGVK.GroupVersion().WithKind(dex.KindConnector + "List"))
	return l
}

func TestSyncConnectorsProjectsAndInjectsSecrets(t *testing.T) {
	t.Parallel()

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "connector-github", Namespace: testNS},
		Data:       map[string][]byte{"clientSecret": []byte("s3cr3t")},
	}
	c := fake.NewClientBuilder().WithScheme(kube.Scheme).WithObjects(secret).Build()

	rawConfig, err := json.Marshal(map[string]any{
		"clientID":    "cid",
		"redirectURI": "https://example/cb",
	})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	conns := []authv1alpha1.Connector{
		{
			ObjectMeta: metav1.ObjectMeta{Name: "github"},
			Spec: authv1alpha1.ConnectorSpec{
				ID:     "github",
				Type:   authv1alpha1.ConnectorTypeGitHub,
				Name:   "GitHub",
				Config: runtime.RawExtension{Raw: rawConfig},
				SecretRefs: []authv1alpha1.ConnectorSecretInjection{
					{
						Path: "clientSecret",
						SecretRef: authv1alpha1.ConnectorSecretRef{
							Name: "connector-github",
							Key:  "clientSecret",
						},
					},
				},
			},
		},
	}

	changed, err := dex.SyncConnectors(context.Background(), c, dex.KubeSecretResolver(c), testNS, conns)
	if err != nil {
		t.Fatalf("SyncConnectors: %v", err)
	}
	if !changed {
		t.Fatal("expected cluster change on creation")
	}

	list := connectorList()
	if err := c.List(context.Background(), list); err != nil {
		t.Fatalf("List: %v", err)
	}
	if got, want := len(list.Items), 1; got != want {
		t.Fatalf("connector count = %d, want %d", got, want)
	}

	item := list.Items[0]
	if got, _, _ := unstructured.NestedString(item.Object, "id"); got != "github" {
		t.Errorf("id = %q", got)
	}
	if got, _, _ := unstructured.NestedString(item.Object, "type"); got != "github" {
		t.Errorf("type = %q", got)
	}
	b64cfg, _, _ := unstructured.NestedString(item.Object, "config")
	cfgJSON, err := base64.StdEncoding.DecodeString(b64cfg)
	if err != nil {
		t.Fatalf("decode config: %v", err)
	}
	var cfg map[string]any
	if err := json.Unmarshal(cfgJSON, &cfg); err != nil {
		t.Fatalf("unmarshal config: %v", err)
	}
	if cfg["clientSecret"] != "s3cr3t" {
		t.Errorf("injected secret missing or wrong; got %q", cfg["clientSecret"])
	}
	if cfg["clientID"] != "cid" {
		t.Errorf("config round-trip lost keys: %+v", cfg)
	}

	// Second sync without changes leaves the cluster alone.
	changed, err = dex.SyncConnectors(context.Background(), c, dex.KubeSecretResolver(c), testNS, conns)
	if err != nil {
		t.Fatalf("SyncConnectors (2): %v", err)
	}
	if changed {
		t.Errorf("expected no change on idempotent re-sync")
	}
}

func TestSyncConnectorsPrunesDisabledAndRemoved(t *testing.T) {
	t.Parallel()

	c := fake.NewClientBuilder().WithScheme(kube.Scheme).Build()

	enabled := authv1alpha1.Connector{
		ObjectMeta: metav1.ObjectMeta{Name: "keycloak"},
		Spec: authv1alpha1.ConnectorSpec{
			ID:     "keycloak",
			Type:   authv1alpha1.ConnectorTypeOIDC,
			Name:   "Keycloak",
			Config: runtime.RawExtension{Raw: []byte(`{}`)},
		},
	}
	if _, err := dex.SyncConnectors(context.Background(), c, nil, testNS, []authv1alpha1.Connector{enabled}); err != nil {
		t.Fatalf("seed: %v", err)
	}
	enabled.Spec.Disabled = true
	changed, err := dex.SyncConnectors(context.Background(), c, nil, testNS, []authv1alpha1.Connector{enabled})
	if err != nil {
		t.Fatalf("resync: %v", err)
	}
	if !changed {
		t.Fatal("disabling should mutate the cluster")
	}
	list := connectorList()
	if err := c.List(context.Background(), list); err != nil {
		t.Fatalf("List: %v", err)
	}
	if got := len(list.Items); got != 0 {
		t.Errorf("expected all connectors pruned, got %d", got)
	}
}
