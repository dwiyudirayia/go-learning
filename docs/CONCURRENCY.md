# 🔀 Konkurensi Go

Ringkasan pola & aturan main konkurensi. Mendalam di modul **07-concurrency** & **38-concurrency-advanced**.

> Filosofi: *"Don't communicate by sharing memory; share memory by communicating."*
> Konkurensi (struktur) ≠ paralelisme (eksekusi bersamaan). Go memberi konkurensi murah; paralelisme diatur `GOMAXPROCS`.

---

## Goroutine

```go
go f()            // jalankan f konkuren; sangat murah (stack awal ~2KB, tumbuh)
```

**Aturan:** setiap goroutine WAJIB punya jalan keluar, atau ia **bocor** (leak). `main` keluar → semua goroutine mati tanpa cleanup.

## Channel

```go
ch := make(chan int)      // unbuffered: kirim & terima BERSINKRON (rendezvous)
ch := make(chan int, 3)   // buffered: kirim tak blok sampai penuh
ch <- v                   // kirim
v := <-ch                 // terima
v, ok := <-ch             // ok=false bila channel ditutup & kosong
close(ch)                 // hanya PENGIRIM yang menutup
for v := range ch { ... } // baca sampai ditutup
```

- Kirim ke channel penuh / terima dari kosong → **blok**.
- Kirim ke channel tertutup → **panic**. Terima dari tertutup → langsung dapat zero value.
- Channel `nil` → blok selamanya (berguna di `select` untuk menonaktifkan case).

## `select`

```go
select {
case v := <-ch1: ...
case ch2 <- x: ...
case <-ctx.Done(): return ctx.Err() // pembatalan
case <-time.After(time.Second): ... // timeout
default: ...                        // non-blocking (tanpa ini, select blok)
}
```

## `sync`

```go
var wg sync.WaitGroup
for _, job := range jobs {
    wg.Add(1)
    go func(j Job) { defer wg.Done(); proses(j) }(job)
}
wg.Wait()

var mu sync.Mutex      // atau sync.RWMutex (RLock untuk baca konkuren)
mu.Lock(); defer mu.Unlock()

var once sync.Once
once.Do(func() { /* init sekali */ })

var pool sync.Pool     // reuse objek panas (isi bisa di-GC kapan saja)
```

## `sync/atomic` (Go 1.19+ bertipe)

```go
var n atomic.Int64
n.Add(1); n.Load()
var p atomic.Pointer[T]
ok := n.CompareAndSwap(old, new)
```

## `context` — pembatalan & deadline

```go
func Ambil(ctx context.Context, id string) (T, error) { // ctx = param PERTAMA
    select {
    case <-ctx.Done(): return zero, ctx.Err()
    case r := <-hasil: return r, nil
    }
}

ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
defer cancel() // SELALU panggil cancel (hindari leak)
```

Deadline **mengalir ke bawah**: batal di atas membatalkan seluruh sub-panggilan (inti ketahanan microservices → `17-studi-kasus-microservices`).

---

## Pola penting

### Pipeline
```go
func gen(nums ...int) <-chan int { out := make(chan int); go func(){ defer close(out); for _, n := range nums { out <- n } }(); return out }
```
Rangkaian tahap, tiap tahap goroutine, dihubungkan channel.

### Fan-out / Fan-in
Banyak worker baca dari satu channel (fan-out); hasil digabung ke satu channel (fan-in).

### Worker pool
```go
jobs := make(chan Job); results := make(chan Result)
for w := 0; w < 3; w++ { go worker(jobs, results) } // 3 worker
```
Batasi konkurensi = ukuran pool.

### Semaphore (buffered channel)
```go
sem := make(chan struct{}, maxKonkuren)
sem <- struct{}{}      // acquire (blok bila penuh)
defer func(){ <-sem }() // release
```

### Graceful drain
Hentikan produksi → tutup channel → `wg.Wait()` → tak ada pekerjaan terpotong. 📍 `25-background-jobs/advanced`

---

## Lanjutan (modul 38)

- **`errgroup`**: goroutine bersama + `SetLimit(n)` + cancel-on-first-error + `Wait()` kembalikan error pertama.
- **`semaphore.Weighted`**: batas resource berbobot.
- **`singleflight`**: gabungkan panggilan duplikat konkuren jadi satu (anti cache stampede → `22-caching`).
- **False sharing**: padding struct agar field yang sering ditulis beda cache line.

---

## Go Memory Model (ringkas)

Perubahan satu goroutine **tak dijamin terlihat** goroutine lain **tanpa sinkronisasi**. Sinkronisasi sah **hanya** lewat: channel, `sync`, `sync/atomic`. Segala akses bersama tanpa itu = **data race** (perilaku tak terdefinisi).

## Alat wajib

```bash
go test -race ./...   # detektor data race — WAJIB untuk kode konkuren
go run -race ./NN/...
```
`runtime.NumGoroutine()` untuk mendeteksi kebocoran.

---

## Kapan JANGAN pakai konkurensi

Kalau tugasnya cepat & sekuensial, goroutine cuma menambah kompleksitas & risiko race. Ukur dulu — konkurensi untuk **I/O-bound** (tunggu jaringan/disk) atau **CPU-bound yang bisa dipecah**.
