package webui_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bzzdevlabs/crossplane-ui/services/gateway/internal/webui"
)

func TestHandlerServesIndexForRoot(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	webui.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rr.Code)
	}
	body, _ := io.ReadAll(rr.Body)
	if !strings.Contains(string(body), "<html") {
		t.Fatalf("expected HTML in body, got %q", body)
	}
}

func TestHandlerFallsBackToIndexForSPARoute(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/deeply/nested/vue/route", nil)
	webui.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("want 200 (SPA fallback), got %d", rr.Code)
	}
	body, _ := io.ReadAll(rr.Body)
	if !strings.Contains(string(body), "<html") {
		t.Fatalf("expected index.html content, got %q", body)
	}
}
