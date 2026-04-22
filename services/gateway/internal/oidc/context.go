// Package oidc verifies ID tokens issued by the gateway's OpenID Connect
// provider (Dex in our default topology) and exposes the resulting user
// through the request context.
package oidc

import "context"

// User is the authenticated subject extracted from a verified ID token.
// Every field is a direct projection of a standard OIDC claim; consumers
// should treat zero values as "claim absent" rather than as an error.
type User struct {
	// Subject mirrors the `sub` claim: a stable, opaque identifier.
	Subject string
	// PreferredUsername mirrors the `preferred_username` claim.
	PreferredUsername string
	// Email mirrors the `email` claim.
	Email string
	// Groups mirrors the `groups` claim (non-standard but widely used by
	// IdPs such as Dex and Keycloak).
	Groups []string
}

// Kubernetes returns the impersonation user name and group list used by the
// gateway's Kubernetes client. We prefer the preferred username when
// present, fall back to the email, and finally to the opaque subject so
// that RoleBindings authored against any of those three shapes keep working.
func (u User) Kubernetes() (string, []string) {
	switch {
	case u.PreferredUsername != "":
		return u.PreferredUsername, u.Groups
	case u.Email != "":
		return u.Email, u.Groups
	default:
		return u.Subject, u.Groups
	}
}

type userKey struct{}

// WithUser returns a child context carrying u.
func WithUser(ctx context.Context, u User) context.Context {
	return context.WithValue(ctx, userKey{}, u)
}

// UserFromContext returns the authenticated user, or ok=false when none
// has been attached by the auth middleware.
func UserFromContext(ctx context.Context) (User, bool) {
	u, ok := ctx.Value(userKey{}).(User)
	return u, ok
}
