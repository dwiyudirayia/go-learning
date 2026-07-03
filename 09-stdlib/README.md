# 09 — Standard Library Penting

Jalankan:
```bash
go run ./09-stdlib
```

Go punya stdlib yang sangat lengkap — banyak hal tak butuh library eksternal. Modul ini fokus paket yang paling sering dipakai di backend: `encoding/json`, `time`, `io`, `os`, `net/http`.

## 1. `encoding/json`

```go
type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email,omitempty"` // dihilangkan bila kosong
	pass  string // huruf kecil = unexported = TIDAK ikut JSON
}

b, _ := json.Marshal(u)                 // struct -> []byte JSON
json.Unmarshal(b, &u)                   // JSON -> struct
b, _ := json.MarshalIndent(u, "", "  ") // JSON rapi (pretty)
```
- **Struct tag** `json:"nama"` menentukan nama field di JSON.
- **`omitempty`** menghilangkan field bila nilainya zero value.
- Hanya field **exported** (huruf besar) yang ikut serialisasi.
- JSON dinamis / tak dikenal → `map[string]any` atau `json.RawMessage`.
- Angka JSON tanpa tipe target masuk ke `float64` (hati-hati saat pakai `any`).

## 2. `time`

**Layout referensi Go itu unik**: `Mon Jan 2 15:04:05 MST 2006` (angka 1-2-3-4-5-6-7). Kamu memformat dengan "menulis ulang" tanggal acuan itu:
```go
t.Format("2006-01-02 15:04:05")     // -> 2026-07-01 15:04:05
time.Parse("2006-01-02", "2026-07-01")
time.Now(); t.Add(24 * time.Hour); t2.Sub(t1) // Duration
```
- `time.Duration` adalah tipe: `2 * time.Second`, `d.Seconds()`.
- Untuk perbandingan: `t.Before(u)`, `t.After(u)`.

## 3. `io`

Abstraksi stream universal: `io.Reader` (baca) & `io.Writer` (tulis). Banyak tipe memenuhinya (file, koneksi, buffer, HTTP body).
```go
io.Copy(dst, src)            // salin dari Reader ke Writer
var buf bytes.Buffer         // implementasi Reader+Writer di memori
r := strings.NewReader("hi") // string sebagai Reader
data, _ := io.ReadAll(r)     // baca habis sebuah Reader
```

## 4. `os`

```go
os.Args                     // argumen CLI ([]string)
os.Getenv("KEY"); os.Setenv // environment variable
os.ReadFile / os.WriteFile  // baca/tulis file sekali jalan
f, _ := os.CreateTemp("", "prefix-*") // file sementara
os.Exit(1)                  // keluar dgn kode (defer TIDAK jalan!)
```

## 5. `net/http`

**Client:**
```go
resp, _ := http.Get(url)
defer resp.Body.Close()
body, _ := io.ReadAll(resp.Body)
```
**Server:**
```go
mux := http.NewServeMux()
mux.HandleFunc("GET /users/{id}", handler) // routing method+path (Go 1.22+)
http.ListenAndServe(":8080", mux)
```
> Sejak Go 1.22, `ServeMux` bawaan sudah mendukung **method + path parameter** (`r.PathValue("id")`) — sering cukup tanpa framework untuk kasus sederhana.

Di modul ini server didemokan dengan `httptest` agar tidak nge-block.

## Latihan (di `09-stdlib/latihan/`)
1. Definisikan struct `Book{Title, Author string; Year int; Tags []string}` dengan tag JSON; marshal & unmarshal.
2. Parse tanggal `"2026-07-01"` lalu tambahkan 90 hari; format hasilnya `"02 Jan 2006"`.
3. Tulis fungsi `wordCount(r io.Reader) map[string]int` yang membaca dari Reader apa pun.
4. Tulis & baca kembali sebuah file JSON sementara berisi daftar `Book`.
5. Buat handler HTTP kecil yang mengembalikan JSON, uji dengan `httptest`.

Kunci jawaban: `go run ./09-stdlib/latihan`.

## Kapan & Di Mana Dipakai (Studi Kasus Nyata)

| Paket | Cara pakai di dunia nyata | Contoh kasus |
|-------|---------------------------|--------------|
| **`encoding/json`** | Serialisasi request/response API, config, cache | Body REST API, simpan config `.json`, komunikasi antar service |
| **`time`** | Timestamp, timeout, penjadwalan, TTL | `created_at`, token expiry, rate-limit window, cron |
| **`io`** | Stream data tanpa muat semua ke memori | Upload/download file besar, proxy body, kompresi |
| **`os`** | Baca config env, argumen, file | 12-factor config (`os.Getenv`), baca file kredensial |
| **`net/http`** | Client & server HTTP | Panggil API pihak ketiga; server sederhana tanpa framework |

**Contoh nyata — config dari environment (pola 12-factor, standar produksi):**
```go
port := os.Getenv("PORT")
if port == "" { port = "8080" } // default
dbURL := os.Getenv("DATABASE_URL")
```

**Contoh nyata — panggil API eksternal & decode JSON:**
```go
resp, err := http.Get("https://api.example.com/users/1")
if err != nil { return err }
defer resp.Body.Close()
var u User
if err := json.NewDecoder(resp.Body).Decode(&u); err != nil { return err }
```

**Catatan penting yang muncul di output:**
- `json.Marshal` **meng-escape** `&`, `<`, `>` jadi `&` dst (perlindungan HTML). Untuk mematikannya pakai `json.Encoder` + `SetEscapeHTML(false)`.
- Angka JSON yang di-`Unmarshal` ke `any` selalu jadi **`float64`** — sumber bug saat kamu harap `int`.
- `os.Exit` membuat **`defer` TIDAK jalan** — hati-hati bila ada cleanup.
- **Selalu `defer resp.Body.Close()`** setelah request berhasil, atau koneksi bocor.

**Kaitan ke depan:** `net/http` + `encoding/json` yang kamu lihat di sini adalah **fondasi** REST API. Framework (Gin/Echo) hanya membungkus ini agar lebih ringkas — memahami versi murninya membuatmu tak "buta" saat pakai framework.

**Cocok dipakai saat:** hampir semua backend. Untuk kasus sederhana, `net/http` bawaan (dengan routing method+path Go 1.22) sudah cukup tanpa framework apa pun.
