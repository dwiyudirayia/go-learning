# 🗄️ Real-Case Tech Stack per Modul

Modul-modul di repo ini memakai **in-memory double** (miniredis, bufconn, NATS embedded, SQLite `:memory:`, map) supaya **jalan tanpa infra** dan test tetap hijau di mesin mana pun. Dokumen ini memetakan tiap double itu ke **tech stack produksi sungguhan**: apa yang dipakai di dunia nyata, kenapa, dan bagaimana menjalankannya betulan.

> 🎯 **Prinsip:** double dan stack nyata berbagi **kode interface yang sama**. Yang berganti hanya *driver/adapter*-nya. Itulah gunanya arsitektur berlapis + interface (modul 04, 29) — tukar implementasi tanpa menyentuh domain.

> 🧪 **Cara menjalankan versi nyata** mengikuti pola repo: client asli + **auto-skip bila env kosong** (`os.Getenv`) + `docker-compose` referensi. Contoh yang sudah ada: `23-message-queue/integration_test.go` (skip tanpa `RABBITMQ_URL`/`KAFKA_BROKERS`).

---

## Ringkasan cepat (peta double → produksi)

| Kebutuhan | Double di repo | Stack produksi umum |
|-----------|----------------|---------------------|
| Relational DB | SQLite (`modernc`) `:memory:` | **PostgreSQL** (utama) / MySQL, pool via **pgbouncer** |
| Cache / rate-limit / lock | **miniredis** | **Redis** (Cluster / Sentinel) / Memcached |
| Message queue / event bus | broker in-memory | **Kafka** (log/streaming) / **RabbitMQ** (routing) / **NATS JetStream** / SQS |
| gRPC transport | **bufconn** | gRPC + **Envoy/xDS**, service mesh (Istio/Linkerd) |
| Tracing | `tracetest` SpanRecorder | **OpenTelemetry** → **Jaeger/Tempo** (OTLP) |
| Metrics | in-proc registry | **Prometheus** + **Grafana** |
| Read-model / search | map / tabel SQL | **Elasticsearch/OpenSearch** / materialized view |
| Vector search (RAG) | slice + substring | **pgvector** / Qdrant / Weaviate / Pinecone |
| Orkestrasi | goroutine/timer | **Kubernetes**, **Temporal** (workflow/saga) |

---

## Fase 1–3 — Fundamental (01–10): **tanpa infra**

Modul **01–09** murni bahasa & runtime Go (slice, interface, error, generics, concurrency, testing, stdlib) — **tak ada stack infra**. Yang "produksi" di sini adalah *tooling*: `golangci-lint`, `go test -race` di CI, coverage gate.

**10 project-layout** → distribusi biner dgn **goreleaser**, modul privat via **GOPRIVATE**/Athens proxy. Tetap tanpa DB.

---

## Fase 4–6 — Backend & Microservices (11–17)

| Modul | Pola | Stack produksi | Kenapa |
|-------|------|----------------|--------|
| 11 CLI (Cobra) | CLI tool | Cobra + Viper; rilis via **goreleaser**; config di **Consul KV**/env | standar CLI Go |
| 12 net/http | HTTP server | di belakang **Nginx/Traefik** (TLS, gzip, LB) | terminasi TLS & routing di edge |
| 13 Fiber | HTTP framework | Fiber + **Redis** (session) + **PostgreSQL**; **Traefik** di depan | throughput tinggi |
| 14 database | `database/sql` + ORM | **PostgreSQL** (utama) / MySQL; pool via **pgbouncer**; replika baca | SQLite hanya untuk lokal/embedded |
| 15 REST+JWT | auth berlapis | **PostgreSQL** (user) + **Redis** (refresh token/blacklist) + reverse proxy | refresh-token butuh store cepat & bisa di-revoke |
| 16 gRPC | RPC | gRPC + **Envoy** (LB/xDS) + **buf** (schema registry/lint) | bufconn hanya untuk test |
| 17 microservices | order↔inventory | **Kubernetes** + service mesh (**Istio/Linkerd**), discovery **Consul/etcd** | mTLS, retry, LB di mesh |

---

## Fase 7 — Production-Readiness (18–21)

| Modul | Stack produksi | Catatan |
|-------|----------------|---------|
| 18 observability | **Prometheus** + **Grafana** + **Loki** (log) + **Tempo/Jaeger** (trace); **OTel Collector** sbagai pipeline | `slog` JSON → Loki; histogram → Prometheus |
| 19 config | **Viper** + **HashiCorp Vault** / **AWS Secrets Manager** / Consul KV | secret JANGAN di file; rotasi otomatis |
| 20 graceful shutdown | **Kubernetes** (SIGTERM + `preStop` + `terminationGracePeriodSeconds`) | kode `signal.NotifyContext` sudah produksi-ready |
| 21 migrations | **golang-migrate** + **PostgreSQL**; dijalankan di CI/CD (job terpisah, bukan saat boot) | strategi expand/contract untuk zero-downtime |

---

## Fase 8 — Skala & Integrasi (22–25)

| Modul | Double | Stack produksi | Kenapa |
|-------|--------|----------------|--------|
| **22 caching** | miniredis | **Redis** (Cluster untuk sharding, Sentinel untuk HA) / Memcached | `go-redis` yang sama; cukup ganti `Addr` |
| **23 message-queue** | broker in-mem | **Kafka** (event log, replay, high-throughput) · **RabbitMQ** (routing kompleks, per-message ack) · **NATS JetStream** (ringan, low-latency) · **SQS** (managed) | pilih per kebutuhan: log vs routing vs simple |
| 24 websocket | httptest | WS + **Redis Pub/Sub** (fan-out lintas instance) atau **Centrifugo** | broadcast butuh backplane antar-pod |
| 25 background-jobs | worker in-mem | **Asynq** (Redis) · **River** (Postgres) · **Temporal** (workflow durable) · Kafka consumer (stream) | job harus PERSIST & survive restart |

---

## Fase 9 — Kualitas/Keamanan/Deploy (26–30)

| Modul | Stack produksi |
|-------|----------------|
| 26 profiling | **Pyroscope/Parca** (continuous profiling) + `pprof`; benchmark di CI dgn `benchstat` |
| 27 security | **Vault** (secret) + **Redis** (rate-limit token bucket lintas pod) + WAF/reverse-proxy + OAuth provider |
| 28 api-docs | **OpenAPI** + **Swagger UI/Redoc**; codegen **oapi-codegen**; di-serve via API gateway |
| 29 clean-architecture | *(arsitektur, bukan infra)* — adapter repository nyata → **PostgreSQL**; sekunder → Redis/ES |
| 30 deployment | **Docker** (distroless) + **Kubernetes** + **Helm** + **ArgoCD** (GitOps); HPA + probes |

---

## Fase 10 — Distributed Systems (31–33)

| Modul | Stack produksi | Kenapa |
|-------|----------------|--------|
| **31 saga-outbox** | outbox di **PostgreSQL/MySQL** + **Debezium** (CDC) → **Kafka**; atau **Temporal** untuk orkestrasi saga | CDC menghindari polling; Temporal menangani state & retry saga |
| 32 resiliency | **Istio** (retry/timeout/circuit-break di mesh) atau library **sony/gobreaker**; distributed lock via **Redis (Redlock)** / **etcd** | mesh = tanpa ubah kode; lib = kontrol halus |
| 33 tracing | **OpenTelemetry SDK** → **OTLP** → **Jaeger/Tempo**; sampling di **OTel Collector** | `tracetest` hanya untuk test |

---

## Fase 11 — API & Data (34–36)

| Modul | Stack produksi |
|-------|----------------|
| 34 grpc-advanced | gRPC + **Envoy/xDS** (LB, retry, deadline) + **buf** registry; keepalive |
| 35 graphql | **gqlgen** + **DataLoader**; federation via **Apollo Router**; data di **PostgreSQL** |
| 36 sqlc | **PostgreSQL** + **pgx** + **sqlc** codegen; **pgbouncer**; `CopyFrom` untuk bulk |

---

## Fase 12 — Kualitas/Cloud (37–40)

| Modul | Stack produksi | Kenapa |
|-------|----------------|--------|
| 37 advanced-testing | **testcontainers** (Postgres/Kafka nyata di CI) + **k6** (load) + `go test -fuzz` | integration test pakai container asli, bukan mock |
| 38 concurrency-advanced | *(library `x/sync`, bukan infra)* | pola dipakai di dalam service apa pun |
| 39 cloud-native | **Kubernetes** + **controller-runtime/Operator SDK** + **Helm**; serverless via **Knative**/Lambda | reconcile loop = pola controller K8s |
| **40 llm-integration** | **Anthropic API** (`claude-opus-4-8`) + **vector DB**: **pgvector** / **Qdrant** / **Weaviate** / **Pinecone** untuk RAG; **Redis** untuk cache respons/embedding | retrieval nyata butuh index vektor, bukan substring |

---

## Fase 13–14 — Penguasaan & Spesialisasi (41–48)

| Modul | Stack produksi |
|-------|----------------|
| 41 capstone | **PostgreSQL** + **Redis** (cache-aside) + Fiber + **Kubernetes**; observability lengkap |
| 42 go-internals | *(runtime/tuning: `GOGC`, `GOMEMLIMIT`, pprof)* — tanpa infra |
| 43 advanced-generics | *(library)* — tanpa infra |
| **44 auth-advanced** | **Keycloak / Auth0 / Ory Hydra** (OIDC/OAuth2) + **OPA / Casbin** (policy engine) + **Redis** (session) + **PostgreSQL** | RBAC/ABAC nyata pakai policy engine, bukan if-else |
| **45 event-sourcing-cqrs** | **write**: **EventStoreDB** *atau* **PostgreSQL/MySQL** (tabel events) — **read**: **Elasticsearch** (query kaya/full-text/agregasi) — **bus**: **Kafka** | ⬇️ lihat deep-dive di bawah |
| 46 service-integrations | 3rd-party asli (**Stripe**, **AWS S3**, **SES**) + webhook via **API Gateway**; idempotency-key di **Redis/Postgres** |
| 47 tui-wasm | TUI (tanpa infra); **WASM** di-serve via **CDN**; biner kecil pakai **TinyGo** |
| 48 grpc-gateway | **grpc-gateway** + **Envoy**; API gateway (**Kong/Apigee**) untuk auth/rate-limit |

---

## 🔬 Deep-dive: CQRS real-case (Modul 45) — MySQL + Elasticsearch

Contoh yang kamu minta. Pola CQRS memisahkan **sisi tulis** dan **sisi baca**, masing-masing pakai penyimpanan yang paling cocok:

```
        COMMAND                                              QUERY
  (tulis, konsistensi)                                (baca, fleksibel/cepat)
          │                                                    ▲
          ▼                                                    │
  ┌───────────────┐   event    ┌──────────┐  index   ┌──────────────────┐
  │  MySQL/Postgres│──────────▶│  Kafka /  │─────────▶│  Elasticsearch    │
  │  (event store, │  (CDC via  │ Debezium) │ projector│  (read model:     │
  │   append-only) │  outbox)   └──────────┘  (consumer)│ full-text, agg)  │
  └───────────────┘                                    └──────────────────┘
```

**Kenapa stack ini:**
- **MySQL/PostgreSQL** untuk *write model* → transaksi ACID, constraint `UNIQUE(aggregate_id, version)` menegakkan **optimistic concurrency** (persis yang didemokan di `45-.../real-case/` versi SQLite — SQLite = stand-in lokal untuk MySQL).
- **Elasticsearch** untuk *read model* → query yang di SQL mahal jadi murah: full-text search, agregasi, faceting, sorting relevansi. Read-model itu **derivatif** (boleh dibuang & di-*rebuild* dari event) — cocok ditaruh di store yang dioptimalkan baca.
- **Kafka + Debezium (CDC)** → *projector* menyusul event baru dari MySQL secara real-time lalu meng-index ke ES, tanpa polling. Alternatif lebih sederhana: **transactional outbox** (modul 31) lalu relay ke ES.

**Konsistensi:** read-model ES *eventually consistent* (tertinggal milidetik–detik dari write). Rancang UI/idempotensi untuk itu.

**Menjalankan nyata (referensi `docker-compose`):**
```yaml
services:
  mysql:         { image: mysql:8,        environment: { MYSQL_ROOT_PASSWORD: secret } }
  elasticsearch: { image: elasticsearch:8.13.0, environment: { discovery.type: single-node, xpack.security.enabled: "false" } }
  kafka:         { image: bitnami/kafka:3.7 }
  connect:       { image: debezium/connect:2.6 }   # CDC MySQL -> Kafka
```
Kode Go: `go-sql-driver/mysql` (write) + `elastic/go-elasticsearch` (read) + `segmentio/kafka-go` (bus). Guard dgn `os.Getenv("MYSQL_DSN")`, `os.Getenv("ES_URL")` → **skip/print instruksi** bila kosong, agar `go build/test ./...` tetap hijau tanpa infra.

> Versi **runnable lokal tanpa infra** ada di [`45-event-sourcing-cqrs/real-case/`](45-event-sourcing-cqrs/real-case) — SQLite sebagai write store + tabel `account_read` sebagai read model. Konsepnya identik; di produksi tukar driver MySQL + read model ke Elasticsearch seperti di atas.

---

## Cara aku bisa lanjut

Untuk modul yang **client-nya sudah ada** di `go.mod` (Redis, Kafka, RabbitMQ, NATS), real-case bisa langsung **runnable dengan docker** (env-guarded, tanpa nambah dependency): mis. **22 caching→Redis**, **23 MQ→Kafka/RabbitMQ**, **25 jobs→Asynq/Redis**.

Untuk stack yang **perlu dependency baru** (MySQL, Elasticsearch, MongoDB, pgvector), aku tambah lewat `go get` + sediakan `docker-compose` + kode env-guarded. Sebut modul mana yang mau diprioritaskan.
