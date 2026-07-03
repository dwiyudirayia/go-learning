# 32 — Resiliency Patterns

Di sistem terdistribusi, kegagalan itu **normal** (service lambat, jaringan putus, beban tinggi). Modul ini menutup pola bertahan hidup: **circuit breaker**, **bulkhead**, dan **distributed lock**. Melengkapi retry/backoff (Modul 23).

Jalankan:
```bash
go run ./32-resiliency-patterns
```
Verifikasi otomatis: `go test ./32-resiliency-patterns`

## 1. Circuit Breaker (`circuitbreaker.go`)

Seperti **sekring listrik**: bila service downstream terus gagal, "putus" agar aplikasi tak membuang waktu memanggil yang rusak (**fail fast**), lalu sesekali mencoba lagi.

```
CLOSED ──(gagal ≥ threshold)──► OPEN ──(setelah timeout)──► HALF-OPEN
  ▲                                                             │
  └──────────────(sukses)─────────────────────────────────────┘
                          HALF-OPEN ──(gagal)──► OPEN
```
Output membuktikan: 3 gagal → **OPEN** → call 4 & 5 ditolak **tanpa memanggil** service (`circuit breaker terbuka`).

**Kenapa penting:** tanpa ini, request menumpuk menunggu service mati → thread/goroutine habis → **cascading failure** (satu service jatuh menjatuhkan semua). Circuit breaker memutus rantai itu.

> Produksi: pakai library matang seperti [sony/gobreaker](https://github.com/sony/gobreaker) atau resilience4go.

## 2. Bulkhead (`bulkhead.go`)

Seperti **sekat kedap air di kapal**: batasi operasi bersamaan agar satu dependency lambat tak menghabiskan semua resource.
```go
bh := NewBulkhead(2)          // maks 2 bersamaan
bh.TryExecute(fn)             // langsung tolak (ErrBulkheadFull) bila penuh
bh.Execute(ctx, fn)           // atau tunggu slot / ctx habis
```
Implementasi: **semaphore** (channel berkapasitas N, Modul 7). Output: 5 request, maks 2 → 3 ditolak.

**Manfaat:** isolasi. Panggilan ke service A yang lambat tak menghabiskan slot untuk service B.

## 3. Distributed Lock (`lock.go`)

Memastikan hanya **satu instance** (dari banyak replika) menjalankan tugas tertentu — mis. cron job yang tak boleh dobel. Memakai Redis:
```go
token, ok, _ := lock.Acquire(ctx, "cron:cleanup", 30*time.Second) // SET NX PX
if ok { defer lock.Release(ctx, "cron:cleanup", token); doJob() }
```
Tiga jaminan penting:
- **`SET NX`** — hanya berhasil bila belum ada (mutual exclusion).
- **TTL** — bila pemegang lock crash, lock **kadaluarsa otomatis** (tak deadlock).
- **Release aman (Lua)** — hanya hapus bila **token cocok** → tak menghapus lock milik instance lain (test membuktikan token salah tak melepas lock).

> Untuk jaminan lebih kuat lintas banyak node Redis, pelajari **Redlock**. Alternatif: lock via etcd/ZooKeeper.

## Pola resiliensi lain (rangkuman)
| Pola | Tujuan | Modul |
|------|--------|-------|
| Retry + backoff + jitter | tahan gangguan sesaat | 23 |
| **Circuit breaker** | fail fast, cegah cascading | ini |
| **Bulkhead** | isolasi resource | ini |
| Timeout / deadline | jangan tunggu selamanya | 7 (`context`) |
| **Distributed lock** | eksklusi antar instance | ini |
| Rate limiting | lindungi dari lonjakan | 27 |
| Graceful degradation | fitur turun, bukan mati | desain |

Gabungan mereka = "defense in depth" untuk keandalan.

## Kapan & Di Mana Dipakai
- Setiap panggilan ke dependency eksternal (service lain, DB, API pihak ketiga).
- Cron/scheduler di aplikasi multi-replika (distributed lock).

## Latihan
1. Tambah `HalfOpenMaxCalls` (izinkan N trial di half-open, bukan 1).
2. Gabungkan circuit breaker + retry (Modul 23): retry hanya bila circuit closed.
3. Bungkus panggilan gRPC Modul 17 dengan circuit breaker + bulkhead.
4. Tambah metrik (Modul 18): state circuit, jumlah reject bulkhead.
5. Implementasikan auto-renew (lease) pada distributed lock untuk job yang lama.
