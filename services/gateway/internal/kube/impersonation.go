package kube

import (
	"errors"
	"fmt"

	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// ClientFactory derives per-request clients that impersonate a given user on
// the cluster. It takes a base *rest.Config — held by the gateway — and
// clones it per call so concurrent requests cannot step on each other's
// headers.
type ClientFactory struct {
	base *rest.Config
}

// NewClientFactory constructs a ClientFactory from the supplied base config.
// The base config MUST grant the `impersonate` verb on users, groups and
// serviceaccounts; otherwise every call this factory makes will 403.
func NewClientFactory(base *rest.Config) *ClientFactory {
	return &ClientFactory{base: base}
}

// For returns a Clientset whose outbound requests carry Impersonate-User
// and Impersonate-Group headers derived from user + groups. The return
// type is the broader kubernetes.Interface so that consumers can swap in
// a fake client for tests.
func (f *ClientFactory) For(user string, groups []string) (kubernetes.Interface, error) {
	cfg, err := f.impersonating(user, groups)
	if err != nil {
		return nil, err
	}
	return NewClientset(cfg)
}

// Dynamic returns a dynamic.Interface stamped with the same impersonation
// headers. Used for resources the gateway does not carry typed clients for
// (Crossplane CRDs, arbitrary managed resources, …).
func (f *ClientFactory) Dynamic(user string, groups []string) (dynamic.Interface, error) {
	cfg, err := f.impersonating(user, groups)
	if err != nil {
		return nil, err
	}
	c, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("new dynamic client: %w", err)
	}
	return c, nil
}

// Discovery returns a discovery.DiscoveryInterface stamped with the
// impersonation headers. Callers use it to enumerate CRDs and look up the
// resources that back a given GVR.
func (f *ClientFactory) Discovery(user string, groups []string) (discovery.DiscoveryInterface, error) {
	cfg, err := f.impersonating(user, groups)
	if err != nil {
		return nil, err
	}
	c, err := discovery.NewDiscoveryClientForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("new discovery client: %w", err)
	}
	return c, nil
}

func (f *ClientFactory) impersonating(user string, groups []string) (*rest.Config, error) {
	if user == "" {
		return nil, errors.New("kube impersonation: empty user")
	}
	cfg := rest.CopyConfig(f.base)
	cfg.Impersonate = rest.ImpersonationConfig{
		UserName: user,
		Groups:   groups,
	}
	return cfg, nil
}
