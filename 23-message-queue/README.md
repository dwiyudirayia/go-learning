# 23 — Message Queue: Multi-Broker + Skema Resiliensi

Arsitektur **event-driven** (service mengumumkan event, service lain bereaksi) yang **loosely coupled**. Modul ini mendukung **banyak broker** lewat satu abstraksi, plus **skema menangani koneksi putus** (reconnect), retry, dan jaminan pengiriman.

Jalankan (default in-memory, tanpa server):
```bash
go run ./23-message-queue                 # in-memory + demo resiliensi
BROKER=nats go run ./23-message-queue     # NATS embedded
```
Verifikasi otomatis: `go test ./23-message-queue`

---

## 1. Abstraksi `Broker` (agar bebas ganti broker)

Kode aplikasi hanya kenal interface `Broker` — tak terikat NATS/RabbitMQ/Kafka:
```go
type Broker interface {
    Publish(ctx, topic string, data []byte) error
    Subscribe(ctx, topic, group string, handler Handler) error // group="" broadcast; group!="" load-balance
    Close() error
}
```
Ganti broker = ganti **satu baris** di composition root. Ini dependency inversion (Modul 4 & 29).

| Broker | File | Butuh server? | Resiliensi |
|--------|------|---------------|-----------|
| **In-Memory** | `inmemory.go` | tidak | bisa simulasi putus (untuk test) |
| **NATS** | `nats_broker.go` | ya (atau embedded) | **auto-reconnect built-in** client |
| **RabbitMQ** | `rabbitmq_broker.go` | ya | **reconnect manual** (NotifyClose) |
| **Kafka** | `kafka_broker.go` | ya | reconnect internal Reader/Writer |

Pilih via env `BROKER` (lihat komentar di `main.go`).

---

## 2. 🔌 Skema Menangani Koneksi Putus (inti permintaan)

Broker/jaringan **pasti** akan putus (deploy, restart, network blip). Sistem harus **pulih sendiri**. Tiga lapis pertahanan di modul ini:

### a. Reconnect KONSUMEN — `superviseConsumer` (`resilience.go`)
Supervisor yang terus menyambung ulang saat konsumen terputus:
```go
for {
    err := connectAndServe(ctx) // sambung + proses pesan sampai putus
    if ctx.Err() != nil { return }  // shutdown normal -> berhenti
    time.Sleep(backoff(cfg, attempt)) // tunggu (eksponensial + jitter)
    // ulangi -> sambung lagi
}
```
- **RabbitMQ**: memantau `conn.NotifyClose` → keluar loop → supervisor menyambung ulang & mendeklarasi ulang queue/binding.
- **Kafka**: `ReadMessage` error → supervisor bikin Reader baru.
- **NATS**: tak perlu supervisor — client-nya auto-reconnect sendiri (`MaxReconnects(-1)`).

### b. Retry PUBLISH — `PublishWithRetry` (`resilience.go`)
Publish yang gagal (broker sedang down) **diulang** dengan backoff sampai sukses:
```go
PublishWithRetry(ctx, broker, topic, data, logger) // retry sampai terkirim / ctx habis
```
Demo membuktikan: broker "mati" → publisher retry → broker "hidup" → **pesan tidak hilang**.

### c. Exponential Backoff + Jitter — `backoff()`
Jeda antar percobaan **bertambah** (100ms → 200ms → 400ms ...) dan diberi **jitter** acak → mencegah semua client menyerbu broker serempak saat pulih ("thundering herd").

---

## 3. ⚙️ Jaminan Pengiriman & Konsep Penting Lainnya

### Ack / Nack (at-least-once)
Konsumen meng-**ack** setelah SUKSES memproses. Bila handler gagal (return error) → **nack + requeue** → diproses ulang.
```go
if err := handler(ctx, d.Body); err != nil { d.Nack(false, true) /*requeue*/ } else { d.Ack(false) }
```
Konsekuensi: pesan bisa terkirim **>1 kali** → handler wajib **idempotent** (Modul 25).

### Delivery semantics
| Jaminan | Arti | Cara |
|---------|------|------|
| **At-most-once** | maksimal 1× (bisa hilang) | ack sebelum proses |
| **At-least-once** | minimal 1× (bisa ganda) | ack SETELAH proses (modul ini) |
| **Exactly-once** | tepat 1× | idempotensi + dedup / transaksi (mahal) |

### Dead Letter Queue (DLQ)
Pesan yang gagal terus (retry habis) dipindah ke queue khusus untuk investigasi — bukan dibuang/di-loop selamanya. (RabbitMQ: `x-dead-letter-exchange`; lihat latihan.)

### Durability & ordering
- **Persistent messages** (RabbitMQ `DeliveryMode: Persistent`, Kafka `RequireAll`) → tak hilang saat broker restart.
- **Urutan**: Kafka menjamin urutan per-partisi; NATS/RabbitMQ umumnya tidak lintas konsumen.

### Backpressure
Batasi konsumen memproses terlalu banyak sekaligus: RabbitMQ `Qos(1,...)` (fair dispatch), Kafka via ukuran fetch. Cegah konsumen kehabisan memori.

---

## 4. Menjalankan dengan Broker Sungguhan (Docker)

```bash
# RabbitMQ
docker run -d --rm -p 5672:5672 -p 15672:15672 rabbitmq:3-management
BROKER=rabbitmq RABBITMQ_URL=amqp://guest:guest@localhost:5672/ go run ./23-message-queue

# Kafka
docker run -d --rm -p 9092:9092 apache/kafka:latest
BROKER=kafka KAFKA_BROKERS=localhost:9092 go run ./23-message-queue
```

### Integration test (otomatis di-skip tanpa env)
```bash
RABBITMQ_URL=amqp://guest:guest@localhost:5672/ go test -run TestRabbitMQ ./23-message-queue
KAFKA_BROKERS=localhost:9092                    go test -run TestKafka    ./23-message-queue
```
Tanpa env, kedua test **SKIP** (CI tetap hijau). In-memory & NATS diuji penuh.

---

## Broker mana yang dipilih?
| | NATS | RabbitMQ | Kafka |
|-|------|----------|-------|
| Sifat | ringan, cepat | routing kaya, matang | log terurut, throughput sangat tinggi |
| Cocok | pub/sub, RPC, microservice | work queue, routing kompleks | event streaming, event sourcing, analytics |
| Persistensi | JetStream | ya (queue) | ya (retensi lama) |

Karena kode pakai interface `Broker`, kamu bisa mulai dengan yang simpel (NATS) dan pindah tanpa mengubah logika bisnis.

## Kapan & Di Mana Dipakai
- Notifikasi async, sinkronisasi antar service, meredam lonjakan (buffer), event sourcing, audit log, pipeline data.

## Latihan
1. Implementasikan `Broker` untuk Redis Pub/Sub (Modul 22) atau Redis Streams.
2. Tambah **DLQ**: setelah N nack, publish pesan ke `topic.failed`.
3. Tambah metrik (Modul 18): jumlah publish/ack/nack/reconnect.
4. Tambah header/metadata pesan (mis. `trace_id`) di seluruh broker.
5. Jalankan RabbitMQ via Docker, matikan container saat konsumen jalan, hidupkan lagi → amati reconnect otomatis.

## ✅ Solusi Latihan (Pembahasan)

1. **Broker Redis** — implementasikan interface `Broker` dengan Redis: `Publish` = `rdb.Publish(ctx, topic, data)`; `Subscribe` = `rdb.Subscribe(ctx, topic).Channel()`. Untuk durability pakai **Redis Streams** (`XADD`/`XREADGROUP`) yang mendukung consumer group + ack.
2. **DLQ setelah N nack** — hitung percobaan per pesan (header `x-retries`); bila `>= N`, alih-alih nack lagi, `Publish(topic+".failed", msg)` lalu ack pesan asli agar tak loop selamanya.
3. **Metrik** — 4 Counter (Modul 18): `publish_total`, `ack_total`, `nack_total`, `reconnect_total`. Naikkan di titik masing-masing pada `resilience.go`.
4. **Header/metadata** — bungkus payload jadi `struct{ Meta map[string]string; Body []byte }` (mis. `Meta["trace_id"]`). Semua broker meng-encode struct ini (JSON) sehingga trace_id ikut lintas service (korelasi dengan Modul 33).
5. **Uji reconnect** — jalankan RabbitMQ via Docker, mulai konsumen, `docker stop rabbitmq` saat jalan → `superviseConsumer` masuk loop reconnect (backoff+jitter); `docker start` → konsumen tersambung lagi otomatis tanpa kehilangan langganan.
