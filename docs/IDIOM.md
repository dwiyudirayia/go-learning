# 🎯 Idiom Go

"Idiom Go" = cara menulis kode yang **dianggap benar/lazim** oleh komunitas Go — bukan sekadar "yang penting jalan", tapi pola yang membuat kode mudah dibaca orang Go lain.

> Prinsip payung: **"Clear is better than clever."** Kalau ragu antara dua cara, pilih yang paling mudah dibaca.

---

## 1. Error sebagai nilai (bukan exception)

Go tak punya `try/catch`. Error adalah nilai yang dikembalikan & **wajib diperiksa**.

```go
f, err := os.Open("data.txt")
if err != nil {
    return fmt.Errorf("buka config: %w", err) // bungkus + tambah konteks
}
defer f.Close()
```

- `if err != nil` **langsung** setelah pemanggilan — jalur error terlihat jelas.
- `%w` **membungkus** (rantai bisa ditelusuri `errors.Is/As`); `%v` **menyegel** (sembunyikan detail).
- **Sentinel**: `var ErrNotFound = errors.New(...)` → cek `errors.Is(err, ErrNotFound)`.
- **Typed error**: `errors.As(err, &ve)` untuk ambil struct + datanya.
- **Multi-error**: `errors.Join(e1, e2)`.

**Anti-pola:** `return err` telanjang untuk error yang butuh konteks. **Idiomatik:** bungkus dengan pesan lokasi.

📍 `05-errors/advanced`

---

## 2. Accept interfaces, return structs

Interface dipenuhi **implisit** (tanpa `implements`). Buat interface **sekecil mungkin**, definisikan di pihak yang **memakai**.

```go
func Simpan(w io.Writer, b []byte) error { ... } // terima interface sempit
func NewBuffer() *bytes.Buffer { ... }           // kembalikan tipe konkret
```

- Interface 1-method diakhiri `-er`: `Reader`, `Writer`, `Stringer`.
- Karena implisit, tipe apa pun ber-method `Read` otomatis jadi `io.Reader`.

📍 `04-interfaces/advanced`

---

## 3. Zero value harus berguna

Tipe idealnya langsung pakai **tanpa init**.

```go
var mu sync.Mutex    // siap Lock()
var buf bytes.Buffer // siap Write()
var s []int          // nil slice — append tetap jalan
s = append(s, 1)
```

`nil` map boleh **dibaca** (`m["x"]` → zero value) tapi **tak boleh ditulis** (panic).

📍 `02-collections/advanced`, `07-concurrency/advanced`

---

## 4. `defer` untuk cleanup yang dijamin

```go
tx, _ := db.Begin()
defer tx.Rollback() // no-op setelah Commit -> jaring pengaman
return tx.Commit()
```

- Argumen `defer` **dievaluasi saat baris ditulis**, **dieksekusi saat fungsi keluar** (LIFO).
- Kombinasi **named return + recover** mengubah panic jadi error di boundary.

📍 `01-basics/advanced`, `05-errors/advanced`

---

## 5. Komposisi > pewarisan (embedding)

```go
type Person struct{ Name string }
func (p Person) Greet() string { return "Halo " + p.Name }

type Employee struct {
    Person     // embedded -> promoted method
    Salary int
}
emp.Greet() // dari Person; bisa di-override di Employee
```

📍 `03-structs-methods/advanced`

---

## 6. Value vs pointer receiver

```go
func (c Counter) Baca() int { return c.n } // value: baca-saja (dapat salinan)
func (c *Counter) Naik()    { c.n++ }       // pointer: mengubah state
```

**Jebakan:** method pointer-receiver hanya di *method set* `*T`. `Counter{}` (nilai) bisa **tak memenuhi** interface yang butuh `Naik()`. **Aturan:** konsisten — kalau satu pointer, semua pointer.

📍 `03-structs-methods/advanced`

---

## 7. Konkurensi: berbagi dengan berkomunikasi

> "Don't communicate by sharing memory; share memory by communicating."

```go
ch := make(chan int)
go func() { ch <- kerja() }()
hasil := <-ch
```

- **`context.Context` = parameter PERTAMA** untuk operasi yang bisa lama/dibatalkan.
- **`sync.WaitGroup`** untuk menunggu banyak goroutine.
- **Tiap goroutine wajib punya jalan keluar** (`ctx.Done()`/channel tutup) → hindari *leak*.
- **Selalu `go test -race`.**

📍 `07-concurrency/advanced`, `38-concurrency-advanced/advanced`

---

## 8. `iota` untuk enum + `Stringer`

```go
type Level int
const (
    Debug Level = iota // 0
    Info               // 1
    Warn               // 2
)
func (l Level) String() string { /* "DEBUG"/"INFO"/... */ }
```

- Enum tanpa angka manual. Untuk flag: `1 << iota` (bitmask).
- Implement `String()` → otomatis dipakai `fmt`/`%v`.

📍 `01-basics/advanced`, `04-interfaces/advanced`

---

## 9. Konvensi penamaan (aksen bahasa Go)

| Aturan | Contoh |
|--------|--------|
| Ekspor = **Kapital**, privat = kecil | `User` / `user` |
| Nama pendek untuk scope pendek | `i`, `r`, `buf`, `ctx` |
| **Tanpa `Get`** pada getter | `u.Name()` |
| Sentinel error diawali `Err` | `ErrNotFound` |
| Interface 1-method `-er` | `Reader` |
| Receiver singkat & konsisten | `func (s *Store)` |
| Package: singkat, lowercase, tanpa `_` | `http`, `strconv` |

`gofmt` menyeragamkan format; `go vet` menangkap kesalahan umum — **jalankan selalu**.

---

## 10. Slice/map sehari-hari

```go
v, ok := m["k"]               // comma-ok: cek keberadaan
for i, x := range s { ... }   // range, bukan index manual
out := make([]T, 0, n)        // prealokasi cap
s = append(s[:i], s[i+1:]...) // hapus elemen (urutan terjaga)
```

**Jebakan:** slice **berbagi backing array**; `append` bisa menimpa. `s[a:b:c]` mematok `cap` agar aman.

📍 `02-collections/advanced`

---

## 11. Konstruktor `NewXxx` + functional options

```go
type Server struct{ host string; port int }
type Option func(*Server)

func WithPort(p int) Option { return func(s *Server) { s.port = p } }

func NewServer(opts ...Option) *Server {
    s := &Server{host: "localhost", port: 8080} // default masuk akal
    for _, o := range opts { o(s) }
    return s
}
NewServer(WithPort(9090))
```

Menghindari konstruktor berparameter banyak; menambah opsi tak memecah caller lama.

📍 `10-project-layout/advanced`, `43-advanced-generics/advanced`

---

## Bacaan lanjut

- Jebakan konkret → [PITFALLS.md](PITFALLS.md)
- Sintaks cepat → [CHEATSHEET.md](CHEATSHEET.md)
- Konkurensi mendalam → [CONCURRENCY.md](CONCURRENCY.md)
- *Effective Go* & *Go Code Review Comments* (dokumen resmi) adalah sumber idiom kanonik.
