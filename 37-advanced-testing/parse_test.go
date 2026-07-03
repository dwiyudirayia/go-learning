package main

import (
	"testing"
	"unicode/utf8"
)

// --- Table-driven test biasa (Modul 8) ---
func TestParseRange(t *testing.T) {
	tests := []struct {
		in      string
		lo, hi  int
		wantErr bool
	}{
		{"2-5", 2, 5, false},
		{" 1 - 10 ", 1, 10, false},
		{"5-2", 0, 0, true}, // lo > hi
		{"abc", 0, 0, true},
		{"3", 0, 0, true},
	}
	for _, tc := range tests {
		lo, hi, err := ParseRange(tc.in)
		if (err != nil) != tc.wantErr {
			t.Errorf("ParseRange(%q) err = %v; wantErr %v", tc.in, err, tc.wantErr)
			continue
		}
		if !tc.wantErr && (lo != tc.lo || hi != tc.hi) {
			t.Errorf("ParseRange(%q) = %d,%d; want %d,%d", tc.in, lo, hi, tc.lo, tc.hi)
		}
	}
}

// --- FUZZING: Go otomatis membangkitkan input acak untuk menemukan crash ---
//
// Jalankan seed corpus (cepat, ikut `go test`):  go test ./37-advanced-testing
// Jalankan fuzzing sungguhan (cari input baru):   go test -fuzz FuzzParseRange -fuzztime 20s ./37-advanced-testing

func FuzzParseRange(f *testing.F) {
	// Seed corpus: contoh awal untuk memandu fuzzer.
	f.Add("2-5")
	f.Add("")
	f.Add("-")
	f.Add("999999999999999999999-1")

	f.Fuzz(func(t *testing.T, s string) {
		// INVARIAN: fungsi tidak boleh CRASH (panic) untuk input apa pun.
		// Jika panic, fuzzer akan melaporkan input yang memicunya.
		lo, hi, err := ParseRange(s)
		if err == nil && lo > hi {
			t.Errorf("ParseRange(%q) sukses tapi lo(%d) > hi(%d)", s, lo, hi)
		}
	})
}

func FuzzReverse(f *testing.F) {
	f.Add("halo")
	f.Add("世界")
	f.Add("café")

	f.Fuzz(func(t *testing.T, s string) {
		rev := Reverse(s)
		// INVARIAN 1: membalik dua kali kembali ke asal — TAPI hanya berlaku untuk
		// UTF-8 valid. (Fuzzing menemukan: untuk byte UTF-8 tak valid, []rune bersifat
		// LOSSY -> mengganti byte rusak dengan U+FFFD, sehingga round-trip tak sama.
		// Pelajaran: invariant kita semula terlalu kuat.)
		if utf8.ValidString(s) {
			if double := Reverse(rev); double != s {
				t.Errorf("Reverse(Reverse(%q)) = %q; want %q", s, double, s)
			}
			// INVARIAN 2: hasil tetap UTF-8 valid bila input valid.
			if !utf8.ValidString(rev) {
				t.Errorf("Reverse merusak UTF-8: input %q", s)
			}
		}
	})
}
