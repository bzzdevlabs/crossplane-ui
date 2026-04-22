package server

import (
	"net/http"
	"sync/atomic"
)

func (s *Server) registerOperational(mux *http.ServeMux) {
	mux.HandleFunc("/healthz", healthz)
	mux.HandleFunc("/readyz", readyz)
	mux.HandleFunc("/metrics", metricsPlaceholder)
}

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

// metricsPlaceholder serves a minimal up gauge on /metrics. The authoritative
// Prometheus endpoint is provided by the controller-runtime manager on a
// separate listener (METRICS_ADDR, default :8082); this endpoint stays so
// that probes and load balancers targeting the HTTP port keep returning 200.
func metricsPlaceholder(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("# HELP crossplane_ui_auth_up 1 when the process is up.\n" +
		"# TYPE crossplane_ui_auth_up gauge\ncrossplane_ui_auth_up 1\n"))
}
