package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"
)

// Recover catches panics raised by downstream handlers, logs the stack and
// returns a 500 response without leaking internals to the client.
func Recover(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			defer func() {
				rec := recover()
				if rec == nil {
					return
				}
				logger.ErrorContext(ctx, "panic in http handler",
					slog.Any("panic", rec),
					slog.String("stack", string(debug.Stack())),
					slog.String("request_id", RequestIDFrom(ctx)),
				)
				http.Error(w, `{"error":"internal_server_error"}`, http.StatusInternalServerError)
			}()
			next.ServeHTTP(w, r)
		})
	}
}
