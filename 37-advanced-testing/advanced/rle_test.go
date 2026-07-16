package advanced

import (
	"bytes"
	"math/rand"
	"os"
	"testing"
)

// ===========================================================================
// 1. FUZZING — mesin mencari input yang MELANGGAR invariant
// ===========================================================================
// Invariant round-trip: Decode(Encode(x)) HARUS sama dengan x, untuk SEMUA x.
// Fuzzer mengoceh ribuan input acak (termasuk byte aneh, run panjang, kosong)
// mencari yang membuat invariant gagal. Bug UTF-8/boundary sering ketahuan di sini.
//
// 🔍 Analogi: invariant round-trip itu seperti MENERJEMAHKAN BOLAK-BALIK —
// kalimat apa pun yang diterjemahkan ke bahasa lain (Encode) lalu
// diterjemahkan kembali (Decode) harus utuh seperti semula. Fuzzer = penguji
// iseng yang menyodorkan jutaan kalimat aneh mencari satu yang membuat
// penerjemahnya keseleo.
func FuzzRoundTrip(f *testing.F) {
	f.Add([]byte("aaabbbcccc"))
	f.Add([]byte(""))
	f.Add(bytes.Repeat([]byte{0x00}, 300)) // run > 255 -> uji pemecahan run
	f.Fuzz(func(t *testing.T, data []byte) {
		hasil := Decode(Encode(data))
		if !bytes.Equal(hasil, data) {
			t.Errorf("round-trip rusak: len(in)=%d len(out)=%d", len(data), len(hasil))
		}
	})
}

// ===========================================================================
// 2. PROPERTY-BASED — verifikasi properti pada banyak input acak
// ===========================================================================
// Alih-alih beberapa contoh tetap, kita generasikan 1000 input acak dan cek
// properti yang harus SELALU benar. Pelengkap fuzzing yang cepat & deterministik.
//
// 🔍 Analogi: contoh tetap itu seperti menguji timbangan dengan 3 batu yang
// itu-itu saja; property-based test seperti menimbang 1000 benda acak dan
// memastikan HUKUMNYA selalu berlaku (naik-turun timbangan konsisten).
// Seed tetap (42) = urutan benda acaknya bisa diulang persis saat debug.
func TestPropertyRoundTrip(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	for i := 0; i < 1000; i++ {
		n := rng.Intn(500)
		buf := make([]byte, n)
		for j := range buf {
			buf[j] = byte(rng.Intn(4)) // sedikit variasi -> banyak run berulang
		}
		if got := Decode(Encode(buf)); !bytes.Equal(got, buf) {
			t.Fatalf("properti gagal pada input acak ke-%d (len %d)", i, n)
		}
	}
}

// ===========================================================================
// 3. INTEGRATION TEST yang AUTO-SKIP tanpa infra (pola testcontainers)
// ===========================================================================
// Test integрasi berat (butuh DB/broker nyata) harus otomatis di-skip bila
// environment tak tersedia -> `go test ./...` tetap hijau di mesin mana pun.
//
// 🔍 Analogi: auto-skip itu seperti KELAS RENANG yang otomatis diliburkan
// bila kolamnya tutup (env tak ada) — murid lain (unit test) tetap belajar
// seperti biasa, bukannya seluruh sekolah diliburkan (test suite merah).
func TestIntegrasiButuhDB(t *testing.T) {
	dsn := os.Getenv("DB_URL")
	if dsn == "" {
		t.Skip("lewati: set DB_URL untuk menjalankan test integrasi (mis. via testcontainers)")
	}
	// ... di sini normalnya membuka koneksi nyata & menguji end-to-end ...
	t.Log("menjalankan test integrasi terhadap", dsn)
}

// Benchmark kecil (jalankan: go test -bench . -benchmem ./37-advanced-testing/advanced).
func BenchmarkEncode(b *testing.B) {
	data := bytes.Repeat([]byte("aaabbc"), 100)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = Encode(data)
	}
}
