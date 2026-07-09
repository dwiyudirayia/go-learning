# 15 — 📦 Studi Kasus: REST API Task Manager + Auth JWT

Menggabungkan **semua** yang sudah dipelajari: Fiber (13) + GORM (14) + arsitektur berlapis (10) + interface (4) + error handling (5) + testing (8) + concurrency-safe (7).

Jalankan:
```bash
go run ./15-studi-kasus-rest            # :3000
PORT=8099 go run ./15-studi-kasus-rest  # port lain
```
Verifikasi otomatis (integration test end-to-end): `go test ./15-studi-kasus-rest`

## Alur pemakaian
```bash
# 1. Daftar
curl -X POST localhost:3000/auth/register -H 'Content-Type: application/json' \
  -d '{"name":"Ana","email":"ana@mail.id","password":"secret123"}'

# 2. Login -> dapat token
curl -X POST localhost:3000/auth/login -H 'Content-Type: application/json' \
  -d '{"email":"ana@mail.id","password":"secret123"}'
# {"token":"eyJ..."}

# 3. Pakai token untuk endpoint terproteksi
TOKEN=eyJ...
curl localhost:3000/tasks/ -H "Authorization: Bearer $TOKEN"
curl -X POST localhost:3000/tasks/ -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' -d '{"title":"belajar JWT"}'
curl -X PATCH  localhost:3000/tasks/1/done -H "Authorization: Bearer $TOKEN"
curl -X DELETE localhost:3000/tasks/1      -H "Authorization: Bearer $TOKEN"
```

## Arsitektur berlapis

```
        HTTP (Fiber)
handler  ──►  service  ──►  repository  ──►  GORM/DB
(parse,      (logika       (akses data,
 status)      bisnis,       interface)
              otorisasi)
   ▲             ▲              ▲
   │         middleware     model (entity)
   │          (JWT auth)
 token util (JWT generate/parse)
```

| Lapisan | Package | Tanggung jawab |
|---------|---------|----------------|
| Model | `internal/model` | entity domain (User, Task) |
| Repository | `internal/repository` | akses DB via **interface** (mudah di-mock) |
| Service | `internal/service` | logika bisnis + **otorisasi kepemilikan** |
| Handler | `internal/handler` | HTTP: parse, panggil service, map error→status |
| Middleware | `internal/middleware` | proteksi JWT (`Bearer`) |
| Token | `internal/token` | buat/verifikasi JWT (HS256) |

**Kunci desain:** lapisan atas bergantung pada **interface** lapisan bawah (Modul 4), bukan implementasi konkret. Ini yang membuat setiap lapis **testable** dan implementasi (mis. GORM→Postgres) bisa ditukar tanpa mengubah service/handler.

## Konsep keamanan yang diterapkan

1. **Password di-hash** dengan `bcrypt` — tidak pernah disimpan/dibalas plaintext. Field `PasswordHash` punya tag `json:"-"` → tak pernah bocor ke response.
2. **JWT** (HS256) berisi `userID`, berlaku 24 jam. Secret dari `JWT_SECRET` (env).
3. **Cegah "alg=none"** — parser menolak token yang bukan HMAC.
4. **Otorisasi kepemilikan** — `service.ownedTask` memastikan user hanya bisa mengubah/menghapus task **miliknya** (test membuktikan user B → 403 atas task user A).
5. **Pesan login ambigu** — "email atau password salah" (tak membocorkan email mana yang ada).

## Pemetaan error → status HTTP
| Error bisnis | Status |
|--------------|--------|
| `ErrEmailTaken` | 409 Conflict |
| `ErrInvalidCredentials` | 401 Unauthorized |
| `ErrForbidden` | 403 Forbidden |
| `ErrNotFound` | 404 Not Found |
| lainnya | 500 |

## Menuju produksi (checklist)
- [ ] `JWT_SECRET` kuat dari env/secret manager (bukan default).
- [ ] Ganti SQLite → PostgreSQL (cukup ganti driver + DSN, lihat Modul 14).
- [ ] Migrasi eksplisit (bukan `AutoMigrate`) + tool seperti `golang-migrate`.
- [ ] Rate limiting & CORS (middleware Fiber).
- [ ] Refresh token & logout (token blacklist / rotasi).
- [ ] Structured logging + request ID + graceful shutdown.
- [ ] Validasi input lebih ketat (validator, Modul 13).

## Latihan
1. Tambah endpoint `GET /me` yang mengembalikan profil user dari token.
2. Tambah `PATCH /tasks/:id` untuk mengubah judul (dengan cek kepemilikan).
3. Tulis **unit test** `service` memakai **mock repository** (bukan DB) — buktikan otorisasi tanpa menyentuh GORM.
4. Tambah pagination pada `GET /tasks`.
5. Tambah refresh token.

## ✅ Status Solusi Latihan
Latihan **1, 2, 3 sudah diselesaikan**: `GET /me`, `PATCH /tasks/:id` (ubah judul), dan **unit test service dengan mock repo** (internal/service/service_test.go). Latihan 4 & 5 (pagination, refresh token) sebagai tantangan.

---

## 🚀 Teknik Advanced (Level Up)
> 💻 **Contoh runnable + komentar detail** untuk teknik di bawah ada di folder [`advanced/`](advanced). Jalankan: `go run ./15-studi-kasus-rest/advanced`


- **Uji berlapis dengan mock repo** — service diuji terhadap interface repository palsu (tanpa DB). Handler diuji via `httptest`/`app.Test`.
- **JWT lanjutan** — access token pendek + refresh token dengan **rotasi** (revoke lama saat dipakai), simpan hash refresh. Lihat [[27-security]].
- **bcrypt cost** — pilih cost seimbang (mis. 10–12); ukur latensi. Bandingkan argon2id untuk keamanan lebih tinggi.
- **Otorisasi** — middleware RBAC/ABAC; jangan cek role di handler tersebar. Lihat [[44-auth-advanced]].
- **Validasi & DTO** — pisahkan model domain dari request/response DTO; validasi di boundary.
- **Idempotency & pagination** — idempotency-key untuk POST yang tak boleh ganda; pagination berbasis cursor (bukan offset) untuk data besar.
- **Error → HTTP konsisten** — satu tempat pemetaan sentinel error → status + body terstruktur.
