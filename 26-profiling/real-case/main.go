// REAL-CASE Modul 26 (profiling) — endpoint pprof LANGSUNG via HTTP.
//
// Versi advanced/ menangkap profil secara programatik. Di produksi, cara umum
// adalah mengekspos handler net/http/pprof pada server HTTP (diamankan), lalu
// menariknya dengan `go tool pprof`. File ini menjalankan server ber-pprof dan
// menariknya sendiri — berjalan lokal tanpa infra.
//
// Jalankan:
//
//	go run ./26-profiling/real-case
//
// Analisis interaktif terhadap server yang hidup:
//
//	go tool pprof http://<host>/debug/pprof/heap
//	go tool pprof http://<host>/debug/pprof/profile?seconds=30   # CPU 30 dtk
package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	_ "net/http/pprof" // EFEK IMPOR: init() mendaftarkan handler /debug/pprof/* ke DefaultServeMux
	"runtime"
)

// bebanMemori sengaja mengalokasikan banyak objek agar TERLIHAT di heap profile.
// Nilai dikembalikan (bukan dibuang) supaya tak langsung di-GC.
//
// Return: slice yang menahan ~4MB agar tetap hidup saat profil diambil.
func bebanMemori() [][]byte {
	var data [][]byte
	for i := 0; i < 4000; i++ {
		data = append(data, make([]byte, 1024)) // 1KB x 4000
	}
	return data
}

// ambilProfil menarik satu profil dari endpoint pprof server yang hidup.
//
// Param:
//   - base : URL dasar server (mis. http://127.0.0.1:PORT)
//   - nama : nama profil pprof ("heap", "goroutine", "allocs", dll)
//
// Return jumlah byte profil yang diterima (bukti endpoint berfungsi).
func ambilProfil(base, nama string) int {
	resp, err := http.Get(base + "/debug/pprof/" + nama + "?debug=0")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	return len(b)
}

func main() {
	// httptest.NewServer(nil) memakai http.DefaultServeMux — tempat paket
	// net/http/pprof sudah mendaftarkan handler-nya lewat blank import di atas.
	srv := httptest.NewServer(nil)
	defer srv.Close()
	fmt.Println("== server dengan pprof di", srv.URL, "==")

	// Buat beban agar profil punya isi bermakna.
	data := bebanMemori()
	runtime.KeepAlive(data) // pastikan 'data' masih hidup saat heap diprofil

	// Tarik beberapa profil dari server yang hidup (seperti go tool pprof).
	fmt.Println("== profil ditarik dari endpoint (byte) ==")
	for _, nama := range []string{"heap", "goroutine", "allocs"} {
		fmt.Printf("  /debug/pprof/%-10s -> %d byte\n", nama, ambilProfil(srv.URL, nama))
	}

	fmt.Println("== endpoint tersedia ==")
	fmt.Println("  /debug/pprof/            (indeks)")
	fmt.Println("  /debug/pprof/profile     (CPU, ?seconds=N)")
	fmt.Println("  /debug/pprof/heap        (memori)")
	fmt.Println("  /debug/pprof/goroutine   (semua goroutine)")
	fmt.Println("  produksi: LINDUNGI endpoint ini (auth/network), jangan publik.")
	fmt.Println("  continuous profiling: Pyroscope/Parca menarik profil ini otomatis.")
}
