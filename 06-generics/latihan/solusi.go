// Solusi latihan Modul 06 — Generics.
// Jalankan: go run ./06-generics/latihan
package main

import "fmt"

func main() {
	fmt.Println("=== Solusi Latihan Modul 06 ===")
	latihan1()
	latihan2()
	latihan3()
	latihan4()
	latihan5()
}

// ------------------------------------------------------------------
// Latihan 1: Map[T, U any]
// ------------------------------------------------------------------
func Map[T, U any](s []T, f func(T) U) []U {
	out := make([]U, len(s))
	for i, v := range s {
		out[i] = f(v)
	}
	return out
}

func latihan1() {
	fmt.Println("\n-- Latihan 1: Map (int -> string) --")
	nums := []int{1, 2, 3}
	strs := Map(nums, func(n int) string { return fmt.Sprintf("angka-%d", n) })
	fmt.Printf("%v -> %v\n", nums, strs)
}

// ------------------------------------------------------------------
// Latihan 2: Filter[T any]
// ------------------------------------------------------------------
func Filter[T any](s []T, keep func(T) bool) []T {
	out := make([]T, 0, len(s))
	for _, v := range s {
		if keep(v) {
			out = append(out, v)
		}
	}
	return out
}

func latihan2() {
	fmt.Println("\n-- Latihan 2: Filter --")
	nums := []int{1, 2, 3, 4, 5, 6}
	genap := Filter(nums, func(n int) bool { return n%2 == 0 })
	fmt.Printf("genap  -> %v\n", genap)

	kata := []string{"go", "rust", "c", "python"}
	pendek := Filter(kata, func(s string) bool { return len(s) <= 2 })
	fmt.Printf("pendek -> %v\n", pendek)
}

// ------------------------------------------------------------------
// Latihan 3: Reduce[T, U any]
// ------------------------------------------------------------------
func Reduce[T, U any](s []T, init U, f func(U, T) U) U {
	acc := init
	for _, v := range s {
		acc = f(acc, v)
	}
	return acc
}

func latihan3() {
	fmt.Println("\n-- Latihan 3: Reduce (sum) --")
	nums := []int{1, 2, 3, 4, 5}
	sum := Reduce(nums, 0, func(acc, n int) int { return acc + n })
	fmt.Printf("sum(%v) = %d\n", nums, sum)

	// Bonus: Reduce juga bisa gabung string
	words := []string{"Go", "itu", "asyik"}
	kalimat := Reduce(words, "", func(acc, w string) string {
		if acc == "" {
			return w
		}
		return acc + " " + w
	})
	fmt.Printf("join -> %q\n", kalimat)
}

// ------------------------------------------------------------------
// Latihan 4: constraint Number + Sum
// ------------------------------------------------------------------
type Number interface {
	~int | ~float64
}

func Sum[T Number](s []T) T {
	var total T
	for _, v := range s {
		total += v
	}
	return total
}

func latihan4() {
	fmt.Println("\n-- Latihan 4: Number constraint + Sum --")
	fmt.Printf("Sum([]int)     = %d\n", Sum([]int{10, 20, 30}))
	fmt.Printf("Sum([]float64) = %.2f\n", Sum([]float64{1.5, 2.5, 3.0}))
}

// ------------------------------------------------------------------
// Latihan 5: Pair[K comparable, V any] + Swap
// ------------------------------------------------------------------
// Catatan constraint: soal menyebut Pair[K comparable, V any]. TAPI begitu kita
// menambah Swap() yang menukar Key<->Val, nilai V akan menjadi Key pada hasil
// (Pair[V, K]) — dan Key WAJIB comparable. Jadi V pun harus comparable.
// Ini contoh nyata bagaimana kebutuhan sebuah method bisa memaksa constraint.
type Pair[K, V comparable] struct {
	Key K
	Val V
}

// Swap menukar Key & Val, menghasilkan Pair dengan tipe tertukar (Pair[V, K]).
// Tipe hasilnya berbeda, jadi Swap mengembalikan Pair baru (bukan pointer receiver).
func (p Pair[K, V]) Swap() Pair[V, K] {
	return Pair[V, K]{Key: p.Val, Val: p.Key}
}

func latihan5() {
	fmt.Println("\n-- Latihan 5: Pair generik + Swap --")
	p := Pair[string, int]{Key: "umur", Val: 20}
	fmt.Printf("asli    -> %+v\n", p)
	fmt.Printf("swapped -> %+v\n", p.Swap())
}
