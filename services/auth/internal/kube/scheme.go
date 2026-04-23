// Package kube wires the Kubernetes client bits shared by the auth service:
// a runtime scheme that carries our CRDs plus the kubeconfig loader.
package kube

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	authv1alpha1 "github.com/bzzdevlabs/crossplane-ui/pkg/apis/auth/v1alpha1"
	"github.com/bzzdevlabs/crossplane-ui/services/auth/internal/dex"
)

// Scheme is the runtime scheme shared by the manager and any ad-hoc clients.
// It carries the core Kubernetes kinds, our auth.crossplane-ui.io types and
// Dex's storage kinds (Password, OAuth2Client) as unstructured.
var Scheme = runtime.NewScheme()

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(Scheme))
	utilruntime.Must(authv1alpha1.AddToScheme(Scheme))
	dex.AddToScheme(Scheme)
}

// LoadConfig returns a *rest.Config. When path is empty the in-cluster
// ServiceAccount token is used; otherwise the kubeconfig file is read.
func LoadConfig(path string) (*rest.Config, error) {
	if path == "" {
		cfg, err := rest.InClusterConfig()
		if err != nil {
			return nil, fmt.Errorf("in-cluster config: %w", err)
		}
		return cfg, nil
	}
	cfg, err := clientcmd.BuildConfigFromFlags("", path)
	if err != nil {
		return nil, fmt.Errorf("build kubeconfig %q: %w", path, err)
	}
	return cfg, nil
}
