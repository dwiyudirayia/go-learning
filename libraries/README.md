# 📦 libraries/ — Katalog Library Go yang Sering Dipakai

Kumpulan **library pihak ketiga** yang lazim dipakai programmer Go di dunia nyata, lengkap dengan **kapan pakai**, **alternatif**, dan **di modul mana** ia sudah muncul di kurikulum ini.

Folder ini **di luar** urutan modul `01`–`48` (modul dikunci di 48). Anggap ini **etalase**: satu tempat untuk menjawab "library apa yang biasa dipakai orang Go, dan kenapa?".

## Cara pakai folder ini

- **Baca katalog di bawah** untuk peta besarnya — kategori, fungsi, kapan dipakai.
- **16 library** punya contoh **runnable + test** di subfolder. Jalankan & baca kodenya:

  ```bash
  go run ./libraries/uuid          # jalankan demo
  go test ./libraries/uuid         # jalankan test (inti pelajarannya sering di sini)
  go test ./libraries/...          # semua contoh sekaligus
  ```

- Tiap file berkomentar **`// 🔍 Analogi:`** seperti materi modul — konsep dijelaskan dari nol untuk orang awam.
- Library yang **sudah dipakai modul** cukup ditautkan ke modulnya (kolom "Dipakai di modul") agar tak ada duplikasi materi.

> **Prinsip memilih library di Go:** komunitas Go menghargai **standard library** dan **dependensi minimal**. Sebelum menambah library, tanyakan: "apakah `net/http`, `log/slog`, `slices`, `database/sql` sudah cukup?" Sering kali cukup. Library di bawah dipakai saat stdlib benar-benar merepotkan atau saat butuh fitur yang tak ada bawaannya. Tiap contoh runnable memuat bagian **"kapan JANGAN pakai"**.

Legenda: **⭐ = ada contoh runnable** di `libraries/<nama>/`.

---

## 1. CLI & Konfigurasi

| Library | Fungsi | Kapan pakai | Alternatif | Dipakai di modul |
|---------|--------|-------------|------------|------------------|
| ⭐ **spf13/cobra** | Kerangka aplikasi CLI (perintah, sub-perintah, flag, bantuan otomatis) | CLI dengan banyak perintah bergaya `git`/`kubectl` | `flag` (stdlib, untuk CLI sederhana), `urfave/cli` | [11](../11-cli-cobra/) |
| ⭐ **spf13/viper** | Konfigurasi berlapis: default → file → env | Aplikasi dengan banyak pengaturan dari berbagai sumber | `env` + `flag` (stdlib), `envconfig`, `koanf` | [19](../19-config/) |
| ⭐ **caarlos0/env** | Konfig dari environment → struct lewat tag | Aplikasi 12-factor/cloud-native (HANYA env, tanpa file) | Viper (kalau butuh berlapis), `envconfig` | — |
| **spf13/pflag** | Flag gaya POSIX (`--nama`, `-n`) | Otomatis terpasang bersama cobra | `flag` (stdlib) | (transitif via cobra) |

## 2. Web & HTTP

| Library | Fungsi | Kapan pakai | Alternatif | Dipakai di modul |
|---------|--------|-------------|------------|------------------|
| **gofiber/fiber** | Web framework cepat (routing, middleware, parsing) | API throughput tinggi; sintaks ringkas ala Express | `net/http` + `chi`, `gin`, `echo` | [13](../13-fiber/), [15](../15-studi-kasus-rest/), [17](../17-studi-kasus-microservices/) |
| ⭐ **go-chi/chi** | Router HTTP idiomatik (param URL, grup, middleware) | Butuh lebih dari ServeMux tapi ingin tetap dekat net/http | `net/http` ServeMux (Go 1.22+), `gorilla/mux` | — |
| **net/http** | Server & client HTTP bawaan | **Default.** Cukup untuk kebanyakan kebutuhan, nol dependensi | — | [12](../12-http-stdlib/) |
| ⭐ **go-resty/resty** | HTTP **client** enak pakai (retry, timeout, JSON, middleware) | Aplikasi yang memanggil banyak API eksternal | `net/http` (stdlib) | — |
| ⭐ **go-playground/validator** | Validasi struct lewat tag | Memvalidasi input di batas luar (handler HTTP) | validasi manual, `ozzo-validation` | [13](../13-fiber/) |
| **coder/websocket** | WebSocket minimalis & modern | Komunikasi dua arah real-time | `gorilla/websocket`, `nhooyr/websocket` | [24](../24-websocket/) |
| **graphql-go/graphql** | Server GraphQL | Klien butuh memilih sendiri field yang diambil | `gqlgen` (schema-first, lebih populer) | [35](../35-graphql/) |

## 3. gRPC & Protobuf

| Library | Fungsi | Kapan pakai | Alternatif | Dipakai di modul |
|---------|--------|-------------|------------|------------------|
| **google.golang.org/grpc** | RPC lintas layanan berbasis HTTP/2 | Komunikasi antar-microservice, kontrak ketat | REST/JSON, Connect, tRPC | [16](../16-grpc/), [17](../17-studi-kasus-microservices/), [34](../34-grpc-advanced/), [48](../48-grpc-gateway/) |
| **google.golang.org/protobuf** | Serialisasi Protocol Buffers | Format pesan gRPC; data biner ringkas | JSON, MessagePack, Avro | (bersama grpc) |

## 4. Database & ORM

| Library | Fungsi | Kapan pakai | Alternatif | Dipakai di modul |
|---------|--------|-------------|------------|------------------|
| ⭐ **gorm.io/gorm** | ORM: struct ↔ tabel, relasi, migrasi | 90% CRUD yang membosankan; prototipe cepat | `sqlc`, `sqlx`, `database/sql` mentah | [14](../14-database/), [15](../15-studi-kasus-rest/) |
| ⭐ **jmoiron/sqlx** | Perpanjangan `database/sql`: scan ke struct | Ingin SQL mentah tapi malas menulis `rows.Scan` berulang | GORM, `sqlc`, `database/sql` | — |
| **database/sql** | Antarmuka SQL bawaan | Kendali penuh atas query; nol dependensi | — | [14](../14-database/) |
| **sqlc** | Generate kode Go type-safe dari SQL | Ingin SQL mentah TAPI aman di waktu kompilasi | GORM, `sqlx` | [36](../36-sqlc-advanced-db/) |
| **jackc/pgx** | Driver & toolkit PostgreSQL | Aplikasi khusus Postgres (fitur & performa penuh) | `lib/pq` (mode maintenance) | [14](../14-database/), [29](../29-clean-architecture/), [36](../36-sqlc-advanced-db/) |
| **modernc.org/sqlite** | SQLite **pure-Go** (tanpa cgo) | SQLite tanpa ribet toolchain C; test & embedded | `mattn/go-sqlite3` (butuh cgo) | [14](../14-database/), [21](../21-migrations/), [36](../36-sqlc-advanced-db/), [45](../45-event-sourcing-cqrs/) |
| **glebarez/sqlite** | Driver SQLite pure-Go untuk GORM | Memakai GORM tanpa cgo | `gorm.io/driver/sqlite` (cgo) | [14](../14-database/), [15](../15-studi-kasus-rest/) |
| **go-sql-driver/mysql** | Driver MySQL | Aplikasi berbasis MySQL/MariaDB | — | [45](../45-event-sourcing-cqrs/) |
| **golang-migrate/migrate** | Migrasi skema database bernomor | Perubahan skema terkontrol & bisa rollback (produksi) | `goose`, `atlas`, GORM AutoMigrate (dev saja) | [15](../15-studi-kasus-rest/), [21](../21-migrations/), [29](../29-clean-architecture/) |

## 5. Cache & Message Queue

| Library | Fungsi | Kapan pakai | Alternatif | Dipakai di modul |
|---------|--------|-------------|------------|------------------|
| ⭐ **redis/go-redis** | Klien Redis | Cache, penghitung atomik, rate limit, sesi | `valkey-go`, `rueidis` | [22](../22-caching/), [24](../24-websocket/), [25](../25-background-jobs/), [27](../27-security/), [32](../32-resiliency-patterns/) |
| **alicebob/miniredis** | Redis palsu in-memory (untuk test) | Menguji kode Redis tanpa server sungguhan | `testcontainers`, `dockertest` | [22](../22-caching/), [32](../32-resiliency-patterns/), [41](../41-capstone/) |
| **nats-io/nats.go** | Klien NATS (pub/sub, JetStream) | Messaging ringan & cepat antar-layanan | RabbitMQ, Kafka, Redis Streams | [23](../23-message-queue/) |
| **rabbitmq/amqp091-go** | Klien RabbitMQ (AMQP) | Antrean kerja dengan routing & jaminan kuat | NATS, Kafka | [23](../23-message-queue/) |
| **segmentio/kafka-go** | Klien Apache Kafka | Log peristiwa throughput tinggi, event streaming | NATS JetStream, Redpanda | [23](../23-message-queue/), [31](../31-saga-outbox/) |

## 6. Observability

| Library | Fungsi | Kapan pakai | Alternatif | Dipakai di modul |
|---------|--------|-------------|------------|------------------|
| ⭐ **prometheus/client_golang** | Metrik (counter, gauge, histogram) + endpoint `/metrics` | Memantau kesehatan & performa layanan | OpenTelemetry Metrics, StatsD | [18](../18-observability/) |
| **go.opentelemetry.io/otel** | Distributed tracing (span, propagasi) | Melacak satu request menembus banyak layanan | Jaeger client, Zipkin | [33](../33-distributed-tracing/) |
| ⭐ **rs/zerolog** | Logging terstruktur (JSON) nyaris nol-alokasi | Log throughput tinggi yang bisa difilter mesin | `log/slog` (stdlib), `zap`, `logrus` | — |
| **log/slog** | Logging terstruktur bawaan (Go 1.21+) | **Default** untuk proyek baru; nol dependensi | zerolog, zap | [18](../18-observability/) |

## 7. Autentikasi & Keamanan

| Library | Fungsi | Kapan pakai | Alternatif | Dipakai di modul |
|---------|--------|-------------|------------|------------------|
| ⭐ **golang-jwt/jwt** | Menerbitkan & memverifikasi JWT | Autentikasi stateless (access + refresh token) | sesi berbasis cookie, PASETO | [15](../15-studi-kasus-rest/), [27](../27-security/), [44](../44-auth-advanced/) |
| ⭐ **x/crypto/bcrypt** | Hash kata sandi (searah, bergaram, lambat) | **Menyimpan kata sandi** (jangan pernah plaintext/MD5/SHA) | `argon2`, `scrypt` (juga di x/crypto) | [41](../41-capstone/) |
| **golang.org/x/crypto** | Kriptografi tambahan (bcrypt, argon2, dll) | Hash kata sandi, kriptografi di luar stdlib | `crypto/*` (stdlib) | [15](../15-studi-kasus-rest/), [41](../41-capstone/) |
| ⭐ **golang.org/x/time/rate** | Rate limiting (token bucket) | Membatasi laju request per pengguna/IP | implementasi manual, `uber-go/ratelimit` | [27](../27-security/) |
| **standard-webhooks** | Verifikasi webhook (HMAC, anti-replay) | Menerima webhook dari layanan pihak ketiga | verifikasi HMAC manual | [46](../46-service-integrations/) |

## 8. Concurrency & Utility

| Library | Fungsi | Kapan pakai | Alternatif | Dipakai di modul |
|---------|--------|-------------|------------|------------------|
| ⭐ **golang.org/x/sync** | `errgroup`, `singleflight`, `semaphore` | Goroutine paralel dengan error & batas konkurensi | `sync.WaitGroup` (stdlib, lebih dasar) | [22](../22-caching/), [38](../38-concurrency-advanced/) |
| ⭐ **google/uuid** | Membuat & mengurai UUID (v4, v7) | ID unik global tanpa koordinasi database | `oklog/ulid`, `rs/xid`, auto-increment | (transitif; dipromosikan di sini) |
| ⭐ **oklog/ulid** | ID unik urut-waktu, ringkas & ramah-URL | ID urut waktu yang lebih pendek dari UUID (26 huruf) | UUIDv7 (`google/uuid`), `rs/xid` | — |
| ⭐ **samber/lo** | Helper generik (Map, Filter, GroupBy, ...) | Transformasi koleksi bergaya fungsional | `slices`/`maps` (stdlib), for-loop biasa | — |
| ⭐ **shopspring/decimal** | Angka desimal presisi tetap | **Uang** & perhitungan finansial (jangan `float64`!) | `int64` satuan sen, `math/big.Rat` | — |
| ⭐ **robfig/cron** | Penjadwalan tugas berulang di dalam proses | Cron job in-app (bersih cache, laporan harian) | `time.Ticker` (stdlib), Kubernetes CronJob | — |
| ⭐ **tidwall/gjson + sjson** | Baca/tulis JSON **tanpa struct** (path bertitik) | JSON dinamis/tak dikenal (webhook, config, ambil 2-3 field) | `encoding/json` (stdlib, untuk bentuk tetap) | — |

## 9. Testing

| Library | Fungsi | Kapan pakai | Alternatif | Dipakai di modul |
|---------|--------|-------------|------------|------------------|
| ⭐ **stretchr/testify** | Assertion, mock, suite | Test lebih ringkas; tim terbiasa dengannya | `testing` bawaan (banyak proyek besar cukup ini) | — |
| ⭐ **google/go-cmp** | Perbandingan mendalam + **laporan diff** | Membandingkan struct rumit di test (diff yang jelas) | `reflect.DeepEqual`, `testify` | — |
| **testing** | Framework test bawaan | **Default.** Table-driven test idiomatik Go | testify | (semua modul) |
| **net/http/httptest** | Server & recorder HTTP untuk test | Menguji handler tanpa jaringan sungguhan | — | [09](../09-stdlib/), [12](../12-http-stdlib/) |

## 10. TUI & Frontend

| Library | Fungsi | Kapan pakai | Alternatif | Dipakai di modul |
|---------|--------|-------------|------------|------------------|
| **charmbracelet/bubbletea** | Framework TUI (arsitektur ELM) | Aplikasi terminal interaktif | `tview`, `termui` | [47](../47-tui-wasm/) |

## 11. AI / LLM

| Library | Fungsi | Kapan pakai | Alternatif | Dipakai di modul |
|---------|--------|-------------|------------|------------------|
| **anthropics/anthropic-sdk-go** | SDK resmi Claude API | Integrasi LLM (chat, streaming, tool use) | panggilan HTTP manual | [40](../40-llm-integration/) |

## 12. Dev Tooling (bukan import — dipasang terpisah)

Perkakas ini **bukan** dependensi kode, melainkan program yang kamu jalankan di terminal. Tak ada di `go.mod`, tapi bagian tak terpisahkan dari alur kerja Go profesional.

| Alat | Fungsi | Perintah |
|------|--------|----------|
| **gofmt / goimports** | Format kode otomatis (bawaan) | `gofmt -w .`, `goimports -w .` |
| **go vet** | Analisis statis bawaan | `go vet ./...` |
| **golangci-lint** | Meta-linter (menggabungkan puluhan linter) | `golangci-lint run` |
| **staticcheck** | Analisis statis mendalam | `staticcheck ./...` |
| **air** | Live-reload saat pengembangan | `air` |
| **mockery / moq** | Generate mock dari interface | `mockery --all` |
| **sqlc** | Generate kode dari SQL | `sqlc generate` |
| **golang-migrate** | CLI migrasi database | `migrate -path ... up` |
| **govulncheck** | Pindai kerentanan keamanan | `govulncheck ./...` |
| **dlv (delve)** | Debugger Go | `dlv debug` |

Lihat [`../docs/TOOLING.md`](../docs/TOOLING.md) untuk detail toolchain.

---

## Daftar contoh runnable (23)

| Folder | Library | Sorotan pelajaran |
|--------|---------|-------------------|
| [`uuid/`](uuid/) | google/uuid | v4 vs **v7** (urut waktu, ramah index DB), jebakan `uuid.Nil` |
| [`ulid/`](ulid/) | oklog/ulid | ID urut-waktu ringkas, baca waktu dari ID, monotonic, `Parse` vs `ParseStrict` |
| [`testify/`](testify/) | stretchr/testify | `assert` vs `require`, mock, suite — & kapan cukup `testing` bawaan |
| [`gocmp/`](gocmp/) | google/go-cmp | diff yang jelas, IgnoreFields, AllowUnexported, toleransi float |
| [`zerolog/`](zerolog/) | rs/zerolog | log terstruktur bisa **diuji**, jebakan lupa `.Msg()`, vs `slog` |
| [`lo/`](lo/) | samber/lo | Map/Filter/Reduce/GroupBy — & **kapan for-loop lebih baik** |
| [`decimal/`](decimal/) | shopspring/decimal | kenapa `float64` **haram** untuk uang, banker's rounding, bagi rata tanpa sen hilang |
| [`gjson/`](gjson/) | tidwall/gjson+sjson | baca/tulis JSON tanpa struct, query & filter array, jebakan `Valid()` |
| [`cron/`](cron/) | robfig/cron | format spec, zona waktu, jebakan **banyak replika** |
| [`resty/`](resty/) | go-resty/resty | error transport vs status HTTP, **retry hanya yang pantas** |
| [`validator/`](validator/) | go-playground/validator | tag, aturan buatan sendiri, lintas-field, jebakan `required` vs nol |
| [`cobra/`](cobra/) | spf13/cobra | pohon perintah, flag persisten, **CLI yang bisa diuji** |
| [`viper/`](viper/) | spf13/viper | urutan lapisan, jebakan tag `mapstructure` vs `json` |
| [`env/`](env/) | caarlos0/env | env→struct, konversi tipe, `required` vs `notEmpty`, nilai kosong→default |
| [`jwt/`](jwt/) | golang-jwt/jwt | access+refresh, **serangan alg:none**, isi token tidak rahasia |
| [`bcrypt/`](bcrypt/) | x/crypto/bcrypt | hash searah+garam, cost & rehash saat login, jebakan batas 72 byte |
| [`chi/`](chi/) | go-chi/chi | param URL, grup middleware, 404 vs 405 — tetap `http.Handler` biasa |
| [`sqlx/`](sqlx/) | jmoiron/sqlx | Get/Select ke struct, named query, `sqlx.In`, transaksi |
| [`errgroup/`](errgroup/) | x/sync | errgroup, **singleflight** (cache stampede), semaphore berbobot |
| [`ratelimit/`](ratelimit/) | x/time/rate | token bucket, pembatas per-IP + pembersihan (anti bocor memori) |
| [`redis/`](redis/) | go-redis + miniredis | cache-aside, jebakan `redis.Nil`, TTL, pipeline |
| [`gorm/`](gorm/) | gorm.io/gorm | jebakan update nilai nol, soft delete, **N+1 & Preload** |
| [`prometheus/`](prometheus/) | client_golang | counter/gauge/histogram, **jebakan kardinalitas label** |

---

## Verifikasi

```bash
go test ./libraries/...     # semua contoh + test-nya harus lulus
gofmt -l ./libraries/       # harus kosong
go vet ./libraries/...      # harus bersih
```

> Semua contoh dirancang **berjalan tanpa infrastruktur eksternal**: Redis dipalsukan `miniredis`, database pakai SQLite in-memory, API dipalsukan `httptest`. Cukup `go run` / `go test`.
