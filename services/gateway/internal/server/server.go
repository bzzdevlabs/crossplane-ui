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

	"github.com/bzzdevlabs/crossplane-ui/services/gateway/internal/api"
	"github.com/bzzdevlabs/crossplane-ui/services/gateway/internal/config"
	"github.com/bzzdevlabs/crossplane-ui/services/gateway/internal/metrics"
	"github.com/bzzdevlabs/crossplane-ui/services/gateway/internal/middleware"
	"github.com/bzzdevlabs/crossplane-ui/services/gateway/internal/webui"
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
	// CrossplaneFactory builds impersonated dynamic + discovery clients used
	// by the Crossplane dashboard endpoints. Optional; nil disables the
	// /api/v1/crossplane/* routes.
	CrossplaneFactory api.CrossplaneFactory
	// Version is the gateway's build tag, surfaced via /api/v1/config so the
	// UI can show it and pin against mismatched clients.
	Version string
}

// Server is the gateway HTTP server.
type Server struct {
	deps Deps
	http *http.Server
}

// New builds a Server ready to listen.
func New(deps *Deps) *Server {
	s := &Server{deps: *deps}

	mux := http.NewServeMux()
	s.registerOperational(mux)
	s.registerAPI(mux)
	// SPA catch-all; longest-prefix match means /api/v1/*, /healthz, /readyz
	// and /metrics still win over "/".
	mux.Handle("/", webui.Handler())

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

// registerAPI mounts the /api/v1/* routes. The public config endpoint is
// exposed without auth (the UI fetches it at boot to configure OIDC); every
// other route sits behind the auth middleware.
func (s *Server) registerAPI(mux *http.ServeMux) {
	mux.Handle("/api/v1/config", api.ConfigHandler(s.deps.Logger, s.deps.Config, s.deps.Version))

	apiMux := http.NewServeMux()
	apiMux.Handle("/api/v1/namespaces", api.NamespacesHandler(s.deps.Logger, s.deps.ClientFactory))
	if s.deps.CrossplaneFactory != nil {
		apiMux.Handle("/api/v1/crossplane/resources",
			api.CrossplaneResourcesHandler(s.deps.Logger, s.deps.CrossplaneFactory))
		apiMux.Handle("/api/v1/crossplane/resource",
			api.ResourceHandler(s.deps.Logger, s.deps.CrossplaneFactory))
		apiMux.Handle("/api/v1/auth/connectors",
			api.ConnectorsHandler(s.deps.Logger, s.deps.CrossplaneFactory))
		apiMux.Handle("/api/v1/auth/users",
			api.UsersHandler(s.deps.Logger, s.deps.CrossplaneFactory))
		apiMux.Handle("/api/v1/auth/groups",
			api.GroupsHandler(s.deps.Logger, s.deps.CrossplaneFactory))
	}
	apiMux.Handle("/api/v1/auth/connector-secrets",
		api.ConnectorSecretsHandler(s.deps.Logger, s.deps.ClientFactory))
	apiMux.Handle("/api/v1/auth/user-password",
		api.UserPasswordHandler(s.deps.Logger, s.deps.ClientFactory))

	mux.Handle("/api/v1/", s.deps.AuthMiddleware(apiMux))
}
