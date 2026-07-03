// Tempat latihanmu untuk Modul 05.
// Jalankan: go run ./05-errors/jawaban-saya
// Isi bagian TODO, lalu bandingkan dengan ../latihan/solusi.go.
package main

import "fmt"

func main() {
	fmt.Println("=== Jawaban saya - Modul 05 ===")
	latihan1()
	latihan2()
	latihan3()
	latihan4()
	latihan5()
}

// Latihan 1: var ErrInsufficientFunds + fungsi Withdraw; cek dengan errors.Is.
func latihan1() {
	fmt.Println("\n-- Latihan 1: sentinel + errors.Is --")
	// TODO
}

// Latihan 2: ValidationError{Field, Msg}; ekstrak dengan errors.As lalu cetak Field.
func latihan2() {
	fmt.Println("\n-- Latihan 2: errors.As --")
	// TODO
}

// Latihan 3: error berlapis repo -> service -> handler, tiap lapis bungkus %w.
func latihan3() {
	fmt.Println("\n-- Latihan 3: wrapping %w --")
	// TODO
}

// Latihan 4: safeDivide(a, b int) (int, error) memakai recover untuk panic bagi nol.
func latihan4() {
	fmt.Println("\n-- Latihan 4: recover --")
	// TODO
}

// Latihan 5: gabungkan beberapa error validasi dengan errors.Join.
func latihan5() {
	fmt.Println("\n-- Latihan 5: errors.Join --")
	// TODO
}
