// Command auth is the local-user management service of crossplane-ui.
//
// It reconciles User and Group custom resources, bootstraps the first
// administrator on startup and keeps the Dex password DB in sync with
// those custom resources.
//
// The current build is a scaffold: only health, readiness and metrics
// endpoints are wired. The reconciler and Dex sync loop are added in M3.
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

	"gitlab.telespazio-digital-factory.fr/icdo/tpzf/crossplane-ui/services/auth/internal/buildinfo"
	"gitlab.telespazio-digital-factory.fr/icdo/tpzf/crossplane-ui/services/auth/internal/config"
	"gitlab.telespazio-digital-factory.fr/icdo/tpzf/crossplane-ui/services/auth/internal/logging"
	"gitlab.telespazio-digital-factory.fr/icdo/tpzf/crossplane-ui/services/auth/internal/server"
)

const shutdownTimeout = 30 * time.Second

func main() {
	if err := run(); err != nil {
		//nolint:sloglint // the logger may not be initialised yet.
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

	logger.Info("starting auth",
		slog.String("version", buildinfo.Version),
		slog.String("commit", buildinfo.Commit),
		slog.String("build_date", buildinfo.Date),
		slog.String("http_addr", cfg.HTTPAddr),
	)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	srv := server.New(logger, cfg)

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

	logger.Info("auth stopped cleanly")
	return nil
}
