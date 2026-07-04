# 29 — Clean / Hexagonal Architecture

Pendalaman arsitektur berlapis (Modul 10 & 15). **Hexagonal** (Ports & Adapters) menempatkan **logika bisnis di inti** yang tak tahu apa-apa soal HTTP, database, atau framework. Dunia luar terhubung lewat **port** (interface) & **adapter** (implementasi).

Jalankan:
```bash
go run ./29-clean-architecture
go test ./29-clean-architecture/...
```

## Aturan emas: Dependency Rule

> Dependency selalu menunjuk **ke dalam** (menuju core). Core tidak pernah tahu soal dunia luar.

```
        ┌─────────────────────────────────────┐
        │   Adapter (driving): HTTP / gRPC / CLI│  ← "menggerakkan" app
        │   ┌─────────────────────────────┐    │
        │   │   Use Case (service)         │    │  ← logika aplikasi
        │   │   ┌─────────────────────┐    │    │
        │   │   │   Domain (entity +   │    │    │  ← aturan bisnis MURNI
        │   │   │   port interfaces)   │    │    │
        │   │   └─────────────────────┘    │    │
        │   └─────────────────────────────┘    │
        │   Adapter (driven): DB / cache / MQ    │  ← "digerakkan" oleh app
        └─────────────────────────────────────┘
```

## Lapisan di modul ini

| Lapisan | Package | Bergantung pada | Tahu soal HTTP/DB? |
|---------|---------|-----------------|--------------------|
| **Domain** | `internal/domain` | — (tidak apa pun) | ❌ tidak |
| **Use case** | `internal/service` | domain (port) | ❌ tidak |
| **Adapter driven** | `internal/adapter/memory` | domain | ✅ (DB) |
| **Adapter driving** | `internal/adapter/rest` | service, domain | ✅ (HTTP) |
| **Composition root** | `main.go` | semua | merakit |

### Port (interface) ada di CORE
```go
// domain/note.go — core mendefinisikan APA yang dibutuhkan
type NoteRepository interface { Save(*Note) error; FindByID(int) (*Note, error) }
```
Adapter di luar **mengimplementasikan** port itu:
```go
// adapter/memory — var _ domain.NoteRepository = (*Repo)(nil)  // dijamin cocok
```

## Kenapa repot begini?

- **Testable**: use case diuji dengan repo **palsu**, tanpa DB/HTTP (lihat `service/note_test.go`) — cepat & deterministik.
- **Ganti teknologi tanpa sentuh bisnis**: memory → Postgres = tulis adapter baru, logika tak berubah.
- **Ganti antarmuka**: REST → gRPC = adapter driving baru, core tetap.
- **Bisnis terlindungi** dari perubahan framework (framework datang & pergi).

## Kapan pakai (dan tidak)
- ✅ Aplikasi menengah–besar, logika bisnis kompleks, umur panjang, banyak integrasi.
- ❌ Script kecil / CRUD sederhana → over-engineering. Layered biasa (Modul 15) sudah cukup.

Jangan kultus arsitektur — pilih sesuai ukuran masalah.

## Kaitan dengan modul lain
Ini versi lebih ketat dari Modul 15 (yang layered). Bandingkan keduanya: 15 = pragmatis, 29 = pemisahan maksimal. Keduanya valid.

## Latihan
1. Tambah adapter **Postgres/GORM** (Modul 14) yang mengimplementasikan `NoteRepository` — main tinggal ganti satu baris.
2. Tambah adapter **gRPC** (Modul 16) sebagai driving adapter kedua, memakai service yang sama.
3. Tambah use case `UpdateNote` + `DeleteNote`.
4. Tambah port kedua `Notifier` (mis. kirim event saat note dibuat) + adapter NATS (Modul 23).
5. Tulis test untuk adapter `rest` memakai `httptest`.

## ✅ Solusi Latihan (Pembahasan)

1. **Adapter Postgres/GORM** — buat tipe baru yang mengimplementasikan interface `NoteRepository` (port). Di `main`, cukup ganti `repo := NewMemRepo()` → `repo := NewGormRepo(db)`. Inti (use case) **tak berubah** — itulah gunanya port/adapter.
2. **Adapter gRPC (driving)** — server gRPC memanggil use case yang sama seperti handler REST. Dua "driving adapter" (REST + gRPC), satu inti bisnis.
3. **`UpdateNote` + `DeleteNote`** — tambah method di interface use case + implementasi; adapter tinggal memetakan request→use case.
4. **Port `Notifier` + adapter NATS** — definisikan `type Notifier interface{ NoteCreated(Note) }`; inti memanggilnya tanpa tahu implementasinya. Adapter NATS (Modul 23) publish event; test pakai mock Notifier.
5. **Test adapter `rest`** — `httptest.NewRequest` + `httptest.NewRecorder`, inject use case (atau mock repo), assert status & body JSON. Adapter diuji terpisah dari inti.
