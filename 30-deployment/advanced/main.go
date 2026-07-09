// Teknik Advanced Modul 30 (deployment) — contoh runnable + penjelasan detail.
//
// Semantik PROBE Kubernetes: liveness vs readiness, dan koordinasinya dengan
// graceful shutdown. Jalankan:
//
//	go run ./30-deployment/advanced
package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"runtime/debug"
	"sync/atomic"
)

// health memisahkan dua kondisi berbeda:
//   - LIVENESS  : "proses masih waras?" -> jika gagal, K8s RESTART pod.
//   - READINESS : "siap terima traffic?" -> jika gagal, K8s CABUT dari Service
//     (tak restart). Dipakai saat warm-up ATAU saat shutdown (drain).
type health struct {
	ready atomic.Bool
}

func main() {
	h := &health{}

	mux := http.NewServeMux()
	// Liveness: selalu 200 selama proses hidup (jangan sertakan dependency di sini,
	// nanti dependency lambat malah memicu restart beruntun).
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "alive")
	})
	// Readiness: 200 hanya bila ready=true; jika tidak -> 503 (traffic dialihkan).
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		if h.ready.Load() {
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, "ready")
			return
		}
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprint(w, "not ready")
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	probe := func(path string) string {
		resp, err := http.Get(srv.URL + path)
		if err != nil {
			return "err: " + err.Error()
		}
		defer resp.Body.Close()
		b, _ := io.ReadAll(resp.Body)
		return fmt.Sprintf("%d %s", resp.StatusCode, b)
	}

	fmt.Println("== fase 1: warm-up (belum ready) ==")
	fmt.Println("  /healthz:", probe("/healthz")) // 200 (hidup)
	fmt.Println("  /readyz :", probe("/readyz"))  // 503 (belum siap -> tak dapat traffic)

	h.ready.Store(true) // selesai inisialisasi (koneksi DB dibuka, cache warm, dst)
	fmt.Println("== fase 2: siap melayani ==")
	fmt.Println("  /healthz:", probe("/healthz")) // 200
	fmt.Println("  /readyz :", probe("/readyz"))  // 200

	// Saat SIGTERM: set ready=false LEBIH DULU agar load balancer berhenti
	// mengirim traffic, baru drain koneksi & tutup dependency (lihat modul 20).
	h.ready.Store(false)
	fmt.Println("== fase 3: shutdown (drain) ==")
	fmt.Println("  /healthz:", probe("/healthz")) // 200 (masih memproses in-flight)
	fmt.Println("  /readyz :", probe("/readyz"))  // 503 (jangan kirim request baru)

	// Metadata build (untuk endpoint /version). Di Docker: build multi-stage ->
	// distroless/scratch, CGO_ENABLED=0, -ldflags='-s -w' agar image kecil.
	if bi, ok := debug.ReadBuildInfo(); ok {
		fmt.Println("== build ==")
		fmt.Println("  go:", bi.GoVersion, "| module:", bi.Main.Path)
	}
}
