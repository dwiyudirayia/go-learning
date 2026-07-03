package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// Tiga jenis metrik yang paling umum:
//   - Counter  : hanya naik (jumlah request)
//   - Histogram: distribusi (durasi request) -> hitung p50/p95/p99
//   - Gauge    : naik-turun (jumlah request in-flight)
var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Jumlah total request HTTP.",
		},
		[]string{"method", "path", "status"}, // label: dimensi untuk mengiris data
	)

	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Durasi request HTTP dalam detik.",
			Buckets: prometheus.DefBuckets, // 0.005s .. 10s
		},
		[]string{"method", "path"},
	)

	httpInFlight = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "http_in_flight_requests",
		Help: "Jumlah request yang sedang diproses.",
	})
)

// registry lokal (bukan global default) supaya rapi & mudah di-test.
func newRegistry() *prometheus.Registry {
	reg := prometheus.NewRegistry()
	reg.MustRegister(httpRequestsTotal, httpRequestDuration, httpInFlight)
	return reg
}

// statusRecorder membungkus ResponseWriter untuk menangkap status code
// (karena http.ResponseWriter tidak menyimpannya).
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// metricsMiddleware mencatat metrik untuk setiap request.
func metricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		httpInFlight.Inc()
		defer httpInFlight.Dec()

		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(rec, r)

		// Pakai r.Pattern (Go 1.23+) agar label "path" tidak meledak oleh nilai dinamis.
		path := r.Pattern
		if path == "" {
			path = r.URL.Path
		}
		dur := time.Since(start).Seconds()
		httpRequestsTotal.WithLabelValues(r.Method, path, strconv.Itoa(rec.status)).Inc()
		httpRequestDuration.WithLabelValues(r.Method, path).Observe(dur)
	})
}
