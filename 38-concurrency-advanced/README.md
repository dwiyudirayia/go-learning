# 38 — Concurrency Advanced

Lanjutan Modul 7. Empat alat konkurensi tingkat lanjut dari `golang.org/x/sync` + stdlib yang membuat kode paralel lebih ringkas, aman, dan efisien.

Jalankan:
```bash
go run ./38-concurrency-advanced
go test -race ./38-concurrency-advanced
```

## 1. `errgroup` — paralel + error handling

Menggantikan `WaitGroup` + channel error manual (Modul 7). Jalankan banyak tugas paralel; **error pertama membatalkan sisanya** via context.
```go
g, ctx := errgroup.WithContext(ctx)
for _, id := range ids {
    g.Go(func() error {
        r, err := fetch(ctx, id)   // ctx dibatalkan bila ada yang gagal
        if err != nil { return err }
        results[i] = r
        return nil
    })
}
err := g.Wait()   // tunggu semua; kembalikan error pertama
```
Output: fetch paralel sukses; saat "id 2 gagal", sisanya dibatalkan. Jauh lebih rapi daripada WaitGroup manual.

## 2. `semaphore` berbobot — batasi konkurensi

Seperti bulkhead (Modul 32), tapi mendukung **bobot** (satu operasi bisa ambil >1 slot — mis. proporsional ke penggunaan memori).
```go
sem := semaphore.NewWeighted(3)
sem.Acquire(ctx, 1); defer sem.Release(1)   // maks 3 bersamaan
```
Output: 6 task, maks 2 → puncak bersamaan = **2**.

## 3. `singleflight` — cegah cache stampede ⭐

Menggabungkan panggilan **identik yang bersamaan** menjadi **satu**. Saat cache miss serempak (Modul 22), tanpa ini 1000 request menyerbu DB bersamaan.
```go
v, err, _ := group.Do(key, func() (any, error) { return loadFromDB(key) })
```
Output menakjubkan: **100 request** key sama → loader dipanggil **1 kali**. 99 lainnya menunggu & berbagi hasil yang sama.

## 4. `sync.Pool` — daur ulang objek (kurangi GC)

Menyimpan objek yang bisa dipakai ulang → mengurangi alokasi & tekanan garbage collector (Modul 26).
```go
var bufPool = sync.Pool{New: func() any { return new(bytes.Buffer) }}
buf := bufPool.Get().(*bytes.Buffer)
buf.Reset()               // WAJIB reset — objek bisa bekas
defer bufPool.Put(buf)    // kembalikan
```
Cocok untuk objek yang sering dibuat-buang di jalur panas (buffer, encoder). ⚠️ Jangan pakai untuk objek dengan state yang harus persist; isi pool bisa di-GC kapan saja.

## Ringkasan kapan pakai apa
| Alat | Untuk |
|------|-------|
| `errgroup` | banyak tugas paralel yang bisa gagal |
| `semaphore` | batasi konkurensi (dengan bobot) |
| `singleflight` | dedupe panggilan identik (cache stampede) |
| `sync.Pool` | kurangi alokasi objek di hot path |
| `sync.Once` (Modul 7) | inisialisasi sekali |
| `atomic` (Modul 7) | counter/flag ringan |

## Tips konkurensi lanjutan
- **Selalu** uji dengan `-race` (Modul 7).
- Batasi jumlah goroutine (worker pool/semaphore) — jangan spawn tak terbatas.
- Propagasikan `context` untuk cancellation di seluruh rantai.
- Hindari share memory; utamakan channel — tapi mutex/atomic tepat untuk counter/state kecil.

## Kapan & Di Mana Dipakai
- `errgroup`: panggil banyak service paralel (Modul 17), proses batch.
- `singleflight`: layer cache (Modul 22), dedupe request mahal.
- `semaphore`/`sync.Pool`: kontrol resource & performa di jalur panas.

## Latihan
1. Ganti worker pool manual Modul 7 dengan `errgroup` + `semaphore`.
2. Tambahkan `singleflight` ke cache Modul 22 (cegah stampede saat miss).
3. Ukur efek `sync.Pool` dengan benchmark (`-benchmem`, Modul 26).
4. Buat `errgroup.SetLimit(n)` untuk batasi goroutine sekaligus.
5. Gabungkan `errgroup` + `context.WithTimeout` untuk deadline seluruh batch.

## ✅ Solusi Latihan (Pembahasan)

1. **`errgroup` + `semaphore`** — ganti worker pool manual (Modul 7) :
   ```go
   g, ctx := errgroup.WithContext(ctx)
   for _, job := range jobs { job := job; g.Go(func() error { return proses(ctx, job) }) }
   err := g.Wait() // error pertama membatalkan sisanya
   ```
2. **`singleflight` ke cache** — bungkus loader cache-miss (Modul 22) dengan `g.Do(key, load)` agar miss bersamaan hanya 1× ke DB.
3. **`sync.Pool` benchmark** — bandingkan alokasi buffer dengan & tanpa `Pool` via `-benchmem` (Modul 26). Pool mengurangi tekanan GC pada objek yang sering dibuat-buang.
4. **`errgroup.SetLimit(n)`** — batasi goroutine berjalan bersamaan jadi n (mirip semaphore built-in). `g.SetLimit(10)` sebelum loop `g.Go`.
5. **`errgroup` + timeout** — `ctx, cancel := context.WithTimeout(ctx, 5*time.Second)`; `errgroup.WithContext(ctx)` → seluruh batch punya deadline; yang lewat batas dibatalkan serempak.

---

## 🚀 Teknik Advanced (Level Up)
> 💻 **Contoh runnable + komentar detail** untuk teknik di bawah ada di folder [`advanced/`](advanced). Jalankan: `go run ./38-concurrency-advanced/advanced`


- **`errgroup` + `SetLimit`** — paralelisme terbatas dengan agregasi error & cancel otomatis saat satu gagal.
- **Weighted semaphore** — `golang.org/x/sync/semaphore` untuk batasi resource berbobot (bukan sekadar hitung).
- **`singleflight`** — gabungkan panggilan duplikat concurrent jadi satu (anti cache stampede [[22-caching]]).
- **`sync.Pool`** — reuse objek panas; awas: objek bisa di-GC kapan saja, jangan simpan state.
- **False sharing** — padding struct agar field yang sering ditulis beda cache line (perf multi-core). Lihat [[42-go-internals]].
- **Atomic & lock-free** — `atomic.Pointer[T]`, compare-and-swap untuk struktur lock-free (mahir; ukur vs mutex).
- **`rate.Limiter`** — throttle idiomatik.
