# 🧪 Strategi Test Go

Cara menguji kode Go secara idiomatik & efektif. Mendalam di modul **08-testing** & **37-advanced-testing**.

> Filosofi Go: testing itu **bawaan bahasa** (paket `testing` + perintah `go test`), bukan framework pihak ketiga. Tak ada `assert` bawaan — pakai `if got != want { t.Errorf(...) }`. Sederhana & eksplisit.

---

## 0. Apa yang diuji (dan seberapa banyak)

**Uji PERILAKU, bukan implementasi.** Test yang bagus bertahan saat refactor internal.

**Piramida test** (banyak → sedikit):
1. **Unit** — fungsi/tipe murni, cepat, tanpa I/O. Mayoritas test di sini.
2. **Integrasi** — dengan DB/broker nyata (auto-skip tanpa infra, lihat §11).
3. **End-to-end** — seluruh alur lewat API (mis. `app.Test`).

Prioritaskan menguji: aturan bisnis, boundary/edge case, jalur error, invariant. Jangan uji: getter sepele, kode pihak ketiga.

---

## 1. Anatomi & aturan dasar

```go
// file: kalkulator_test.go  (akhiran _test.go WAJIB)
package kalkulator

import "testing"

func TestTambah(t *testing.T) {        // Test<Nama>, terima *testing.T
    got := Tambah(2, 3)
    if got != 5 {
        t.Errorf("Tambah(2,3) = %d, mau %d", got, 5) // Errorf: catat, lanjut
    }
}
```

- `t.Errorf` → tandai gagal, **lanjut**. `t.Fatalf` → gagal & **hentikan test ini** (untuk prasyarat).
- Format pesan idiomatik: `"NamaFungsi(input) = got, mau want"`.
- Jalankan: `go test ./...`, `go test -v ./pkg`, `go test -run TestTambah ./pkg`.

**Dua pilihan package:**
- `package kalkulator` → **white-box**: akses simbol internal (privat).
- `package kalkulator_test` → **black-box**: hanya API publik (menguji dari sudut pandang pengguna). Boleh dua-duanya berdampingan.

---

## 2. Table-driven test (idiom WAJIB Go)

Satu daftar kasus, satu loop. Menambah kasus = menambah baris.

```go
func TestIsPalindrome(t *testing.T) {
    kasus := []struct {
        nama string
        in   string
        want bool
    }{
        {"kosong", "", true},
        {"unicode", "katak", true},
        {"bukan", "golang", false},
    }
    for _, k := range kasus {
        t.Run(k.nama, func(t *testing.T) {   // subtest: muncul terpisah, bisa di-filter
            t.Parallel()                      // subtest paralel (opsional)
            if got := IsPalindrome(k.in); got != k.want {
                t.Errorf("IsPalindrome(%q) = %v, mau %v", k.in, got, k.want)
            }
        })
    }
}
```

- `t.Run(nama, ...)` → subtest bernama (`go test -run TestIsPalindrome/unicode`).
- `t.Parallel()` → jalankan paralel; mempercepat & menyingkap race.

📍 `08-testing/advanced/palindrome_test.go`

---

## 3. Helper: `t.Helper`, `t.Cleanup`, `t.TempDir`

```go
func harusValid(t *testing.T, in string) {
    t.Helper() // laporan gagal menunjuk PEMANGGIL, bukan baris ini
    if !Valid(in) {
        t.Errorf("%q seharusnya valid", in)
    }
}

func TestFile(t *testing.T) {
    dir := t.TempDir()          // folder unik, auto-hapus saat test selesai
    t.Cleanup(func() { /* ... */ }) // pembersihan (LIFO), lebih rapi dari defer tersebar
    _ = dir
}
```

---

## 4. Test double via interface (mock/stub/fake)

Go tak butuh library mock — cukup **interface + implementasi palsu**.

```go
// Kode bergantung pada interface sempit:
type UserRepo interface{ FindByEmail(string) (User, error) }

// Test mengisi dengan fake in-memory:
type fakeRepo struct{ users map[string]User }
func (f fakeRepo) FindByEmail(e string) (User, error) {
    u, ok := f.users[e]
    if !ok { return User{}, ErrNotFound }
    return u, nil
}

func TestLogin(t *testing.T) {
    svc := NewService(fakeRepo{users: map[string]User{"a@b.com": {...}}})
    // ... uji svc tanpa DB nyata
}
```

Istilah: **stub** (balas nilai tetap), **fake** (implementasi ringan sungguhan, mis. in-memory), **mock** (verifikasi interaksi/pemanggilan), **spy** (catat pemanggilan). Di Go, *fake* & *stub* paling umum.

📍 `15-studi-kasus-rest/advanced` (mock repo), `40-llm-integration/advanced` (`MockChatter`)

---

## 5. Testable Example (dokumentasi + verifikasi)

Muncul di `go doc` **dan** diverifikasi terhadap stdout.

```go
func ExampleReverse() {
    fmt.Println(Reverse("halo"))
    // Output: olah
}
```

Bila output tak cocok → test gagal. Gunakan `// Unordered output:` untuk urutan bebas.

---

## 6. Benchmark

```go
func BenchmarkEncode(b *testing.B) {
    data := bytes.Repeat([]byte("x"), 1000)
    b.ReportAllocs()             // laporkan alokasi/op
    for i := 0; i < b.N; i++ {   // Go menyetel b.N otomatis
        _ = Encode(data)
    }
}
```

```bash
go test -bench . -benchmem ./pkg     # + memori (allocs/op, B/op)
go test -bench . -count=10 ./pkg | tee new.txt
benchstat old.txt new.txt            # bandingkan STATISTIK, bukan satu angka
```

Aturan: **profil dulu** (`docs/TOOLING.md` → pprof), optimasi berdasar data, jangan tebak.

📍 `26-profiling`, `08-testing/advanced` (`BenchmarkIsPalindrome`)

---

## 7. Fuzzing (Go 1.18+)

Uji **properti/invariant** dengan ribuan input acak, bukan contoh tunggal.

```go
func FuzzRoundTrip(f *testing.F) {
    f.Add([]byte("seed"))            // seed corpus
    f.Fuzz(func(t *testing.T, data []byte) {
        if !bytes.Equal(Decode(Encode(data)), data) {
            t.Errorf("round-trip rusak untuk %q", data)
        }
    })
}
```

```bash
go test -run x -fuzz FuzzRoundTrip -fuzztime 10s ./pkg
```

Fuzzer menyimpan input pemecah ke `testdata/fuzz/` sebagai regression test. (Di repo ini, fuzzer 08 & 37 pernah menemukan bug UTF-8 nyata.)

📍 `37-advanced-testing/advanced/rle_test.go`

---

## 8. Menguji HTTP

```go
// Handler tanpa server (unit):
func TestHandler(t *testing.T) {
    req := httptest.NewRequest("GET", "/users/42", nil)
    rec := httptest.NewRecorder()
    Handler(rec, req)
    if rec.Code != 200 { t.Fatalf("status %d", rec.Code) }
}

// Server sementara (integrasi ringan):
srv := httptest.NewServer(mux)
defer srv.Close()
resp, _ := http.Get(srv.URL + "/health")
```

Fiber: `app.Test(req)` menembus seluruh middleware+handler tanpa port.
📍 `12-http-stdlib/advanced`, `13-fiber/advanced`, `15-studi-kasus-rest/advanced`

---

## 9. Golden files (output besar/berstruktur)

Simpan output ekspektasi di `testdata/`, perbarui via flag.

```go
var update = flag.Bool("update", false, "perbarui golden files")

func TestRender(t *testing.T) {
    got := Render(input)
    golden := "testdata/render.golden"
    if *update { os.WriteFile(golden, got, 0o644); return }
    want, _ := os.ReadFile(golden)
    if !bytes.Equal(got, want) {
        t.Errorf("output beda dari golden (jalankan -update bila memang berubah)")
    }
}
```
Konvensi: folder `testdata/` diabaikan tools Go.

---

## 10. Menguji konkurensi

```bash
go test -race ./...     # detektor data race — WAJIB untuk kode konkuren
```

- `t.Parallel()` + `-race` menyingkap race pada state bersama.
- **`testing/synctest`** (Go 1.24+, eksperimental) → uji kode berbasis waktu **deterministik** tanpa `time.Sleep` nyata (fake clock).
- Untuk kebocoran goroutine: bandingkan `runtime.NumGoroutine()` sebelum/sesudah, atau `go.uber.org/goleak`.

📍 `07-concurrency`, `38-concurrency-advanced`

---

## 11. Integration test yang AUTO-SKIP (pola repo ini)

Test yang butuh infra nyata harus **otomatis di-skip** bila env tak ada → `go test ./...` tetap hijau di mesin mana pun.

```go
func TestIntegrasiPostgres(t *testing.T) {
    dsn := os.Getenv("POSTGRES_DSN")
    if dsn == "" {
        t.Skip("lewati: set POSTGRES_DSN (mis. via testcontainers) untuk test ini")
    }
    // ... buka koneksi nyata, uji end-to-end ...
}
```

- **testcontainers-go** menyalakan DB/broker asli dalam container untuk CI.
- Alternatif tanpa infra: **in-memory double** (miniredis, bufconn gRPC, NATS embedded, SQLite `:memory:`).

📍 `23-message-queue/integration_test.go`, `37-advanced-testing`, semua `NN/real-case/` (env-guarded)

---

## 12. Coverage

```bash
go test -cover ./...
go test -coverprofile=c.out ./... && go tool cover -func=c.out   # per-fungsi
go tool cover -html=c.out                                        # visual HTML
```

Coverage = **panduan, bukan target**. 100% cabang trivial tak berarti; utamakan cabang **berisiko** (error handling, boundary). Waspadai *coverage theater*.

---

## 13. Jebakan test

- **Menguji implementasi, bukan perilaku** → test rapuh saat refactor.
- **Test bergantung urutan/waktu** → flaky. Buat deterministik (seed `rand.New(rand.NewSource(1))`, fake clock).
- **Loop tanpa subtest** → kegagalan tak jelas kasus mana. Pakai `t.Run`.
- **Lupa `t.Helper()`** → baris gagal menunjuk helper, bukan pemanggil.
- **State global bersama antar-test** → bocor. Isolasi (fresh fixture, `t.Cleanup`).
- **`-race` tak dijalankan** untuk kode konkuren → race lolos.

---

## Alur ringkas

```bash
go test ./...              # cepat, tiap simpan
go test -race ./...        # sebelum commit (kode konkuren)
go test -cover ./...       # cek cakupan
go test -bench . -benchmem # saat peduli performa
go test -run x -fuzz F -fuzztime 30s  # cari edge case
```
> Repo ini: `go test ./...` harus **hijau semua paket** sebelum menyatakan selesai.
