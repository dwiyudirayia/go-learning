# 41 — 📦 CAPSTONE: URL Shortener

Studi kasus **besar** yang menggabungkan banyak modul menjadi satu aplikasi siap-produksi (mini). Inilah tujuan belajar: bukan konsep terpisah, tapi **merangkainya**.

Jalankan (tanpa infra apa pun — SQLite temp + Redis in-memory):
```bash
go run ./41-capstone
```
Verifikasi otomatis: `go test ./41-capstone`

Coba:
```bash
curl -X POST localhost:8080/auth/register -H 'Content-Type: application/json' \
  -d '{"email":"a@b.c","password":"secret123"}'
TOKEN=$(curl -s -X POST localhost:8080/auth/login -H 'Content-Type: application/json' \
  -d '{"email":"a@b.c","password":"secret123"}' | sed -n 's/.*"token":"\([^"]*\)".*/\1/p')
curl -X POST localhost:8080/api/shorten -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' -d '{"url":"https://go.dev"}'
# -> {"code":"aB3x_9", "short":"/aB3x_9", ...}
curl -i localhost:8080/aB3x_9      # -> 302 Location: https://go.dev
```

## 🧩 Modul apa saja yang digabung di sini?

| Fitur di capstone | Berasal dari modul | File |
|-------------------|--------------------|------|
| REST API (Fiber, middleware, error handler) | 13 | `handler.go` |
| Database (`database/sql`, SQLite) | 14 | `store.go` |
| Auth JWT + bcrypt | 15, 27 | `auth.go` |
| **Cache-aside** (Redis) | 22 | `cache.go`, `service.go` |
| Config via environment | 19 | `main.go` (`getenv`) |
| Graceful shutdown | 20 | `main.go` |
| Structured logging (`slog`) | 18 | seluruh |
| Health check | 30 | `handler.go` (`/healthz`) |
| Arsitektur berlapis (store→service→handler) | 10, 29 | struktur file |
| Interface + mock untuk test | 4, 8 | `app_test.go` |

## Alur permintaan (arsitektur)

```
POST /api/shorten (JWT) ──► handler ──► service ──► store (INSERT link)
GET  /:code              ──► handler ──► service ──► cache HIT? ─► redirect
                                                  └─ MISS ─► store ─► isi cache ─► redirect
```

### Pola kunci: Cache-Aside pada redirect
Redirect adalah jalur panas (hot path) — dioptimalkan dengan cache:
```go
func Resolve(code) {
    if url, hit := cache.Get(code); hit { return url }   // cepat
    link := store.GetByCode(code)                         // miss -> DB
    cache.Set(code, link.URL)                             // isi cache
    return link.URL
}
```
Test membuktikan redirect kedua dilayani dari cache.

## Cara mempelajari capstone ini (untuk fokus belajar)

1. **Baca `store.go`** — lapisan data paling dasar. Pahami tabel & query.
2. **Baca `service.go`** — logika bisnis + cache-aside. Ini "otak" aplikasi.
3. **Baca `auth.go`** — bagaimana JWT & bcrypt dipakai.
4. **Baca `handler.go`** — bagaimana HTTP dipetakan ke service.
5. **Baca `main.go`** — bagaimana semuanya **dirakit** (composition root) + shutdown.
6. **Baca `app_test.go`** — bagaimana menguji integrasi tanpa infra.

Lalu jalankan `go test -v ./41-capstone` dan ikuti alurnya.

## Menuju produksi (latihan bertahap)
- [ ] Hitung klik **asinkron** via worker queue (Modul 25) agar redirect makin cepat.
- [ ] Tambah metrics Prometheus (Modul 18) + `/metrics`.
- [ ] Ganti SQLite → PostgreSQL + sqlc (Modul 36) + migrasi (Modul 21).
- [ ] Tambah rate limiting (Modul 27) pada `/api/shorten`.
- [ ] Tambah custom alias, expiry, & analitik per-link.
- [ ] Tambah distributed tracing (Modul 33) & deploy K8s (Modul 30, 39).

## Latihan
1. Tambah endpoint `GET /api/links` (daftar link milik user, dari token).
2. Tambah `DELETE /api/links/:code` (hapus, cek kepemilikan).
3. Tambah invalidasi cache saat link dihapus (Modul 22).
4. Tambah unit test `service` dengan **mock store & cache** (bukan DB nyata).
5. Tambah `Dockerfile` (Modul 30) & jalankan dengan PostgreSQL + Redis via docker-compose.

## ✅ Solusi Latihan (Pembahasan)

1. **`GET /api/links`** — endpoint terproteksi; baca `userID := c.Locals("userID")`, tambah `store.ListLinksByUser(userID)` (`SELECT ... WHERE user_id = ?`), kembalikan JSON array.
2. **`DELETE /api/links/:code`** — ambil link by code, **cek kepemilikan** (`link.UserID == userID`) → kalau bukan, 403; kalau ya, `DELETE FROM links WHERE code = ?`.
3. **Invalidasi cache saat hapus** — setelah delete di DB, `cache.Del(ctx, "url:"+code)` (Modul 22) agar redirect lama tak melayani dari cache basi.
4. **Unit test service dengan mock** — definisikan interface `storeIface` & `cacheIface`, buat mock in-memory, uji `Shorten`/`Resolve` **tanpa** DB/Redis nyata (cepat, deterministik — Modul 08, 15).
5. **Dockerfile + compose** — multi-stage build (Modul 30) + `docker-compose.yml` berisi service app, `postgres`, `redis`; app baca `DB_PATH`/`REDIS_ADDR` dari env (Modul 19). Ganti driver SQLite→pgx untuk Postgres.
