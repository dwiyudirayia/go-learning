// Package advanced berisi kode kecil sebagai bahan mendemonstrasikan TEKNIK
// pengujian lanjutan. Fokusnya ada di palindrome_test.go — file ini hanya
// menyediakan fungsi yang akan diuji.
//
// Jalankan test biasa   : go test ./08-testing/advanced
// Jalankan + verbose     : go test -v ./08-testing/advanced
// Jalankan fuzzing 10 dtk: go test -run x -fuzz FuzzIsPalindrome -fuzztime 10s ./08-testing/advanced
package advanced

import (
	"strings"
	"unicode"
)

// normalisasi menyisakan huruf/angka dalam huruf kecil, membuang spasi & tanda
// baca. Jadi "Kasur, ini rusak!" bisa dikenali sebagai palindrom.
func normalisasi(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// IsPalindrome true bila s (setelah dinormalisasi) sama dibaca maju & mundur.
// Aman untuk Unicode karena bekerja pada level rune, bukan byte.
func IsPalindrome(s string) bool {
	r := []rune(normalisasi(s))
	for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
		if r[i] != r[j] {
			return false
		}
	}
	return true
}
