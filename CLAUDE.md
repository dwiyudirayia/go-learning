# CLAUDE.md

Panduan untuk Claude Code saat bekerja di repo ini. **Bahasa jawaban: Indonesia.**

## Apa ini

`go-learning` — kurikulum belajar Go menyeluruh (**48 modul**), untuk user yang sudah paham bahasa lain tapi baru di Go. Fokus: idiom khas Go, tempo cepat, tiap konsep dijelaskan dari nol. Ini **repo belajar**, bukan aplikasi produksi — tujuan tiap file adalah **mengajar**, jadi utamakan kejelasan & komentar yang mendidik di atas keringkasan.

Go **1.26.4**, satu `go.mod` (module `go-learning`). Roadmap & progres di `README.md` root; panduan cara belajar di `LEARNING.md`.

## Perintah

```bash
go run ./NN-nama          # jalankan satu modul (mis. go run ./01-basics)
go test ./...             # semua test (harus 48 paket lulus, 0 gagal)
go test -race ./NN-nama   # test satu modul + deteksi data race
gofmt -l .                # cek format (harus kosong)
go vet ./...              # harus bersih
make run MOD=01-basics    # alias make untuk run
make test-race            # semua test + race
```

**Sebelum menyatakan selesai**, selalu jalankan: `gofmt -l .` (kosong), `go vet ./...`, `go build ./...`, `go test ./...` — semua harus hijau.

## Struktur & konvensi tiap modul

```
NN-nama/
├── README.md                  ← konsep + "Kapan & Di Mana Dipakai" + soal Latihan
├── main.go (atau *.go)        ← MATERI: kode berkomentar, runnable
├── latihan/solusi.go          ← kunci jawaban (modul 1–9); package main terisolasi
└── jawaban-saya/main.go       ← workspace latihan user (template TODO)
```

- Tiap subfolder = **package terpisah**. `latihan/` dan `jawaban-saya/` adalah `package main` yang berdiri sendiri — jangan impor antar-modul.
- **Kunci jawaban** letaknya beda per rentang: modul **1–9** → `latihan/solusi.go`; modul **11–17** → bagian README "Status Solusi Latihan"; modul **10, 18–48** → bagian README `## ✅ Solusi Latihan (Pembahasan)`. Jaga konsistensi ini saat menambah/mengubah.
- README selalu punya bagian **"Kapan & Di Mana Dipakai"** (studi kasus nyata) + **Latihan**.
- File `*.pb.go` di-**generate** (modul 16, 17, 48) — jangan diedit tangan; regen via `make proto`.

## Framework & keputusan yang sudah dikunci

- **Web framework = Fiber v2** (bukan Gin). Modul 13 & 15 pakai Fiber. Jangan ganti tanpa diminta.
- **SQLite pure-Go**: `modernc.org/sqlite` (dan `glebarez` di modul 14) — tanpa cgo.
- Migrasi: `golang-migrate` v4 (modul 21). Cache: `go-redis` + `miniredis`. LLM: `anthropic-sdk-go`, model default `claude-opus-4-8` (modul 40 — **baca skill `claude-api` dulu** untuk model ID, jangan menebak).
- Test infra memakai **in-memory doubles** agar jalan tanpa infra eksternal: `bufconn` (gRPC), `miniredis`, NATS embedded, `httptest`. Test integrasi berat auto-skip bila env (mis. `RABBITMQ_URL`, `KAFKA_BROKERS`) tak ada.

## Catatan lingkungan (WSL)

- `go get` / `go mod download` **bekerja** (proxy.golang.org content-addressed). Tapi **SHA256 tarball Go tak reliable** di WSL ini (proxy re-gzip) — verifikasi instalasi Go secara fungsional (`go build`), bukan lewat hash.
- Port lokal sering sibuk (3000, 8099). Server baca env `PORT`/`GRPC_ADDR`/`INVENTORY_ADDR`.
- Docker mungkin tak terpasang — Dockerfile/compose disediakan sebagai referensi; verifikasi lewat `go test`, bukan menjalankan container.
- gopls kadang menampilkan diagnostics context `[js,wasm]` (modul 47) — itu artefak, **bukan error host**. `go build/vet/test ./...` di host tetap bersih.

## Gaya kerja di repo ini

- Kode & komentar **dalam Bahasa Indonesia**, idiomatik Go (table-driven test, sentinel error + `errors.Is/As`, interface untuk mock, layered arch handler/service/repository).
- Modul **DIKUNCI di 48** — jangan menambah modul baru kecuali user memintanya eksplisit.
- Saat menambah/mengubah, jaga repo tetap hijau: setiap `package main` baru harus valid & lolos build/vet/test.
