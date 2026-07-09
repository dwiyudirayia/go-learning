# 18 — Observability (Logging, Metrics)

**Observability** = kemampuan memahami apa yang terjadi di dalam sistem dari luar. Tiga pilarnya: **Logs** (apa yang terjadi), **Metrics** (angka/tren), **Traces** (alur satu request lintas service). Modul ini fokus **Logs** & **Metrics**.

Jalankan:
```bash
go run ./18-observability
curl localhost:8080/hello?name=Ana
curl localhost:8080/metrics       # data untuk Prometheus
```
Verifikasi otomatis: `go test ./18-observability`

## 1. Structured Logging dengan `log/slog` (stdlib, Go 1.21+)

Jangan `fmt.Println` di produksi. Pakai **log terstruktur** (key=value) agar bisa difilter/di-query mesin.
```go
logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
logger.Info("http_request",
    slog.String("method", "GET"),
    slog.Int("status", 200),
    slog.Duration("durasi", d),
)
// -> {"time":"...","level":"INFO","msg":"http_request","method":"GET","status":200,...}
```
- **JSON handler** → siap dikonsumsi Loki/ELK/Datadog/CloudWatch.
- **Level**: Debug/Info/Warn/Error. Set lewat `HandlerOptions{Level: ...}`.
- `slog.String/Int/Duration/Any` membuat field bertipe — bukan string mentah.

**Kenapa penting:** saat produksi error jam 2 pagi, kamu bisa `status=500 AND path=/orders` di dashboard, bukan `grep` teks acak.

## 2. Metrics dengan Prometheus

Tiga tipe metrik utama:
| Tipe | Sifat | Contoh |
|------|-------|--------|
| **Counter** | hanya naik | jumlah request, jumlah error |
| **Histogram** | distribusi | durasi request → p50/p95/p99 |
| **Gauge** | naik-turun | request in-flight, koneksi DB aktif |

```go
httpRequestsTotal = prometheus.NewCounterVec(
    prometheus.CounterOpts{Name: "http_requests_total", Help: "..."},
    []string{"method", "path", "status"}, // label = dimensi
)
httpRequestsTotal.WithLabelValues("GET", "/hello", "200").Inc()
```
- **Label** = dimensi untuk mengiris data (per method/path/status).
- ⚠️ **Jangan pakai nilai tak terbatas** (mis. user ID, `{id}` mentah) sebagai label → "cardinality explosion". Modul ini pakai **`r.Pattern`** (`GET /hello`), bukan path mentah.
- `/metrics` diekspos via `promhttp.HandlerFor(reg, ...)` → di-*scrape* Prometheus tiap interval.

### Middleware metrics
`metricsMiddleware` (di `metrics.go`) membungkus tiap request: naikkan gauge in-flight, catat durasi (histogram) & jumlah (counter). `statusRecorder` dipakai untuk **menangkap status code** (karena `http.ResponseWriter` tak menyimpannya).

## 3. Pilar ketiga: Tracing (sekilas)

**Distributed tracing** (OpenTelemetry) melacak satu request yang melewati banyak service (order → inventory → payment), tiap langkah jadi "span". Sangat berguna di microservices (Modul 17). Setup OTel lebih berat (exporter + collector) — di luar cakupan modul ini, tapi polanya: bungkus handler & gRPC call dengan interceptor OTel, lalu kirim ke Jaeger/Tempo.

## Kapan & Di Mana Dipakai
- **SEMUA** service produksi butuh minimal: structured log + metrics + health check.
- Metrics → dashboard (Grafana) & alert ("error rate > 1%").
- Logs → investigasi insiden.
- Traces → temukan service lambat di rantai microservice.

## Latihan
1. Tambah metrik **Gauge** `db_connections_active` dan naikkan/turunkan secara simulasi.
2. Tambah label `status` juga pada histogram durasi.
3. Ganti log level via env `LOG_LEVEL` (debug/info/warn).
4. Tambah `slog.With("request_id", id)` untuk menyertakan request ID di semua log satu request.
5. Tambah endpoint `/healthz` (liveness) & `/readyz` (readiness) — lihat Modul 20.

## ✅ Solusi Latihan (Pembahasan)

1. **Gauge `db_connections_active`** — Gauge bisa naik & turun (beda dari Counter):
   ```go
   dbConns := promauto.NewGauge(prometheus.GaugeOpts{Name: "db_connections_active"})
   dbConns.Inc()  // saat koneksi dipakai
   dbConns.Dec()  // saat dilepas   (atau dbConns.Set(float64(pool.Stats().InUse)))
   ```
2. **Label `status` pada histogram** — ubah jadi `*HistogramVec` dan sertakan label saat observe:
   ```go
   reqDur := promauto.NewHistogramVec(prometheus.HistogramOpts{Name: "http_duration_seconds"}, []string{"status"})
   reqDur.WithLabelValues(strconv.Itoa(code)).Observe(elapsed.Seconds())
   ```
   Hati-hati kardinalitas: pakai kelas status (`2xx`,`4xx`,`5xx`), jangan kode mentah tak terbatas.
3. **`LOG_LEVEL` via env** — petakan string ke `slog.Level`:
   ```go
   var lvl slog.Level
   lvl.UnmarshalText([]byte(os.Getenv("LOG_LEVEL"))) // "debug"/"info"/"warn"/"error"
   h := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: lvl})
   ```
4. **Request ID di semua log** — buat child logger lalu taruh di context:
   ```go
   reqLog := logger.With("request_id", id)
   ctx := context.WithValue(r.Context(), logKey{}, reqLog)
   ```
   Semua log dari `reqLog` otomatis membawa `request_id` — korelasi satu request jadi mudah.
5. **`/healthz` & `/readyz`** — lihat Modul 20: `/healthz` selalu 200 bila proses hidup; `/readyz` cek dependency (DB ping) dan balas 503 bila belum siap / sedang shutdown.

---

## 🚀 Teknik Advanced (Level Up)
> 💻 **Contoh runnable + komentar detail** untuk teknik di bawah ada di folder [`advanced/`](advanced). Jalankan: `go run ./18-observability/advanced`


- **Tiga pilar** — logs (`slog`), metrics (Prometheus), traces (OTel [[33-distributed-tracing]]) — korelasikan via trace-id.
- **`slog` handler kustom** — tambah atribut dari context (request-id, user), redaksi field sensitif, output JSON untuk agregator.
- **Metrik: kardinalitas label** — hindari label ber-kardinalitas tinggi (user-id, url mentah) → ledakan time-series. Gunakan label terbatas.
- **Histogram vs Summary** — histogram (bucket) bisa diagregasi lintas instance; pilih bucket sesuai SLO latensi.
- **Metodologi** — **RED** (Rate, Errors, Duration) untuk service; **USE** (Utilization, Saturation, Errors) untuk resource.
- **Exemplars** — tautkan sampel trace ke bucket histogram untuk drill-down.
- **`pprof` endpoint** — expose `/debug/pprof` (diamankan) untuk profil produksi. Lihat [[26-profiling]].
