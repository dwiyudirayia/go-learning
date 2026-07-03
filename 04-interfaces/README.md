# 04 — Interface

Jalankan:
```bash
go run ./04-interfaces
```

Interface di Go **beda dari Java/C#**. Tidak ada keyword `implements` — sebuah tipe memenuhi interface **secara otomatis** hanya dengan punya method-nya (structural / "duck typing").

## 1. Interface = kumpulan method

```go
type Shape interface {
	Area() float64
	Perimeter() float64
}
```
Tipe apa pun yang punya method `Area()` dan `Perimeter()` **otomatis** adalah `Shape`. Tidak perlu deklarasi apa pun. Ini memungkinkan kamu mendefinisikan interface **setelah** tipe konkretnya ada — bahkan untuk tipe dari package lain.

## 2. Prinsip idiomatik

- **Interface kecil lebih baik.** Idiom Go: interface 1–3 method. Contoh standar: `io.Reader`, `io.Writer`, `fmt.Stringer` (semua 1 method).
- **"Accept interfaces, return structs."** Fungsi menerima interface (fleksibel), tapi mengembalikan tipe konkret.
- **Definisikan interface di sisi yang MEMAKAI**, bukan di sisi yang mengimplementasikan.

## 3. Interface value = (tipe, nilai)

Sebuah nilai interface menyimpan **dua hal**: tipe dinamis + nilai. Ini penting untuk memahami `nil`.

### ⚠️ Jebakan "typed nil"
```go
var p *T = nil
var i interface{} = p
i == nil   // FALSE! i berisi (tipe=*T, nilai=nil), jadi bukan nil interface
```
Interface `nil` hanya jika **tipe dan nilai dua-duanya kosong**. Ini sumber bug klasik saat mengembalikan error.

## 4. Type assertion & type switch

```go
v, ok := i.(string)   // comma-ok: ok=false kalau bukan string (tidak panic)
s := i.(string)       // tanpa ok: PANIC kalau bukan string

switch v := i.(type) { // type switch
case string:
	...
case int:
	...
default:
	...
}
```

## 5. `any` (empty interface)

`any` (alias `interface{}` sejak Go 1.18) = "tipe apa saja", karena tak punya method sama sekali. Berguna untuk container generik pra-generics & `fmt.Println(...)`. **Tapi** pakai secukupnya — kehilangan type-safety.

## 6. Interface bawaan yang wajib kenal

- `error` → `Error() string`
- `fmt.Stringer` → `String() string` (dipakai `%v`/`Println`)
- `io.Writer` → `Write([]byte) (int, error)`
- `io.Reader` → `Read([]byte) (int, error)`
- `sort.Interface` → `Len`, `Less`, `Swap`

## Latihan
1. Buat interface `Shape{Area() float64}`. Implementasikan `Circle` & `Rectangle`. Buat fungsi `totalArea(shapes []Shape) float64`.
2. Buat fungsi `describe(i any)` yang memakai **type switch** untuk mencetak jenis & nilai (tangani `int`, `string`, `bool`, `Shape`, default).
3. Implementasikan `fmt.Stringer` pada sebuah tipe `Money` (mis. mencetak `Rp1.000`).
4. Buat tipe yang mengimplementasikan `io.Writer` sederhana (mis. menghitung total byte yang ditulis), lalu pakai dengan `fmt.Fprintf`.
5. Tunjukkan jebakan **typed nil**: buat fungsi yang mengembalikan `error` dari `*MyError` nil dan buktikan `err != nil`.

Kerjakan di `04-interfaces/jawaban-saya/`, lalu minta saya review.

## Kapan & Di Mana Dipakai (Studi Kasus Nyata)

| Konsep | Cara pakai di dunia nyata | Contoh kasus |
|--------|---------------------------|--------------|
| **Interface sebagai kontrak** | Dependency inversion → kode tak terikat implementasi | Service bergantung pada `UserRepository` (interface), bukan Postgres langsung |
| **Mocking untuk test** | Ganti implementasi asli dengan palsu saat test | `mockRepo` yang implement interface → unit test tanpa DB |
| **`io.Writer`/`io.Reader`** | Abstraksi I/O universal | Tulis ke file, HTTP response, buffer, gzip — semua sama |
| **`fmt.Stringer`** | Format tampilan custom untuk logging/debug | `Money` → `Rp1.500.000`, `Status` → nama |
| **Type switch** | Tangani data heterogen | Decode JSON ke `any`, event handler multi-tipe |
| **⚠️ Typed nil** | Hindari bug perbandingan error `== nil` | Selalu return `nil` literal, bukan pointer bertipe |

**Contoh nyata — interface untuk testability (INI alasan utama interface di backend Go):**
```go
// Definisikan di sisi yang MEMAKAI (service), bukan di repository.
type UserRepository interface {
    FindByID(id int) (*User, error)
    Save(u *User) error
}

// Produksi: implementasi Postgres.  Test: implementasi in-memory/mock.
type UserService struct { repo UserRepository }

// Saat test:
type fakeRepo struct{ users map[int]*User }
func (f *fakeRepo) FindByID(id int) (*User, error) { return f.users[id], nil }
// -> UserService bisa diuji TANPA database nyata.
```

**Kapan bikin interface?** Jangan buat interface "untuk jaga-jaga". Buat saat: (1) butuh mengganti implementasi (test/mock, ganti provider), (2) ada >1 implementasi nyata, (3) memisahkan lapisan (handler↔service↔repo). Kalau cuma 1 implementasi dan tak perlu di-mock → pakai struct langsung.

**Cocok dipakai saat:** membangun sistem berlapis yang **testable** dan **fleksibel** — inti arsitektur clean/hexagonal di Go. Ini akan sangat terpakai di modul REST API & microservices nanti.
