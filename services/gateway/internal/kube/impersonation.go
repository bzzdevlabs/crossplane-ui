package kube

import (
	"fmt"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// ClientFactory derives a clientset that impersonates a given user on the
// cluster. It takes a base *rest.Config — held by the gateway — and clones
// it per call so concurrent requests cannot step on each other's headers.
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
	if user == "" {
		return nil, fmt.Errorf("kube impersonation: empty user")
	}
	cfg := rest.CopyConfig(f.base)
	cfg.Impersonate = rest.ImpersonationConfig{
		UserName: user,
		Groups:   groups,
	}
	return NewClientset(cfg)
}
