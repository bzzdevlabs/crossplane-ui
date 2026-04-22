package api

import (
	"log/slog"
	"net/http"

	"gitlab.telespazio-digital-factory.fr/icdo/tpzf/crossplane-ui/services/gateway/internal/config"
)

// ConfigResponse is the public-facing subset of the gateway's configuration
// the web UI needs at boot time to bootstrap its OIDC client and feature
// flags. Do NOT leak secrets through this payload.
type ConfigResponse struct {
	// Auth carries the OIDC parameters the SPA uses with PKCE.
	Auth AuthConfig `json:"auth"`
	// Version is the semver tag of the gateway build (empty in dev).
	Version string `json:"version,omitempty"`
}

// AuthConfig describes how the UI should talk to the identity provider.
type AuthConfig struct {
	// Enabled reports whether OIDC is active. When false the gateway runs in
	// dev-passthrough mode and no login flow is required.
	Enabled bool `json:"enabled"`
	// IssuerURL is the Dex issuer the browser redirects to.
	IssuerURL string `json:"issuerURL,omitempty"`
	// ClientID is the Dex OAuth2Client identifier registered for the SPA.
	ClientID string `json:"clientID,omitempty"`
	// Scopes is the list of OIDC scopes to request.
	Scopes []string `json:"scopes,omitempty"`
}

// ConfigHandler returns the UI bootstrap document. Served without auth so the
// login page itself can fetch it.
func ConfigHandler(logger *slog.Logger, cfg *config.Config, version string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed")
			return
		}
		payload := ConfigResponse{
			Version: version,
			Auth: AuthConfig{
				Enabled:   cfg.AuthEnabled(),
				IssuerURL: cfg.OIDCIssuerURL,
				ClientID:  cfg.OIDCClientID,
				Scopes:    []string{"openid", "profile", "email", "groups", "offline_access"},
			},
		}
		writeJSON(w, logger, http.StatusOK, payload)
	})
}
