package oidc

import (
	"context"
	"errors"
	"fmt"

	coreosoidc "github.com/coreos/go-oidc/v3/oidc"
)

// Config drives the construction of an *coreosoidc.IDTokenVerifier.
//
// The `DiscoveryURL` and `ExpectedIssuer` distinction matters in compose
// setups where Dex advertises itself to the browser under a hostname the
// gateway cannot resolve (e.g. `http://localhost:5556/dex`) while the
// gateway itself reaches Dex via a docker-internal name
// (`http://dex:5556/dex`). The gateway fetches discovery from the latter
// but must still accept tokens whose `iss` matches the former.
type Config struct {
	// DiscoveryURL is where the gateway fetches
	// `/.well-known/openid-configuration`. Required.
	DiscoveryURL string
	// ExpectedIssuer is the value the `iss` claim must match.
	// When empty, defaults to DiscoveryURL.
	ExpectedIssuer string
	// ClientID is the OIDC client identifier against which the `aud` claim
	// is validated. Required.
	ClientID string
	// SkipIssuerCheck disables the `iss` check entirely. Dev-only escape
	// hatch; never set to true in production.
	SkipIssuerCheck bool
}

// Verifier is the interface the middleware consumes. Having our own minimal
// interface lets tests inject a fake without importing go-oidc.
type Verifier interface {
	Verify(ctx context.Context, rawIDToken string) (IDToken, error)
}

// IDToken is the minimal surface the middleware needs from a verified
// token — decoding claims.
type IDToken interface {
	Claims(v any) error
}

// NewVerifier builds a Verifier by performing OIDC discovery against cfg.
// The ctx is used for the discovery fetch only.
func NewVerifier(ctx context.Context, cfg Config) (Verifier, error) {
	if cfg.DiscoveryURL == "" {
		return nil, errors.New("oidc: DiscoveryURL is required")
	}
	if cfg.ClientID == "" {
		return nil, errors.New("oidc: ClientID is required")
	}

	expected := cfg.ExpectedIssuer
	if expected == "" {
		expected = cfg.DiscoveryURL
	}

	discoveryCtx := ctx
	if expected != cfg.DiscoveryURL {
		discoveryCtx = coreosoidc.InsecureIssuerURLContext(ctx, expected)
	}

	provider, err := coreosoidc.NewProvider(discoveryCtx, cfg.DiscoveryURL)
	if err != nil {
		return nil, fmt.Errorf("oidc discovery: %w", err)
	}

	inner := provider.Verifier(&coreosoidc.Config{
		ClientID:        cfg.ClientID,
		SkipIssuerCheck: cfg.SkipIssuerCheck,
	})
	return &coreosVerifier{inner: inner}, nil
}

type coreosVerifier struct {
	inner *coreosoidc.IDTokenVerifier
}

func (v *coreosVerifier) Verify(ctx context.Context, rawIDToken string) (IDToken, error) {
	tok, err := v.inner.Verify(ctx, rawIDToken)
	if err != nil {
		return nil, err
	}
	return &coreosIDToken{inner: tok}, nil
}

type coreosIDToken struct {
	inner *coreosoidc.IDToken
}

func (t *coreosIDToken) Claims(v any) error { return t.inner.Claims(v) }
