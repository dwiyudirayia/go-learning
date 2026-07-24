# CLAUDE.md

Panduan untuk Claude Code saat bekerja di repo ini. **Bahasa jawaban: Indonesia.**

## Apa ini

`go-learning` — kurikulum belajar Go menyeluruh (**48 modul**), untuk user yang sudah paham bahasa lain tapi baru di Go. Fokus: idiom khas Go, tempo cepat, tiap konsep dijelaskan dari nol. Ini **repo belajar**, bukan aplikasi produksi — tujuan tiap file adalah **mengajar**, jadi utamakan kejelasan & komentar yang mendidik di atas keringkasan.

Go **1.26.4**, satu `go.mod` (module `go-learning`). Roadmap & progres di `README.md` root; panduan cara belajar di `LEARNING.md`.

## Perintah

```bash
go run ./NN-nama          # jalankan satu modul (mis. go run ./01-basics)
go run ./NN-nama/advanced # jalankan contoh Teknik Advanced (08 & 37: pakai go test)
go test ./...             # semua test (harus 54 paket lulus, 0 gagal)
go test -race ./NN-nama   # test satu modul + deteksi data race
gofmt -l .                # cek format (harus kosong)
go vet ./...              # harus bersih
make run MOD=01-basics    # alias make untuk run
make advanced MOD=01-basics      # run contoh advanced
make realcase-up MOD=22-caching  # nyalakan compose infra real-case
make test-race            # semua test + race
```

**Sebelum menyatakan selesai**, selalu jalankan: `gofmt -l .` (kosong), `go vet ./...`, `go build ./...`, `go test ./...` — semua harus hijau.

## Struktur & konvensi tiap modul

```
NN-nama/
├── README.md                  ← konsep + "Kapan & Di Mana Dipakai" + Latihan + "🚀 Teknik Advanced"
├── main.go (atau *.go)        ← MATERI: kode berkomentar, runnable
├── advanced/                  ← contoh runnable Teknik Advanced (SEMUA 48 modul punya)
├── real-case/                 ← versi tech stack produksi (24 modul; env-guarded auto-skip)
├── latihan/solusi.go          ← kunci jawaban (modul 1–9); package main terisolasi
└── jawaban-saya/main.go       ← workspace latihan user (template TODO) — JANGAN disentuh
```

- Tiap subfolder = **package terpisah**. `latihan/`, `advanced/`, `real-case/`, `jawaban-saya/` adalah package yang berdiri sendiri — jangan impor antar-modul. `advanced/` umumnya `package main` (`go run`), kecuali modul **08 & 37** = package non-main (`go test`).
- **Kunci jawaban** letaknya beda per rentang: modul **1–9** → `latihan/solusi.go`; modul **11–17** → bagian README "Status Solusi Latihan"; modul **10, 18–48** → bagian README `## ✅ Solusi Latihan (Pembahasan)`. Jaga konsistensi ini saat menambah/mengubah.
- README selalu punya bagian **"Kapan & Di Mana Dipakai"** (studi kasus nyata) + **Latihan** + **"🚀 Teknik Advanced (Level Up)"** (tertaut ke folder `advanced/`).
- **Konvensi analogi**: konsep di file materi utama **dan** `advanced/` diberi komentar berpenanda **`// 🔍 Analogi:`** — analogi sehari-hari ramah pemula (mis. circuit breaker = sekring, JWT = gelang konser). **Materi/teknik BARU wajib ikut konvensi ini.**
- File `*.pb.go` di-**generate** (modul 16, 17, 48) — jangan diedit tangan; regen via `make proto`.
- Folder pendukung root: `docs/` (9 rujukan: IDIOM, CHEATSHEET, MEMBACA-TIPE, PITFALLS, CONCURRENCY, TESTING, TOOLING, GLOSSARY + README indeks), `REAL-CASE-STACKS.md` (peta double→stack produksi per modul), `41-capstone/workshop/` (capstone guided step-by-step, baru Langkah 1).
- `libraries/` (DI LUAR modul 01–48, tak menambah hitungan modul): katalog library Go populer (`README.md`) + **23 subfolder contoh runnable + test**, tiap subfolder `package main` mandiri bergaya `// 🔍 Analogi:` (uuid, ulid, testify, gocmp, zerolog, lo, decimal, gjson, cron, resty, validator, cobra, viper, env, jwt, bcrypt, chi, sqlx, errgroup, ratelimit, redis, gorm, prometheus). Ikut `go test ./...` — jaga tetap hijau. Deps baru: testify, zerolog, samber/lo, shopspring/decimal, robfig/cron, go-resty/resty, google/uuid, go-chi/chi, jmoiron/sqlx, oklog/ulid, google/go-cmp, caarlos0/env (semua dipromosikan jadi direct bila perlu). bcrypt pakai x/crypto & gjson/sjson pakai tidwall yang sudah ada.

## Framework & keputusan yang sudah dikunci

- **Web framework = Fiber v2** (bukan Gin). Modul 13 & 15 pakai Fiber. Jangan ganti tanpa diminta.
- **SQLite pure-Go**: `modernc.org/sqlite` (dan `glebarez` di modul 14) — tanpa cgo.
- Migrasi: `golang-migrate` v4 (modul 21). Cache: `go-redis` + `miniredis`. LLM: `anthropic-sdk-go`, model default `claude-opus-4-8` (modul 40 — **baca skill `claude-api` dulu** untuk model ID, jangan menebak).
- Test infra memakai **in-memory doubles** agar jalan tanpa infra eksternal: `bufconn` (gRPC), `miniredis`, NATS embedded, `httptest`. Test integrasi berat auto-skip bila env (mis. `RABBITMQ_URL`, `KAFKA_BROKERS`) tak ada.

## Catatan lingkungan (WSL + editor Windows)

- `go get` / `go mod download` **bekerja** (proxy.golang.org content-addressed), tapi kadang `connection reset by peer` — retry beberapa kali. **SHA256 tarball Go tak reliable** di WSL ini (proxy re-gzip) — verifikasi instalasi Go secara fungsional (`go build`), bukan lewat hash.
- Port lokal sering sibuk (3000, 8099). Server baca env `PORT`/`GRPC_ADDR`/`INVENTORY_ADDR`.
- Docker mungkin tak terpasang — Dockerfile/compose disediakan sebagai referensi; verifikasi lewat `go test`, bukan menjalankan container.
- gopls kadang menampilkan diagnostics context `[js,wasm]` (modul 47) — itu artefak, **bukan error host**. `go build/vet/test ./...` di host tetap bersih.
- **User juga membuka repo dari VS Code sisi Windows** (`C:\AMB\...`) — gopls di sana menganalisis dgn `GOOS=windows`. Kode harus **compile lintas platform**: hindari API Unix-only (mis. `syscall.Kill` → pakai `os.FindProcess(os.Getpid()).Signal(syscall.SIGTERM)`). Bila ragu, cek dgn `GOOS=windows go build ./NN-nama/...`.

## Jebakan yang pernah bikin merah (jangan diulang)

- Literal `%w`/`%v` di string `fmt.Println` → vet gagal; tulis 'verb-w'/'verb-v' di komentar/teks.
- `http.Get(url)` lalu pakai `resp` tanpa cek `err` dulu → analyzer `httpresponse` gagal.
- Komentar dgn indentasi spasi >1 setelah `//` bisa kena reflow gofmt → cek `gofmt -l` setelah edit komentar.
- `rand.Seed` deprecated → pakai `rand.New(rand.NewSource(...))`.
- Cara cek `jawaban-saya/` masih template: `diff jawaban-saya/main.go latihan/solusi.go` (bukan hitung jumlah func).

## Gaya kerja di repo ini

- Kode & komentar **dalam Bahasa Indonesia**, idiomatik Go (table-driven test, sentinel error + `errors.Is/As`, interface untuk mock, layered arch handler/service/repository).
- Modul **DIKUNCI di 48** — jangan menambah modul baru kecuali user memintanya eksplisit.
- Saat menambah/mengubah, jaga repo tetap hijau: setiap `package main` baru harus valid & lolos build/vet/test.
