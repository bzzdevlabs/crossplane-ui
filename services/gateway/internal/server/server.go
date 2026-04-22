// Package server wires the HTTP router and middlewares for the gateway.
//
// Composition:
//
//	Recover → RequestID → AccessLog → Metrics → CORS → Mux
//	                                                      │
//	                                                      ├── /healthz, /readyz, /metrics
//	                                                      └── /api/v1/* (auth mw + handlers)
package server

import (
	"context"
	"log/slog"
	"net/http"

	"gitlab.telespazio-digital-factory.fr/icdo/tpzf/crossplane-ui/services/gateway/internal/api"
	"gitlab.telespazio-digital-factory.fr/icdo/tpzf/crossplane-ui/services/gateway/internal/config"
	"gitlab.telespazio-digital-factory.fr/icdo/tpzf/crossplane-ui/services/gateway/internal/metrics"
	"gitlab.telespazio-digital-factory.fr/icdo/tpzf/crossplane-ui/services/gateway/internal/middleware"
)

// Deps bundles the heavyweight collaborators the server needs to run.
// Callers build them once in main() and hand them over.
type Deps struct {
	Logger        *slog.Logger
	Config        *config.Config
	Registry      *metrics.Registry
	ClientFactory api.ClientFactory
	// AuthMiddleware must either enforce OIDC or act as the dev pass-through
	// (injecting a synthetic user) so downstream handlers always find a user
	// in context.
	AuthMiddleware func(http.Handler) http.Handler
}

// Server is the gateway HTTP server.
type Server struct {
	deps Deps
	http *http.Server
}

// New builds a Server ready to listen.
func New(deps Deps) *Server {
	s := &Server{deps: deps}

	mux := http.NewServeMux()
	s.registerOperational(mux)
	s.registerAPI(mux)

	// Outer chain applied to every route. Order matters:
	//   Recover   — captures panics from every later layer
	//   RequestID — attaches the identifier every log/metric expects
	//   AccessLog — observes the final status + duration
	//   Metrics   — records request counts and latency
	//   CORS      — browser integration; no-op when the request is same-origin
	chain := middleware.Chain(
		middleware.Recover(deps.Logger),
		middleware.RequestID,
		middleware.AccessLog(deps.Logger),
		middleware.Metrics(deps.Registry.HTTPRequests, deps.Registry.HTTPDuration),
		middleware.CORS(deps.Config.CORSAllowedOrigins),
	)

	s.http = &http.Server{
		Addr:              deps.Config.HTTPAddr,
		Handler:           chain(mux),
		ReadHeaderTimeout: deps.Config.ReadHeaderTimeout,
	}
	return s
}

// ListenAndServe starts the server. It blocks until the server exits.
func (s *Server) ListenAndServe() error {
	s.deps.Logger.Info("http server listening",
		slog.String("addr", s.deps.Config.HTTPAddr),
		slog.Bool("auth_enabled", s.deps.Config.AuthEnabled()),
	)
	return s.http.ListenAndServe()
}

// Shutdown gracefully terminates the server.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.http.Shutdown(ctx)
}

// registerAPI mounts the /api/v1/* routes behind the auth middleware.
func (s *Server) registerAPI(mux *http.ServeMux) {
	apiMux := http.NewServeMux()
	apiMux.Handle("/api/v1/namespaces", api.NamespacesHandler(s.deps.Logger, s.deps.ClientFactory))

	mux.Handle("/api/v1/", s.deps.AuthMiddleware(apiMux))
}
