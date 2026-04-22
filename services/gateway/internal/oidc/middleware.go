package oidc

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
)

// Middleware rejects requests that do not carry a valid Bearer token and
// injects the resolved User into the request context on success.
//
// The middleware intentionally leaks very little through the API: a
// missing, malformed or invalid token all collapse to 401 with a static
// JSON body. Details surface only in the server logs.
func Middleware(v Verifier, logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			raw := extractBearer(r)
			if raw == "" {
				writeUnauthorized(w, "missing_bearer_token")
				return
			}
			tok, err := v.Verify(r.Context(), raw)
			if err != nil {
				logger.WarnContext(r.Context(), "oidc verification failed",
					slog.String("error", err.Error()))
				writeUnauthorized(w, "invalid_token")
				return
			}
			var claims struct {
				Sub               string   `json:"sub"`
				Email             string   `json:"email"`
				PreferredUsername string   `json:"preferred_username"`
				Groups            []string `json:"groups"`
			}
			if err := tok.Claims(&claims); err != nil {
				logger.WarnContext(r.Context(), "oidc claims decode failed",
					slog.String("error", err.Error()))
				writeUnauthorized(w, "invalid_claims")
				return
			}
			user := User{
				Subject:           claims.Sub,
				Email:             claims.Email,
				PreferredUsername: claims.PreferredUsername,
				Groups:            claims.Groups,
			}
			next.ServeHTTP(w, r.WithContext(WithUser(r.Context(), user)))
		})
	}
}

// DevPassthrough is the no-op auth middleware used when `OIDC_ISSUER_URL`
// is empty. It injects a synthetic "dev-admin" user with cluster-admin
// groups so the request can still be impersonated against the local k3d
// cluster. It MUST NOT be enabled in production.
func DevPassthrough(next http.Handler) http.Handler {
	devUser := User{
		Subject:           "dev",
		PreferredUsername: "dev-admin",
		Email:             "dev-admin@local",
		Groups:            []string{"system:masters"},
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r.WithContext(WithUser(r.Context(), devUser)))
	})
}

func extractBearer(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if !strings.HasPrefix(h, "Bearer ") {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(h, "Bearer "))
}

func writeUnauthorized(w http.ResponseWriter, code string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": code})
}
