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

// 🔍 Analogi besar: profiling itu seperti CEK KESEHATAN / MRI untuk program. Alih-alih MENEBAK
// "mungkin fungsi ini yang lambat", pprof mengukur FAKTA: fungsi mana makan CPU/memori terbanyak.
// Aturan emas optimasi: "ukur dulu, baru optimasi" — sering biang lambatnya di tempat tak terduga.
//
// 🔍 Analogi import "_" (blank): pola "_ net/http/pprof" itu seperti MENYALAKAN alat hanya dengan
// memasang colokannya. Kita tak memanggil fungsinya langsung; cukup meng-import-nya, dan paket itu
// diam-diam mendaftarkan endpoint /debug/pprof/* saat start (efek samping init()). Tanda "_" berarti
// "aku sengaja import ini demi efek sampingnya, bukan untuk memakai isinya".

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
