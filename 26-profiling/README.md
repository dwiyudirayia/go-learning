# 26 — Profiling & Optimization

> *"Premature optimization is the root of all evil."* — **Ukur dulu, baru optimasi.** Jangan menebak di mana bottleneck; Go punya tools kelas dunia untuk mengukurnya.

Jalankan:
```bash
go test -bench . -benchmem ./26-profiling   # bandingkan performa & alokasi
go run ./26-profiling                        # server dengan endpoint pprof
```

## 1. Benchmark (`testing.B`)

```go
func BenchmarkBuildFast(b *testing.B) {
    for i := 0; i < b.N; i++ { BuildFast(1000) }
}
```
`b.N` diatur otomatis. Jalankan dengan `-benchmem` untuk lihat alokasi:
```
BenchmarkBuildSlow-12   210983 ns/op   530283 B/op   999 allocs/op
BenchmarkBuildFast-12     1081 ns/op     1024 B/op     1 allocs/op
```
**Bukti nyata:** `strings.Builder` + `Grow()` mengubah **999 alokasi → 1**, dan **195× lebih cepat**. Ini kenapa Modul 2 menekankan `strings.Builder`.

### Kenapa `+=` lambat?
String **immutable** → tiap `s += "x"` mengalokasikan string baru & menyalin semua isi lama = O(n²). `strings.Builder` menulis ke buffer yang tumbuh.

## 2. pprof (profil runtime)

Import side-effect mendaftarkan endpoint:
```go
import _ "net/http/pprof"   // -> /debug/pprof/*
```
Ambil & analisa profil:
```bash
# CPU: apa yang paling banyak makan CPU?
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=5

# Memori: apa yang paling banyak alokasi?
go tool pprof http://localhost:6060/debug/pprof/heap

# Goroutine: ada kebocoran goroutine?
curl http://localhost:6060/debug/pprof/goroutine?debug=1
```
Di dalam `pprof`:
- `top` — fungsi dengan konsumsi terbesar.
- `list NamaFungsi` — baris kode mana yang berat.
- `web` — visualisasi graf (butuh graphviz).

Profil dari benchmark:
```bash
go test -bench . -cpuprofile cpu.out -memprofile mem.out ./26-profiling
go tool pprof cpu.out
```

## 3. Escape Analysis (alokasi heap vs stack)

Nilai yang "lolos" dari fungsi (mis. pointer yang dikembalikan) dialokasikan di **heap** (dikelola GC, lebih mahal). Cek:
```bash
go build -gcflags='-m' ./26-profiling   # "escapes to heap" = alokasi heap
```
Mengurangi alokasi heap = mengurangi tekanan GC = lebih cepat.

## Alur optimasi yang benar
1. **Ukur** (benchmark/pprof) — temukan bottleneck nyata.
2. **Optimasi** bagian itu saja.
3. **Ukur lagi** — buktikan membaik.
4. Jaga korektnes (test tetap hijau).

Jangan optimasi kode yang bukan bottleneck — buang waktu & bikin kode rumit.

## Kapan & Di Mana Dipakai
- Saat ada masalah performa nyata (latensi tinggi, memori bengkak, CPU 100%).
- Sebelum rilis fitur yang jalur-panas (hot path).
- Mencari **goroutine leak** & **memory leak** di produksi.

## Latihan
1. Tambah `BenchmarkBuildFastNoGrow` (tanpa `Grow`) & bandingkan.
2. Jalankan `go build -gcflags='-m'` dan temukan yang "escapes to heap".
3. Buat fungsi dengan goroutine leak, temukan lewat `/debug/pprof/goroutine`.
4. Profil `SumSquares` dengan `-cpuprofile`, buka dengan `go tool pprof`.
5. Optimasi sebuah fungsi di modul lain (mis. reverse string Modul 2) & buktikan dengan benchmark.

## ✅ Solusi Latihan (Pembahasan)

1. **`BenchmarkBuildFastNoGrow`** — salin benchmark tanpa `sb.Grow(n)`. Jalankan `go test -bench . -benchmem` → versi tanpa Grow punya lebih banyak `allocs/op` (realokasi berulang saat buffer tumbuh).
2. **Escape analysis** — `go build -gcflags='-m' ./26-profiling 2>&1 | grep escapes`. Nilai yang alamatnya dikembalikan / disimpan di interface biasanya "escapes to heap".
3. **Goroutine leak** — buat goroutine yang `<-make(chan int)` (blok selamanya). Buka `/debug/pprof/goroutine?debug=1` → jumlah goroutine terus naik. Perbaiki dengan `context` cancel.
4. **CPU profile** — `go test -bench BenchmarkSumSquares -cpuprofile cpu.out` lalu `go tool pprof cpu.out` → `top`, `list SumSquares` untuk lihat baris terpanas.
5. **Optimasi reverse (Modul 2)** — benchmark versi `[]rune` vs versi dua-pointer in-place; buktikan mana lebih sedikit alokasi. Aturannya: **ukur dulu**, jangan tebak.

---

## 🚀 Teknik Advanced (Level Up)
> 💻 **Contoh runnable + komentar detail** untuk teknik di bawah ada di folder [`advanced/`](advanced). Jalankan: `go run ./26-profiling/advanced`


- **`pprof` profil** — CPU, heap, goroutine, block, mutex. `go tool pprof`, visual `-http=:8080` (flamegraph). Profil di produksi via endpoint aman.
- **Benchmark disiplin** — `-benchmem` (allocs/op), jalankan berkali (`-count`), bandingkan dengan **`benchstat`** (signifikansi statistik, bukan satu angka).
- **Execution tracer** — `runtime/trace` + `go tool trace` untuk lihat scheduler, GC, blocking secara timeline.
- **Escape analysis** — `go build -gcflags='-m'` lihat apa yang lolos ke heap; kurangi alokasi panas. Lihat [[42-go-internals]].
- **Tuning GC** — `GOGC` (frekuensi GC) & **`GOMEMLIMIT`** (Go 1.19, soft memory limit) untuk kontrol jejak memori vs CPU.
- **`sync.Pool`** untuk reuse objek panas; ukur dulu—jangan tebak.
- **Optimasi berbasis data** — profil dulu, baru optimasi; hindari micro-optimization tanpa bukti.
