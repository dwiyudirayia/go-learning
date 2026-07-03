# 33 — Distributed Tracing (OpenTelemetry)

Pilar ketiga observability (setelah logs & metrics, Modul 18). **Tracing** melacak **satu request** yang melintasi banyak fungsi/service, menampilkan tiap langkah sebagai **span** dalam pohon — sehingga kamu bisa melihat **di mana** waktu terbuang.

Jalankan:
```bash
go run ./33-distributed-tracing
```
Verifikasi otomatis: `go test ./33-distributed-tracing`

Output:
```
trace untuk order 42:
  HandleOrder (9ms)
    validateOrder (2ms)
    chargePayment (5ms)   <- langkah paling lambat, langsung terlihat
    saveToDatabase (1ms)
```

## Konsep

### Trace & Span
- **Trace** = seluruh perjalanan satu request (punya `TraceID` unik).
- **Span** = satu operasi dalam trace (nama, waktu mulai/selesai, atribut, status). Punya **parent** → membentuk pohon.

```go
ctx, span := tracer().Start(ctx, "HandleOrder")   // buka span
defer span.End()                                   // tutup (durasi terhitung)
span.SetAttributes(attribute.Int("order.id", 42))  // metadata untuk filter
span.SetStatus(codes.Error, "pesan")               // tandai gagal
```

### Context propagation = kunci parent-child
Span anak menjadi anak karena `ctx` yang dibawa mengandung span induk:
```go
ctx, parent := tracer().Start(ctx, "HandleOrder") // induk
child := chargePayment(ctx)                        // ctx bawa induk -> child jadi anak
```
Test membuktikan `chargePayment` adalah **anak** `HandleOrder`, dan semua span punya **TraceID sama**.

### Lintas service (yang bikin "distributed")
Antar service, konteks trace dikirim lewat **header HTTP `traceparent`** (W3C Trace Context). `otel.SetTextMapPropagator(propagation.TraceContext{})` mengaturnya. Sehingga span di `order-service` (Modul 17) dan `inventory-service` masuk **satu trace** — kamu lihat seluruh alur lintas service di satu layar.

## Setup Exporter (produksi)

Modul ini pakai **in-memory recorder** untuk demo. Di produksi, ganti span processor dengan **exporter**:
```go
// OTLP -> Jaeger / Tempo / Grafana / Datadog
import "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
exp, _ := otlptracegrpc.New(ctx, otlptracegrpc.WithEndpoint("localhost:4317"), otlptracegrpc.WithInsecure())
tp := sdktrace.NewTracerProvider(sdktrace.WithBatcher(exp))
```
Lalu jalankan Jaeger:
```bash
docker run -d -p 16686:16686 -p 4317:4317 jaegertracing/all-in-one
# buka http://localhost:16686 untuk melihat trace
```
Instrumentasi otomatis tersedia: `otelhttp` (HTTP), `otelgrpc` (gRPC), `otelgorm` (DB) — bungkus handler/client dan span dibuat otomatis.

## Sampling
Merekam **semua** trace mahal di traffic tinggi. Gunakan **sampling** (mis. 1%): `sdktrace.WithSampler(sdktrace.TraceIDRatioBased(0.01))`.

## Tiga pilar observability (lengkap)
| Pilar | Menjawab | Modul |
|-------|----------|-------|
| **Logs** | apa yang terjadi (detail) | 18 |
| **Metrics** | berapa & tren (agregat) | 18 |
| **Traces** | di mana lambat (alur 1 request) | ini |

Idealnya ketiganya **berkorelasi** (log menyertakan `trace_id`).

## Kapan & Di Mana Dipakai
- Microservices (Modul 17): temukan service lambat di rantai panggilan.
- Debug latensi p99, N+1 query, panggilan eksternal lambat.

## Latihan
1. Tambah span pada panggilan gRPC Modul 17 (pakai `otelgrpc`).
2. Tambahkan exporter OTLP + jalankan Jaeger via Docker, lihat trace-nya.
3. Sisipkan `trace_id` ke structured log (Modul 18) untuk korelasi.
4. Tambah sampling 10% dan amati efeknya.
5. Buat span error (`SetStatus(codes.Error)`) & lihat bagaimana ia ditandai.
