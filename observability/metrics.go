package observability

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	RequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "HTTP request duration in seconds.",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "path", "status"})

	RequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of HTTP requests.",
	}, []string{"method", "path", "status"})

	TransfersTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "transfers_total",
		Help: "Total number of transfers executed.",
	})

	TransferAmount = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "transfer_amount",
		Help:    "Distribution of transfer amounts.",
		Buckets: []float64{10, 50, 100, 500, 1000, 5000, 10000},
	})
)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		start := time.Now()

		next.ServeHTTP(rec, r)

		duration := time.Since(start).Seconds()
		status := strconv.Itoa(rec.status)

		RequestDuration.WithLabelValues(r.Method, r.URL.Path, status).Observe(duration)
		RequestsTotal.WithLabelValues(r.Method, r.URL.Path, status).Inc()
	})
}
