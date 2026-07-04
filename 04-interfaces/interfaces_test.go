package main

import (
	"math"
	"testing"
)

// Test demonstrasi untuk Modul 04.
// Jalankan: go test -v ./04-interfaces
//
// totalArea menerima []Shape — bukti kekuatan interface: satu fungsi bekerja
// untuk tipe konkret apa pun yang punya method Area().

func TestTotalArea(t *testing.T) {
	shapes := []Shape{
		Rectangle{W: 3, H: 4}, // 12
		Circle{R: 1},          // π
		Rectangle{W: 2, H: 5}, // 10
	}
	want := 12 + math.Pi + 10
	got := totalArea(shapes)
	if math.Abs(got-want) > 1e-9 {
		t.Errorf("totalArea = %v, mau %v", got, want)
	}
}

func TestTotalAreaKosong(t *testing.T) {
	if got := totalArea(nil); got != 0 {
		t.Errorf("totalArea(nil) = %v, mau 0", got)
	}
}

// Menegaskan tiap tipe konkret memenuhi interface Shape (dicek saat kompilasi juga).
func TestShapeImplementations(t *testing.T) {
	var _ Shape = Circle{}
	var _ Shape = Rectangle{}

	if got := (Circle{R: 2}).Area(); math.Abs(got-(math.Pi*4)) > 1e-9 {
		t.Errorf("Circle{2}.Area() = %v, mau %v", got, math.Pi*4)
	}
}
