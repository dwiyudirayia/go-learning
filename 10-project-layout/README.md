# 10 — Struktur Proyek, Modules & Makefile

Setelah paham bahasa, kamu perlu tahu **cara menata proyek Go** yang idiomatik & skalabel. Ini fondasi sebelum bikin aplikasi nyata (CLI, REST API, microservices).

Jalankan contoh:
```bash
go run ./10-project-layout/example/cmd/greeter -name Ana
go run ./10-project-layout/example/cmd/greeter -name Budi -shout
```

## 1. Layout standar komunitas

Tidak ada aturan resmi, tapi pola ini sangat umum (mirip *golang-standards/project-layout*):

```
myapp/
├── go.mod, go.sum
├── Makefile
├── cmd/                # entry point aplikasi (main package)
│   └── myapp/main.go   # 1 folder per binary
├── internal/          # kode privat — TAK BISA di-import proyek lain
│   ├── handler/       # lapisan HTTP
│   ├── service/       # logika bisnis
│   └── repository/    # akses data
├── pkg/               # kode yang BOLEH dipakai proyek lain (opsional)
├── api/               # spec: OpenAPI, .proto
├── configs/           # file konfigurasi
└── migrations/        # migrasi database
```

### `internal/` itu spesial
Folder bernama `internal/` **dipaksa oleh compiler**: hanya boleh di-import oleh kode di dalam parent-nya. Ini cara Go menegakkan enkapsulasi antar-modul — taruh detail implementasi di sini.

### `cmd/`
Tiap subfolder di `cmd/` = satu binary. Contoh: `cmd/server`, `cmd/worker`, `cmd/migrate`. Isi `main.go`-nya **tipis** — hanya wiring; logika ada di `internal/`.

## 2. Arsitektur berlapis (layered)

Pola paling umum untuk backend Go:
```
handler (HTTP)  ->  service (bisnis)  ->  repository (data)
     |                   |                     |
  parse req         aturan bisnis         query DB
  tulis resp        validasi              simpan/ambil
```
Tiap lapis bergantung pada **interface** lapis di bawahnya (Modul 4) → mudah di-test (Modul 8). Kita pakai pola ini di studi kasus REST nanti.

## 3. Go Modules

```bash
go mod init github.com/user/myapp   # buat modul (nama = import path)
go get github.com/gofiber/fiber/v2  # tambah dependency
go get -u ./...                     # update dependency
go mod tidy                         # rapikan go.mod & go.sum (buang tak terpakai)
go mod download                     # unduh semua ke cache
go mod vendor                       # (opsional) salin deps ke ./vendor
```
- `go.mod` — daftar modul & versi dependency langsung.
- `go.sum` — checksum kriptografis (integritas). **Commit keduanya.**
- Versi mengikuti **SemVer**; major v2+ masuk ke import path (`/v2`).

## 4. Makefile (otomasi perintah)

Menyederhanakan perintah yang sering dipakai. Lihat `example/Makefile`:
```make
run:        ; go run ./cmd/greeter
test:       ; go test ./...
build:      ; go build -o bin/app ./cmd/greeter
lint:       ; go vet ./...
```
Jalankan: `make run`, `make test`, dst.

## Contoh di modul ini
`example/` mendemokan struktur mini yang benar:
```
example/
├── Makefile
├── cmd/greeter/main.go          # entry point (tipis)
└── internal/greeting/greeting.go # logika (privat)
```
`main.go` mem-parse flag lalu memanggil `internal/greeting` — persis pola `cmd` → `internal`.

## Latihan
1. Tambah binary kedua `cmd/farewell/` yang memakai package `internal/greeting` yang sama.
2. Tambah fungsi baru di `internal/greeting` + test-nya (`greeting_test.go`).
3. Tambah target `make build` dan coba hasilkan binary ke `bin/`.
4. Jelaskan (di komentar) kenapa `internal/greeting` tak bisa di-import dari luar `example/`.
