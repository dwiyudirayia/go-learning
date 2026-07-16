// Teknik Advanced Modul 43 (advanced generics) — contoh runnable + penjelasan detail.
//
// iter.Seq (range-over-func, Go 1.23) + struktur data generik + functional
// options generik. Jalankan:
//
//	go run ./43-advanced-generics/advanced
package main

import (
	"fmt"
	"iter"
	"slices"
)

// ===========================================================================
// 1. iter.Seq[T] — ITERATOR LAZY via range-over-func (Go 1.23)
// ===========================================================================
// iter.Seq[T] = func(yield func(T) bool). `for v := range seq {}` memanggil
// yield tiap elemen; mengembalikan false (mis. karena break) menghentikannya.
// Nilai dihasilkan MALAS (on-demand) -> bisa "tak hingga" tanpa boros memori.
//
// 🔍 Analogi: slice itu seperti MEMASAK SEMUA PESANAN DI MUKA lalu menaruhnya
// di meja (boros bila tamu cuma makan 5); iterator lazy seperti KOKI TEPPANYAKI
// — memasak satu porsi TEPAT saat diminta (yield), dan langsung berhenti
// begitu tamu bilang cukup (break). Makanya deret "tak hingga" pun aman.

// Hitung menghasilkan 0,1,2,... hingga tak terbatas (lazy!).
func Hitung() iter.Seq[int] {
	return func(yield func(int) bool) {
		for i := 0; ; i++ {
			if !yield(i) { // konsumen break -> berhenti
				return
			}
		}
	}
}

// Filter membungkus iterator lain -> hanya meneruskan yang lolos predikat.
// Komposisi iterator: rantai transformasi tanpa mengalokasikan slice antara.
//
// 🔍 Analogi: komposisi iterator itu seperti MENYAMBUNG SARINGAN di ujung
// selang — air mengalir sekali lewat semua saringan berurutan; tak perlu
// menampung ke ember (slice antara) tiap ganti saringan.
func Filter[T any](seq iter.Seq[T], keep func(T) bool) iter.Seq[T] {
	return func(yield func(T) bool) {
		for v := range seq {
			if keep(v) && !yield(v) {
				return
			}
		}
	}
}

// ===========================================================================
// 2. Struktur data generik: Set[T] dengan iterator sendiri
// ===========================================================================
type Set[T comparable] struct{ m map[T]struct{} }

func NewSet[T comparable](xs ...T) *Set[T] {
	s := &Set[T]{m: make(map[T]struct{})}
	for _, x := range xs {
		s.m[x] = struct{}{}
	}
	return s
}
func (s *Set[T]) All() iter.Seq[T] {
	return func(yield func(T) bool) {
		for k := range s.m {
			if !yield(k) {
				return
			}
		}
	}
}

// ===========================================================================
// 3. Functional options GENERIK
// ===========================================================================
type Option[T any] func(*T)

func Build[T any](opts ...Option[T]) T {
	var t T
	for _, o := range opts {
		o(&t)
	}
	return t
}

type Server struct {
	Host string
	Port int
}

func main() {
	// 1. iterator lazy: ambil 5 bilangan genap pertama dari deret TAK HINGGA
	fmt.Println("== 1. iter.Seq lazy (genap, ambil 5) ==")
	genap := Filter(Hitung(), func(n int) bool { return n%2 == 0 })
	var lima []int
	for n := range genap { // range-over-func
		lima = append(lima, n)
		if len(lima) == 5 {
			break // menghentikan iterator tak hingga
		}
	}
	fmt.Println("  ", lima)

	// 2. Set generik + kumpulkan iteratornya jadi slice terurut
	fmt.Println("== 2. Set[T] + slices.Sorted(iter) ==")
	s := NewSet("go", "rust", "go", "zig", "c")
	fmt.Println("   anggota unik terurut:", slices.Sorted(s.All()))

	// 3. functional options generik
	fmt.Println("== 3. functional options generik ==")
	srv := Build(
		func(s *Server) { s.Host = "localhost" },
		func(s *Server) { s.Port = 8080 },
	)
	fmt.Printf("   %+v\n", srv)
}
