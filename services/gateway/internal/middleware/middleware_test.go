package middleware_test

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"

	"github.com/bzzdevlabs/crossplane-ui/services/gateway/internal/middleware"
)

func TestRequestIDGeneratesWhenAbsent(t *testing.T) {
	t.Parallel()

	var captured string
	h := middleware.RequestID(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		captured = middleware.RequestIDFrom(r.Context())
	}))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if captured == "" {
		t.Fatal("expected request id in context")
	}
	if rec.Header().Get(middleware.HeaderRequestID) != captured {
		t.Errorf("response header mismatch: %q vs %q",
			rec.Header().Get(middleware.HeaderRequestID), captured)
	}
}

func TestRequestIDPropagatesWhenProvided(t *testing.T) {
	t.Parallel()

	var captured string
	h := middleware.RequestID(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		captured = middleware.RequestIDFrom(r.Context())
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(middleware.HeaderRequestID, "client-supplied-id")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if captured != "client-supplied-id" {
		t.Errorf("expected propagated id, got %q", captured)
	}
}

func TestRecoverTurnsPanicIntoInternalError(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))

	h := middleware.Recover(logger)(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		panic("boom")
	}))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
	if !strings.Contains(buf.String(), "panic in http handler") {
		t.Errorf("expected panic log, got %q", buf.String())
	}
}

func TestCORSAllowsConfiguredOrigin(t *testing.T) {
	t.Parallel()

	inner := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	h := middleware.CORS([]string{"http://localhost:5173"})(inner)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Header().Get("Access-Control-Allow-Origin") != "http://localhost:5173" {
		t.Errorf("missing CORS header: %v", rec.Header())
	}
}

func TestCORSRejectsUnknownOrigin(t *testing.T) {
	t.Parallel()

	inner := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	h := middleware.CORS([]string{"http://localhost:5173"})(inner)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://evil.example.com")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("unexpected CORS header %q", got)
	}
}

func TestCORSHandlesPreflight(t *testing.T) {
	t.Parallel()

	h := middleware.CORS([]string{"*"})(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Fatal("handler should not be called on preflight")
	}))

	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set("Origin", "http://example.com")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("preflight status = %d, want 204", rec.Code)
	}
}

func TestMetricsInstrumentsRequest(t *testing.T) {
	t.Parallel()

	reg := prometheus.NewRegistry()
	reqCtr := prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "test_requests"},
		[]string{"method", "path", "status"},
	)
	durVec := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{Name: "test_duration"},
		[]string{"method", "path"},
	)
	reg.MustRegister(reqCtr, durVec)

	h := middleware.Metrics(reqCtr, durVec)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/x", nil))

	counter, err := reqCtr.GetMetricWithLabelValues("POST", "/x", "201")
	if err != nil {
		t.Fatalf("get metric: %v", err)
	}
	if got := testutil.ToFloat64(counter); got != 1 {
		t.Errorf("counter = %v, want 1", got)
	}
}
