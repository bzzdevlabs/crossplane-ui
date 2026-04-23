package api_test

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bzzdevlabs/crossplane-ui/services/gateway/internal/api"
	"github.com/bzzdevlabs/crossplane-ui/services/gateway/internal/config"
)

func TestConfigHandlerReturnsOIDCDetailsWhenEnabled(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		OIDCIssuerURL: "http://dex.local/dex",
		OIDCClientID:  "crossplane-ui",
	}
	handler := api.ConfigHandler(slog.New(slog.NewTextHandler(io.Discard, nil)), cfg, "1.2.3")

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/config", nil)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d, want %d", rr.Code, http.StatusOK)
	}

	var body api.ConfigResponse
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Version != "1.2.3" {
		t.Errorf("version: got %q, want %q", body.Version, "1.2.3")
	}
	if !body.Auth.Enabled {
		t.Error("auth.enabled: got false, want true")
	}
	if body.Auth.IssuerURL != "http://dex.local/dex" {
		t.Errorf("auth.issuerURL: got %q", body.Auth.IssuerURL)
	}
	if body.Auth.ClientID != "crossplane-ui" {
		t.Errorf("auth.clientID: got %q", body.Auth.ClientID)
	}
	if len(body.Auth.Scopes) == 0 {
		t.Error("auth.scopes should not be empty")
	}
}

func TestConfigHandlerReportsDisabledWhenIssuerEmpty(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{}
	handler := api.ConfigHandler(slog.New(slog.NewTextHandler(io.Discard, nil)), cfg, "")

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/config", nil)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d, want %d", rr.Code, http.StatusOK)
	}
	var body api.ConfigResponse
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Auth.Enabled {
		t.Error("auth.enabled: got true, want false")
	}
	if body.Auth.IssuerURL != "" {
		t.Errorf("auth.issuerURL: want empty, got %q", body.Auth.IssuerURL)
	}
}

func TestConfigHandlerRejectsNonGET(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{}
	handler := api.ConfigHandler(slog.New(slog.NewTextHandler(io.Discard, nil)), cfg, "")

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/config", nil)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status: got %d, want %d", rr.Code, http.StatusMethodNotAllowed)
	}
}
