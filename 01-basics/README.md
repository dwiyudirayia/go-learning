# 01 — Dasar Go & Idiom

Karena kamu sudah paham programming, fokus modul ini adalah **apa yang beda/khas di Go**, bukan konsep umum.

Jalankan:
```bash
go run ./01-basics
```

## Yang khas di Go (rangkuman)

1. **Deklarasi variabel dari kanan ke kiri** — tipe di belakang nama: `var x int`. Idiom yang paling sering: `:=` (short declaration) di dalam fungsi.
2. **Zero value** — setiap tipe punya nilai default: `int`→`0`, `string`→`""`, `bool`→`false`, pointer/slice/map→`nil`. Tidak ada "undefined".
3. **Tidak ada variabel/import yang tak terpakai** — itu **error kompilasi**, bukan warning. Ini disengaja untuk menjaga kode bersih.
4. **Multiple return values** — idiom `value, err := f()` dipakai di mana-mana, menggantikan exception.
5. **`defer`** — jadwalkan eksekusi saat fungsi selesai (LIFO). Untuk cleanup: tutup file, unlock mutex.
6. **Exported = huruf besar** — identifier diawali huruf kapital (`Println`) bersifat publik; huruf kecil = privat ke package. Tidak ada keyword `public`/`private`.
7. **`iota`** — generator konstanta berurutan, untuk bikin enum.
8. **Satu cara loop** — hanya ada `for`. `while` = `for kondisi {}`, infinite = `for {}`.

## Konsep penting

### Variabel & konstanta
```go
var a int = 10       // eksplisit
var b = 10           // tipe di-infer
c := 10              // short form (hanya di dalam fungsi)
const Pi = 3.14      // konstanta, dievaluasi saat kompilasi
```

### Tipe dasar
`bool`, `string`, `int`/`int8..64`, `uint`, `float32/64`, `complex`, `byte` (alias `uint8`), `rune` (alias `int32`).
Go **tidak melakukan konversi tipe otomatis** — `int` + `float64` harus dikonversi eksplisit: `float64(i) + f`.

### Fungsi
```go
func add(a, b int) int { return a + b }          // param bertipe sama boleh digabung
func divmod(a, b int) (int, int) { ... }          // multiple return
func split(sum int) (x, y int) { ... }            // named return
```

## Latihan
1. Buat fungsi `celsiusToFahrenheit(c float64) float64` dan cetak hasil untuk 0, 37, dan 100.
2. Buat konstanta hari dalam seminggu memakai `iota` (Senin=0 ... Minggu=6).
3. Tulis fungsi `minMax(nums ...int) (min, max int)` memakai variadic + named return.
4. Buat program yang pakai `defer` untuk mencetak "selesai" setelah fungsi `main` berjalan.

Kunci jawaban tersedia di `01-basics/latihan/solusi.go` (`go run ./01-basics/latihan`). Coba kerjakan dulu di `01-basics/jawaban-saya/` sebelum melihatnya.

## Kapan & Di Mana Dipakai (Studi Kasus Nyata)

| Konsep | Cara pakai di dunia nyata | Contoh kasus |
|--------|---------------------------|--------------|
| **`iota` (enum)** | Bikin status/tipe yang terbatas & jelas | `OrderStatus` (Pending, Paid, Shipped) di e-commerce; level log (Debug, Info, Error); role user |
| **`iota` + bit shift** | Flag/permission yang bisa digabung | Hak akses: `Read=1<<0, Write=1<<1, Admin=1<<2`, lalu `Read|Write` |
| **`defer`** | Cleanup yang **dijamin jalan** | `defer resp.Body.Close()` (HTTP), `defer rows.Close()` (DB), `defer mu.Unlock()`, `defer tx.Rollback()` |
| **Multiple return `(val, err)`** | Standar SEMUA fungsi yang bisa gagal | `user, err := repo.Find(id)`, `n, err := conn.Read(buf)` |
| **Named return + `defer`** | Ubah/log hasil sebelum return | Pola `recover` mengisi `err` (lihat Modul 5) |
| **`const`** | Nilai konfigurasi tetap saat kompilasi | Versi API, timeout default, nama header (`"Authorization"`) |

**Contoh nyata `defer` (HTTP handler):**
```go
resp, err := http.Get(url)
if err != nil { return err }
defer resp.Body.Close() // dijamin tertutup walau ada return/panic di bawah
```

**Cocok dipakai saat:** hampir SETIAP program Go — ini fondasi. `defer` khususnya wajib tiap kali kamu membuka resource (file, koneksi, lock).

---

## 🚀 Teknik Advanced (Level Up)
> 💻 **Contoh runnable + komentar detail** untuk semua teknik di bawah ada di folder [`advanced/`](advanced). Jalankan: `go run ./01-basics/advanced`


Materi modul ini fondasi. Ini teknik lanjutan yang tetap "dasar" tapi sering luput:

- **`iota` untuk enum & bitmask** — `type Level int; const (Debug Level = iota; Info; Warn)`. Untuk flag pakai geser bit: `const (Read = 1 << iota; Write; Exec)` lalu gabung `Read|Write`.
- **Konstanta *untyped* & presisi arbitrer** — `const Big = 1 << 62` boleh, evaluasi saat compile. Overflow konstanta ketahuan saat build, bukan runtime.
- **Named return values + `defer`** — return bernama bisa dimodifikasi di `defer` (dipakai untuk membungkus error). Hati-hati *naked return* di fungsi panjang: kurangi keterbacaan.
- **Labeled `break`/`continue`** — keluar dari loop bersarang: `Outer: for {...break Outer}`. Lebih bersih daripada flag boolean.
- **Variable shadowing** — `:=` di scope baru (mis. dalam `if err := ...; err != nil`) bikin variabel baru. Ini jebakan klasik; `go vet -vettool=shadow` bisa mendeteksi.
- **`switch` tanpa kondisi** = rantai `if-else` rapi; `fallthrough` untuk lanjut ke case berikut (jarang).
