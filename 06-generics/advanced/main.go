// Teknik Advanced Modul 06 (generics) — contoh runnable + penjelasan detail.
//
// Jalankan: go run ./06-generics/advanced
package main

import "fmt"

// ===========================================================================
// 1. Type set & union constraint dengan tilde (~)
// ===========================================================================
// Constraint adalah interface berisi HIMPUNAN tipe yang diizinkan.
// - "int | float64" = tepat tipe itu.
// - "~int"          = semua tipe yang UNDERLYING-nya int (termasuk type MyInt int).
// Tanpa tilde, "type Celsius float64" TIDAK akan cocok dengan "float64".

type Angka interface {
	~int | ~int64 | ~float64
}

// Max bekerja untuk semua tipe di dalam constraint Angka. Operator < valid
// karena semua anggota constraint mendukung pengurutan.
func Max[T Angka](a, b T) T {
	if a > b {
		return a
	}
	return b
}

// Celsius punya underlying float64 -> cocok dengan ~float64.
type Celsius float64

// ===========================================================================
// 2. Struktur data generik type-safe: Set[T]
// ===========================================================================
// comparable = constraint bawaan untuk tipe yang bisa dibandingkan dgn ==
// (syarat menjadi kunci map). Dengan generics, Set aman tipe tanpa any+assertion.

type Set[T comparable] struct {
	m map[T]struct{}
}

func NewSet[T comparable](items ...T) *Set[T] {
	s := &Set[T]{m: make(map[T]struct{}, len(items))}
	for _, it := range items {
		s.Add(it)
	}
	return s
}

func (s *Set[T]) Add(v T)      { s.m[v] = struct{}{} }
func (s *Set[T]) Has(v T) bool { _, ok := s.m[v]; return ok }
func (s *Set[T]) Len() int     { return len(s.m) }

// ===========================================================================
// 3. Fungsi generik higher-order: Map (transformasi tipe A -> B)
// ===========================================================================
// Perhatikan DUA type parameter [A, B any]: input dan output boleh beda tipe.
// Catatan penting: METHOD tidak boleh punya type parameter sendiri di Go —
// makanya Map ditulis sebagai fungsi bebas, bukan method Set.
func Map[A, B any](in []A, f func(A) B) []B {
	out := make([]B, len(in))
	for i, v := range in {
		out[i] = f(v)
	}
	return out
}

func main() {
	fmt.Println("== 1. union constraint + tilde ==")
	fmt.Println("  Max(3, 7)          =", Max(3, 7))
	fmt.Println("  Max(2.5, 1.5)      =", Max(2.5, 1.5))
	// Inference gagal untuk tipe kustom? beri argumen tipe eksplisit:
	fmt.Println("  Max[Celsius](20,25) =", Max[Celsius](20, 25))

	fmt.Println("== 2. Set[T] generik ==")
	s := NewSet("go", "rust", "go", "c")
	fmt.Printf("  anggota unik=%d, punya \"rust\"=%v\n", s.Len(), s.Has("rust"))

	fmt.Println("== 3. Map generik (int -> string) ==")
	angka := []int{1, 2, 3}
	teks := Map(angka, func(n int) string { return fmt.Sprintf("#%d", n) })
	fmt.Println("  hasil:", teks)

	// Catatan performa: generics di Go memakai "GC shape stenciling" (berbagi
	// kode per bentuk memori via dictionary), BUKAN monomorfisasi penuh. Kadang
	// sedikit lebih lambat dari kode konkret — ukur (benchmark) sebelum menggeneralkan.
}
