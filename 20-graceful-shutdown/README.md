# 20 — Graceful Shutdown & Lifecycle

Saat server dihentikan (deploy ulang, scale down, Ctrl+C), request yang **sedang berjalan** tidak boleh terputus. **Graceful shutdown** = berhenti menerima koneksi baru, tunggu request in-flight selesai, baru mati.

Jalankan:
```bash
go run ./20-graceful-shutdown
# terminal lain:
curl localhost:8080/slow    # butuh 300ms
# tekan Ctrl+C di server saat /slow berjalan -> request tetap diselesaikan
```
Verifikasi otomatis: `go test ./20-graceful-shutdown` (membuktikan request in-flight tetap 200).

## Pola inti

### 1. Tangkap sinyal OS
```go
ctx, stop := signal.NotifyContext(context.Background(),
    syscall.SIGINT,  // Ctrl+C
    syscall.SIGTERM) // dikirim Docker/Kubernetes saat menghentikan container
defer stop()
```
`SIGTERM` adalah yang dikirim orchestrator (K8s/Docker) sebelum `SIGKILL`. Kamu punya jendela waktu (default K8s 30 detik) untuk berhenti rapi.

### 2. Shutdown dengan timeout
```go
<-ctx.Done()                     // tunggu sinyal
shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
srv.Shutdown(shutdownCtx)        // berhenti terima koneksi baru; tunggu in-flight selesai
```
- `srv.Serve` mengembalikan `http.ErrServerClosed` saat `Shutdown` dipanggil — itu **normal**, bukan error.
- Bila melewati timeout (request macet), `srv.Close()` memaksa tutup.

## Kenapa penting

Tanpa graceful shutdown, setiap deploy/restart bisa:
- Memutus request user di tengah (error 502).
- Meninggalkan transaksi DB setengah jalan.
- Kehilangan pesan yang sedang diproses.

Di Kubernetes, ini wajib dipasangkan dengan **readiness probe**: saat menerima SIGTERM, tandai "not ready" agar load balancer berhenti mengirim traffic baru, sambil menyelesaikan yang lama.

## Checklist shutdown lengkap (produksi)
Urutan menutup resource (kebalikan urutan membuka):
1. Berhenti terima request baru (`srv.Shutdown`).
2. Tunggu worker/background job selesai (`sync.WaitGroup`, Modul 7).
3. Flush buffer (log, metrics, trace).
4. Tutup koneksi DB, Redis, message queue.

## Kapan & Di Mana Dipakai
- **Setiap** service HTTP/gRPC/worker di produksi. Ini pembeda service "amatir" vs "siap deploy".

## Latihan
1. Tambahkan penutupan koneksi DB (`db.Close()`) setelah `Shutdown`.
2. Tambah background worker (goroutine + `WaitGroup`) yang juga ditunggu saat shutdown.
3. Terapkan pola ini ke server Fiber Modul 13 (`app.ShutdownWithContext(ctx)`).
4. Tambah readiness flag: `/readyz` mengembalikan 503 setelah sinyal diterima.
5. Uji dengan `kill -TERM <pid>` (bukan hanya Ctrl+C).
