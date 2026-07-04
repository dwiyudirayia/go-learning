package main

import (
	"math"
	"testing"
)

// Test demonstrasi untuk Modul 03.
// Jalankan: go test -v ./03-structs-methods
//
// Menguji constructor yang mengembalikan error (idiom "NewX") + method value-receiver.

func TestNewRectangle(t *testing.T) {
	cases := []struct {
		name    string
		w, h    float64
		wantErr bool
	}{
		{"valid", 3, 4, false},
		{"lebar nol", 0, 4, true},
		{"tinggi negatif", 3, -1, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r, err := NewRectangle(c.w, c.h)
			if c.wantErr {
				if err == nil {
					t.Fatalf("NewRectangle(%v,%v) mau error, dapat nil", c.w, c.h)
				}
				return // pada kasus error, r harus nil — tak perlu cek lanjut
			}
			if err != nil {
				t.Fatalf("NewRectangle(%v,%v) tak terduga error: %v", c.w, c.h, err)
			}
			if r.Area() != c.w*c.h {
				t.Errorf("Area() = %v, mau %v", r.Area(), c.w*c.h)
			}
		})
	}
}

// Scale memakai pointer receiver -> mengubah state asli.
func TestRectangleScale(t *testing.T) {
	r, _ := NewRectangle(2, 3)
	r.Scale(2)
	if got := r.Area(); math.Abs(got-24) > 1e-9 { // (2*2)*(3*2)=24
		t.Errorf("setelah Scale(2), Area() = %v, mau 24", got)
	}
}
