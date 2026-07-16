package advanced

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// ===========================================================================
// 1. Table-driven + SUBTEST + t.Parallel()
// ===========================================================================
// Table-driven = satu daftar kasus, satu loop. t.Run membuat subtest bernama
// (muncul terpisah di output & bisa dijalankan sendiri via -run). t.Parallel()
// menjalankan subtest secara paralel -> lebih cepat & menyingkap race.
//
// 🔍 Analogi: table-driven test itu seperti UJIAN DENGAN LEMBAR SOAL — satu
// prosedur koreksi (loop), banyak soal (baris tabel). t.Run memberi tiap soal
// NOMOR — yang salah langsung ketahuan "soal nomor berapa", dan t.Parallel()
// seperti membagi koreksi ke beberapa korektor sekaligus.
func TestIsPalindrome(t *testing.T) {
	kasus := []struct {
		nama string
		in   string
		want bool
	}{
		{"kata sederhana", "kasur rusak", true},
		{"abaikan tanda baca & huruf besar", "Kasur, ini rusak!", true},
		{"unicode", "katak", true},
		{"kosong", "", true},
		{"bukan palindrom", "golang", false},
	}
	for _, k := range kasus {
		k := k // (kebiasaan aman; sejak Go 1.22 loop var sudah per-iterasi)
		t.Run(k.nama, func(t *testing.T) {
			t.Parallel() // subtest ini boleh jalan paralel dgn subtest lain
			if got := IsPalindrome(k.in); got != k.want {
				t.Errorf("IsPalindrome(%q) = %v, mau %v", k.in, got, k.want)
			}
		})
	}
}

// ===========================================================================
// 2. t.Helper() — laporan gagal menunjuk pemanggil, bukan baris helper
// ===========================================================================
// Tanpa t.Helper(), kegagalan akan menunjuk ke dalam fungsi bantu ini.
// Dengan t.Helper(), Go melompatinya sehingga baris yang error adalah baris
// pemanggil — jauh lebih mudah ditelusuri.
//
// 🔍 Analogi: t.Helper() itu seperti KURIR yang menulis alamat PENGIRIM asli
// di paket komplain — tanpa itu, komplain tercatat "dari kantor kurir"
// (baris helper), padahal sumber masalahnya si pengirim (baris pemanggil).
func harusPalindrom(t *testing.T, s string) {
	t.Helper()
	if !IsPalindrome(s) {
		t.Errorf("%q seharusnya palindrom", s)
	}
}

func TestDenganHelper(t *testing.T) {
	harusPalindrom(t, "malam")
	harusPalindrom(t, "ini")
}

// ===========================================================================
// 3. t.TempDir() + t.Cleanup() — resource sementara yang auto-bersih
// ===========================================================================
// t.TempDir() memberi folder unik yang otomatis dihapus saat test selesai.
// t.Cleanup() mendaftarkan pembersihan (dieksekusi LIFO) — lebih rapi dari
// defer untuk setup yang tersebar.
//
// 🔍 Analogi: t.TempDir() itu seperti KAMAR HOTEL — kamu bebas berantakan
// selama menginap (test berjalan), dan housekeeping otomatis membereskan
// begitu kamu check-out. t.Cleanup() = daftar pesan ke housekeeping,
// dikerjakan dari pesan TERAKHIR ke pertama (LIFO).
func TestTulisBacaFile(t *testing.T) {
	dir := t.TempDir() // otomatis terhapus setelah test
	path := filepath.Join(dir, "data.txt")

	t.Cleanup(func() { t.Log("cleanup dijalankan (LIFO) setelah test") })

	if err := os.WriteFile(path, []byte("malam"), 0o600); err != nil {
		t.Fatalf("tulis file: %v", err) // Fatalf menghentikan test ini
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("baca file: %v", err)
	}
	if !IsPalindrome(string(b)) {
		t.Errorf("isi file %q bukan palindrom", b)
	}
}

// ===========================================================================
// 4. Example test — dokumentasi yang SEKALIGUS diverifikasi
// ===========================================================================
// Komentar // Output: dibandingkan dengan stdout. Jika beda, test gagal.
// Contoh ini juga muncul di dokumentasi `go doc`.
//
// 🔍 Analogi: Example test itu seperti FOTO MASAKAN DI MENU yang dijamin
// undang-undang — kalau hidangan asli (output program) tak sama dengan
// fotonya (komentar // Output:), restoran kena tegur (test gagal). Dokumentasi
// jadi mustahil basi.
func ExampleIsPalindrome() {
	fmt.Println(IsPalindrome("Kasur ini rusak")) // tanda baca & kapital diabaikan
	fmt.Println(IsPalindrome("golang"))
	// Output:
	// true
	// false
}

// ===========================================================================
// 5. Fuzzing (Go 1.18+) — mesin mencari input pemecah invariant
// ===========================================================================
// Kita uji PROPERTY, bukan contoh tunggal: membalik urutan rune pada input
// tak boleh mengubah status palindrom (karena normalisasi bersifat simetris).
// Fuzzer akan mengoceh ribuan input acak mencari yang melanggar.
//
// 🔍 Analogi: test biasa itu seperti menguji jembatan dengan BEBERAPA truk
// yang kamu pilih sendiri; fuzzing seperti melepas RIBUAN kendaraan acak —
// termasuk yang aneh-aneh — dan membiarkan mesin mencari satu yang bikin
// jembatan retak (input pemecah invariant).
func FuzzIsPalindrome(f *testing.F) {
	f.Add("kasur rusak") // seed corpus
	f.Add("golang")
	f.Add("Ω≈ç√")
	f.Fuzz(func(t *testing.T, s string) {
		rs := []rune(s)
		for i, j := 0, len(rs)-1; i < j; i, j = i+1, j-1 {
			rs[i], rs[j] = rs[j], rs[i]
		}
		if IsPalindrome(s) != IsPalindrome(string(rs)) {
			t.Errorf("invariant rusak: input %q vs terbalik %q", s, string(rs))
		}
	})
}

// ===========================================================================
// 6. Benchmark — ukur performa, laporkan alokasi (-benchmem)
// ===========================================================================
// Jalankan: go test -bench . -benchmem ./08-testing/advanced
//
// 🔍 Analogi: benchmark itu seperti STOPWATCH PELATIH LARI — atlet (fungsi)
// disuruh lari berulang-ulang (b.N kali) sampai catatan waktunya stabil,
// bukan dinilai dari satu kali sprint yang bisa kebetulan cepat/lambat.
func BenchmarkIsPalindrome(b *testing.B) {
	s := "Kasur ini benar-benar rusak sudah"
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = IsPalindrome(s)
	}
}
