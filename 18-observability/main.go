// Modul 18 — Observability: structured logging (slog) + metrics (Prometheus).
//
// Jalankan: go run ./18-observability
//
//	curl localhost:8080/hello?name=Ana
//	curl localhost:8080/metrics        # data Prometheus
//
// Verifikasi otomatis: go test ./18-observability
package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// buildHandler merangkai router + middleware (logging & metrics). Dipisah dari
// main agar bisa diuji dengan httptest.
func buildHandler(logger *slog.Logger, reg *prometheus.Registry) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /hello", func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Query().Get("name")
		if name == "" {
			name = "dunia"
		}
		w.Write([]byte("Halo, " + name + "!\n"))
	})

	mux.HandleFunc("GET /boom", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "sengaja error", http.StatusInternalServerError)
	})

	// /metrics mengekspos data untuk di-scrape Prometheus.
	mux.Handle("GET /metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))

	// Rantai middleware: logging di luar, metrics di dalam.
	return loggingMiddleware(logger)(metricsMiddleware(mux))
}

// loggingMiddleware mencatat tiap request sebagai structured log (JSON).
func loggingMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rec, r)
			// Field terstruktur (key=value), bukan string mentah -> mudah difilter/di-query.
			logger.Info("http_request",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", rec.status),
				slog.Duration("durasi", time.Since(start)),
			)
		})
	}
}

func main() {
	// slog dengan handler JSON: siap dikonsumsi Loki/ELK/Datadog.
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	reg := newRegistry()
	handler := buildHandler(logger, reg)

	addr := ":8080"
	logger.Info("server_start", slog.String("addr", addr))
	if err := http.ListenAndServe(addr, handler); err != nil {
		logger.Error("server_error", slog.Any("err", err))
		os.Exit(1)
	}
}
