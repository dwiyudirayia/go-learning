// Modul 37 — Advanced Testing: fuzzing, integration, load test.
package main

import (
	"fmt"
	"strconv"
	"strings"
)

// 🔍 Analogi besar FUZZING: unit test biasa itu kamu memberi CONTOH SOAL pilihan sendiri ("2-5
// harus jadi 2,5"). Fuzzing itu seperti menyewa MONYET PENGETIK yang melempar RIBUAN input aneh
// & acak ("", "--", "9999999999999", emoji) ke fungsimu untuk mencari yang bikin CRASH/panic.
// Alih-alih menguji contoh spesifik, fuzzing menguji INVARIAN — sifat yang harus SELALU benar apa
// pun inputnya (mis. "Reverse(Reverse(s)) harus sama dengan s"). Di repo ini, fuzzer pernah menemukan
// bug UTF-8 nyata. Fungsi yang mengurai input mentah seperti ParseRange = target fuzzing paling pas.

// ParseRange mem-parse "2-5" menjadi (2, 5). Fungsi seperti ini rawan crash oleh
// input tak terduga -> kandidat sempurna untuk FUZZING.
func ParseRange(s string) (lo, hi int, err error) {
	parts := strings.SplitN(s, "-", 2)
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("format harus lo-hi: %q", s)
	}
	lo, err = strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return 0, 0, fmt.Errorf("lo tidak valid: %w", err)
	}
	hi, err = strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return 0, 0, fmt.Errorf("hi tidak valid: %w", err)
	}
	if lo > hi {
		return 0, 0, fmt.Errorf("lo (%d) > hi (%d)", lo, hi)
	}
	return lo, hi, nil
}

// Reverse membalik string per-rune (aman multi-byte). INVARIAN yang bisa diuji
// fuzzing: Reverse(Reverse(s)) == s, dan hasilnya tetap UTF-8 valid.
func Reverse(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}
