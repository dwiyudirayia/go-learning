# 07 — Concurrency

Jalankan:
```bash
go run ./07-concurrency
# WAJIB dicoba: deteksi data race
go run -race ./07-concurrency
```

Ini **kekuatan utama Go**. Motto Go: *"Don't communicate by sharing memory; share memory by communicating"* — utamakan **channel** untuk komunikasi antar goroutine, mutex hanya bila perlu lindungi state bersama.

## 1. Goroutine

Fungsi yang dijalankan konkuren, cukup dengan `go`:
```go
go doWork()   // jalan di goroutine baru, main tidak menunggu
```
- Sangat ringan (ribuan–jutaan goroutine OK; bukan OS thread 1:1).
- **`main` tidak menunggu goroutine.** Kalau `main` selesai, semua goroutine mati. Perlu sinkronisasi (`WaitGroup`/channel).

## 2. Channel

Pipa berjenis untuk mengirim nilai antar goroutine:
```go
ch := make(chan int)      // unbuffered: kirim & terima harus "bertemu" (sinkron)
ch := make(chan int, 3)   // buffered: kirim tak blok selama buffer belum penuh
ch <- 5                   // kirim
v := <-ch                 // terima
close(ch)                 // tutup (hanya pengirim yang menutup)
for v := range ch { ... } // terima sampai channel ditutup
```
- **Arah channel** memperjelas maksud: `chan<- int` (kirim saja), `<-chan int` (terima saja).
- Menerima dari channel tertutup langsung mengembalikan zero value; `v, ok := <-ch` → `ok=false` bila sudah tutup & kosong.
- **Aturan:** yang **menutup** channel adalah **pengirim**, bukan penerima. Menutup 2x atau mengirim ke channel tertutup → panic.

## 3. `select`

Menunggu beberapa operasi channel sekaligus:
```go
select {
case v := <-ch1:
	...
case ch2 <- x:
	...
case <-time.After(time.Second):
	// timeout
default:
	// tak ada yang siap (non-blocking)
}
```

## 4. Paket `sync`

- **`sync.WaitGroup`** — tunggu sekumpulan goroutine selesai (`Add`/`Done`/`Wait`).
- **`sync.Mutex` / `RWMutex`** — lindungi state bersama dari akses bersamaan.
- **`sync.Once`** — jalankan sesuatu **tepat sekali** (mis. inisialisasi singleton).
- **`sync/atomic`** — operasi atomik ringan untuk counter sederhana.

## 5. `context`

Cara idiomatik untuk **cancellation, timeout, deadline**, dan membawa nilai request-scoped. Wajib di server & pemanggilan berantai:
```go
ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
defer cancel()
select {
case <-ctx.Done():
	return ctx.Err() // context deadline exceeded / canceled
case res := <-work:
	...
}
```

## 6. Data race & `-race`

Dua goroutine mengakses variabel sama dan minimal satu menulis, tanpa sinkronisasi = **data race** (bug non-deterministik). Selalu uji dengan:
```bash
go run -race ./07-concurrency
go test -race ./...
```

## 7. Pola konkurensi penting

- **Worker pool** — N worker mengambil job dari satu channel (batasi paralelisme).
- **Pipeline** — rangkaian tahap, tiap tahap goroutine, dihubungkan channel.
- **Fan-out / Fan-in** — sebar kerja ke banyak goroutine (fan-out), gabungkan hasil (fan-in).

## Latihan
1. Jalankan 5 goroutine yang masing-masing mencetak ID-nya; pakai `WaitGroup` agar `main` menunggu semuanya.
2. Buat generator: fungsi yang mengembalikan `<-chan int` berisi 1..n, lalu konsumsi dengan `range`.
3. Buat `Counter` aman-konkuren memakai `sync.Mutex`; naikkan dari 100 goroutine, hasil harus tepat 100 (uji `-race`).
4. Buat worker pool: 3 worker memproses 9 job (kuadratkan angka), kumpulkan hasilnya.
5. Buat fungsi `fetchWithTimeout(d time.Duration)` yang memakai `context.WithTimeout` + `select`; kembalikan hasil bila cepat, atau error timeout bila lambat.

Kerjakan di `07-concurrency/jawaban-saya/`, lalu minta saya review.

## Kapan & Di Mana Dipakai (Studi Kasus Nyata)

| Konsep | Cara pakai di dunia nyata | Contoh kasus |
|--------|---------------------------|--------------|
| **Goroutine** | Kerjakan hal-hal paralel/background | Tangani tiap request HTTP di goroutine sendiri; kirim email async |
| **Channel** | Komunikasi & sinkronisasi antar goroutine | Antre job, streaming hasil, sinyal selesai |
| **`WaitGroup`** | Tunggu banyak tugas paralel selesai | Panggil 5 API sekaligus, tunggu semua sebelum gabung hasil |
| **`select` + timeout** | Batasi waktu tunggu / multiplex channel | Timeout panggilan ke service lain; graceful shutdown |
| **`Mutex`/`atomic`** | Lindungi state bersama | Counter metrics, cache in-memory, connection pool |
| **`sync.Once`** | Inisialisasi sekali (thread-safe) | Singleton config, koneksi DB global, lazy init |
| **`context`** | Cancellation, timeout, deadline berantai | Batalkan query DB saat client putus; deadline seluruh request |
| **Worker pool** | Batasi paralelisme (jangan spawn tak terbatas) | Proses 10.000 file dgn 8 worker; rate-limit panggilan API |
| **Pipeline / fan-in-out** | Alirkan data bertahap, sebar-gabung kerja | ETL: baca→transform→tulis; proses gambar paralel |

**Contoh nyata — panggil beberapa API paralel lalu gabung (pola sangat umum di backend):**
```go
var wg sync.WaitGroup
results := make([]Result, 3)
for i, url := range urls {
    wg.Add(1)
    go func(i int, url string) {
        defer wg.Done()
        results[i] = fetch(url) // tiap goroutine tulis slot berbeda -> aman
    }(i, url)
}
wg.Wait() // total waktu = API terlambat, BUKAN jumlah semuanya
```

**Contoh nyata — `context` di HTTP handler (wajib di produksi):**
```go
func handler(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context() // otomatis dibatalkan jika client memutus koneksi
    rows, err := db.QueryContext(ctx, "SELECT ...") // query ikut dibatalkan
    // ...
}
```

**⚠️ Kesalahan umum yang harus dihindari:**
- **Goroutine leak** — goroutine yang menunggu channel selamanya karena tak ada yang mengirim/menutup. Selalu pastikan ada jalan keluar (close/context).
- **Spawn goroutine tak terbatas** per item → pakai **worker pool** untuk membatasi.
- **Data race** — selalu uji `go test -race`. Lindungi state bersama dengan channel atau mutex.
- **Lupa `wg.Add` sebelum `go`**, atau memanggil `Add` di dalam goroutine → race/hitungan salah.

**Cocok dipakai saat:** server yang menangani banyak request, memanggil banyak service (microservices!), memproses data besar secara paralel, atau butuh timeout/cancellation. Ini yang membuat Go unggul untuk backend & sistem berskala.
