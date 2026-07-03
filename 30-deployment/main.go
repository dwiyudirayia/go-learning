// Modul 30 — Deployment: aplikasi siap-produksi (health probes, config env, version).
//
// Jalankan: go run ./30-deployment
// Verifikasi otomatis: go test ./30-deployment
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync/atomic"
)

// version diisi saat build via ldflags:
//
//	go build -ldflags "-X main.version=1.2.3" ./30-deployment
var version = "dev"

// ready menandai apakah app siap menerima traffic (readiness probe).
// Saat shutdown (Modul 20), set ke false agar load balancer berhenti mengirim traffic.
var ready atomic.Bool

func buildHandler() http.Handler {
	ready.Store(true)
	mux := http.NewServeMux()

	// Liveness: apakah proses hidup? K8s me-restart pod bila ini gagal.
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	// Readiness: apakah siap menerima traffic? K8s berhenti kirim traffic bila gagal
	// (tanpa restart). Berguna saat warming up atau graceful shutdown.
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		if !ready.Load() {
			http.Error(w, "not ready", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ready"))
	})

	// Info versi (berguna untuk verifikasi deploy).
	mux.HandleFunc("GET /version", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"version": version})
	})

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Halo dari aplikasi ter-deploy!\n"))
	})

	return mux
}

func main() {
	port := os.Getenv("PORT") // config via env (Modul 19)
	if port == "" {
		port = "8080"
	}
	log.Printf("versi %s berjalan di :%s", version, port)
	if err := http.ListenAndServe(":"+port, buildHandler()); err != nil {
		log.Fatal(err)
	}
}
