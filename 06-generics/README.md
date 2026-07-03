# 06 — Generics

Jalankan:
```bash
go run ./06-generics
```

Generics (sejak Go 1.18) memungkinkan menulis fungsi/tipe yang bekerja untuk **banyak tipe** tanpa `any` + type assertion. Tetap **type-safe** dan dicek saat kompilasi.

## 1. Type parameter

```go
func Max[T cmp.Ordered](a, b T) T {
	if a > b { return a }
	return b
}
Max(3, 5)        // T di-infer = int
Max("a", "b")    // T di-infer = string
Max(1.5, 2.5)    // T = float64
```
`[T cmp.Ordered]` = parameter tipe `T` dengan **constraint** `cmp.Ordered`.

## 2. Constraint

Constraint = interface yang membatasi tipe apa yang boleh dipakai:
- **`any`** — tipe apa saja (tanpa batasan operasi).
- **`comparable`** — tipe yang bisa `==`/`!=` (untuk key map, dll).
- **`cmp.Ordered`** (paket `cmp`, Go 1.21+) — tipe yang mendukung `< > <= >=` (angka & string).
- **Custom** dengan union `|` dan `~` (underlying type):
  ```go
  type Number interface {
  	~int | ~int64 | ~float64
  }
  ```
  `~int` berarti "int atau tipe apa pun yang underlying-nya int" (mis. `type Celsius int`).

## 3. Tipe generik (struct)

```go
type Stack[T any] struct { items []T }
func (s *Stack[T]) Push(v T) { s.items = append(s.items, v) }
func (s *Stack[T]) Pop() (T, bool) { ... }
```

## 4. Type inference

Sering kamu **tak perlu** menulis `[T]` — Go menyimpulkannya dari argumen. Tulis eksplisit hanya bila ambigu: `MakeSlice[int]()`.

## 5. Paket standar berbasis generics (Go 1.21+)

- **`slices`** — `slices.Sort`, `slices.Contains`, `slices.Index`, `slices.Max`, `slices.SortFunc`, dll.
- **`maps`** — `maps.Keys`, `maps.Values` (mengembalikan iterator di Go 1.23+).
- **`cmp`** — `cmp.Ordered`, `cmp.Compare`, `cmp.Or`.
- **builtin** `min(a, b)` & `max(a, b)` (sejak Go 1.21) untuk kasus sederhana.

## 6. Kapan JANGAN pakai generics

- Kalau cukup satu tipe → jangan digenerik-kan (over-engineering).
- Kalau butuh perilaku beda per tipe → pakai **interface** (method), bukan generics.
- Aturan: generics untuk **struktur data & algoritma** yang identik lintas tipe (container, Map/Filter/Reduce).

## Latihan
1. Tulis `Map[T, U any](s []T, f func(T) U) []U` dan pakai untuk mengubah `[]int` jadi `[]string`.
2. Tulis `Filter[T any](s []T, keep func(T) bool) []T`.
3. Tulis `Reduce[T, U any](s []T, init U, f func(U, T) U) U`; pakai untuk menjumlah `[]int`.
4. Buat `Number` constraint (union `~int | ~float64`) dan `Sum[T Number](s []T) T`.
5. Buat tipe generik `Pair[K comparable, V any]` dengan method `Swap()`.

Kerjakan di `06-generics/jawaban-saya/`, lalu minta saya review.

## Kapan & Di Mana Dipakai (Studi Kasus Nyata)

| Konsep | Cara pakai di dunia nyata | Contoh kasus |
|--------|---------------------------|--------------|
| **Container generik** | Struktur data yang identik lintas tipe | `Stack[T]`, `Queue[T]`, `Set[T]`, cache `Cache[K,V]` |
| **Map/Filter/Reduce** | Transformasi koleksi yang ringkas | Ubah `[]User` → `[]UserDTO`, filter aktif, jumlahkan total |
| **Constraint numerik** | Utilitas matematis lintas tipe angka | `Sum`, `Average`, `Clamp` untuk int & float sekaligus |
| **`slices` / `maps` / `cmp`** | Operasi koleksi harian tanpa boilerplate | `slices.Sort`, `slices.Contains`, `slices.SortFunc` |
| **Repository generik** | CRUD dasar dipakai ulang untuk banyak entity | `Repository[T]` dengan `GetByID`, `Save`, `Delete` |

**Contoh nyata — transformasi DTO (sangat umum di REST API):**
```go
type User struct { ID int; Name, Password string }
type UserDTO struct { ID int; Name string } // tanpa Password

// Ubah entity domain -> DTO aman untuk response, satu baris:
dtos := Map(users, func(u User) UserDTO {
    return UserDTO{ID: u.ID, Name: u.Name}
})
```

**Contoh nyata — repository generik (hemat kode untuk banyak entity):**
```go
type Entity interface { GetID() int }

type Repository[T Entity] struct { data map[int]T }

func (r *Repository[T]) Save(e T)  { r.data[e.GetID()] = e }
func (r *Repository[T]) Get(id int) (T, bool) { v, ok := r.data[id]; return v, ok }
// -> Repository[User], Repository[Product] tanpa menulis ulang.
```

**⚠️ Jangan berlebihan:** kalau cuma butuh SATU tipe, atau butuh perilaku berbeda per tipe (→ pakai **interface**/method), jangan pakai generics. Generics ideal untuk **wadah data & algoritma yang benar-benar sama** lintas tipe.

**Cocok dipakai saat:** membuat utility/library koleksi, transformasi data (mapping entity↔DTO), dan struktur data reusable. Untuk logika bisnis yang beragam per tipe, interface tetap pilihan utama.
