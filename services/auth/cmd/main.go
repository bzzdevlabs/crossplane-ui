// Command auth is the local-user management service of crossplane-ui.
//
// It reconciles User and Group custom resources, bootstraps the first
// administrator on startup and keeps the Dex password DB in sync with
// those custom resources.
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

	"github.com/go-logr/logr"
	"golang.org/x/sync/errgroup"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logsv1 "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/bzzdevlabs/crossplane-ui/services/auth/internal/bootstrap"
	"github.com/bzzdevlabs/crossplane-ui/services/auth/internal/buildinfo"
	"github.com/bzzdevlabs/crossplane-ui/services/auth/internal/config"
	"github.com/bzzdevlabs/crossplane-ui/services/auth/internal/dex"
	"github.com/bzzdevlabs/crossplane-ui/services/auth/internal/kube"
	"github.com/bzzdevlabs/crossplane-ui/services/auth/internal/logging"
	mgrpkg "github.com/bzzdevlabs/crossplane-ui/services/auth/internal/manager"
	"github.com/bzzdevlabs/crossplane-ui/services/auth/internal/server"
)

const shutdownTimeout = 30 * time.Second

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
	// Forward controller-runtime's structured logs through the same slog
	// handler so everything ends up on stderr in one format.
	ctrlLogger := logr.FromSlogHandler(logger.Handler())
	logsv1.SetLogger(ctrlLogger)
	ctrl.SetLogger(ctrlLogger)

	logger.Info("starting auth",
		slog.String("version", buildinfo.Version),
		slog.String("commit", buildinfo.Commit),
		slog.String("build_date", buildinfo.Date),
		slog.String("http_addr", cfg.HTTPAddr),
		slog.String("namespace", cfg.Namespace),
	)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	restCfg, err := kube.LoadConfig(cfg.KubeconfigPath)
	if err != nil {
		return fmt.Errorf("kube config: %w", err)
	}

	// One-shot bootstrap: create admin User + scrub secret.
	directClient, err := mgrpkg.NewDirectClient(restCfg)
	if err != nil {
		return fmt.Errorf("direct client: %w", err)
	}
	if err := bootstrap.Run(ctx, directClient, bootstrap.Config{
		Namespace:       cfg.Namespace,
		SecretName:      cfg.BootstrapAdminPasswordSecret,
		DefaultUsername: cfg.BootstrapAdminUsername,
	}); err != nil {
		logger.Warn("bootstrap skipped or failed — controller will retry on first event",
			slog.String("error", err.Error()))
	}

	if err := ensureOAuth2Client(ctx, directClient, cfg, logger); err != nil {
		logger.Warn("oauth2 client bootstrap skipped or failed — login will be unavailable until resolved",
			slog.String("error", err.Error()))
	}

	mgr, err := mgrpkg.New(mgrpkg.Options{
		RestConfig: restCfg,
		Config:     cfg,
		Logger:     ctrlLogger,
	})
	if err != nil {
		return fmt.Errorf("build manager: %w", err)
	}

	srv := server.New(logger, cfg)

	g, gCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		if err := mgr.Start(gCtx); err != nil {
			return fmt.Errorf("manager: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		logger.Info("http server listening", slog.String("addr", cfg.HTTPAddr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("http server: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		<-gCtx.Done()
		// Detach from the cancelled parent so Shutdown gets the full timeout
		// to drain in-flight requests; contextcheck expects the detached
		// context to be derived from the original.
		shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(gCtx), shutdownTimeout)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("graceful shutdown: %w", err)
		}
		return nil
	})

	if err := g.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		return err
	}

	logger.Info("auth stopped cleanly")
	return nil
}

// ensureOAuth2Client reads the shared OIDC client secret from a Kubernetes
// Secret and upserts the matching OAuth2Client Dex CR. Retries with a capped
// backoff so first-install orderings (where Dex has not yet finished creating
// its CRDs) do not leave the gateway unable to authenticate.
func ensureOAuth2Client(ctx context.Context, c client.Client, cfg *config.Config, logger *slog.Logger) error {
	if cfg.OIDCClientID == "" || cfg.OIDCClientSecretSecret == "" {
		logger.InfoContext(ctx, "oauth2 client bootstrap skipped",
			slog.String("reason", "OIDC_CLIENT_ID or OIDC_CLIENT_SECRET_NAME not set"))
		return nil
	}
	if len(cfg.OIDCRedirectURIs) == 0 {
		return errors.New("OIDC_REDIRECT_URIS must be set when OIDC_CLIENT_ID is set")
	}

	var sec corev1.Secret
	err := c.Get(ctx, types.NamespacedName{Namespace: cfg.Namespace, Name: cfg.OIDCClientSecretSecret}, &sec)
	if apierrors.IsNotFound(err) {
		return fmt.Errorf("oidc client secret %q not found in %q", cfg.OIDCClientSecretSecret, cfg.Namespace)
	}
	if err != nil {
		return fmt.Errorf("get oidc client secret: %w", err)
	}
	raw, ok := sec.Data[cfg.OIDCClientSecretKey]
	if !ok || len(raw) == 0 {
		return fmt.Errorf("oidc client secret %q missing key %q", cfg.OIDCClientSecretSecret, cfg.OIDCClientSecretKey)
	}

	clientCfg := &dex.OAuth2ClientConfig{
		Namespace:    cfg.Namespace,
		ID:           cfg.OIDCClientID,
		Secret:       string(raw),
		Name:         cfg.OIDCClientName,
		RedirectURIs: cfg.OIDCRedirectURIs,
	}

	const maxAttempts = 8
	backoff := time.Second
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		changed, err := dex.EnsureOAuth2Client(ctx, c, clientCfg)
		if err == nil {
			logger.InfoContext(ctx, "oauth2 client reconciled",
				slog.String("client_id", cfg.OIDCClientID),
				slog.Any("redirect_uris", cfg.OIDCRedirectURIs),
				slog.Bool("changed", changed),
				slog.Int("attempt", attempt))
			return nil
		}
		if attempt == maxAttempts {
			return fmt.Errorf("ensure oauth2client after %d attempts: %w", maxAttempts, err)
		}
		logger.WarnContext(ctx, "oauth2 client upsert failed, retrying",
			slog.String("error", err.Error()),
			slog.Int("attempt", attempt),
			slog.Duration("backoff", backoff))
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(backoff):
		}
		if backoff < 30*time.Second {
			backoff *= 2
		}
	}
	return nil
}
