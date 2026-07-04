package main

import "testing"

// Test demonstrasi untuk Modul 02 — fokus reverseString.
// Jalankan: go test -v ./02-collections
//
// Perhatikan kasus "unicode": reverseString bekerja per-rune (bukan per-byte),
// jadi karakter multi-byte (é, 世) tidak rusak. Ini bug klasik kalau pakai []byte.

func TestReverseString(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"ascii", "halo", "olah"},
		{"kosong", "", ""},
		{"satu huruf", "a", "a"},
		{"palindrom", "kodok", "kodok"},
		{"unicode", "héllo", "olléh"}, // é = 2 byte, harus tetap utuh
		{"cjk", "世界", "界世"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := reverseString(c.in)
			if got != c.want {
				t.Errorf("reverseString(%q) = %q, mau %q", c.in, got, c.want)
			}
		})
	}
}

// Membalik dua kali harus kembali ke asal (properti/invariant).
func TestReverseStringTwiceIsIdentity(t *testing.T) {
	for _, s := range []string{"halo", "héllo", "世界", ""} {
		if got := reverseString(reverseString(s)); got != s {
			t.Errorf("reverse(reverse(%q)) = %q, mau %q", s, got, s)
		}
	}
}
