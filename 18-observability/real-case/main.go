// REAL-CASE Modul 18 (observability) — METRIK PROMETHEUS via /metrics.
//
// Versi advanced/ fokus pada slog. Versi ini mengekspos metrik ala produksi:
// aplikasi menghitung request (counter) & latensi (histogram), lalu Prometheus
// mengambilnya (scrape) dari endpoint /metrics. Bagian Go BERJALAN LOKAL (kita
// scrape /metrics sendiri); docker-compose menyediakan Prometheus asli.
//
// Jalankan:
//
//	go run ./18-observability/real-case
//
// Dengan Prometheus asli:
//
//	docker compose -f 18-observability/real-case/docker-compose.yml up -d
//	# lalu jalankan app ini di :2112 dan buka http://localhost:9090
package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// reqTotal = COUNTER: nilai yang hanya NAIK. Label "route" & "status" memungkinkan
// query per-endpoint/per-status. HATI-HATI kardinalitas label (jangan pakai
// user_id / URL mentah -> ledakan time-series).
var reqTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Jumlah total HTTP request diproses.",
	},
	[]string{"route", "status"},
)

// reqDuration = HISTOGRAM: membagi observasi latensi ke dalam "bucket" agar bisa
// dihitung persentil (p50/p95/p99) dan diagregasi lintas instance. Pilih bucket
// sesuai SLO latensi aplikasi.
var reqDuration = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "Distribusi latensi HTTP request.",
		Buckets: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25},
	},
	[]string{"route"},
)

// instrument membungkus sebuah handler untuk MEREKAM metrik secara otomatis:
// mencatat durasi (histogram) dan menaikkan counter dengan label status.
//
// Param:
//   - route : nama logis endpoint (dipakai sebagai label, bukan URL mentah).
//   - next  : handler asli yang ingin diukur.
//
// Return http.HandlerFunc baru yang terinstrumentasi.
func instrument(route string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		// rec menangkap status code yang ditulis handler (untuk label metrik).
		rec := &statusRecorder{ResponseWriter: w, status: 200}
		next(rec, r)
		reqDuration.WithLabelValues(route).Observe(time.Since(start).Seconds())
		reqTotal.WithLabelValues(route, fmt.Sprint(rec.status)).Inc()
	}
}

// statusRecorder membungkus http.ResponseWriter hanya untuk MEREKAM status code
// (ResponseWriter standar tak menyediakan cara membacanya kembali).
type statusRecorder struct {
	http.ResponseWriter
	status int
}

// WriteHeader mencatat status lalu meneruskan ke writer asli.
func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

func main() {
	// Daftarkan metrik ke registry default Prometheus.
	prometheus.MustRegister(reqTotal, reqDuration)

	mux := http.NewServeMux()
	// Endpoint bisnis, dibungkus instrument() agar terukur.
	mux.HandleFunc("/api/ok", instrument("/api/ok", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(8 * time.Millisecond)
		fmt.Fprintln(w, "ok")
	}))
	mux.HandleFunc("/api/error", instrument("/api/error", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	// /metrics = endpoint yang di-SCRAPE Prometheus. promhttp.Handler() otomatis
	// menyerialkan semua metrik terdaftar ke format teks Prometheus.
	mux.Handle("/metrics", promhttp.Handler())

	// httptest server = server HTTP nyata di port lokal (tanpa perlu :2112 tetap).
	srv := httptest.NewServer(mux)
	defer srv.Close()

	// Bangkitkan sedikit trafik agar metrik terisi.
	for i := 0; i < 5; i++ {
		http.Get(srv.URL + "/api/ok")
	}
	http.Get(srv.URL + "/api/error")

	// Scrape /metrics sendiri (persis yang dilakukan Prometheus tiap interval).
	resp, _ := http.Get(srv.URL + "/metrics")
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	fmt.Println("== cuplikan /metrics (yang di-scrape Prometheus) ==")
	for _, line := range strings.Split(string(body), "\n") {
		if strings.HasPrefix(line, "http_requests_total") ||
			strings.HasPrefix(line, "http_request_duration_seconds_count") {
			fmt.Println("  " + line)
		}
	}
	fmt.Println("== Grafana memvisualkan ini; alert berbasis rate()/histogram_quantile() ==")
}
