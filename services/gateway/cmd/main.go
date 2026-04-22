// Command gateway is the HTTP entrypoint for crossplane-ui.
//
// It serves the embedded Vue UI, verifies OIDC tokens against Dex, and
// forwards API calls to the Kubernetes API with user impersonation.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gitlab.telespazio-digital-factory.fr/icdo/tpzf/crossplane-ui/services/gateway/internal/buildinfo"
	"gitlab.telespazio-digital-factory.fr/icdo/tpzf/crossplane-ui/services/gateway/internal/config"
	"gitlab.telespazio-digital-factory.fr/icdo/tpzf/crossplane-ui/services/gateway/internal/kube"
	"gitlab.telespazio-digital-factory.fr/icdo/tpzf/crossplane-ui/services/gateway/internal/logging"
	"gitlab.telespazio-digital-factory.fr/icdo/tpzf/crossplane-ui/services/gateway/internal/metrics"
	"gitlab.telespazio-digital-factory.fr/icdo/tpzf/crossplane-ui/services/gateway/internal/oidc"
	"gitlab.telespazio-digital-factory.fr/icdo/tpzf/crossplane-ui/services/gateway/internal/server"
)

const (
	// shutdownTimeout bounds how long we wait for in-flight requests to finish
	// after receiving SIGTERM before we force-close connections.
	shutdownTimeout = 30 * time.Second
	// oidcDiscoveryTimeout bounds the blocking OIDC discovery fetch at startup.
	oidcDiscoveryTimeout = 15 * time.Second
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "fatal: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	logger := logging.New(cfg.LogLevel, cfg.LogFormat)
	slog.SetDefault(logger)

	logger.Info("starting gateway",
		slog.String("version", buildinfo.Version),
		slog.String("commit", buildinfo.Commit),
		slog.String("build_date", buildinfo.Date),
		slog.String("http_addr", cfg.HTTPAddr),
		slog.Bool("auth_enabled", cfg.AuthEnabled()),
	)

	kubeCfg, err := kube.LoadConfig(cfg.KubeconfigPath)
	if err != nil {
		return fmt.Errorf("kube config: %w", err)
	}
	factory := kube.NewClientFactory(kubeCfg)

	reg := metrics.New()

	authMW, err := buildAuthMiddleware(cfg, logger)
	if err != nil {
		return fmt.Errorf("auth middleware: %w", err)
	}

	srv := server.New(server.Deps{
		Logger:         logger,
		Config:         cfg,
		Registry:       reg,
		ClientFactory:  factory,
		AuthMiddleware: authMW,
	})

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- fmt.Errorf("http server: %w", err)
			return
		}
		errCh <- nil
	}()

	select {
	case <-ctx.Done():
		logger.Info("shutdown signal received")
	case err := <-errCh:
		if err != nil {
			return err
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("graceful shutdown: %w", err)
	}

	logger.Info("gateway stopped cleanly")
	return nil
}

func buildAuthMiddleware(cfg *config.Config, logger *slog.Logger) (func(http.Handler) http.Handler, error) {
	if !cfg.AuthEnabled() {
		logger.Warn("OIDC_ISSUER_URL is empty — running with dev pass-through (NO AUTHENTICATION); " +
			"do not use this mode outside local development")
		return oidc.DevPassthrough, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), oidcDiscoveryTimeout)
	defer cancel()

	verifier, err := oidc.NewVerifier(ctx, oidc.Config{
		DiscoveryURL:    cfg.EffectiveDiscoveryURL(),
		ExpectedIssuer:  cfg.OIDCIssuerURL,
		ClientID:        cfg.OIDCClientID,
		SkipIssuerCheck: cfg.OIDCSkipIssuerCheck,
	})
	if err != nil {
		return nil, err
	}
	logger.Info("oidc verifier ready",
		slog.String("discovery_url", cfg.EffectiveDiscoveryURL()),
		slog.String("expected_issuer", cfg.OIDCIssuerURL),
		slog.String("client_id", cfg.OIDCClientID),
	)
	return oidc.Middleware(verifier, logger), nil
}
