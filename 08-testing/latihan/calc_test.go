// Kunci jawaban latihan Modul 08 — Testing.
// Jalankan: go test -v ./08-testing/latihan/
package calc

import (
	"fmt"
	"testing"
)

// Latihan 1: table-driven Divide termasuk pembagi nol.
func TestDivide(t *testing.T) {
	tests := []struct {
		name    string
		a, b    int
		want    int
		wantErr bool
	}{
		{"bagi normal", 10, 2, 5, false},
		{"pembulatan ke bawah", 7, 2, 3, false},
		{"bagi nol", 1, 0, 0, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Divide(tc.a, tc.b)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("Divide(%d,%d): mengharapkan error", tc.a, tc.b)
				}
				return
			}
			if err != nil {
				t.Fatalf("Divide(%d,%d): tak terduga error: %v", tc.a, tc.b, err)
			}
			if got != tc.want {
				t.Errorf("Divide(%d,%d) = %d; want %d", tc.a, tc.b, got, tc.want)
			}
		})
	}
}

// Latihan 2: table-driven IsPalindrome termasuk non-ASCII.
func TestIsPalindrome(t *testing.T) {
	tests := []struct {
		in   string
		want bool
	}{
		{"katak", true},
		{"Ana", true},   // huruf besar diabaikan
		{"malam", true}, // m-a-l-a-m
		{"golang", false},
		{"level", true},
		{"世a世", true}, // per-rune, bukan per-byte
	}
	for _, tc := range tests {
		t.Run(tc.in, func(t *testing.T) {
			if got := IsPalindrome(tc.in); got != tc.want {
				t.Errorf("IsPalindrome(%q) = %t; want %t", tc.in, got, tc.want)
			}
		})
	}
}

// Latihan 3: FizzBuzz untuk 3, 5, 15, dan angka biasa.
func TestFizzBuzz(t *testing.T) {
	tests := []struct {
		n    int
		want string
	}{
		{1, "1"},
		{3, "Fizz"},
		{5, "Buzz"},
		{9, "Fizz"},
		{10, "Buzz"},
		{15, "FizzBuzz"},
		{30, "FizzBuzz"},
	}
	for _, tc := range tests {
		t.Run(fmt.Sprintf("n=%d", tc.n), func(t *testing.T) {
			if got := FizzBuzz(tc.n); got != tc.want {
				t.Errorf("FizzBuzz(%d) = %q; want %q", tc.n, got, tc.want)
			}
		})
	}
}

// Latihan 4: benchmark.
func BenchmarkFizzBuzz(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = FizzBuzz(i)
	}
}

// Latihan 5: example dengan // Output:.
func ExampleFizzBuzz() {
	fmt.Println(FizzBuzz(3))
	fmt.Println(FizzBuzz(5))
	fmt.Println(FizzBuzz(15))
	fmt.Println(FizzBuzz(7))
	// Output:
	// Fizz
	// Buzz
	// FizzBuzz
	// 7
}
