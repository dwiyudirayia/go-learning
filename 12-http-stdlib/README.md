# 12 — REST API dengan `net/http` Murni

Sebelum pakai framework (Fiber), penting paham cara bikin REST API **tanpa** framework. Sejak Go 1.22, `net/http` bawaan sudah cukup untuk banyak kasus.

Jalankan server:
```bash
go run ./12-http-stdlib          # dengar di :8080
```
Uji dengan curl:
```bash
curl localhost:8080/books
curl -X POST localhost:8080/books -d '{"title":"Go","author":"RP"}'
curl localhost:8080/books/1
curl -X PUT localhost:8080/books/1 -d '{"title":"Go v2","author":"RP"}'
curl -X DELETE localhost:8080/books/1
```
Verifikasi otomatis (tanpa server): `go test ./12-http-stdlib`

## Konsep

### Routing method + path (Go 1.22+)
```go
mux := http.NewServeMux()
mux.HandleFunc("GET /books", handleList)
mux.HandleFunc("POST /books", handleCreate)
mux.HandleFunc("GET /books/{id}", handleGet) // {id} = path parameter
// ambil: id := r.PathValue("id")
```
Sebelum 1.22 kamu perlu router pihak ketiga (gorilla/mux, chi) untuk ini. Sekarang bawaan.

### Struktur handler
- Handler = `func(w http.ResponseWriter, r *http.Request)`.
- Dijadikan **method dari struct `server`** yang memegang dependency (`store`) → mudah di-test & bebas variabel global.
- `writeJSON` / `writeError` helper untuk konsistensi response + status code.

### Status code yang benar (penting di REST)
| Aksi | Sukses | Gagal |
|------|--------|-------|
| GET | 200 OK | 404 Not Found |
| POST | 201 Created | 400 Bad Request |
| PUT | 200 OK | 404 / 400 |
| DELETE | 204 No Content | 404 |

### Middleware
Fungsi yang membungkus `http.Handler`:
```go
func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
```
Rantai: `logging(auth(router))`. Pola ini identik di semua framework — hanya sintaksnya beda.

### Arsitektur
`handler` (parse req, tulis resp) → `store` (data). Store pakai `sync.RWMutex` agar aman-konkuren (tiap request = goroutine, Modul 7). Di Modul 14 store diganti database.

## `net/http` vs Framework (Fiber)
- **`net/http`** — nol dependency, cukup untuk API kecil–menengah, kontrol penuh.
- **Fiber (Modul 13)** — routing lebih ringkas, middleware siap pakai (CORS, JWT, logger), binding & validasi, performa tinggi (fasthttp). Cocok saat API tumbuh besar.

Memahami versi murni ini membuatmu tahu **apa yang dilakukan framework di balik layar**.

## Latihan
1. Tambah endpoint `GET /books/search?author=X` (baca query string via `r.URL.Query()`).
2. Tambah middleware `recover` (Modul 5) agar panic tak menjatuhkan server.
3. Tambah validasi: tolak `author` kosong juga.
4. Tambah pagination pada `GET /books` (`?limit=&offset=`).
5. Tambah test untuk endpoint search.

Test ada di `api_test.go` sebagai contoh pola httptest.

## Kapan & Di Mana Dipakai (Studi Kasus Nyata)

| Konsep | Cara pakai di dunia nyata | Contoh kasus |
|--------|---------------------------|--------------|
| **`net/http` murni** | API kecil–menengah tanpa dependency | microservice ringan, internal tool, webhook receiver |
| **Routing method+path** | endpoint RESTful | `GET /users/{id}`, `POST /orders` |
| **Handler = method struct** | injeksi dependency, bebas global var | handler pegang `store`/`db`/`logger` |
| **Middleware** | lintas-cutting concern | logging, auth, CORS, rate limit, recover |
| **Status code tepat** | kontrak REST yang benar | 201 saat create, 404 saat tak ada, 204 saat delete |
| **`sync.RWMutex` di store** | state bersama aman-konkuren | cache in-memory, counter, sesi |

**Contoh nyata — merangkai middleware (pola universal):**
```go
handler := logging(auth(recoverMw(router)))  // eksekusi: logging -> auth -> recover -> router
http.ListenAndServe(":8080", handler)
```

**Kapan cukup `net/http`, kapan pindah framework?**
- **Cukup `net/http`**: endpoint sedikit, tim kecil, ingin nol dependency & kontrol penuh.
- **Pindah ke Fiber/Echo/chi (Modul 13)**: butuh middleware siap pakai, binding+validasi otomatis, atau API tumbuh besar.

**Cocok dipakai saat:** memulai backend apa pun. Karena kamu paham versi "murni" ini, kamu akan mengerti apa yang framework lakukan di balik layar — dan bisa memilih dengan sadar, bukan ikut-ikutan.

## ✅ Status Solusi Latihan
Latihan **1, 2, 4, 5 sudah diselesaikan**: search (`/books/search`), recover middleware, pagination (`?limit=&offset=`), dan test-nya (lihat `latihan.go` + `api_test.go`). Latihan 3 (validasi author) dibiarkan sebagai latihan singkat.

---

## 🚀 Teknik Advanced (Level Up)
> 💻 **Contoh runnable + komentar detail** untuk teknik di bawah ada di folder [`advanced/`](advanced). Jalankan: `go run ./12-http-stdlib/advanced`


- **Routing Go 1.22 `ServeMux`** — pola dengan **method + wildcard**: `mux.HandleFunc("GET /users/{id}", h)`, ambil via `r.PathValue("id")`. Tak perlu router pihak ketiga untuk kasus sederhana.
- **Middleware chaining** — `func(http.Handler) http.Handler`; rantai dengan helper `Chain(h, m1, m2)`. Simpan nilai di `context` (auth, request-id).
- **Timeout berlapis** — `http.Server{ReadHeaderTimeout, ReadTimeout, WriteTimeout}` + `http.TimeoutHandler` per-route + `context.WithTimeout` per-request.
- **Graceful shutdown** — `srv.Shutdown(ctx)` berhenti terima koneksi baru & tunggu in-flight. Lihat [[20-graceful-shutdown]].
- **Streaming** — `http.ResponseController` (Go 1.20) untuk `Flush`/deadline per-request; SSE via flush berkala.
- **Batasi body** — `http.MaxBytesReader` cegah abuse memori.
- **`httptest`** untuk unit test handler tanpa jaringan nyata.
