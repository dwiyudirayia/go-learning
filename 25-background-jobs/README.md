# 25 — Background Jobs & Scheduler

Tidak semua kerja harus selesai saat request (sinkron). Kirim email, resize gambar, generate laporan → lebih baik dijalankan di **latar belakang** agar response cepat. Modul ini membangun **worker queue** dengan retry, backoff, dan idempotensi, plus **scheduler** berkala.

Jalankan:
```bash
go run ./25-background-jobs
```
Verifikasi otomatis: `go test -race ./25-background-jobs`

## Worker Queue (`queue.go`)

```
Enqueue(job) ──► [ channel antrean ] ──► worker 1 ┐
                                          worker 2 ├─ proses konkuren
                                          worker 3 ┘
```
Berbasis pola dari Modul 7 (channel + worker pool), ditambah kemampuan produksi:

### 1. Retry + Backoff
Job yang gagal dicoba ulang sampai `MaxRetries`, dengan jeda yang **bertambah** tiap percobaan:
```go
for attempt := 1; attempt <= MaxRetries+1; attempt++ {
    if err := handler(); err == nil { return }  // sukses
    time.Sleep(backoff * time.Duration(attempt)) // backoff
}
// exhausted -> dead letter
```
Output: job "flaky" gagal 2×, sukses di percobaan ke-3.

### 2. Dead Letter
Job yang gagal permanen (retry habis) tidak dibuang diam-diam — dicatat sebagai "failed" untuk investigasi/alert. Di produksi: simpan ke tabel/queue khusus.

### 3. Idempotensi (WAJIB di sistem terdistribusi)
Message queue bisa mengirim pesan **ganda** (at-least-once delivery). Handler harus aman dijalankan >1× untuk input sama:
```go
if q.processed[job.ID] { return } // sudah pernah sukses -> lewati
```
Test membuktikan: job dengan ID sama di-enqueue 2×, handler dipanggil **1×**.

## Scheduler (`scheduler.go`)
Menjalankan task berkala (cron sederhana) dengan `time.Ticker`, berhenti rapi via `context` (Modul 20):
```go
NewScheduler(5*time.Minute, cleanupTask).Run(ctx)
```
Untuk ekspresi cron sungguhan (`"0 */5 * * *"`), pakai [robfig/cron](https://github.com/robfig/cron).

## Kapan pakai apa?

| Kebutuhan | Solusi |
|-----------|--------|
| Kerja async dalam 1 proses | worker queue in-memory (modul ini) |
| Antar service / tahan restart | message queue (NATS/Kafka, Modul 23) + worker |
| Terjadwal (harian, tiap 5 menit) | scheduler/cron |
| Job berat & terdistribusi | Asynq, Machinery, River (queue di atas Redis/Postgres) |

⚠️ Worker in-memory **hilang saat proses mati**. Untuk job penting, pakai queue persisten (Redis/DB/Kafka).

## Kapan & Di Mana Dipakai
- Kirim email/notifikasi, proses upload, generate PDF/laporan, sinkronisasi data, cleanup berkala, retry panggilan API pihak ketiga.

## Latihan
1. Tambah **prioritas** job (2 channel: high & normal).
2. Ganti backoff linear jadi **eksponensial + jitter**.
3. Simpan dead letter ke file/DB, bukan hanya menghitung.
4. Integrasikan dengan NATS (Modul 23): job masuk dari event, bukan `Enqueue` langsung.
5. Pakai `robfig/cron` untuk jadwal "tiap hari jam 2 pagi".
