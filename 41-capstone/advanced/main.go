// Teknik Advanced Modul 41 (capstone) — contoh runnable + penjelasan detail.
//
// Capstone-in-miniature: gabungan LAYERED (store->service->handler) + Fiber +
// integration test via app.Test — inti yang mematangkan semua modul. Jalankan:
//
//	go run ./41-capstone/advanced
package main

import (
	"errors"
	"fmt"
	"io"
	"net/http/httptest"
	"strings"
	"sync"

	"github.com/gofiber/fiber/v2"
)

var ErrNotFound = errors.New("kode tidak ditemukan")

// 🔍 Analogi: arsitektur berlapis itu seperti RESTORAN — gudang bahan (store),
// dapur dengan resep (service), dan pelayan yang menghadapi tamu (handler).
// Tamu tak pernah masuk gudang; pelayan tak pernah memasak. Tiap lapisan bisa
// diganti (gudang in-memory -> SQLite) tanpa melatih ulang seluruh restoran.

// ---- Lapisan STORE (data) — di sini in-memory; di capstone asli: SQLite ----
type Store struct {
	mu   sync.RWMutex
	data map[string]string // code -> url
}

func NewStore() *Store { return &Store{data: map[string]string{}} }

func (s *Store) Save(code, url string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[code] = url
}
func (s *Store) Get(code string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if u, ok := s.data[code]; ok {
		return u, nil
	}
	return "", ErrNotFound
}

// ---- Lapisan SERVICE (aturan bisnis) — bergantung pada store lewat pointer --
type Service struct {
	store *Store
	seq   int
}

func (svc *Service) Shorten(url string) (string, error) {
	if url == "" {
		return "", errors.New("url kosong")
	}
	svc.seq++
	code := fmt.Sprintf("c%d", svc.seq) // generator kode sederhana
	svc.store.Save(code, url)
	return code, nil
}
func (svc *Service) Resolve(code string) (string, error) { return svc.store.Get(code) }

// ---- Lapisan HANDLER (HTTP) — memetakan error domain -> status --------------
func buildApp(svc *Service) *fiber.App {
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if errors.Is(err, ErrNotFound) {
				code = fiber.StatusNotFound
			}
			return c.Status(code).JSON(fiber.Map{"error": err.Error()})
		},
	})
	app.Post("/shorten", func(c *fiber.Ctx) error {
		var body struct {
			URL string `json:"url"`
		}
		_ = c.BodyParser(&body)
		code, err := svc.Shorten(body.URL)
		if err != nil {
			return err
		}
		return c.Status(fiber.StatusCreated).JSON(fiber.Map{"code": code})
	})
	app.Get("/:code", func(c *fiber.Ctx) error {
		url, err := svc.Resolve(c.Params("code"))
		if err != nil {
			return err // -> ErrorHandler memetakan ke 404
		}
		return c.JSON(fiber.Map{"url": url})
	})
	return app
}

func main() {
	// COMPOSITION ROOT: rakit store -> service -> handler.
	svc := &Service{store: NewStore()}
	app := buildApp(svc)

	// =====================================================================
	// INTEGRATION TEST via app.Test — menembus SEMUA lapisan tanpa port nyata.
	// Inilah cara memvalidasi bahwa store+service+handler bekerja bersama.
	//
	// 🔍 Analogi: unit test menguji tiap alat musik sendiri-sendiri;
	// integration test itu GLADI BERSIH SATU ORKESTRA — semua lapisan main
	// bersama dari partitur asli, tapi di ruang latihan (tanpa penonton/port
	// nyata). Salah kompak antar-pemain baru ketahuan di sini.
	// =====================================================================
	fmt.Println("== 1. POST /shorten ==")
	fmt.Println(" ", call(app, "POST", "/shorten", `{"url":"https://go.dev"}`))

	fmt.Println("== 2. GET /c1 (resolve) ==")
	fmt.Println(" ", call(app, "GET", "/c1", ""))

	fmt.Println("== 3. GET /tidak-ada -> 404 lewat ErrorHandler ==")
	fmt.Println(" ", call(app, "GET", "/zzz", ""))

	// Di capstone penuh, tambahkan: cache-aside (Redis/miniredis) di depan store,
	// JWT+bcrypt untuk auth, config via env, observability, dan graceful shutdown.
}

func call(app *fiber.App, method, path, body string) string {
	var r io.Reader
	if body != "" {
		r = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, r)
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		return "err: " + err.Error()
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	return fmt.Sprintf("%d %s", resp.StatusCode, b)
}
