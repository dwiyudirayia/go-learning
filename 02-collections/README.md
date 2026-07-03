# 02 — Collections: Array, Slice, Map, String & Rune

Jalankan:
```bash
go run ./02-collections
```

Ini modul dengan **paling banyak jebakan** buat pendatang dari bahasa lain. Pahami betul model memorinya.

## 1. Array vs Slice (INI PENTING)

- **Array** `[3]int` — ukuran **tetap** & bagian dari tipe. `[3]int` dan `[4]int` tipe berbeda. Array **di-copy saat di-assign / dikirim ke fungsi** (value semantics).
- **Slice** `[]int` — **view** ke atas sebuah array (backing array). Punya 3 komponen: **pointer** ke elemen awal, **len** (panjang), **cap** (kapasitas). Slice adalah yang **hampir selalu kamu pakai**.

```go
a := [3]int{1, 2, 3}   // array
s := []int{1, 2, 3}     // slice
s2 := make([]int, 0, 8) // slice: len=0, cap=8
```

## 2. `append` & jebakan backing array

`append` bisa mengembalikan slice dengan **backing array yang sama** (kalau cap masih cukup) atau **array baru** (kalau realokasi). Karena itu:

```go
s = append(s, x)   // SELALU tampung balik ke variabel; hasil append bisa pindah array
```

Jebakan klasik: dua slice berbagi backing array → ubah satu, yang lain ikut berubah. Modul ini mendemokannya.

## 3. `copy` & slice tricks

```go
copy(dst, src)                 // menyalin min(len(dst),len(src)) elemen
s = append(s[:i], s[i+1:]...)  // hapus elemen index i
```

## 4. Map

```go
m := map[string]int{"a": 1}
v, ok := m["x"]   // comma-ok: ok=false kalau key tak ada (v = zero value)
delete(m, "a")
```
- Iterasi map **urutannya acak** (disengaja!). Jangan andalkan urutan.
- **Map nil** boleh dibaca (hasil zero value) tapi **panic kalau ditulis**. Selalu `make` dulu sebelum menulis.

## 5. String, byte, rune (UTF-8)

- `string` di Go = **untaian byte read-only** berenkode UTF-8, **bukan** untaian karakter.
- `len(s)` = jumlah **byte**, bukan jumlah karakter.
- `s[i]` mengembalikan **byte** (`uint8`).
- `range s` meng-iterasi **rune** (code point), dengan index byte-nya.
- `rune` = alias `int32`, satu code point Unicode.

```go
s := "Halo, 世界"
len(s)            // 12 byte (bukan 8): "世" & "界" masing-masing 3 byte
[]rune(s)         // konversi ke slice rune -> len 8
for i, r := range s { ... }  // r bertipe rune
```

Untuk menyusun string efisien, pakai `strings.Builder` (bukan `+=` berulang).

## Latihan
1. Buat slice `[]int` berisi 1..10, lalu buat slice baru berisi hanya bilangan genap.
2. Tunjukkan jebakan backing array: bikin `b := a[:2]`, ubah `b[0]`, cetak `a`.
3. Hitung frekuensi tiap kata dari sebuah kalimat memakai `map[string]int`.
4. Hitung jumlah **karakter** (rune) vs jumlah **byte** dari string `"naïve café 世界"`.
5. Balik (reverse) sebuah string dengan benar (hati-hati multi-byte rune) — hasilnya harus tetap valid.

Kerjakan di `02-collections/jawaban-saya/` (buat sendiri), lalu minta saya review.

## Kapan & Di Mana Dipakai (Studi Kasus Nyata)

| Konsep | Cara pakai di dunia nyata | Contoh kasus |
|--------|---------------------------|--------------|
| **Slice `[]T`** | Struktur data default untuk "daftar apa pun" | List user dari query DB, batch item, hasil JSON array |
| **Preallocate `make([]T, 0, n)`** | Optimasi di jalur panas (hot path) | Bangun response 10.000 baris tanpa realokasi berulang |
| **`map[K]V`** | Lookup cepat O(1), index, cache | Cache in-memory `map[int]User`, hitung frekuensi, dedup |
| **`map[K]struct{}`** | Set (himpunan unik) hemat memori | Daftar ID unik, cek keanggotaan (`_, ada := set[x]`) |
| **`comma-ok`** | Cek key ada tanpa salah tafsir zero value | `if v, ok := cache[id]; ok { ... }` |
| **`copy` / clone** | Cegah pemanggil mengubah data internalmu | Repository mengembalikan salinan slice, bukan referensi asli |
| **byte vs rune** | Validasi & manipulasi teks yang benar | Batas panjang username (hitung rune!), potong teks, parsing |
| **`strings.Builder`** | Susun string besar efisien | Bangun query, render template, gabung ribuan baris log |

**Contoh nyata — cache & set:**
```go
// Cache lookup
if u, ok := userCache[id]; ok { return u } // hit

// Set: kumpulan ID unik
seen := map[int]struct{}{}
for _, id := range ids {
    if _, dup := seen[id]; dup { continue }
    seen[id] = struct{}{}
}
```

**⚠️ Jebakan produksi:** jangan kembalikan slice internal apa adanya dari sebuah struct — pemanggil bisa mengubah data di baliknya. Kembalikan hasil `copy` bila perlu aman.

**Cocok dipakai saat:** mengolah koleksi data (99% program). Map untuk lookup/caching/agregasi; slice untuk daftar berurut; rune saat menyentuh input teks user.
