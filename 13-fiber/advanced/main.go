// Teknik Advanced Modul 13 (Fiber) — contoh runnable + penjelasan detail.
//
// Memakai app.Test untuk menguji IN-PROCESS (tanpa port). Jalankan:
//
//	go run ./13-fiber/advanced
package main

import (
	"errors"
	"fmt"
	"io"
	"net/http/httptest"

	"github.com/gofiber/fiber/v2"
)

// Sentinel error domain -> dipetakan ke status HTTP di satu tempat.
var ErrTidakDitemukan = errors.New("data tidak ditemukan")

func main() {
	// =====================================================================
	// 1. Custom ErrorHandler terpusat
	// =====================================================================
	// Semua error yang dikembalikan handler mengalir ke sini. Kita petakan
	// sentinel error & *fiber.Error ke status yang tepat -> handler jadi bersih
	// (cukup `return ErrTidakDitemukan`).
	//
	// 🔍 Analogi: ErrorHandler terpusat itu seperti CUSTOMER SERVICE satu
	// pintu — semua keluhan dari lantai mana pun (handler) diteruskan ke satu
	// meja yang tahu cara menjawabnya (memetakan error -> status HTTP).
	// Karyawan lantai cukup bilang "ada masalah X", tak perlu hafal skrip
	// jawaban resmi.
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if errors.Is(err, ErrTidakDitemukan) {
				code = fiber.StatusNotFound
			}
			var fe *fiber.Error
			if errors.As(err, &fe) { // fiber.NewError(...) membawa kode sendiri
				code = fe.Code
			}
			return c.Status(code).JSON(fiber.Map{"error": err.Error()})
		},
	})

	// =====================================================================
	// 2. Middleware + c.Locals — oper data antar-lapis dalam satu request
	// =====================================================================
	// 🔍 Analogi: c.Locals itu seperti GELANG PASIEN di rumah sakit — dipasang
	// di pintu masuk (middleware), lalu setiap petugas berikutnya (handler)
	// tinggal membaca gelangnya tanpa bertanya ulang. Gelang dilepas saat
	// pasien pulang (request selesai) — tak bocor ke pasien lain.
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("request_id", "req-abc-123") // simpan untuk handler di bawahnya
		return c.Next()                       // WAJIB Next() agar lanjut
	})

	// =====================================================================
	// 3. Route group — prefix + middleware bersama
	// =====================================================================
	// 🔍 Analogi: route group itu seperti SAYAP GEDUNG dengan satu pintu
	// masuk — semua ruangan di sayap "/api" otomatis berbagi alamat depan dan
	// penjaga pintu (middleware) yang sama, tanpa ditulis ulang per ruangan.
	api := app.Group("/api")
	api.Get("/users/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		if id == "0" {
			return ErrTidakDitemukan // -> ErrorHandler memetakan ke 404
		}
		return c.JSON(fiber.Map{
			"id":         id,
			"request_id": c.Locals("request_id"),
		})
	})

	// =====================================================================
	// app.Test: kirim *http.Request langsung ke app tanpa jaringan nyata.
	// =====================================================================
	fmt.Println("== 1. sukses (200) ==")
	fmt.Println(" ", panggil(app, "/api/users/7"))

	fmt.Println("== 2. sentinel -> 404 lewat ErrorHandler ==")
	fmt.Println(" ", panggil(app, "/api/users/0"))

	fmt.Println("== 3. rute tak ada -> 404 default Fiber ==")
	fmt.Println(" ", panggil(app, "/api/tidak-ada"))
}

func panggil(app *fiber.App, path string) string {
	req := httptest.NewRequest("GET", path, nil)
	resp, err := app.Test(req) // in-process; -1 default timeout
	if err != nil {
		return "err: " + err.Error()
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	return fmt.Sprintf("%d %s", resp.StatusCode, b)
}
