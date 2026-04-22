package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// Metrics wires HTTP request counters and a latency histogram backed by
// pre-registered Prometheus vectors. The caller keeps ownership of the
// vectors so that tests and server construction can inspect them.
//
// Cardinality note: the `path` label is the raw URL.Path. Handlers mounted
// with high-cardinality dynamic segments (e.g. resource names) should be
// registered behind a label-normalising wrapper to keep series counts sane.
func Metrics(requests *prometheus.CounterVec, duration *prometheus.HistogramVec) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := newStatusRecorder(w)
			next.ServeHTTP(rec, r)

			requests.WithLabelValues(r.Method, r.URL.Path, strconv.Itoa(rec.status)).Inc()
			duration.WithLabelValues(r.Method, r.URL.Path).Observe(time.Since(start).Seconds())
		})
	}
}
