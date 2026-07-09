# ⚠️ Jebakan Umum Go (Pitfalls)

Bug halus yang sering menimpa pendatang Go. Tiap entri: **masalah → contoh → solusi**.

---

## 1. Typed nil pada interface

Interface menyimpan `(tipe, nilai)`; ia `== nil` **hanya** jika keduanya nil.

```go
func f() error {
    var e *MyErr // nil pointer BERTIPE
    return e     // interface != nil !!
}
if f() != nil { /* MASUK, walau "tak ada error" */ }
```

**Solusi:** kembalikan `nil` literal untuk interface, jangan pointer nil bertipe.
📍 `04-interfaces/advanced`

---

## 2. Menulis ke nil map → panic

```go
var m map[string]int
m["a"] = 1 // panic: assignment to entry in nil map
```

**Solusi:** `m := make(map[string]int)` atau `map[string]int{}`. (Membaca nil map aman → zero value.)

---

## 3. Slice berbagi backing array (aliasing)

```go
a := []int{1, 2, 3, 4, 5}
b := a[:2]
b = append(b, 99) // menimpa a[2] jadi 99!
```

**Solusi:** three-index untuk patok cap → `b := a[:2:2]` (append memaksa array baru), atau `slices.Clone`.
📍 `02-collections/advanced`

---

## 4. `range` mengopi nilai

```go
for _, v := range items {
    v.Field = 1 // mengubah SALINAN, tak berefek ke items
}
```

**Solusi:** akses lewat indeks → `for i := range items { items[i].Field = 1 }`.

---

## 5. Loop variable capture (goroutine/closure)

Sebelum Go 1.22, variabel loop **dipakai bersama**:

```go
for _, v := range s {
    go func() { fmt.Println(v) }() // (pra-1.22) semua cetak nilai TERAKHIR
}
```

**Solusi:** Go **1.22+** memperbaiki ini (tiap iterasi variabel baru). Untuk kode lama: `v := v` di dalam loop. Repo ini Go 1.26 → sudah aman, tapi pahami polanya.

---

## 6. Pembagian integer

```go
c := 100
f := c * 9 / 5 + 32 // integer! 9/5 = 1
```

**Solusi:** pakai float → `float64(c)*9.0/5.0 + 32`.
📍 `01-basics`

---

## 7. `defer` di dalam loop

```go
for _, name := range files {
    f, _ := os.Open(name)
    defer f.Close() // TERTUNDA sampai fungsi keluar -> file menumpuk
}
```

**Solusi:** bungkus badan loop dalam fungsi/closure, atau `Close()` eksplisit tiap iterasi.

---

## 8. Goroutine leak

```go
ch := make(chan int) // unbuffered
go func() { ch <- 1 }() // blok selamanya jika tak ada yang menerima
return                  // goroutine bocor
```

**Solusi:** pastikan setiap goroutine punya jalan keluar (`ctx.Done()`, channel ditutup, atau buffer cukup). Uji `go test -race` & pantau `runtime.NumGoroutine()`.
📍 `07-concurrency/advanced`

---

## 9. Data race

```go
var n int
for i := 0; i < 100; i++ { go func() { n++ }() } // race!
```

**Solusi:** `sync.Mutex` atau `atomic.Int64`. **Selalu** `go test -race`. Race = bug walau "kelihatan benar".

---

## 10. Membandingkan float dengan `==`

```go
if 0.1+0.2 == 0.3 { ... } // false (presisi biner)
```

**Solusi:** bandingkan selisih terhadap epsilon: `math.Abs(a-b) < 1e-9`.

---

## 11. Field struct tak diekspor tak ter-marshal

```go
type User struct {
    Name string  // ✅ ter-marshal ke JSON
    age  int     // ❌ huruf kecil -> DIABAIKAN encoding/json
}
```

**Solusi:** ekspor field (huruf kapital) + tag → `Name string \`json:"name"\``.
📍 `09-stdlib/advanced`

---

## 12. `time.Timer`/`Ticker` bocor

```go
t := time.NewTicker(time.Second) // tak di-Stop -> resource menggantung
```

**Solusi:** `defer t.Stop()`.

---

## 13. Lupa cek error dari `http.Get` sebelum pakai `resp`

```go
resp, _ := http.Get(url)
defer resp.Body.Close() // panic bila err != nil (resp nil)
```

**Solusi:** cek `err` dulu. (`go vet` menandai ini.) Dan **selalu** set `http.Client{Timeout: ...}` — default tak ada timeout.

---

## 14. Shadowing dengan `:=`

```go
x := 1
if true {
    x := 2 // variabel BARU (shadow), perubahan tak bocor ke luar
    _ = x
}
// x tetap 1
```

**Solusi:** gunakan `=` bila ingin menimpa; `go vet -vettool=shadow` mendeteksi.

---

## 15. `append` mengembalikan slice baru

```go
append(s, 1) // hasilnya HARUS ditampung
s = append(s, 1) // benar
```

---

> **Aturan emas:** jalankan `gofmt`, `go vet`, dan `go test -race` sebelum menyatakan selesai. Banyak jebakan di atas ketahuan otomatis.
