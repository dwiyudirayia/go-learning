package main

import "testing"

// Test demonstrasi untuk Modul 01.
// Jalankan: go test ./01-basics    (atau: go test -v ./01-basics)
//
// Pola "table-driven test" adalah idiom WAJIB di Go: satu daftar kasus,
// satu loop. Menambah kasus = menambah satu baris. (Dibahas penuh di Modul 08.)

func TestBagi(t *testing.T) {
	cases := []struct {
		name                string
		a, b                int
		wantHasil, wantSisa int
	}{
		{"habis dibagi", 10, 2, 5, 0},
		{"ada sisa", 7, 3, 2, 1},
		{"pembilang < penyebut", 3, 5, 0, 3},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			hasil, sisa := bagi(c.a, c.b)
			if hasil != c.wantHasil || sisa != c.wantSisa {
				t.Errorf("bagi(%d,%d) = (%d,%d), mau (%d,%d)",
					c.a, c.b, hasil, sisa, c.wantHasil, c.wantSisa)
			}
		})
	}
}

func TestMinMax(t *testing.T) {
	cases := []struct {
		name             string
		nums             []int
		wantMin, wantMax int
	}{
		{"acak", []int{3, 1, 4, 1, 5, 9, 2}, 1, 9},
		{"satu elemen", []int{42}, 42, 42},
		{"semua sama", []int{7, 7, 7}, 7, 7},
		{"negatif", []int{-3, -1, -8}, -8, -1},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			min, max := minMax(c.nums...)
			if min != c.wantMin || max != c.wantMax {
				t.Errorf("minMax(%v) = (%d,%d), mau (%d,%d)",
					c.nums, min, max, c.wantMin, c.wantMax)
			}
		})
	}
}
