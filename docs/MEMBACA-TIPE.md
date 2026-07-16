# 🔎 MEMBACA-TIPE.md — "Fungsi ini butuh tipe yang tak kukenal, terus gimana?"

Situasi paling umum yang bikin bingung saat baru di Go:

```go
func json.NewDecoder(r io.Reader) *json.Decoder
func rate.NewLimiter(r rate.Limit, b int) *rate.Limiter
func fiber.New(config ...fiber.Config) *fiber.App
```

> *"Apa itu `io.Reader`? `rate.Limit`? Dari mana saya dapat nilainya?!"*

Dokumen ini memberi **alur baku 3 langkah** yang selalu berhasil — untuk tipe **apa pun**, dari library **mana pun**. Setelah terbiasa, langkah ini cuma makan waktu 10–30 detik.

🔍 **Analogi besar:** tipe parameter itu seperti **lubang pada mainan puzzle balok**. Tugasmu bukan menghafal semua balok di dunia, tapi tahu cara **melihat bentuk lubangnya** — begitu bentuknya jelas (bulat/kotak/segitiga), kamu langsung tahu balok mana yang muat.

---

## Langkah 1 — Intip definisi tipenya (5 detik)

Tiga cara, pilih yang paling dekat dengan jarimu:

| Cara | Bagaimana | Kapan |
|------|-----------|-------|
| **Hover** di VS Code | arahkan kursor ke nama tipe | tercepat, selalu mulai dari sini |
| **F12** (Go to Definition) | Ctrl+klik / F12 pada tipe | ingin lihat definisi + method lengkapnya |
| **`go doc`** di terminal | `go doc io.Reader`, `go doc fiber.Config` | di luar editor / ingin output ringkas |

Yang kamu cari **cuma satu kata** di baris pertama definisinya:

```go
type Reader interface { ... }   // ← kata kuncinya: interface
type Config struct { ... }      // ← kata kuncinya: struct
type Limit float64              // ← alias tipe dasar
type HandlerFunc func(...) ...  // ← tipe fungsi
```

Kata itu menentukan **seluruh strategi** di Langkah 2.

---

## Langkah 2 — Cocokkan dengan 5 kasus

### Kasus A: `interface` → JANGAN dibuat, cukup BERIKAN yang cocok

```go
type Reader interface {
	Read(p []byte) (n int, err error)
}
```

🔍 **Analogi:** interface itu **lowongan kerja**, bukan nama orang. `io.Reader` tidak minta barang bernama "Reader" — ia minta *siapa pun yang bisa `Read`*. Jadi pertanyaannya bukan *"bagaimana membuat io.Reader?"* melainkan *"barang apa yang sudah saya pegang yang bisa Read?"*

Jawabannya sering sudah ada di genggamanmu:

```go
json.NewDecoder(strings.NewReader(`{"a":1}`)) // string  → io.Reader
json.NewDecoder(resp.Body)                    // respons HTTP → io.Reader
json.NewDecoder(file)                         // os.File → io.Reader
json.NewDecoder(&bytes.Buffer{})              // buffer  → io.Reader
```

**Cara menemukan "siapa saja yang memenuhi interface X":**
1. Di VS Code: klik kanan nama interface → **Go to Implementations** (atau `Ctrl+F12`).
2. Di [pkg.go.dev](https://pkg.go.dev/io#Reader): tipe populer mencantumkan interface yang dipenuhinya di bagian dokumentasinya.
3. Interface kecil (1–2 method) → **bikin sendiri pun gampang**: cukup tulis tipe dengan method itu. Tak perlu deklarasi `implements` — kalau method-nya cocok, otomatis diterima (*duck typing* saat compile; lihat modul [04-interfaces](../04-interfaces/)).

> ⚠️ Interface yang berisi method **huruf kecil** (mis. `sealed()`) sengaja "disegel" — hanya package itu yang boleh mengimplementasikan. Cari tipe bawaan package-nya. (Contoh di `04-interfaces/advanced/`.)

### Kasus B: `struct` → cari konstruktor `NewXxx`, atau isi literalnya

```go
type Config struct {
	ErrorHandler func(*Ctx, error) error
	// ...banyak field lain
}
```

🔍 **Analogi:** struct itu **formulir**. Dua cara mendapatkannya: minta formulir yang **sudah diisi petugas** (konstruktor `NewXxx` — default aman, kamu tinggal pakai), atau **isi sendiri kolom yang kamu butuhkan** (struct literal — kolom lain otomatis zero value).

```go
// 1) Konstruktor — konvensi Go: fungsi New / NewXxx di package yang sama
rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})

// 2) Literal — isi hanya field yang perlu, sisanya zero value
app := fiber.New(fiber.Config{
	ErrorHandler: myHandler, // field lain? biarkan — sudah ada default
})
```

**Cara menemukan konstruktornya:** `go doc paket NamaStruct` — di bagian bawah output selalu terdaftar `func NewXxx(...) *Xxx` bila ada. Tak ada konstruktor? Berarti memang didesain untuk diisi literal (umum untuk struct `Config`/`Options`).

### Kasus C: alias tipe dasar (`type X int/float64/string`) → cari KONSTANTA atau FUNGSI PEMBUAT

```go
type Limit float64
```

🔍 **Analogi:** ini **mata uang khusus**. `rate.Limit` "cuma" float64, tapi package-nya menyediakan **money changer resmi** — dan memakai itu jauh lebih aman daripada mengarang angka sendiri.

`go doc golang.org/x/time/rate.Limit` menunjukkan langsung di bawah definisinya:

```
func Every(interval time.Duration) Limit
```

Itulah jawabannya:

```go
rate.NewLimiter(rate.Every(100*time.Millisecond), 2) // ✅ persis modul 27
rate.NewLimiter(10, 2)                               // juga sah (10 event/detik) — konversi implisit dari untyped constant
```

Pola serupa yang akan sering kamu temui:

| Tipe | Pembuatnya |
|------|-----------|
| `time.Duration` | `5 * time.Second`, `time.Millisecond` (konstanta) |
| `rate.Limit` | `rate.Every(...)`, `rate.Inf` |
| `slog.Level` | `slog.LevelInfo`, `slog.LevelDebug` (konstanta) |
| `codes.Code` (gRPC) | `codes.NotFound`, `codes.OK` (konstanta) |
| `fiber.StatusXxx` | konstanta `fiber.StatusNotFound` = 404 |

**Trik:** `go doc -all paket | grep "^const\|^func"` menampilkan semua konstanta & fungsi package sekaligus.

### Kasus D: tipe fungsi (`type X func(...)`) → tulis closure dengan signature sama

```go
type HandlerFunc func(*Ctx) error
```

🔍 **Analogi:** ini **colokan berbentuk fungsi**. Tak perlu mencari "barang bernama HandlerFunc" — cukup tulis fungsi anonim yang bentuknya pas, langsung nancap:

```go
app.Get("/users/:id", func(c *fiber.Ctx) error {   // ← closure, selesai
	return c.JSON(fiber.Map{"id": c.Params("id")})
})
```

Fungsi bernama dengan signature sama juga otomatis diterima — tak perlu konversi:

```go
func handleUser(c *fiber.Ctx) error { ... }
app.Get("/users/:id", handleUser) // ✅
```

### Kasus E: pointer (`*X`), slice (`[]X`), map, channel → selesaikan dulu X-nya

Bentuk-bentuk ini cuma **bungkus**. Tentukan dulu cara mendapatkan `X` (pakai Kasus A–D), lalu:

```go
*X   → &x                     // ambil alamatnya; konstruktor biasanya SUDAH mengembalikan pointer
[]X  → []X{x1, x2}            // literal slice
map[K]V → map[K]V{k: v}       // literal map
chan X  → make(chan X)        // selalu make
...X (variadic) → oper 0..n nilai X biasa, atau slice dengan xs...
```

> 💡 Kalau kamu lihat `config ...fiber.Config` (variadic), artinya parameter itu **opsional** — `fiber.New()` tanpa argumen pun sah.

---

## Langkah 3 — Masih ragu? Lihat contoh pemakaian nyata

Urutan paling efisien:

1. **Bagian Example di pkg.go.dev** — buka `pkg.go.dev/<import-path>`, scroll ke fungsi yang kamu pakai. Hampir semua package populer punya contoh yang bisa langsung disalin. (Contoh Example test yang diverifikasi compiler ada di modul [08-testing](../08-testing/).)
2. **Grep repo ini** — 48 modul = bank contoh berbahasa Indonesia:
   ```bash
   grep -rn "NewLimiter" --include="*.go" .      # siapa yang pernah memakainya?
   grep -rn "io.Reader" --include="*.go" . | head
   ```
3. **Baca test milik library-nya** — F12 ke fungsinya, lalu buka file `*_test.go` di sebelahnya. Test adalah dokumentasi pemakaian yang dijamin jalan.
4. **Tanya compiler** — tulis saja tebakannya lalu `go build`. Pesan error Go sangat spesifik:
   ```
   cannot use "halo" (untyped string constant) as io.Reader value:
   string does not implement io.Reader (missing method Read)
   ```
   Error ini bahkan **memberitahu method apa yang kurang** → petunjuk langsung ke Kasus A.

---

## Contoh bedah lengkap: dari bingung → jalan

Misal kamu mau memakai `json.NewDecoder` tapi tak kenal parameternya.

```
1. go doc encoding/json.NewDecoder
   → func NewDecoder(r io.Reader) *Decoder          # butuh io.Reader

2. go doc io.Reader
   → type Reader interface { Read(p []byte) (n int, err error) }
   → KASUS A (interface): berikan apa pun yang bisa Read.

3. Yang kupegang string? → grep/ingat: strings.NewReader mengubah string jadi Reader.

4. Hasil (dipakai persis di 09-stdlib/advanced/):
   dec := json.NewDecoder(strings.NewReader(`{"nama":"A"}`))
```

Satu lagi, tipe milik library pihak ketiga:

```
1. Hover pada parameter di fiber.New → config ...fiber.Config
   → variadic: boleh dikosongkan!  fiber.New() sah.

2. Mau custom? F12 ke fiber.Config → struct dengan puluhan field ber-default.
   → KASUS B: isi literal hanya field yang perlu:
   fiber.New(fiber.Config{ErrorHandler: ...})       # lihat 13-fiber/advanced/
```

---

## Cheat sheet — tempel di dinding 🧠

| Baris pertama definisi | Artinya | Cara dapat nilainya |
|---|---|---|
| `type X interface {...}` | lowongan kerja | berikan apa pun yang punya method-nya (cek stdlib dulu; atau bikin tipe sendiri) |
| `type X struct {...}` | formulir | `NewX(...)` bila ada; kalau tidak → literal `X{Field: ...}` |
| `type X int/float/string` | mata uang khusus | konstanta package (`slog.LevelInfo`) atau fungsi pembuat (`rate.Every`) |
| `type X func(...)` | colokan fungsi | tulis closure dengan signature sama |
| `*X`, `[]X`, `map`, `chan`, `...X` | bungkus | selesaikan X dulu, lalu `&x` / literal / `make` / oper apa adanya |

**Perintah yang wajib hafal:**

```bash
go doc paket.Tipe        # definisi + fungsi terkait (mis. go doc io.Reader)
go doc paket.Fungsi      # signature + dokumentasi (mis. go doc json.NewDecoder)
go doc -all paket | less # semua isi package
```

---

## Terkait

- Modul [04-interfaces](../04-interfaces/) — kenapa interface Go = *duck typing* saat compile.
- Modul [03-structs-methods](../03-structs-methods/) — method set: kapan butuh `*T` vs `T`.
- [GLOSSARY.md](GLOSSARY.md) — istilah (receiver, method set, zero value).
- [PITFALLS.md](PITFALLS.md) — jebakan terkait (typed-nil saat mengembalikan interface, dll).
- [TOOLING.md](TOOLING.md) — perintah `go doc` & teman-temannya lebih lengkap.
