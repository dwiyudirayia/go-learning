// Package calc berisi fungsi-fungsi yang harus kamu tulis test-nya (Modul 08).
package calc

import (
	"errors"
	"fmt"
	"strings"
)

// Add menjumlahkan dua bilangan.
func Add(a, b int) int { return a + b }

// Divide membagi a dengan b; error bila b == 0.
func Divide(a, b int) (int, error) {
	if b == 0 {
		return 0, errors.New("pembagian dengan nol")
	}
	return a / b, nil
}

// IsPalindrome cek apakah s palindrom (per-rune, mengabaikan besar/kecil huruf).
func IsPalindrome(s string) bool {
	r := []rune(strings.ToLower(s))
	for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
		if r[i] != r[j] {
			return false
		}
	}
	return true
}

// FizzBuzz mengembalikan "Fizz" (kelipatan 3), "Buzz" (kelipatan 5),
// "FizzBuzz" (kelipatan 15), atau angka itu sendiri sebagai string.
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
