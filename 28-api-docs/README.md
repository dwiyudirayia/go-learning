# 28 — API Documentation (OpenAPI/Swagger)

API tanpa dokumentasi = sulit dipakai orang lain (frontend, mobile, partner). **OpenAPI** (dulu Swagger) adalah standar mendeskripsikan REST API dalam format mesin, lalu menghasilkan dokumentasi interaktif & SDK client otomatis.

Jalankan:
```bash
go run ./28-api-docs
# http://localhost:8080/docs         -> Swagger UI (coba API dari browser)
# http://localhost:8080/openapi.json -> spesifikasi
```
Verifikasi otomatis: `go test ./28-api-docs`

## 1. Spesifikasi OpenAPI

File `openapi.json` mendeskripsikan endpoint, request/response, & schema:
```json
"/books": {
  "post": {
    "requestBody": { "...": "BookInput" },
    "responses": { "201": {...}, "400": { "...": "Error" } }
  }
}
```
Dari spec ini kamu bisa **generate**: dokumentasi (Swagger UI), client SDK (banyak bahasa), mock server, & test kontrak.

### Dua pendekatan
- **Spec-first** (modul ini): tulis OpenAPI dulu, lalu implementasi mengikuti. Bagus untuk kontrak antar tim.
- **Code-first**: anotasi di kode Go, generate spec dengan [swaggo/swag](https://github.com/swaggo/swag) (`// @Summary ...`).

Spec di-*embed* ke binary (`//go:embed openapi.json`) & disajikan di `/openapi.json`. Swagger UI (dari CDN) memuatnya di `/docs`.

## 2. Format Error yang Konsisten

Semua error memakai **satu bentuk** — klien tak perlu menebak:
```json
{ "error": { "code": "VALIDATION_ERROR", "message": "title & author wajib diisi" } }
```
`code` (stabil, untuk logika program) + `message` (untuk manusia). Test memverifikasi bentuk ini. Bandingkan dengan Modul 5 (pemetaan error) & Modul 13 (error handler terpusat).

## 3. API Versioning

`/api/v1/books` — saat ada **breaking change**, buat `/api/v2` tanpa merusak klien lama.
| Strategi | Contoh |
|----------|--------|
| Path (paling umum) | `/api/v1/...`, `/api/v2/...` |
| Header | `Accept: application/vnd.app.v2+json` |
| Query | `?version=2` |

## Praktik desain REST yang baik
- **Noun, bukan verb**: `/books` (bukan `/getBooks`). Aksi lewat HTTP method.
- **Status code tepat** (Modul 12): 200/201/204/400/401/404/409/500.
- **Pagination** konsisten: `?limit=&offset=` atau cursor.
- **Filtering/sorting**: `?author=X&sort=-created_at`.
- **Idempotensi** untuk PUT/DELETE.
- **HATEOAS** (opsional): sertakan link terkait.

## Kapan & Di Mana Dipakai
- API yang dikonsumsi tim/klien lain, API publik, dokumentasi yang harus selalu sinkron dengan kode.

## Latihan
1. Dokumentasikan endpoint `GET /books/{id}` di `openapi.json` + implementasinya.
2. Tambah kode error `NOT_FOUND` (404) dengan format konsisten.
3. Pakai `swaggo/swag` untuk meng-generate spec dari anotasi kode (code-first).
4. Tambah `/api/v2/books` dengan field tambahan tanpa merusak v1.
5. Generate client SDK dari spec (openapi-generator) untuk satu bahasa.

## ✅ Solusi Latihan (Pembahasan)

1. **Dokumentasikan `GET /books/{id}`** — tambah path item di `openapi.json` dengan `parameters` (path `id`, integer) + response `200` (schema Book) & `404`. Implementasikan handler yang cocok.
2. **Error `NOT_FOUND` konsisten** — definisikan schema error `{code, message}` sekali, referensikan (`$ref`) di semua response 4xx. Handler balas `{"code":"NOT_FOUND","message":"..."}` + status 404.
3. **`swaggo/swag` (code-first)** — anotasi handler dengan komentar `// @Summary`, `// @Success 200 {object} Book`, jalankan `swag init` → generate spec + Swagger UI dari kode. Kebalikan dari design-first.
4. **Versi v2 tanpa merusak v1** — daftarkan route `/api/v2/books` terpisah dengan schema baru (field tambahan). v1 tetap utuh → klien lama tak rusak (backward compatible).
5. **Generate client SDK** — `openapi-generator-cli generate -i openapi.json -g go -o client/`. Dari satu spec dapat SDK banyak bahasa — kontrak jadi sumber kebenaran.

---

## 🚀 Teknik Advanced (Level Up)
> 💻 **Contoh runnable + komentar detail** untuk teknik di bawah ada di folder [`advanced/`](advanced). Jalankan: `go run ./28-api-docs/advanced`


- **Spec-first vs code-first** — spec-first (tulis OpenAPI dulu, generate server/client via `oapi-codegen`) menjaga kontrak; code-first (`swag` dari komentar) lebih cepat tapi rawan drift.
- **Validasi dari spec** — middleware validasi request/response terhadap OpenAPI (kkc) agar dokumen = perilaku.
- **Codegen klien** — hasilkan SDK klien type-safe dari spec untuk konsumen.
- **Versioning API** — `/v1`, `/v2`; deprecation policy & header `Sunset`.
- **Contract testing** — pastikan implementasi tak menyimpang dari kontrak (Pact/dredd).
- **Contoh & error schema** — dokumentasikan bentuk error konsisten (lihat pemetaan error di [[15-studi-kasus-rest]]).
