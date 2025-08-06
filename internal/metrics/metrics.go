package metrics

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	namespace = "image-sharing"
	subsystem = "http"
)

type Metrics struct {
	requestsTotal   *prometheus.CounterVec
	requestDuration *prometheus.HistogramVec
}

func New() *Metrics {
	m := &Metrics{
		// HTTP metrics
		requestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "requests_total",
				Help:      "Total number of HTTP requests processed, partitioned by status code, method and path.",
			},
			[]string{"code", "method", "path"},
		),
		requestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "request_duration_seconds",
				Help:      "The HTTP request latencies in seconds.",
				Buckets:   []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
			},
			[]string{"method", "path"},
		),
	}
	return m
}

func (m *Metrics) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := NewResponseWriter(w)

			next.ServeHTTP(ww, r)

			duration := time.Since(start).Seconds()

			path := chi.RouteContext(r.Context()).RoutePattern()
			if path == "" {
				path = r.URL.Path
			}

			m.requestDuration.WithLabelValues(r.Method, path).Observe(duration)
			m.requestsTotal.WithLabelValues(http.StatusText(ww.status), r.Method, path).Inc()
		})
	}
}

func (m *Metrics) Handler() http.Handler {
	return promhttp.Handler()
}

type ResponseWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func NewResponseWriter(w http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{ResponseWriter: w}
}

func (rw *ResponseWriter) WriteHeader(code int) {
	if !rw.wroteHeader {
		rw.status = code
		rw.wroteHeader = true
	}
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *ResponseWriter) Write(b []byte) (int, error) {
	if !rw.wroteHeader {
		rw.status = http.StatusOK
		rw.wroteHeader = true
	}
	return rw.ResponseWriter.Write(b)
}

func (rw *ResponseWriter) Status() int {
	if !rw.wroteHeader {
		return http.StatusOK
	}
	return rw.status
}

func (rw *ResponseWriter) Unwrap() http.ResponseWriter {
	return rw.ResponseWriter
}
