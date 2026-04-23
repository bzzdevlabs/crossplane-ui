package metrics_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bzzdevlabs/crossplane-ui/services/gateway/internal/metrics"
)

func TestRegistryExposesExpectedMetrics(t *testing.T) {
	t.Parallel()

	reg := metrics.New()
	reg.HTTPRequests.WithLabelValues("GET", "/healthz", "200").Inc()
	reg.HTTPDuration.WithLabelValues("GET", "/healthz").Observe(0.01)
	reg.KubeAPIRequests.WithLabelValues("list", "namespaces", "200").Inc()

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	reg.Handler().ServeHTTP(rec, req)

	body, _ := io.ReadAll(rec.Body)
	text := string(body)

	for _, want := range []string{
		"crossplane_ui_gateway_http_requests_total",
		"crossplane_ui_gateway_http_request_duration_seconds",
		"crossplane_ui_gateway_kube_api_requests_total",
		"go_goroutines",
		"process_cpu_seconds_total",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("metrics output missing %q", want)
		}
	}
}
