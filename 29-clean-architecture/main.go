// Modul 29 — Clean/Hexagonal Architecture.
// main.go = "composition root": satu-satunya tempat semua lapisan DIRAKIT.
// Aliran dependency selalu menuju CORE (domain), tidak pernah keluar.
//
// Jalankan: go run ./29-clean-architecture
// Verifikasi otomatis: go test ./29-clean-architecture/...
package main

import (
	"log"
	"net/http"

	"go-learning/29-clean-architecture/internal/adapter/memory"
	"go-learning/29-clean-architecture/internal/adapter/rest"
	"go-learning/29-clean-architecture/internal/service"
)

// 🔍 Analogi besar: clean architecture itu seperti STEKER & COLOKAN. Inti bisnis (service) itu
// alat listriknya; database & HTTP cuma "colokan" yang bisa diganti (SQLite -> Postgres, REST ->
// gRPC) tanpa membongkar alatnya. Kuncinya: panah ketergantungan selalu menunjuk KE DALAM (ke inti).
// Inti tak tahu-menahu soal Postgres atau HTTP; ia hanya kenal INTERFACE. Untung: mudah diuji &
// ditukar. Analogi lain: inti = otak (aturan bisnis), adapter = tangan/mata (cara berinteraksi dunia luar).

// 🔍 Analogi: main() ini "COMPOSITION ROOT" — seperti panggung tempat semua pemain dirakit sebelum
// pertunjukan. Hanya DI SINI kita memutuskan "pakai penyimpanan memory, antarmuka REST". Ganti
// keputusan? cukup ubah baris di sini, bukan menyebar ke seluruh kode. Satu tempat perakitan.
func main() {
	// Rakit dari dalam ke luar:
	repo := memory.New()     // adapter penyimpanan (bisa ditukar postgres)
	svc := service.New(repo) // use case (logika bisnis)
	handler := rest.New(svc) // adapter HTTP (bisa ditukar gRPC/CLI)

	log.Println("clean-architecture server di http://localhost:8080")
	if err := http.ListenAndServe(":8080", handler.Routes()); err != nil {
		log.Fatal(err)
	}
}
