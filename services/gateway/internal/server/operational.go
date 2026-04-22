package server

import (
	"net/http"
	"sync/atomic"
)

// registerOperational wires the health, readiness and metrics endpoints.
// They intentionally sit outside the auth chain so probes never 401.
func (s *Server) registerOperational(mux *http.ServeMux) {
	mux.HandleFunc("/healthz", healthz)
	mux.HandleFunc("/readyz", readyz)
	mux.Handle("/metrics", s.deps.Registry.Handler())
}

// ready is flipped to true once the server has completed its startup checks.
var ready atomic.Bool

func init() { ready.Store(true) }

func healthz(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func readyz(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	if !ready.Load() {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte("not ready"))
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ready"))
}
