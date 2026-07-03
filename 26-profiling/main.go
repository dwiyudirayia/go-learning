// Jalankan server dengan endpoint pprof: go run ./26-profiling
//
// Lalu ambil profil (di terminal lain):
//
//	go tool pprof http://localhost:6060/debug/pprof/profile?seconds=5   # CPU 5 detik
//	go tool pprof http://localhost:6060/debug/pprof/heap                # memori
//	curl http://localhost:6060/debug/pprof/goroutine?debug=1            # goroutine
//
// Di dalam pprof: ketik `top`, `list NamaFungsi`, atau `web` (butuh graphviz).
package main

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof" // side-effect: mendaftarkan handler /debug/pprof/*
)

func main() {
	mux := http.NewServeMux()

	// Endpoint yang sengaja "sibuk" untuk dilihat di profil CPU.
	mux.HandleFunc("GET /work", func(w http.ResponseWriter, r *http.Request) {
		result := SumSquares(5_000_000)
		fmt.Fprintf(w, "hasil=%d\n", result)
	})

	// Daftarkan handler pprof (yang diregistrasi ke DefaultServeMux oleh import).
	mux.Handle("/debug/pprof/", http.DefaultServeMux)

	log.Println("server + pprof di http://localhost:6060")
	log.Println("coba: go tool pprof http://localhost:6060/debug/pprof/profile?seconds=5")
	if err := http.ListenAndServe(":6060", mux); err != nil {
		log.Fatal(err)
	}
}
