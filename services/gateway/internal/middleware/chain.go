// Package middleware collects the net/http middlewares applied by the gateway.
//
// Every middleware is a plain `func(http.Handler) http.Handler`, which keeps
// the package free of framework lock-in and testable with the standard
// `net/http/httptest` harness.
package middleware

import "net/http"

// Chain composes middlewares so that the first argument is the outermost
// wrapper. `Chain(a, b, c)(h)` runs the request through `a → b → c → h`.
func Chain(mw ...func(http.Handler) http.Handler) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		for i := len(mw) - 1; i >= 0; i-- {
			h = mw[i](h)
		}
		return h
	}
}

// statusRecorder wraps http.ResponseWriter to expose the final status code
// and response size to surrounding middleware.
type statusRecorder struct {
	http.ResponseWriter
	status int
	bytes  int
}

func newStatusRecorder(w http.ResponseWriter) *statusRecorder {
	return &statusRecorder{ResponseWriter: w, status: http.StatusOK}
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

func (s *statusRecorder) Write(b []byte) (int, error) {
	n, err := s.ResponseWriter.Write(b)
	s.bytes += n
	return n, err
}
