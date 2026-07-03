// Tulis test-mu di sini. Jalankan: go test -v ./08-testing/jawaban-saya
// Bandingkan dengan kunci di ../latihan/calc_test.go setelah mencoba sendiri.
package calc

import "testing"

// Latihan 1: table-driven test untuk Divide, termasuk kasus pembagi nol (error).
func TestDivide(t *testing.T) {
	// TODO: buat tabel kasus {a, b, want, wantErr} dan loop dengan t.Run.
	t.Skip("hapus baris ini setelah kamu menulis test-nya")
}

// Latihan 2: table-driven test untuk IsPalindrome (termasuk non-ASCII).
func TestIsPalindrome(t *testing.T) {
	// TODO
	t.Skip("hapus baris ini setelah kamu menulis test-nya")
}

// Latihan 3: test FizzBuzz untuk 3, 5, 15, dan angka biasa.
func TestFizzBuzz(t *testing.T) {
	// TODO
	t.Skip("hapus baris ini setelah kamu menulis test-nya")
}

// Latihan 4: tambah BenchmarkFizzBuzz (func BenchmarkFizzBuzz(b *testing.B)).
// Latihan 5: tambah ExampleFizzBuzz dengan komentar // Output:.
