// REST API dengan Fiber v2 — modul 13.
//
// Jalankan server:
//
//	go run ./13-fiber            # dengar di :3000
//
// Coba:
//
//	curl localhost:3000/api/books
//	curl -X POST localhost:3000/api/books -H 'Content-Type: application/json' -d '{"title":"Go","author":"RP","year":2015}'
//	curl localhost:3000/api/books/1
//	curl -X DELETE localhost:3000/api/books/1
//
// Verifikasi otomatis (tanpa server): go test ./13-fiber
package main

import (
	"errors"
	"log"
	"os"

	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

// server memegang dependency: store + validator.
type server struct {
	store    *BookStore
	validate *validator.Validate
}

// 🔍 Analogi: Fiber itu framework web "siap pakai" — dibanding net/http (modul 12) yang
// seperti masak dari bahan mentah, Fiber seperti dapur yang alat-alatnya sudah lengkap:
// router lebih ringkas, BodyParser otomatis, middleware tinggal pasang. Fiber terinspirasi
// Express (Node.js) & terkenal cepat. Repo ini memakai Fiber (bukan Gin) sebagai standarnya.

// 🔍 Analogi: ErrorHandler terpusat itu seperti BAGIAN KELUHAN SATU PINTU. Semua handler yang
// gagal cukup "return error"; semuanya bermuara ke sini untuk diubah jadi respons JSON rapi —
// jadi tak perlu menulis format error berulang di tiap handler. DRY (Don't Repeat Yourself).
// buildApp merangkai app Fiber: middleware, error handler, dan rute.
func buildApp(s *server) *fiber.App {
	app := fiber.New(fiber.Config{
		// Error handler terpusat: semua error yang dikembalikan handler mampir ke sini.
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			var fe *fiber.Error
			if errors.As(err, &fe) {
				code = fe.Code
			}
			return c.Status(code).JSON(fiber.Map{"error": err.Error()})
		},
	})

	// Middleware global (urutan penting).
	app.Use(recover.New()) // ubah panic jadi 500, server tetap hidup
	app.Use(logger.New())  // log tiap request
	app.Use(cors.New())    // latihan 2: izinkan akses lintas-origin (mis. dari frontend)

	// 🔍 Analogi: route group "/api" itu seperti MEMBERI AWALAN alamat bersama — semua rute di
	// bawahnya otomatis diawali /api (jadi /api/books). Praktis untuk versioning (/api/v1) atau
	// memasang middleware khusus satu kelompok (mis. semua /admin butuh login).
	// Route group: semua di bawah /api.
	api := app.Group("/api")
	api.Get("/books", s.list)
	api.Post("/books", s.create)
	api.Get("/books/:id", s.get)
	api.Put("/books/:id", s.update) // latihan 1: update
	api.Delete("/books/:id", s.delete)

	return app
}

func main() {
	s := &server{store: seed(NewBookStore()), validate: validator.New()}
	app := buildApp(s)

	// Port dari env (12-factor); default 3000. Ganti: PORT=8099 go run ./13-fiber
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	log.Printf("Fiber server di http://localhost:%s", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatal(err)
	}
}

// ---------- Handlers (signature Fiber: func(*fiber.Ctx) error) ----------

func (s *server) list(c *fiber.Ctx) error {
	// Latihan 3: filter opsional ?author=X.
	author := strings.ToLower(c.Query("author"))
	books := s.store.List()
	if author == "" {
		return c.JSON(books)
	}
	filtered := make([]Book, 0, len(books))
	for _, b := range books {
		if strings.Contains(strings.ToLower(b.Author), author) {
			filtered = append(filtered, b)
		}
	}
	return c.JSON(filtered)
}

// Latihan 1: update buku ber-ID.
func (s *server) update(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "id harus angka")
	}
	var b Book
	if err := c.BodyParser(&b); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "body tidak valid")
	}
	if err := s.validate.Struct(b); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "validasi gagal: "+err.Error())
	}
	updated, err := s.store.Update(id, b)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	}
	return c.JSON(updated)
}

func (s *server) create(c *fiber.Ctx) error {
	var b Book
	// 🔍 Analogi: BodyParser itu PENERJEMAH SERBABISA di pintu masuk — entah tamu bicara JSON,
	// form, atau query string, ia otomatis mengubahnya jadi struct Book yang dipahami program.
	// BodyParser otomatis mendeteksi JSON/form/query dan mengisi struct.
	if err := c.BodyParser(&b); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "body tidak valid")
	}
	// 🔍 Analogi: validator itu SATPAM MUTU. Berdasarkan tag `validate:"required,min=1"` di struct,
	// ia menolak data cacat (judul kosong, tahun negatif) SEBELUM masuk ke penyimpanan. Aturan
	// ditulis sekali di struct, dipakai di mana-mana — bersih & konsisten.
	// Validasi berbasis tag `validate:"..."` di struct Book.
	if err := s.validate.Struct(b); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "validasi gagal: "+err.Error())
	}
	return c.Status(fiber.StatusCreated).JSON(s.store.Create(b))
}

func (s *server) get(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id") // ambil & konversi path param
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "id harus angka")
	}
	b, err := s.store.Get(id)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	}
	return c.JSON(b)
}

func (s *server) delete(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "id harus angka")
	}
	if err := s.store.Delete(id); err != nil {
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func seed(s *BookStore) *BookStore {
	s.Create(Book{Title: "The Go Programming Language", Author: "Donovan & Kernighan", Year: 2015})
	s.Create(Book{Title: "Clean Architecture", Author: "Robert C. Martin", Year: 2017})
	return s
}
