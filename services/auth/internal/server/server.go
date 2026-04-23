// Package server wires the HTTP router of the auth service.
//
// The auth service deliberately keeps a small public API surface: it is
// mostly a Kubernetes controller that watches User/Group custom resources.
// Only operational and simple admin endpoints are exposed over HTTP.
package server

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/bzzdevlabs/crossplane-ui/services/auth/internal/config"
)

// Server is the auth HTTP server.
type Server struct {
	logger *slog.Logger
	cfg    *config.Config
	http   *http.Server
}

// New builds a new Server ready to call ListenAndServe on.
func New(logger *slog.Logger, cfg *config.Config) *Server {
	s := &Server{logger: logger, cfg: cfg}

	mux := http.NewServeMux()
	s.registerOperational(mux)
	s.registerAdmin(mux)

	s.http = &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           requestLogger(logger, mux),
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
	}
	return s
}

// ListenAndServe starts the server. It blocks until the server exits.
func (s *Server) ListenAndServe() error {
	return s.http.ListenAndServe()
}

// Shutdown gracefully terminates the server.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.http.Shutdown(ctx)
}

func (s *Server) registerAdmin(mux *http.ServeMux) {
	mux.HandleFunc("/admin/v1/", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotImplemented)
		_, _ = w.Write([]byte(`{"error":"not_implemented","message":"Admin API is wired in milestone M3"}`))
	})
}

func requestLogger(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		logger.LogAttrs(r.Context(), slog.LevelInfo, "http request",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Int("status", rec.status),
			slog.Duration("duration", time.Since(start)),
			slog.String("remote_addr", r.RemoteAddr),
		)
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}
