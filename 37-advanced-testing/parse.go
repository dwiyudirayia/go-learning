// Modul 37 — Advanced Testing: fuzzing, integration, load test.
package main

import (
	"fmt"
	"strconv"
	"strings"
)

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
