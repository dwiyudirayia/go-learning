# 13 — REST API dengan Fiber

[Fiber](https://gofiber.io) adalah framework web Go yang cepat (di atas `fasthttp`), dengan API mirip Express.js. Ringkas untuk routing, middleware, binding, dan validasi.

Jalankan:
```bash
go run ./13-fiber                 # :3000 (default)
PORT=8099 go run ./13-fiber       # port lain (env, 12-factor)
```
Coba:
```bash
curl localhost:3000/api/books
curl -X POST localhost:3000/api/books -H 'Content-Type: application/json' \
  -d '{"title":"Go","author":"RP","year":2015}'
curl localhost:3000/api/books/1
curl -X DELETE localhost:3000/api/books/1
```
Verifikasi otomatis: `go test ./13-fiber` (pakai `app.Test`, tanpa buka port).

## Konsep Fiber

### App, middleware, group
```go
app := fiber.New(fiber.Config{ErrorHandler: ...})
app.Use(recover.New())   // panic -> 500, server tetap hidup
app.Use(logger.New())    // log tiap request
api := app.Group("/api") // prefix untuk sekelompok rute
api.Get("/books", handler)
```

### Handler
Signature: `func(c *fiber.Ctx) error` — **mengembalikan error** (bukan menulis status manual). Error mampir ke `ErrorHandler` terpusat.
```go
func (s *server) get(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")           // path param + konversi
	if err != nil { return fiber.NewError(400, "id harus angka") }
	b, err := s.store.Get(id)
	if err != nil { return fiber.NewError(404, err.Error()) }
	return c.JSON(b)                        // 200 + JSON
}
```

### Binding & Validasi
```go
var b Book
c.BodyParser(&b)              // isi struct dari JSON/form/query otomatis
s.validate.Struct(b)          // validasi tag `validate:"required,min=2"`
```
Validasi pakai `go-playground/validator`. Tag di struct `Book`:
```go
Title  string `validate:"required,min=2"`
Author string `validate:"required"`
Year   int    `validate:"gte=0,lte=2100"`
```

### Error handler terpusat
Semua `return err` dari handler ditangani satu tempat → response error konsisten. Pola idiomatik yang jauh lebih rapi daripada menulis status di tiap handler.

## Fiber vs `net/http` (Modul 12)

| | net/http (12) | Fiber (13) |
|-|---------------|------------|
| Dependency | nol | fasthttp dkk |
| Routing | `GET /books/{id}` | `api.Get("/books/:id")` |
| Middleware | tulis sendiri | siap pakai (logger, recover, cors, jwt, ...) |
| Binding/validasi | manual | `BodyParser` + validator |
| Performa | tinggi | sangat tinggi |
| Kompatibilitas | `http.Handler` standar | fasthttp (bukan net/http) |

> Catatan: Fiber memakai `fasthttp`, **bukan** `net/http`. Jadi middleware ekosistem `net/http` tidak langsung kompatibel. Untuk mayoritas API ini bukan masalah; bila butuh kompatibilitas penuh `net/http`, pertimbangkan Echo/chi/Gin.

## Kapan Dipakai
- API yang butuh throughput tinggi & sintaks ringkas.
- Tim yang suka gaya Express.js.
- Prototipe cepat sampai produksi menengah.

## Latihan
1. Tambah endpoint `PUT /api/books/:id` (update) + validasinya.
2. Tambah middleware `cors` (`github.com/gofiber/fiber/v2/middleware/cors`).
3. Tambah query filter `GET /api/books?author=X` (`c.Query("author")`).
4. Pisahkan handler ke package `internal/handler` (pola Modul 10).
5. Tambah test untuk update & filter.

> Studi kasus lengkap (Fiber + database + JWT) ada di **Modul 15**.

## ✅ Status Solusi Latihan
Latihan **1, 2, 3, 5 sudah diselesaikan**: update (`PUT`), CORS, filter (`?author=`), dan test (`main.go` + `api_test.go`). Latihan 4 (pisah handler ke package) tersedia sebagai tantangan.

---

## 🚀 Teknik Advanced (Level Up)
> 💻 **Contoh runnable + komentar detail** untuk teknik di bawah ada di folder [`advanced/`](advanced). Jalankan: `go run ./13-fiber/advanced`


- **Custom error handler** — `fiber.Config{ErrorHandler: ...}` sentralisasi mapping error → status. Kombinasi dengan sentinel error dari service.
- **Middleware order** — urutan `app.Use` menentukan eksekusi; recover, logger, cors, auth berurutan. Grup route: `api := app.Group("/api", authMw)`.
- **`app.Test(req)`** — uji end-to-end tanpa port nyata (dipakai di [[41-capstone]]).
- **Immutability & `c.Locals`** — nilai dari `c.Params`/`c.Query` valid hanya selama handler; salin jika disimpan. `c.Locals` untuk oper data antar-middleware.
- **Zero-alloc gotcha** — Fiber reuse buffer; jangan simpan reference `[]byte` dari context tanpa `copy`.
- **Streaming** — `c.SendStream` untuk respons besar; WebSocket via `contrib/websocket`. Lihat [[24-websocket]].
- **Validator** — integrasikan `go-playground/validator` untuk validasi DTO.
