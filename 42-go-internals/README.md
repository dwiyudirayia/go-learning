# 42 — Go Internals

Memahami cara kerja Go "di balik layar" — bukan untuk hafalan, tapi agar bisa **menjelaskan performa** & mengambil keputusan tepat (Modul 26, 38). Empat topik: scheduler, GC, memory alignment, escape analysis.

Jalankan:
```bash
go run ./42-go-internals
go build -gcflags='-m' ./42-go-internals   # lihat escape analysis
go test -bench . -benchmem ./42-go-internals
```

## 1. Scheduler — model GMP

Go menjalankan **ribuan goroutine** di atas **sedikit thread OS** dengan scheduler sendiri (bukan OS).

```
G (goroutine)  — unit kerja ringan, ~2KB stack awal (tumbuh dinamis)
M (machine)    — thread OS sungguhan
P (processor)  — konteks penjadwalan; jumlahnya = GOMAXPROCS
```
- **`GOMAXPROCS`** = jumlah P = maksimum goroutine yang benar-benar **paralel** (default = jumlah CPU).
- Scheduler bersifat **cooperative + preemptive**: goroutine berpindah saat blocking (I/O, channel, syscall) atau di titik aman.
- **Work-stealing**: P yang menganggur "mencuri" goroutine dari P lain → beban merata.

Karena goroutine murah, spawn ribuan itu normal (output: 1000 goroutine tanpa masalah). Inilah kenapa Go unggul untuk server berkonkurensi tinggi (Modul 7).

## 2. Garbage Collector

Go memakai GC **concurrent, tri-color mark-and-sweep** dengan pause sangat pendek (biasanya < 1ms).
```go
var m runtime.MemStats
runtime.ReadMemStats(&m)   // m.NumGC, m.HeapAlloc, m.TotalAlloc, m.PauseNs
runtime.GC()               // paksa (jarang perlu; biasanya otomatis)
```
- **`GOGC`** (default 100) = ambang: GC jalan saat heap tumbuh 100% sejak GC terakhir. Naikkan (`GOGC=200`) → GC lebih jarang, pakai memori lebih banyak, CPU lebih hemat.
- **`GOMEMLIMIT`** = batas memori keras (soft) untuk mencegah OOM.
- Mengurangi **alokasi** (Modul 38 `sync.Pool`, Modul 26) mengurangi tekanan GC → aplikasi lebih cepat.

## 3. Memory Alignment & Padding

Compiler menyisipkan **padding** agar tiap field selaras dengan batas ukurannya. **Urutan field memengaruhi ukuran struct**:
```go
type Bad  struct { a bool; b int64; c bool } // 24 byte (padding boros)
type Good struct { b int64; a bool; c bool } // 16 byte (rapi)
```
Output & test membuktikan: **hemat 8 byte** hanya dengan mengurutkan field **besar → kecil**. Untuk jutaan struct (mis. cache, slice besar), ini menghemat memori signifikan. `unsafe.Sizeof` mengungkapnya; alat `fieldalignment` (golang.org/x/tools) mengeceknya otomatis.

## 4. Escape Analysis (stack vs heap)

Compiler memutuskan apakah sebuah nilai bisa di **stack** (murah, otomatis bebas) atau harus **heap** (dikelola GC, lebih mahal):
```go
func escape() *int { x := 42; return &x }  // &x lolos -> HEAP
func stay() int    { x := 42; return x }   // tak lolos -> STACK
```
Lihat keputusannya:
```bash
go build -gcflags='-m' ./42-go-internals   # "moved to heap: x" / "does not escape"
```
Benchmark membuktikan versi heap lebih lambat & mengalokasi (`allocs/op`). Nilai yang "lolos": pointer yang dikembalikan, disimpan di interface/`any`, atau ditangkap closure.

## Konsep lain (sekilas)

### Memory Model (happens-before)
Go menjamin urutan operasi memori antar-goroutine hanya bila disinkronkan (channel, mutex, `sync/atomic`). Tanpa sinkronisasi = **data race** (Modul 7). Karena itu **selalu `go test -race`**.

### `//go:` directives
Instruksi compiler: `//go:embed` (Modul 21), `//go:generate` (jalankan codegen via `go generate`), `//go:noinline`, `//go:build` (build tags). Ditulis sebagai komentar khusus (tanpa spasi setelah `//`).

## Kapan & Di Mana Dipakai
- Saat mengoptimasi performa (setelah profiling, Modul 26): kurangi alokasi heap, rapikan struct, setel GOGC.
- Saat debugging: goroutine leak, memori bengkak, pause GC.
- Wawancara Go tingkat menengah–senior sering menanyakan GMP & escape analysis.

## Latihan
1. Jalankan `go build -gcflags='-m'` pada fungsi di modul lain — temukan yang "escapes to heap".
2. Rapikan sebuah struct di modul lain (urutkan field) & buktikan `unsafe.Sizeof` mengecil.
3. Set `GOGC=off` lalu jalankan demo GC — amati bedanya (`GOGC=off go run ...`).
4. Set `GOMAXPROCS=1` & bandingkan perilaku worker pool (Modul 7).
5. Pasang `fieldalignment` (`go install golang.org/x/tools/go/analysis/passes/fieldalignment/cmd/fieldalignment@latest`) & jalankan pada repo.

## ✅ Solusi Latihan (Pembahasan)

1. **Escape analysis lintas modul** — `go build -gcflags='-m' ./NN-... 2>&1 | grep 'escapes to heap'`. Nilai yang di-return by pointer atau masuk interface umumnya escape ke heap.
2. **Rapikan struct** — urutkan field dari besar→kecil (mis. `int64, int64, bool` bukan `bool, int64, bool`) untuk kurangi padding; buktikan `unsafe.Sizeof` mengecil (contoh modul: 24B→16B).
3. **`GOGC=off`** — `GOGC=off go run ./42-go-internals` menonaktifkan GC → heap terus tumbuh, tak ada jeda GC. Bandingkan `NumGC` di `runtime.MemStats`. Untuk paham trade-off throughput vs memori.
4. **`GOMAXPROCS=1`** — `GOMAXPROCS=1 go run ./07-concurrency` memaksa 1 OS thread → goroutine berjalan konkuren tapi tak paralel; worker pool jadi serial secara efektif.
5. **`fieldalignment`** — `go install golang.org/x/tools/go/analysis/passes/fieldalignment/cmd/fieldalignment@latest` lalu `fieldalignment ./...`; `-fix` untuk auto-urutkan field. Otomatisasi latihan #2.
