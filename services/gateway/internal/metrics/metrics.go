// Package metrics assembles the gateway's Prometheus registry and exposes
// the canonical counters and histograms used across the codebase.
//
// Every metric lives under the namespace "crossplane_ui" and the subsystem
// "gateway", keeping cardinality low and making dashboards portable.
package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Registry bundles a Prometheus registry pre-populated with the collectors
// the gateway reports on. The exported *Vec fields are ready to be labeled
// by the middleware and handlers that own them.
type Registry struct {
	*prometheus.Registry

	HTTPRequests    *prometheus.CounterVec
	HTTPDuration    *prometheus.HistogramVec
	KubeAPIRequests *prometheus.CounterVec
}

// New builds a fresh Registry with Go runtime, process, and gateway-specific
// metrics registered.
func New() *Registry {
	reg := prometheus.NewRegistry()
	reg.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

	const ns, sub = "crossplane_ui", "gateway"

	httpReq := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: ns,
		Subsystem: sub,
		Name:      "http_requests_total",
		Help:      "Total HTTP requests handled by the gateway, labelled by method, path and response status.",
	}, []string{"method", "path", "status"})

	httpDur := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: ns,
		Subsystem: sub,
		Name:      "http_request_duration_seconds",
		Help:      "HTTP request duration histogram, labelled by method and path.",
		Buckets:   prometheus.DefBuckets,
	}, []string{"method", "path"})

	kubeReq := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: ns,
		Subsystem: sub,
		Name:      "kube_api_requests_total",
		Help:      "Outgoing Kubernetes API calls issued by the gateway, labelled by verb, resource and response code.",
	}, []string{"verb", "resource", "code"})

	reg.MustRegister(httpReq, httpDur, kubeReq)

	return &Registry{
		Registry:        reg,
		HTTPRequests:    httpReq,
		HTTPDuration:    httpDur,
		KubeAPIRequests: kubeReq,
	}
}

// Handler returns an http.Handler serving the Prometheus exposition format
// for this registry.
func (r *Registry) Handler() http.Handler {
	return promhttp.HandlerFor(r.Registry, promhttp.HandlerOpts{Registry: r.Registry})
}
