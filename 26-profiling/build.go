// Modul 26 — Profiling & Optimization: mengukur, bukan menebak.
package main

import "strings"

// BuildSlow menggabungkan string dengan `+=`. Setiap iterasi mengalokasikan
// string BARU (string itu immutable) -> O(n^2) alokasi. Boros memori.
func BuildSlow(n int) string {
	s := ""
	for i := 0; i < n; i++ {
		s += "x"
	}
	return s
}

// BuildFast memakai strings.Builder yang menulis ke buffer yang tumbuh -> jauh
// lebih sedikit alokasi. Bandingkan dengan benchmark (-benchmem).
func BuildFast(n int) string {
	var b strings.Builder
	b.Grow(n) // pra-alokasi kapasitas -> nol realokasi
	for i := 0; i < n; i++ {
		b.WriteByte('x')
	}
	return b.String()
}

// SumSquares menjumlahkan kuadrat 0..n-1. Contoh fungsi CPU-bound untuk profil CPU.
func SumSquares(n int) int {
	total := 0
	for i := 0; i < n; i++ {
		total += i * i
	}
	return total
}
