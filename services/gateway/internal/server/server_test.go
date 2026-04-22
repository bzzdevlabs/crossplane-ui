package server_test

import (
	"context"
	"io"
	"log/slog"
	"net"
	"net/http"
	"testing"
	"time"

	"k8s.io/client-go/kubernetes"

	"gitlab.telespazio-digital-factory.fr/icdo/tpzf/crossplane-ui/services/gateway/internal/config"
	"gitlab.telespazio-digital-factory.fr/icdo/tpzf/crossplane-ui/services/gateway/internal/metrics"
	"gitlab.telespazio-digital-factory.fr/icdo/tpzf/crossplane-ui/services/gateway/internal/oidc"
	"gitlab.telespazio-digital-factory.fr/icdo/tpzf/crossplane-ui/services/gateway/internal/server"
)

// nopFactory satisfies api.ClientFactory; unit tests in this package do not
// exercise any /api/v1/* route, so the factory is never actually invoked.
type nopFactory struct{}

func (nopFactory) For(string, []string) (kubernetes.Interface, error) { return nil, nil }

// newTestServer starts a Server on an ephemeral port and returns its base URL
// along with a cleanup function that shuts it down.
func newTestServer(t *testing.T) (baseURL string, shutdown func()) {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	addr := ln.Addr().String()
	_ = ln.Close()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	cfg := &config.Config{
		HTTPAddr:          addr,
		ReadHeaderTimeout: time.Second,
		LogLevel:          "info",
		LogFormat:         "text",
	}
	s := server.New(server.Deps{
		Logger:         logger,
		Config:         cfg,
		Registry:       metrics.New(),
		ClientFactory:  nopFactory{},
		AuthMiddleware: oidc.DevPassthrough,
	})

	done := make(chan struct{})
	go func() {
		defer close(done)
		_ = s.ListenAndServe()
	}()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", addr, 100*time.Millisecond)
		if err == nil {
			_ = conn.Close()
			break
		}
		time.Sleep(20 * time.Millisecond)
	}

	return "http://" + addr, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = s.Shutdown(ctx)
		<-done
	}
}

func TestHealthzReturnsOK(t *testing.T) {
	t.Parallel()
	base, shutdown := newTestServer(t)
	t.Cleanup(shutdown)

	resp, err := http.Get(base + "/healthz") //nolint:noctx // simple test
	if err != nil {
		t.Fatalf("GET /healthz: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestReadyzReturnsOK(t *testing.T) {
	t.Parallel()
	base, shutdown := newTestServer(t)
	t.Cleanup(shutdown)

	resp, err := http.Get(base + "/readyz") //nolint:noctx // simple test
	if err != nil {
		t.Fatalf("GET /readyz: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestMetricsExposesPrometheusText(t *testing.T) {
	t.Parallel()
	base, shutdown := newTestServer(t)
	t.Cleanup(shutdown)

	resp, err := http.Get(base + "/metrics") //nolint:noctx // simple test
	if err != nil {
		t.Fatalf("GET /metrics: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	if len(body) == 0 {
		t.Fatal("empty metrics body")
	}
}
