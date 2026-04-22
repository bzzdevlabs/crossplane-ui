package middleware

import (
	"net/http"
	"slices"
)

// CORS replies with Access-Control-Allow-Origin for pre-approved origins and
// handles OPTIONS preflight. The gateway is same-origin in production, so
// this middleware is mostly useful for the Vite dev server running on
// :5173 while the gateway binds :8080.
func CORS(allowedOrigins []string) func(http.Handler) http.Handler {
	wildcard := slices.Contains(allowedOrigins, "*")
	allowed := func(origin string) bool {
		if origin == "" {
			return false
		}
		return wildcard || slices.Contains(allowedOrigins, origin)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if allowed(origin) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, "+HeaderRequestID)
				w.Header().Set("Access-Control-Expose-Headers", HeaderRequestID)
				w.Header().Add("Vary", "Origin")
			}
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
