# 43 — Advanced Generics

Lanjutan Modul 6. Empat kemampuan generik tingkat lanjut: **struktur data generik**, **iterator** (`iter.Seq`, Go 1.23+), **type sets**, dan pola **functional options**.

Jalankan:
```bash
go run ./43-advanced-generics
go test ./43-advanced-generics
```

## 1. Struktur data generik (`Set[T]`)

Wadah reusable untuk tipe apa pun (`comparable`):
```go
s := NewSet(1, 2, 3)      // Set[int]
s2 := NewSet("a", "b")    // Set[string]
a.Union(b); a.Intersect(b)
```
Satu implementasi, dipakai untuk semua tipe — type-safe, tanpa `any` + assertion.

## 2. Iterator & range-over-func (Go 1.23+) ⭐

Fungsi bertipe **`iter.Seq[T]` = `func(yield func(T) bool)`** bisa dipakai langsung dengan `for v := range seq`:
```go
func Count(n int) iter.Seq[int] {
    return func(yield func(int) bool) {
        for i := 0; i < n; i++ {
            if !yield(i) { return } // yield=false -> konsumen break
        }
    }
}
for v := range Count(5) { ... }   // 0 1 2 3 4
```

**Keunggulan: LAZY & bisa dirangkai** tanpa slice antara:
```go
genap   := Filter(Count(10), isEven)   // belum dihitung
kuadrat := Map(genap, square)          // masih lazy
result  := Collect(kuadrat)            // baru dihitung di sini
```
Test membuktikan lazy: `Map(Count(1000000), ...)` yang di-`break` setelah 3 hanya mengevaluasi ~3 elemen, bukan sejuta. Ini memungkinkan memproses aliran data tak-hingga / besar tanpa memuat semuanya ke memori.

`iter.Seq2[K,V]` untuk pasangan (mis. iterasi map). Stdlib `slices`/`maps` sudah mengembalikan iterator (`slices.Values`, `maps.Keys`) & `slices.Collect`/`slices.Sorted` mengonsumsinya.

## 3. Type Sets (constraint dengan `|` dan `~`)

Membatasi tipe yang boleh dipakai (Modul 6):
```go
type Number interface { ~int | ~int64 | ~float64 }  // union + underlying (~)
func Sum[T Number](s []T) T { ... }
```
`~int` = "int atau tipe apa pun ber-underlying int". Constraint juga bisa mensyaratkan **method** (interface biasa) — menggabungkan generics + interface.

## 4. Functional Options

Pola idiomatik untuk konstruktor dengan banyak parameter opsional:
```go
type Option func(*Server)
func WithPort(p int) Option { return func(s *Server) { s.Port = p } }
srv := NewServer(WithPort(9090), WithTLS())  // sisanya default
```
Lebih baik dari struct config besar (mudah ditambah tanpa breaking) atau konstruktor bertumpuk. Dipakai di banyak library (gRPC, zap, dll).

## Kapan pakai generics vs interface?
| Situasi | Pilih |
|---------|-------|
| Wadah/algoritma identik lintas tipe | **generics** (Set, Map/Filter, iterator) |
| Perilaku berbeda per tipe | **interface** (method) |
| Butuh keduanya | constraint interface + type param |

Jangan generik-kan sesuatu yang cuma dipakai satu tipe (over-engineering).

## Kapan & Di Mana Dipakai
- Library utilitas koleksi, pipeline transformasi data (iterator), konstruktor fleksibel (options), wadah reusable.

## Latihan
1. Tambah method `Set.Difference` dan `Set.Filter(pred)` yang mengembalikan Set baru.
2. Tulis iterator `Take[T](seq, n)` yang mengambil n elemen pertama.
3. Tulis iterator `Zip[A,B]` menggabungkan dua urutan (pakai `iter.Seq2`).
4. Tambah `WithLogger` option + validasi di `NewServer` (return error).
5. Ganti Map/Filter Modul 6 (yang mengembalikan slice) dengan versi iterator lazy ini & bandingkan.

## ✅ Solusi Latihan (Pembahasan)

1. **`Set.Difference` & `Set.Filter`** —
   ```go
   func (s Set[T]) Difference(o Set[T]) Set[T] { r:=NewSet[T](); for v:=range s { if !o.Has(v){ r.Add(v) } }; return r }
   func (s Set[T]) Filter(pred func(T) bool) Set[T] { r:=NewSet[T](); for v:=range s { if pred(v){ r.Add(v) } }; return r }
   ```
2. **Iterator `Take[T]`** — bungkus `iter.Seq[T]`, hentikan setelah n:
   ```go
   func Take[T any](seq iter.Seq[T], n int) iter.Seq[T] {
       return func(yield func(T) bool) { i:=0; for v:=range seq { if i>=n||!yield(v){return}; i++ } }
   }
   ```
3. **`Zip[A,B]`** — pakai `iter.Seq2[A,B]`; tarik dari dua sumber paralel (butuh `next()` via `iter.Pull`). Berhenti saat salah satu habis.
4. **`WithLogger` + validasi** — functional option yang mengembalikan error; `NewServer` menerapkan semua option lalu validasi (mis. logger tak nil) → return `(*Server, error)`.
5. **Ganti Map/Filter Modul 6** — versi iterator **lazy**: tak mengalokasikan slice antara; elemen mengalir satu per satu. Bandingkan alokasi dengan versi slice (Modul 26) — lazy menang untuk pipeline panjang.

---

## 🚀 Teknik Advanced (Level Up)
> 💻 **Contoh runnable + komentar detail** untuk teknik di bawah ada di folder [`advanced/`](advanced). Jalankan: `go run ./43-advanced-generics/advanced`


- **Type sets lanjutan** — constraint sebagai interface berisi union & `~` (underlying type). Rancang constraint sesempit mungkin.
- **`iter.Seq[T]` / `iter.Seq2[K,V]` (Go 1.23)** — iterator lazy via **range-over-func**: `for v := range mySeq {}`. Bangun pipeline malas & hemat memori.
- **Struktur data generik** — `Set[T comparable]`, tree, cache type-safe tanpa `any`+assertion.
- **Functional options generik** — builder type-safe dengan opsi.
- **Phantom/marker types** — tipe untuk enforce invariant di compile-time (mis. unit terukur, state builder).
- **Batas inference** — kadang perlu instansiasi eksplisit; method tak bisa punya type param (taruh di tipe).
