# Belajar Golang: Dari Dasar sampai Mahir 🚀

Kurikulum belajar Go **menyeluruh & bertahap** — dari fundamental, concurrency, CLI, backend/REST (Fiber), microservices/gRPC, production-readiness, distributed systems, integrasi LLM, sampai **penguasaan & spesialisasi**. **48 modul**, semua berisi kode yang bisa dijalankan, penjelasan konsep, studi kasus nyata, dan latihan.

> 🎯 Ditujukan untuk yang **sudah paham programming** (bahasa lain) dan ingin menguasai idiom khas Go dengan cepat. Setiap modul tetap dijelaskan dari nol per konsep.

> 📚 **Bingung mulai dari mana / cara belajarnya?** Baca **[LEARNING.md](LEARNING.md)** — panduan belajar: urutan, estimasi waktu, jalur sesuai tujuan, & cara belajar tiap modul.

> 🧭 **Rujukan cepat Go** ada di **[`docs/`](docs/)** — idiom, cheatsheet sintaks, jebakan umum (pitfalls), konkurensi, tooling, & glosarium. Buka sesuai kebutuhan sambil belajar (tak perlu dihafal).

> 📦 **Library yang sering dipakai** ada di **[`libraries/`](libraries/)** — katalog library Go populer (kapan pakai, alternatif, tautan ke modul) + **23 contoh runnable + test** (uuid, ulid, testify, go-cmp, zerolog, decimal, gjson, jwt, bcrypt, chi, sqlx, gorm, redis, prometheus, env, dll).

---

## ⚡ Mulai Cepat (3 langkah)

```bash
# 1. Pastikan Go terpasang (butuh 1.26+)
go version

# 2. Unduh semua dependency (otomatis dari go.mod)
go mod download

# 3. Jalankan modul pertama
go run ./01-basics
```

Lalu buka **`01-basics/README.md`**, baca, jalankan, kerjakan latihannya. Naik ke modul berikutnya. Selesai. 🎉

---

## 📦 Prasyarat & Setup

| Kebutuhan | Wajib? | Keterangan |
|-----------|--------|------------|
| **Go 1.26+** | ✅ wajib | `go version`. Unduh di [go.dev/dl](https://go.dev/dl) |
| Editor + Go extension | sangat disarankan | VS Code + `gopls`, atau GoLand |
| **Dependency** (Fiber, GORM, Cobra, gRPC, dll) | otomatis | sudah tercatat di `go.mod`/`go.sum`; `go run`/`go build` mengunduhnya sendiri |
| `protoc` (protobuf compiler) | opsional | **hanya** jika ingin **meng-generate ulang** kode `.proto` (Modul 16 & 17). Kode hasil generate (`*.pb.go`) **sudah disertakan**, jadi modul tetap jalan tanpa protoc |
| Docker | opsional | hanya untuk mencoba `docker compose` di Modul 17 |

> 💡 **Tidak perlu instalasi rumit.** Untuk 15 dari 17 modul cukup `go` saja. protoc & Docker hanya pelengkap di modul terakhir.

Regenerasi proto (opsional):
```bash
make tools   # pasang plugin protoc-gen-go (sekali saja)
make proto   # generate ulang *.pb.go  (perlu protoc terpasang)
```

---

## 🗂️ Struktur Tiap Modul (penting dibaca!)

Setiap folder `NN-nama/` mengikuti pola yang **sama** supaya mudah diikuti:

```
NN-nama/
├── README.md          ← 📖 MULAI DARI SINI: konsep + "Kapan Dipakai" + Latihan + "Teknik Advanced"
├── main.go            ← 💻 MATERI: contoh berkomentar, jalankan `go run ./NN-nama`
├── advanced/          ← 🚀 DEMO TEKNIK LANJUTAN (runnable, komentar detail): `go run ./NN-nama/advanced`
├── real-case/         ← 🗄️ IMPLEMENTASI STACK PRODUKSI (Postgres/Redis/Kafka/…) + docker-compose
├── latihan/solusi.go  ← ✅ KUNCI JAWABAN latihan (coba sendiri dulu!)
└── jawaban-saya/      ← ✍️ WORKSPACE-mu: kerjakan latihan di sini SEBELUM melihat kunci
```

- **`advanced/`** ada di **semua 48 modul** — contoh runnable tiap teknik di bagian "🚀 Teknik Advanced" README, dengan komentar penjelasan mendetail. (Modul 08 & 37 berupa `_test.go`: `go test ./NN-nama/advanced`.)
- **`real-case/`** ada di **24 modul yang relevan infra** — implementasi memakai tech stack produksi sungguhan (bukan in-memory), lengkap dengan `docker-compose.yml`. Bersifat *env-guarded*: otomatis mencetak panduan bila infra tak diset, jadi tetap aman di CI. **Peta stack produksi tiap modul ada di [`REAL-CASE-STACKS.md`](REAL-CASE-STACKS.md).**
- Modul backend (12–17) memakai **`_test.go`** sebagai verifikasi; jalankan `go test ./NN-nama/...`.
- Ingin latihan sendiri? Kerjakan di `NN-nama/jawaban-saya/` (template TODO sudah disediakan), lalu bandingkan dengan `latihan/solusi.go`.

---

## 🛠️ Perintah Berguna

Repo ini punya `Makefile` untuk mempermudah:

```bash
make help                 # daftar semua perintah
make run MOD=01-basics    # jalankan sebuah modul
make test                 # jalankan semua test
make test-race            # test + deteksi data race (penting utk modul 07)
make cover                # test + laporan coverage
make fmt                  # rapikan format semua kode
make vet                  # analisa statis
```

Tanpa `make` pun bisa langsung: `go run ./01-basics`, `go test ./...`, `go fmt ./...`, `go vet ./...`.

---

## 🗺️ Roadmap (peta belajar)

Ikuti **berurutan** — tiap modul membangun di atas yang sebelumnya. Centang `[x]` saat selesai.

### Fase 1 — Fundamental & Idiom Go
- [x] **01-basics** — package, variabel, tipe, konstanta & `iota`, kontrol alur, fungsi, multiple return, `defer`
- [x] **02-collections** — array, slice, map, string & rune
- [x] **03-structs-methods** — struct, method, pointer vs value, embedding
- [x] **04-interfaces** — interface, type assertion, type switch, `any`
- [x] **05-errors** — error handling idiomatis, wrapping, custom error, panic/recover
- [x] **06-generics** — type parameter, constraint

### Fase 2 — Concurrency (kekuatan utama Go)
- [x] **07-concurrency** — goroutine, channel, `select`, `sync`, `context`, pola-pola umum

### Fase 3 — Kualitas & Tooling
- [x] **08-testing** — unit test, table-driven, benchmark, coverage, mock
- [x] **09-stdlib** — `io`, `encoding/json`, `time`, `os`, `net/http` dasar
- [x] **10-project-layout** — struktur proyek Go, modules, Makefile

### Fase 4 — CLI & Tooling
- [x] **11-cli-cobra** — bangun CLI tool dengan Cobra

### Fase 5 — Backend / REST API
- [x] **12-http-stdlib** — REST pakai `net/http` murni
- [x] **13-fiber** — REST API dengan Fiber (routing, middleware, validasi)
- [x] **14-database** — `database/sql`, GORM, migrasi (SQLite pure-Go)
- [x] **15-studi-kasus-rest** — 📦 Studi kasus: REST API CRUD + Auth JWT lengkap (Fiber + GORM)

### Fase 6 — Microservices & gRPC
- [x] **16-grpc** — Protobuf, gRPC unary & streaming
- [x] **17-studi-kasus-microservices** — 📦 Studi kasus: 2 service (Fiber ↔ gRPC) + Docker

### Fase 7 — Production-Readiness
- [x] **18-observability** — structured logging (`slog`), metrics Prometheus, request middleware
- [x] **19-config** — Viper (file + env + flag), 12-factor, validasi config
- [x] **20-graceful-shutdown** — signal handling, `context` cancellation, connection draining
- [x] **21-migrations** — migrasi database berversi (embed SQL) + seeding

### Fase 8 — Skala & Integrasi
- [x] **22-caching** — Redis (go-redis), pola cache-aside, TTL, invalidation
- [x] **23-message-queue** — multi-broker (NATS/RabbitMQ/Kafka) + skema resiliensi (reconnect, retry, DLQ)
- [x] **24-websocket** — real-time: WebSocket & Server-Sent Events
- [x] **25-background-jobs** — worker queue, scheduler/cron, retry & idempotensi

### Fase 9 — Kualitas, Keamanan & Deployment
- [x] **26-profiling** — `pprof`, benchmark lanjut, optimasi alokasi
- [x] **27-security** — TLS/mTLS gRPC, rate limiting, security headers, refresh token
- [x] **28-api-docs** — OpenAPI/Swagger, versioning, format error konsisten
- [x] **29-clean-architecture** — ports & adapters (hexagonal), DDD ringan
- [x] **30-deployment** — Docker lanjut, manifest Kubernetes, CI/CD

### Fase 10 — Distributed Systems Patterns
- [x] **31-saga-outbox** — transaksi terdistribusi (saga), outbox pattern, idempotency
- [x] **32-resiliency-patterns** — circuit breaker, bulkhead, distributed lock, retry budget
- [x] **33-distributed-tracing** — OpenTelemetry (span lintas service)

### Fase 11 — API & Data Lanjutan
- [x] **34-grpc-advanced** — streaming (client/bidi), interceptor chain, deadline
- [x] **35-graphql** — gqlgen, resolver, DataLoader (atasi N+1)
- [x] **36-sqlc-advanced-db** — sqlc (SQL type-safe), connection pool, transaksi

### Fase 12 — Kualitas, Performa & Cloud-Native
- [x] **37-advanced-testing** — fuzzing, integration (testcontainers), load test
- [x] **38-concurrency-advanced** — errgroup, semaphore, singleflight, sync.Pool
- [x] **39-cloud-native** — Helm chart, K8s controller pattern, serverless
- [x] **40-llm-integration** — integrasi API Claude: tool use, streaming, RAG (dengan mock)

### Fase 13 — Penguasaan & Pendalaman
- [x] **41-capstone** — 📦 Studi kasus BESAR: URL shortener menggabungkan Fiber+DB+JWT+cache+observability+config+graceful shutdown
- [x] **42-go-internals** — scheduler (GMP), garbage collector, memory model, escape analysis, `go:` directives
- [x] **43-advanced-generics** — type sets, iterator (`iter.Seq`, range-over-func), struktur data generik, functional options

### Fase 14 — Spesialisasi
- [x] **44-auth-advanced** — OAuth2/OIDC flow, RBAC/ABAC, session, multi-tenancy
- [x] **45-event-sourcing-cqrs** — event store, aggregate, projection, read/write model terpisah
- [x] **46-service-integrations** — pola integrasi Stripe/S3/email (interface + mock + webhook)
- [x] **47-tui-wasm** — Bubble Tea (TUI) & Go → WebAssembly (browser)
- [x] **48-grpc-gateway** — satu proto → REST + gRPC sekaligus

---

## 💡 Cara Belajar yang Efektif

1. **Baca `README.md` modul dulu**, baru buka `main.go`. Konsep → kode.
2. **Jalankan** materinya (`go run ./NN-nama`) dan cocokkan output dengan penjelasan.
3. **Kerjakan latihan sendiri** di `jawaban-saya/` SEBELUM melihat `latihan/solusi.go`. Ini bagian terpenting.
4. **Biasakan** `go fmt`, `go vet`, dan `go test` — sama seperti kerja Go profesional.
5. Jangan lompat modul di Fase 1–2; fondasi ini dipakai terus di modul lanjutan.
6. Bagian **"Kapan & Di Mana Dipakai"** di tiap README menghubungkan konsep ke kasus nyata (backend/microservice) — jangan dilewati.

---

## 🧰 Library & Tools yang Dipelajari

| Modul | Library/Tool |
|-------|--------------|
| 11 | [Cobra](https://github.com/spf13/cobra) (CLI) |
| 13, 15, 17 | [Fiber](https://gofiber.io) v2 (web framework) + [validator](https://github.com/go-playground/validator) |
| 14, 15 | [GORM](https://gorm.io) + SQLite pure-Go ([glebarez](https://github.com/glebarez/sqlite)/[modernc](https://modernc.org/sqlite)) |
| 15 | [golang-jwt](https://github.com/golang-jwt/jwt) + bcrypt |
| 16, 17 | [gRPC](https://grpc.io) + Protocol Buffers |
| 18 | [Prometheus](https://github.com/prometheus/client_golang) + `log/slog` |
| 19 | [Viper](https://github.com/spf13/viper) |
| 21 | migrator `embed` (konsep golang-migrate/goose) |
| 22 | [go-redis](https://github.com/redis/go-redis) + [miniredis](https://github.com/alicebob/miniredis) |
| 23 | [NATS](https://nats.io) + [RabbitMQ](https://github.com/rabbitmq/amqp091-go) + [Kafka](https://github.com/segmentio/kafka-go) (multi-broker) |
| 24 | [coder/websocket](https://github.com/coder/websocket) + SSE |
| 26 | `pprof`, benchmark |
| 27 | rate limiter (`x/time/rate`), refresh token |
| 28 | OpenAPI / Swagger UI |
| 30 | Docker (distroless) + Kubernetes |
| 32 | circuit breaker, distributed lock (Redis) |
| 33 | [OpenTelemetry](https://opentelemetry.io) (tracing) |
| 34 | gRPC streaming + interceptor |
| 35 | [graphql-go](https://github.com/graphql-go/graphql) (GraphQL) |
| 36 | [sqlc](https://sqlc.dev) (SQL type-safe codegen) |
| 37 | fuzzing (native) + k6 (load test) |
| 38 | `golang.org/x/sync` (errgroup, singleflight, semaphore) |
| 39 | Helm + K8s controller pattern |
| 40 | [anthropic-sdk-go](https://github.com/anthropics/anthropic-sdk-go) (Claude API) |
| 41 | 📦 Capstone (Fiber+SQLite+JWT+Redis+shutdown) |
| 42 | runtime, `unsafe`, escape analysis |
| 43 | generics lanjut, `iter.Seq` (Go 1.23) |
| 44 | RBAC/ABAC + OAuth2 (crypto/hmac) |
| 45 | event sourcing & CQRS |
| 46 | pola integrasi + webhook HMAC |
| 47 | [Bubble Tea](https://github.com/charmbracelet/bubbletea) (TUI) + WASM |
| 48 | gRPC + REST gateway |

Semua versi terkunci di `go.mod`/`go.sum`.

---

## 📄 Lisensi
[MIT](LICENSE) — bebas dipakai, dimodifikasi, dan dibagikan. Silakan ganti nama pemegang hak cipta di file `LICENSE`.

Selamat belajar! Kalau menemukan hal membingungkan, itu wajar — Go butuh sedikit pembiasaan, tapi setelah "klik" akan terasa sangat rapi. 🐹
