# 08 — Testing

Testing adalah **bawaan bahasa** di Go (paket `testing` + perintah `go test`). Tidak perlu framework eksternal untuk mulai.

Jalankan:
```bash
go test ./08-testing/...            # jalankan semua test
go test -v ./08-testing/...         # verbose (lihat tiap subtest)
go test -race ./08-testing/...      # + deteksi data race
go test -cover ./08-testing/...     # laporan coverage
go test -run TestApplyDiscount ./08-testing/...   # filter test tertentu
go test -bench . ./08-testing/...   # jalankan benchmark
```

## 1. Aturan dasar

- File test **berakhiran `_test.go`**, berada di package yang sama.
- Fungsi test: `func TestXxx(t *testing.T)` (nama diawali `Test`, huruf ke-5 kapital).
- Gagalkan test dengan `t.Errorf` (lanjut) atau `t.Fatalf` (stop test itu).
- Tidak ada `assert` bawaan — pola idiomatik: `if got != want { t.Errorf(...) }`.

```go
func TestApplyDiscount(t *testing.T) {
	got, _ := ApplyDiscount(1000, 10)
	if got != 900 {
		t.Errorf("ApplyDiscount(1000,10) = %d; want 900", got)
	}
}
```

## 2. Table-driven test (idiom WAJIB di Go)

Satu test, banyak kasus, dalam sebuah tabel. Pakai `t.Run` untuk **subtest** (nama & isolasi):
```go
tests := []struct {
	name    string
	price   int
	pct     int
	want    int
	wantErr bool
}{
	{"diskon 10%", 1000, 10, 900, false},
	{"pct invalid", 1000, 150, 0, true},
}
for _, tc := range tests {
	t.Run(tc.name, func(t *testing.T) { ... })
}
```
Keuntungan: tambah kasus = tambah 1 baris; subtest yang gagal terlihat namanya.

## 3. Mocking dengan interface

Ganti dependency asli (DB, API) dengan implementasi palsu saat test — inilah alasan utama interface (Modul 4). Tidak perlu library mock; cukup buat struct yang mengimplementasikan interface.
```go
type fakeRepo struct{ products map[int]Product }
func (f fakeRepo) FindByID(id int) (Product, error) { ... }
// -> uji Catalog TANPA database nyata.
```

## 4. Benchmark

```go
func BenchmarkApplyDiscount(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ApplyDiscount(1000, 10)
	}
}
```
Jalankan: `go test -bench . -benchmem`. `b.N` diatur otomatis oleh runtime.

## 5. Example test (dokumentasi + verifikasi sekaligus)

Fungsi `ExampleXxx` dengan komentar `// Output:` diverifikasi saat `go test`, dan muncul di dokumentasi `go doc`:
```go
func ExampleApplyDiscount() {
	got, _ := ApplyDiscount(1000, 10)
	fmt.Println(got)
	// Output: 900
}
```

## 6. Helper & setup/teardown

- `t.Helper()` menandai fungsi sebagai helper (baris error menunjuk pemanggil).
- `t.Cleanup(fn)` mendaftarkan pembersihan.
- `TestMain(m *testing.M)` untuk setup/teardown global (mis. koneksi DB test).

## File di modul ini
- `store.go` — kode yang diuji (fungsi + service + interface).
- `store_test.go` — contoh lengkap: table-driven, subtest, mock, benchmark, example.

## Latihan (di `08-testing/latihan/`)
- `calc.go` sudah berisi beberapa fungsi (`Add`, `Divide`, `IsPalindrome`, `FizzBuzz`).
- Tulis test-nya di `calc_test.go` (kunci jawaban sudah disediakan — coba dulu sendiri!).
1. Table-driven test untuk `Divide` termasuk kasus pembagi nol (error).
2. Table-driven test untuk `IsPalindrome` (termasuk yang mengandu­ng huruf non-ASCII).
3. Test `FizzBuzz` untuk 3, 5, 15, dan angka biasa.
4. Tambah `BenchmarkFizzBuzz`.
5. Tambah `ExampleFizzBuzz` dengan `// Output:`.

Jalankan kunci jawaban: `go test -v ./08-testing/latihan/`.

## Kapan & Di Mana Dipakai (Studi Kasus Nyata)

| Konsep | Cara pakai di dunia nyata | Contoh kasus |
|--------|---------------------------|--------------|
| **Table-driven test** | Uji banyak kasus ringkas & mudah ditambah | Validasi input, kalkulasi harga/pajak, parsing |
| **Subtest `t.Run`** | Isolasi & nama tiap kasus; bisa `-run Nama/Kasus` | Debug kasus tertentu tanpa jalankan semua |
| **Mock via interface** | Uji logika tanpa DB/API/jaringan nyata | Test service dengan repo palsu; test handler tanpa DB |
| **`-race`** | Tangkap data race di test konkuren | CI wajib: `go test -race ./...` |
| **`-cover`** | Ukur bagian kode yang teruji | Target coverage tim; temukan cabang tak teruji |
| **Benchmark** | Ukur performa & alokasi memori | Bandingkan 2 implementasi; cegah regresi performa |
| **Example test** | Dokumentasi yang dijamin benar | Contoh pemakaian API di `go doc` yang tak pernah basi |
| **`httptest`** | Uji HTTP handler tanpa server sungguhan | (dipakai nanti di modul REST API) |

**Kenapa mock lewat interface itu kunci:** di backend, fungsi sering bergantung pada DB/API eksternal. Kalau test memanggil DB asli → lambat, rapuh, butuh setup. Dengan interface + implementasi palsu, test jadi **cepat, deterministik, dan jalan di mana saja** (termasuk CI).

**Alur kerja profesional:**
```bash
go test ./...            # sebelum commit
go test -race ./...      # sebelum merge (CI)
go test -cover ./...     # pantau cakupan
go test -run TestFoo ./... -v   # saat debug satu test
```

**Tips idiomatik:**
- Format pesan error: `"NamaFungsi(input) = got; want want"` — konsisten & mudah dibaca.
- `t.Fatalf` bila lanjut tak bermakna (nil pointer); `t.Errorf` bila ingin lihat semua kegagalan.
- Simpan test **di package yang sama** (`package store`) untuk akses internal, atau `package store_test` untuk menguji hanya API publik (black-box).

**Cocok dipakai saat:** SELALU. Test adalah jaring pengaman yang membuatmu berani me-refactor. Di Go, menulis test itu murah (bawaan bahasa) — jadikan kebiasaan sejak awal, apalagi menjelang REST API & microservices.
