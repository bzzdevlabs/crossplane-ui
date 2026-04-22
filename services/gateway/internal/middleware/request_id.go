package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

// HeaderRequestID is the HTTP header through which request identifiers flow
// in and out of the gateway.
const HeaderRequestID = "X-Request-ID"

type requestIDKey struct{}

// RequestID ensures every request carries a stable identifier, generating
// one when the client does not supply `X-Request-ID`. The identifier is
// both echoed in the response and attached to the request context.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get(HeaderRequestID)
		if id == "" {
			id = uuid.NewString()
		}
		w.Header().Set(HeaderRequestID, id)
		ctx := context.WithValue(r.Context(), requestIDKey{}, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequestIDFrom retrieves the request identifier stored by the RequestID
// middleware, or the empty string if none was set.
func RequestIDFrom(ctx context.Context) string {
	if v, ok := ctx.Value(requestIDKey{}).(string); ok {
		return v
	}
	return ""
}
