# 05 — Error Handling

Jalankan:
```bash
go run ./05-errors
```

Filosofi Go: **error adalah nilai biasa**, bukan exception. Fungsi mengembalikan `error` sebagai nilai terakhir; pemanggil memeriksanya secara eksplisit dengan `if err != nil`. Ini membuat alur kegagalan **terlihat** di kode.

## 1. Interface `error`

```go
type error interface {
	Error() string
}
```
Membuat error:
```go
errors.New("pesan")
fmt.Errorf("gagal proses id %d", id)
```

## 2. Idiom pemeriksaan

```go
val, err := doSomething()
if err != nil {
	return fmt.Errorf("konteks tambahan: %w", err) // bungkus & teruskan
}
// lanjut pakai val
```
- Tangani error **sedini mungkin**, jangan ditumpuk.
- Tambahkan **konteks** saat meneruskan ke atas, jangan telan diam-diam.

## 3. Sentinel error

Error tetap yang bisa dibandingkan, dideklarasikan sebagai `var`:
```go
var ErrNotFound = errors.New("data tidak ditemukan")
```
Contoh nyata: `io.EOF`, `sql.ErrNoRows`.

## 4. Wrapping: `%w`, `errors.Is`, `errors.As` (INTI MODUL)

- Bungkus dengan `%w` agar error asal tetap bisa "dilihat" lewat rantai:
  ```go
  fmt.Errorf("query user: %w", ErrNotFound)
  ```
- **`errors.Is(err, target)`** — cek apakah di rantai ada sentinel tertentu.
- **`errors.As(err, &target)`** — ekstrak error bertipe tertentu dari rantai.
- **`errors.Join(e1, e2)`** (Go 1.20+) — gabungkan beberapa error jadi satu.

```go
if errors.Is(err, ErrNotFound) { ... }        // cocokkan sentinel
var ve *ValidationError
if errors.As(err, &ve) { ... ve.Field ... }    // ambil tipe konkret
```

## 5. Custom error type

Untuk membawa data tambahan (field, kode), buat tipe yang punya `Error() string`:
```go
type ValidationError struct {
	Field string
	Msg   string
}
func (e *ValidationError) Error() string { return e.Field + ": " + e.Msg }
```

## 6. panic & recover (JARANG dipakai)

- `panic` = kondisi tak terpulihkan (bug programmer). **Bukan** untuk error biasa.
- `recover` (hanya bermakna di dalam `defer`) menangkap panic agar program tak crash.
- Pola umum: library **tidak** boleh panic keluar; ubah panic jadi error di batas API.

```go
func aman() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("pulih dari panic: %v", r)
		}
	}()
	// ... kode yang mungkin panic
	return nil
}
```

**Aturan emas:** untuk kesalahan yang diharapkan (input salah, file tak ada, dsb) → **return error**. `panic` hanya untuk keadaan yang benar-benar mustahil / bug.

## Latihan
1. Buat `var ErrInsufficientFunds` dan fungsi `Withdraw` yang mengembalikannya; cek dengan `errors.Is`.
2. Buat `ValidationError{Field, Msg}` dan fungsi `validateUser` yang mengembalikannya; ekstrak dengan `errors.As` lalu cetak `Field`.
3. Buat fungsi berlapis: `repo -> service -> handler`, tiap lapis membungkus error dengan `%w` menambah konteks. Cetak rantai lengkap & cek sentinel di ujung dengan `errors.Is`.
4. Tulis fungsi `safeDivide(a, b int) (int, error)` yang memakai `recover` untuk menangkap panic pembagian nol dan mengubahnya jadi error.
5. Gabungkan beberapa error validasi dengan `errors.Join` dan tampilkan semuanya.

Kerjakan di `05-errors/jawaban-saya/`, lalu minta saya review.

## Kapan & Di Mana Dipakai (Studi Kasus Nyata)

| Konsep | Cara pakai di dunia nyata | Contoh kasus |
|--------|---------------------------|--------------|
| **Sentinel error** | Kondisi yang dikenali & ditangani khusus | `ErrNotFound` → HTTP 404, `ErrUnauthorized` → 401 |
| **Wrapping `%w`** | Bawa konteks lintas lapisan tanpa hilang asal | `repo → service → handler`, tiap lapis tambah info |
| **`errors.Is`** | Cabang keputusan berdasar jenis error | Map error ke status HTTP; putuskan retry |
| **`errors.As`** | Ambil detail dari error bertipe | `ValidationError` → balas 400 + nama field |
| **`errors.Join`** | Kumpulkan banyak kegagalan sekaligus | Validasi form: laporkan SEMUA field salah |
| **`recover` (middleware)** | Cegah 1 panic menjatuhkan seluruh server | Recovery middleware di HTTP server → balas 500, server tetap hidup |

**Contoh nyata — map error ke HTTP status (pola inti REST API):**
```go
func handleGetUser(w http.ResponseWriter, r *http.Request) {
    user, err := svc.GetUser(id)
    switch {
    case errors.Is(err, ErrNotFound):
        http.Error(w, "user tidak ada", http.StatusNotFound)   // 404
    case errors.As(err, &validationErr):
        http.Error(w, validationErr.Error(), http.StatusBadRequest) // 400
    case err != nil:
        http.Error(w, "server error", http.StatusInternalServerError) // 500
    default:
        json.NewEncoder(w).Encode(user) // 200
    }
}
```

**Contoh nyata — recovery middleware (wajib di server produksi):**
```go
func recoverMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if rec := recover(); rec != nil {
                log.Printf("panic: %v", rec)
                http.Error(w, "internal error", 500) // server TIDAK crash
            }
        }()
        next.ServeHTTP(w, r)
    })
}
```

**Cocok dipakai saat:** SEMUA kode produksi. Pola `sentinel + %w + errors.Is/As` adalah cara standar Go memetakan kegagalan domain ke respons API. `recover` khusus untuk batas terluar (middleware/worker) agar satu bug tak menjatuhkan proses.

---

## 🚀 Teknik Advanced (Level Up)
> 💻 **Contoh runnable + komentar detail** untuk semua teknik di bawah ada di folder [`advanced/`](advanced). Jalankan: `go run ./05-errors/advanced`


- **`errors.Join` (Go 1.20)** — gabung banyak error jadi satu; `errors.Is` menelusuri semuanya. Implementasi kustom via `Unwrap() []error`.
- **`%w` vs `%v`** — `%w` mempertahankan rantai (bisa `errors.Is/As`); `%v` "menyegel" jadi teks (sengaja menyembunyikan detail internal dari caller).
- **Sentinel vs typed error** — sentinel (`var ErrNotFound = errors.New(...)`) untuk kondisi tetap; typed error (struct) saat butuh data tambahan (`errors.As`).
- **`fs.ErrNotExist` bukan `os.IsNotExist`** — cek modern: `errors.Is(err, fs.ErrNotExist)`. `os.IsNotExist` lawas & tak menembus wrap.
- **panic/recover sebagai kontrol lokal** — boleh dalam satu paket (mis. parser rekursif), **jangan** bocor lintas API. Recover di boundary (middleware HTTP) untuk cegah crash server.
- **Stack trace** — stdlib tak simpan stack. Pakai `slog` + `runtime` source, atau `%+v` dari library error tertentu.
- **Jangan `if err != nil { return err }` telanjang** untuk error yang butuh konteks — bungkus: `return fmt.Errorf("baca config: %w", err)`.
