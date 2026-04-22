// Package kube wraps the Kubernetes client-go library to centralise how the
// gateway talks to the cluster.
//
// Two concerns live here:
//   - Building the base *rest.Config (in-cluster or kubeconfig-file-based).
//   - Deriving per-request, impersonation-aware clientsets so that every
//     call to the Kubernetes API carries the authenticated user's identity.
package kube

import (
	"fmt"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// LoadConfig returns a *rest.Config suitable for talking to the Kubernetes
// API server. When path is empty in-cluster configuration is used; this
// matches Helm-deployed gateways where the ServiceAccount token is mounted
// at `/var/run/secrets/kubernetes.io/serviceaccount`.
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

// NewClientset builds a full clientset from cfg.
func NewClientset(cfg *rest.Config) (*kubernetes.Clientset, error) {
	cs, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("new clientset: %w", err)
	}
	return cs, nil
}
