# 37 — Advanced Testing (Fuzzing, Integration, Load)

Lanjutan Modul 8. Tiga teknik testing tingkat lanjut: **fuzzing** (temukan bug otomatis), **integration testing** (uji dengan dependency nyata), dan **load testing** (uji beban).

Jalankan:
```bash
go test ./37-advanced-testing                                  # test biasa + seed corpus
go test -fuzz FuzzReverse -fuzztime 15s ./37-advanced-testing  # fuzzing sungguhan
```

## 1. Fuzzing (bawaan Go 1.18+)

Fuzzer membangkitkan **input acak** untuk menemukan crash/bug yang tak terpikirkan.
```go
func FuzzReverse(f *testing.F) {
    f.Add("halo")          // seed corpus (contoh awal)
    f.Fuzz(func(t *testing.T, s string) {
        // INVARIAN yang harus selalu benar untuk input APA PUN:
        if utf8.ValidString(s) && Reverse(Reverse(s)) != s {
            t.Errorf("round-trip gagal: %q", s)
        }
    })
}
```

### 🎯 Fuzzing menemukan bug NYATA di modul ini
Versi awal menegaskan `Reverse(Reverse(s)) == s` untuk **semua** input. Fuzzer menemukan input **UTF-8 tak valid** yang melanggarnya: `[]rune(s)` bersifat **lossy** — mengganti byte rusak dengan `U+FFFD`, jadi round-trip tak sama.
- **Pelajaran:** invariant kita terlalu kuat; round-trip hanya berlaku untuk UTF-8 valid.
- Input pemicu otomatis tersimpan di `testdata/fuzz/FuzzReverse/` → jadi **regression test permanen** (ikut `go test` selamanya).

Ini kekuatan fuzzing: menemukan edge case yang tak akan kamu tulis manual.

### Cara pakai
```bash
go test -fuzz FuzzXxx -fuzztime 30s ./paket   # cari bug baru
go test ./paket                                # replay corpus (cepat, di CI)
```
Kandidat fuzzing: parser, encoder/decoder, validasi input, apa pun yang menerima data dari luar.

## 2. Integration Testing (testcontainers)

Unit test pakai mock (Modul 8). **Integration test** pakai dependency **nyata** (DB, Redis, Kafka) di container sementara — otomatis start/stop.
```go
pg, _ := postgres.Run(ctx, "postgres:16", postgres.WithDatabase("test"))
defer pg.Terminate(ctx)
dsn, _ := pg.ConnectionString(ctx)
db, _ := sql.Open("pgx", dsn)   // uji terhadap Postgres SUNGGUHAN
```
Library: [`testcontainers-go`](https://github.com/testcontainers/testcontainers-go). Butuh **Docker**. Test di `integration_test.go` **di-skip** kecuali `RUN_INTEGRATION=1`.
```bash
RUN_INTEGRATION=1 go test -run TestWithRealDatabase ./37-advanced-testing
```
Keunggulan: menangkap bug yang mock sembunyikan (dialek SQL, tipe kolom, constraint) — bersih & reproducible di CI.

## 3. Load Testing (k6)

Uji **berapa banyak beban** yang sistem tahan. Pakai [k6](https://k6.io) (`loadtest.js`):
```bash
k6 run 37-advanced-testing/loadtest.js
```
```js
export const options = {
  stages: [{ duration: "10s", target: 50 }, ...],  // naikkan ke 50 user
  thresholds: { http_req_duration: ["p(95)<500"] }, // 95% < 500ms
};
```
Ukur: throughput (req/s), latensi p95/p99, error rate. Kombinasikan dengan profiling (Modul 26) untuk temukan bottleneck. Alternatif Go: [vegeta](https://github.com/tsenart/vegeta).

## Piramida testing
```
        /\        E2E (sedikit, lambat, mahal)
       /  \       Integration (testcontainers) — sedang
      /____\      Unit (banyak, cepat) — mayoritas
```
Tambahan lintas piramida: **fuzzing** (temukan edge case) & **load** (validasi performa).

## Teknik lain (sekilas)
- **Golden files**: bandingkan output dengan file referensi (`-update` untuk regenerasi).
- **Property-based testing**: nyatakan properti yang harus selalu benar (fuzzing adalah bentuknya).
- **Mutation testing**: ukur kualitas test dengan menyisipkan bug buatan.
- **Contract testing** (Pact): jamin kontrak antar service.

## Kapan & Di Mana Dipakai
- Fuzzing: parser, protokol, input dari user/jaringan.
- Integration: repository/DB, klien service eksternal.
- Load: sebelum rilis fitur di jalur panas, kapasitas planning.

## Latihan
1. Jalankan `go test -fuzz FuzzParseRange -fuzztime 20s` — apakah menemukan sesuatu?
2. Tulis fuzz test untuk sebuah fungsi di modul lain (mis. JSON parser).
3. Buat integration test nyata dengan testcontainers + Postgres (butuh Docker).
4. Jalankan k6 terhadap server Modul 15, catat p95 & error rate.
5. Tambah golden file test untuk output yang kompleks.

## ✅ Solusi Latihan (Pembahasan)

1. **Jalankan fuzzer** — `go test -fuzz FuzzParseRange -fuzztime 20s ./37-advanced-testing`. Bila menemukan input yang crash/melanggar invariant, Go menyimpannya di `testdata/fuzz/` sebagai corpus regresi.
2. **Fuzz fungsi lain** — mis. fungsi parse JSON/URL: `f.Fuzz(func(t, s string){ ParseX(s) })` — pastikan tak panic pada input acak. (Modul 2 reverse: cek `utf8.ValidString` sebelum invariant `reverse(reverse(s))==s`.)
3. **testcontainers + Postgres** — `postgres.RunContainer(ctx)` memberi DB nyata sekali pakai; test skip otomatis bila Docker tak ada (`testing.Short()` / cek env). Realistis tapi lebih lambat.
4. **k6** — skrip JS `http.get('http://localhost:8080/...')`; `k6 run script.js` → laporan p95 latency & error rate. Load test terhadap server Modul 15.
5. **Golden file** — simpan output diharapkan di `testdata/x.golden`; test bandingkan output aktual, flag `-update` untuk regenerasi. Cocok untuk output besar/kompleks (HTML, JSON).

---

## 🚀 Teknik Advanced (Level Up)
> 💻 **Contoh runnable + komentar detail** untuk teknik di bawah ada di folder [`advanced/`](advanced). Jalankan: `go test -v ./37-advanced-testing/advanced`


- **Native fuzzing** — `go test -fuzz`; corpus di `testdata/fuzz`; simpan seed penting. Fuzzer nemu bug UTF-8 nyata di repo ini → invariant diperbaiki.
- **Property-based testing** — uji invariant untuk banyak input acak (mis. `Reverse(Reverse(x))==x`) alih-alih contoh tunggal.
- **Testcontainers** — jalankan DB/broker nyata dalam container untuk test integrasi; auto-skip jika Docker tak ada.
- **Load testing (k6)** — verifikasi throughput/latency SLO; masuk CI untuk regresi performa.
- **Golden files & snapshot** — untuk output besar/berstruktur.
- **`synctest`** — uji kode berbasis waktu deterministik (Go 1.24). Selalu `-race` untuk kode konkuren.
