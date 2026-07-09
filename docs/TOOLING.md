# 🧰 Go Tooling

Perintah `go` & alat yang dipakai sehari-hari. Repo ini juga menyediakan `Makefile` (ketik `make help`).

## Menjalankan & membangun

```bash
go run ./01-basics            # kompilasi + jalankan (tak simpan biner)
go build ./...                # kompilasi semua paket (cek error)
go build -o app ./cmd/app     # hasilkan biner bernama app
go install ./cmd/app          # pasang biner ke $GOBIN
```

## Format & analisa statis

```bash
gofmt -l .                    # daftar file belum terformat (harus KOSONG)
gofmt -w .                    # format & tulis
go vet ./...                  # tangkap bug umum (printf salah, dll) — harus BERSIH
```
> Standar repo: `gofmt -l .` kosong, `go vet ./...` bersih, `go build ./...` & `go test ./...` hijau **sebelum** menyatakan selesai.

## Test

```bash
go test ./...                 # semua test
go test -v ./NN-nama          # verbose satu paket
go test -run TestFoo ./pkg    # filter nama test (regex)
go test -race ./...           # detektor data race (kode konkuren)
go test -cover ./...          # ringkas coverage
go test -coverprofile=c.out ./pkg && go tool cover -func=c.out  # per-fungsi
go test -bench . -benchmem ./pkg   # benchmark + alokasi/op
go test -run x -fuzz FuzzFoo -fuzztime 10s ./pkg  # fuzzing
```
Bandingkan benchmark andal dengan `benchstat` (bukan satu angka).

## Modules

```bash
go mod init nama-modul        # buat go.mod
go mod tidy                   # tambah yang dipakai, buang yang tidak
go get pkg@latest             # tambah/upgrade dependency
go get pkg@v1.2.3             # versi spesifik
go mod why pkg                # kenapa dependency ini ada
go mod verify                 # verifikasi integritas (go.sum)
go build -mod=readonly ./...  # gagal bila go.sum tak lengkap (cek CI)
```
> Catatan WSL repo ini: `go get` kadang `connection reset` (proxy) → coba ulang beberapa kali.

## Profiling & trace

```bash
go tool pprof http://host/debug/pprof/heap        # profil memori server hidup
go tool pprof cpu.prof                            # analisa (flamegraph: -http=:8080)
go tool pprof -http=:8080 cpu.prof
go tool trace trace.out                           # timeline scheduler/GC
go build -gcflags='-m' ./pkg                      # escape analysis (heap vs stack)
```
📍 `26-profiling`

## Build tags & kode kondisional

```go
//go:build linux && amd64      // file hanya utk platform/fitur ini
//go:build js && wasm          // target WASM (modul 47)
```

```go
import _ "embed"
//go:embed templates/*         // sematkan file ke biner
var files embed.FS

//go:generate stringer -type=Level  // dijalankan oleh `go generate ./...`
```

## Cross-compile & variabel runtime

```bash
GOOS=linux GOARCH=arm64 go build ./cmd/app   # lintas platform
CGO_ENABLED=0 go build ...                    # biner statik (distroless)
go build -ldflags="-s -w -X main.version=$(git describe)" ./cmd/app
```

| Env | Fungsi |
|-----|--------|
| `GOMAXPROCS` | jumlah P (paralelisme) |
| `GOGC` | agresivitas GC (default 100) |
| `GOMEMLIMIT` | batas memori lunak (Go 1.19+) |
| `GOFLAGS` | flag default (mis. `-mod=readonly`) |
| `GODEBUG=gctrace=1` | trace GC ke stderr |

## Dokumentasi

```bash
go doc fmt.Println            # doc paket/simbol
go doc -all ./pkg             # semua yang diekspor
```

## Linter (opsional, produksi)

```bash
golangci-lint run             # agregator linter populer (bukan bawaan Go)
```

## Makefile repo ini

```bash
make run MOD=01-basics        # jalankan modul
make advanced MOD=07-concurrency  # jalankan demo advanced/
make realcase-up MOD=22-caching   # nyalakan docker-compose real-case
make test-race                # semua test + race
make fmt vet tidy             # format, vet, rapikan modul
```
