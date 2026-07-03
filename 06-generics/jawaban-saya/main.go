// Tempat latihanmu untuk Modul 06.
// Jalankan: go run ./06-generics/jawaban-saya
// Isi bagian TODO, lalu bandingkan dengan ../latihan/solusi.go.
package main

import "fmt"

func main() {
	fmt.Println("=== Jawaban saya - Modul 06 ===")
	latihan1()
	latihan2()
	latihan3()
	latihan4()
	latihan5()
}

// Latihan 1: Map[T, U any](s []T, f func(T) U) []U; pakai []int -> []string.
// TODO: definisikan fungsi generik Map di level package, lalu pakai di sini.
func latihan1() {
	fmt.Println("\n-- Latihan 1: Map --")
	// TODO
}

// Latihan 2: Filter[T any](s []T, keep func(T) bool) []T.
func latihan2() {
	fmt.Println("\n-- Latihan 2: Filter --")
	// TODO
}

// Latihan 3: Reduce[T, U any](s []T, init U, f func(U, T) U) U; jumlahkan []int.
func latihan3() {
	fmt.Println("\n-- Latihan 3: Reduce --")
	// TODO
}

// Latihan 4: constraint Number (~int | ~float64) + Sum[T Number](s []T) T.
func latihan4() {
	fmt.Println("\n-- Latihan 4: Number + Sum --")
	// TODO
}

// Latihan 5: tipe generik Pair[K comparable, V any] dengan method Swap().
func latihan5() {
	fmt.Println("\n-- Latihan 5: Pair + Swap --")
	// TODO
}
