// Tempat latihanmu untuk Modul 08 — tulis TEST untuk fungsi-fungsi di bawah.
// Fungsi di sini sama dengan ../latihan/calc.go. Tulis test di calc_test.go,
// lalu jalankan: go test -v ./08-testing/jawaban-saya
package calc

import (
	"errors"
	"fmt"
	"strings"
)

func Add(a, b int) int { return a + b }

func Divide(a, b int) (int, error) {
	if b == 0 {
		return 0, errors.New("pembagian dengan nol")
	}
	return a / b, nil
}

func IsPalindrome(s string) bool {
	r := []rune(strings.ToLower(s))
	for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
		if r[i] != r[j] {
			return false
		}
	}
	return true
}

func FizzBuzz(n int) string {
	switch {
	case n%15 == 0:
		return "FizzBuzz"
	case n%3 == 0:
		return "Fizz"
	case n%5 == 0:
		return "Buzz"
	default:
		return fmt.Sprintf("%d", n)
	}
}
