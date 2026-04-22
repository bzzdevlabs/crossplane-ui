package dex

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// OAuth2ClientConfig describes the OAuth2Client object to materialise in Dex.
type OAuth2ClientConfig struct {
	// Namespace is Dex's namespace (where it reads CRs from).
	Namespace string
	// ID is the OIDC client identifier (e.g. "crossplane-ui").
	ID string
	// Secret is the shared secret between the client and Dex.
	Secret string
	// Name is a human-friendly label shown on the Dex approval screen.
	Name string
	// RedirectURIs are the allowed OAuth2 callback URLs.
	RedirectURIs []string
}

// EnsureOAuth2Client upserts the OAuth2Client object for the gateway. Returns
// true when the cluster was mutated.
func EnsureOAuth2Client(ctx context.Context, c client.Client, cfg *OAuth2ClientConfig) (bool, error) {
	if cfg.Namespace == "" {
		return false, errors.New("oauth2client: namespace is empty")
	}
	if cfg.ID == "" {
		return false, errors.New("oauth2client: id is empty")
	}
	if cfg.Secret == "" {
		return false, errors.New("oauth2client: secret is empty")
	}
	if len(cfg.RedirectURIs) == 0 {
		return false, errors.New("oauth2client: redirectURIs is empty")
	}

	name := idToName(cfg.ID)
	desired := renderOAuth2Client(name, cfg)

	existing := newOAuth2ClientObject()
	err := c.Get(ctx, client.ObjectKey{Namespace: cfg.Namespace, Name: name}, existing)
	switch {
	case apierrors.IsNotFound(err):
		if err := c.Create(ctx, desired); err != nil {
			return false, fmt.Errorf("create oauth2client: %w", err)
		}
		return true, nil
	case err != nil:
		return false, fmt.Errorf("get oauth2client: %w", err)
	}

	if sameClientPayload(existing, desired) {
		return false, nil
	}
	desired.SetResourceVersion(existing.GetResourceVersion())
	if err := c.Update(ctx, desired); err != nil {
		return false, fmt.Errorf("update oauth2client: %w", err)
	}
	return true, nil
}

// SplitRedirectURIs parses a comma- or whitespace-separated list of URIs.
func SplitRedirectURIs(s string) []string {
	fields := strings.FieldsFunc(s, func(r rune) bool { return r == ',' || r == ' ' || r == '\t' || r == '\n' })
	out := make([]string, 0, len(fields))
	for _, f := range fields {
		if f = strings.TrimSpace(f); f != "" {
			out = append(out, f)
		}
	}
	return out
}

func renderOAuth2Client(name string, cfg *OAuth2ClientConfig) *unstructured.Unstructured {
	obj := newOAuth2ClientObject()
	obj.SetNamespace(cfg.Namespace)
	obj.SetName(name)
	obj.SetLabels(copyLabels(managedLabels))

	obj.Object["id"] = cfg.ID
	obj.Object["secret"] = cfg.Secret
	obj.Object["name"] = displayName(cfg)
	obj.Object["redirectURIs"] = toAny(cfg.RedirectURIs)
	// The client is public (PKCE) so the SPA can complete the authorization
	// code flow without embedding a secret. The `secret` field is still
	// populated so server-side flows (gateway-mediated login, refresh) can
	// use the same registration later without re-creating the CR.
	obj.Object["public"] = true
	return obj
}

func sameClientPayload(got, want *unstructured.Unstructured) bool {
	keys := []string{"id", "secret", "name", "public"}
	for _, k := range keys {
		if got.Object[k] != want.Object[k] {
			return false
		}
	}
	return reflect.DeepEqual(got.Object["redirectURIs"], want.Object["redirectURIs"])
}

func displayName(cfg *OAuth2ClientConfig) string {
	if cfg.Name != "" {
		return cfg.Name
	}
	return cfg.ID
}

func toAny(in []string) []any {
	out := make([]any, len(in))
	for i, s := range in {
		out[i] = s
	}
	return out
}

func newOAuth2ClientObject() *unstructured.Unstructured {
	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(OAuth2ClientGVK)
	return obj
}
