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
