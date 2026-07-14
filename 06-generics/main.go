// Package main untuk modul 06 — Generics.
// Jalankan: go run ./06-generics
package main

import (
	"cmp"
	"fmt"
	"slices"
	"strings"
)

func main() {
	fmt.Println("=== 06 — Generics ===")
	fungsiGenerik()
	constraintKustom()
	tipeGenerik()
	stdlibGenerics()
}

// ------------------------------------------------------------------
// 1. Fungsi generik: Max, Map, Filter, Reduce
// ------------------------------------------------------------------

// 🔍 Analogi besar: generic itu seperti RESEP MASAKAN yang ditulis dengan "bahan X"
// alih-alih menyebut bahan tertentu. Satu resep "tumis X" bisa dipakai untuk ayam,
// tempe, atau sayur — tanpa menulis ulang resepnya per bahan. [T] itulah "bahan X".
// Sebelum generic, kamu harus menyalin fungsi yang sama untuk int, lalu untuk string, dst.

// 🔍 Analogi: constraint (cmp.Ordered) itu "SYARAT bahan". Max butuh bahan yang "bisa
// diurutkan/dibandingkan" (angka, teks) — jadi cmp.Ordered menyaring: tipe yang tak bisa
// pakai operator > ditolak saat kompilasi, bukan meledak saat program jalan.
// Max bekerja untuk semua tipe yang bisa dibandingkan urutannya.
func Max[T cmp.Ordered](a, b T) T {
	if a > b {
		return a
	}
	return b
}

// 🔍 Analogi: Map itu seperti BAN BERJALAN PABRIK — tiap barang (T) lewat satu mesin (f)
// dan keluar jadi barang jenis lain (U). Map angka->label: {1,2,3} lewat mesin "#" jadi
// {"#1","#2","#3"}. Filter = penyaring (buang yang tak lolos). Reduce = melipat semua jadi satu.
// Map mengubah []T menjadi []U lewat fungsi f. Dua type parameter.
func Map[T, U any](s []T, f func(T) U) []U {
	out := make([]U, len(s))
	for i, v := range s {
		out[i] = f(v)
	}
	return out
}

// Filter menyisakan elemen yang lolos predikat keep.
func Filter[T any](s []T, keep func(T) bool) []T {
	out := make([]T, 0, len(s))
	for _, v := range s {
		if keep(v) {
			out = append(out, v)
		}
	}
	return out
}

// Reduce melipat []T menjadi satu nilai U.
func Reduce[T, U any](s []T, init U, f func(U, T) U) U {
	acc := init
	for _, v := range s {
		acc = f(acc, v)
	}
	return acc
}

func fungsiGenerik() {
	fmt.Println("\n-- Fungsi generik --")

	// Type inference: tak perlu tulis [int]/[string].
	fmt.Printf("Max(3,5)=%d  Max(\"a\",\"b\")=%q  Max(1.5,2.5)=%.1f\n",
		Max(3, 5), Max("a", "b"), Max(1.5, 2.5))

	nums := []int{1, 2, 3, 4, 5}

	// int -> string
	labels := Map(nums, func(n int) string { return fmt.Sprintf("#%d", n) })
	fmt.Printf("Map  -> %v\n", labels)

	genap := Filter(nums, func(n int) bool { return n%2 == 0 })
	fmt.Printf("Filter genap -> %v\n", genap)

	total := Reduce(nums, 0, func(acc, n int) int { return acc + n })
	fmt.Printf("Reduce sum -> %d\n", total)
}

// ------------------------------------------------------------------
// 2. Constraint kustom dengan union & ~
// ------------------------------------------------------------------

// 🔍 Analogi: union "|" itu daftar bahan yang BOLEH ("int ATAU int64 ATAU float64").
// Tanda "~" berarti "termasuk tipe turunannya". Celsius adalah 'int' yang diberi nama baru —
// tanpa ~, ia ditolak; dengan ~int, Go bilang "asal DASARNYA int, boleh masuk". Seperti aturan
// "boleh bawa jeruk" yang otomatis juga membolehkan "jeruk bali" karena sama-sama jeruk.
// Number: int/int64/float64 ATAU tipe apa pun yang underlying-nya itu (~).
type Number interface {
	~int | ~int64 | ~float64
}

func Sum[T Number](s []T) T {
	var total T
	for _, v := range s {
		total += v
	}
	return total
}

// Celsius underlying-nya int -> tetap memenuhi Number berkat ~int.
type Celsius int

func constraintKustom() {
	fmt.Println("\n-- Constraint kustom (union & ~) --")

	fmt.Printf("Sum([]int)      = %d\n", Sum([]int{1, 2, 3}))
	fmt.Printf("Sum([]float64)  = %.1f\n", Sum([]float64{1.1, 2.2, 3.3}))
	fmt.Printf("Sum([]Celsius)  = %d (Celsius underlying int, cocok ~int)\n",
		Sum([]Celsius{10, 20, 30}))
}

// ------------------------------------------------------------------
// 3. Tipe generik: Stack[T] & Pair[K,V]
// ------------------------------------------------------------------
// 🔍 Analogi: Stack[T] itu TUMPUKAN PIRING generik. Sekali dibuat, kamu bisa cetak
// Stack[string] (tumpukan teks) atau Stack[int] (tumpukan angka) dari cetakan yang sama.
// Push = taruh piring di atas; Pop = ambil dari atas (yang terakhir masuk, pertama keluar).
type Stack[T any] struct {
	items []T
}

func (s *Stack[T]) Push(v T) { s.items = append(s.items, v) }

func (s *Stack[T]) Pop() (T, bool) {
	// 🔍 Analogi: 'var zero T' itu "nilai netral sesuai jenis tumpukan" — 0 untuk angka,
	// "" untuk teks. Dipakai saat tumpukan kosong: kita balas nilai netral + false, bukan panik.
	var zero T
	if len(s.items) == 0 {
		return zero, false // zero value generik + false
	}
	last := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]
	return last, true
}

type Pair[K comparable, V any] struct {
	Key K
	Val V
}

func tipeGenerik() {
	fmt.Println("\n-- Tipe generik --")

	var st Stack[string]
	st.Push("a")
	st.Push("b")
	st.Push("c")
	for {
		v, ok := st.Pop()
		if !ok {
			break
		}
		fmt.Printf("pop -> %q\n", v)
	}

	p := Pair[string, int]{Key: "umur", Val: 20}
	fmt.Printf("pair -> %+v\n", p)
}

// ------------------------------------------------------------------
// 4. Paket standar berbasis generics (slices, cmp, builtin min/max)
// ------------------------------------------------------------------
func stdlibGenerics() {
	fmt.Println("\n-- stdlib: slices, cmp, min/max --")

	s := []int{5, 2, 8, 1, 9, 3}
	slices.Sort(s) // generic sort, tanpa boilerplate
	fmt.Printf("slices.Sort   -> %v\n", s)
	fmt.Printf("slices.Contains(_,8)=%t  slices.Max=%d\n", slices.Contains(s, 8), slices.Max(s))

	// SortFunc dengan cmp.Compare untuk urut custom (mis. berdasar panjang string).
	kata := []string{"aaa", "b", "cc"}
	slices.SortFunc(kata, func(a, b string) int { return cmp.Compare(len(a), len(b)) })
	fmt.Printf("SortFunc by len -> %v\n", kata)

	// builtin min/max (Go 1.21+) untuk kasus sederhana.
	fmt.Printf("builtin min(3,7)=%d max(3,7)=%d\n", min(3, 7), max(3, 7))

	fmt.Printf("join -> %s\n", strings.Join(kata, ","))
}
