// Tempat latihanmu untuk Modul 04.
// Jalankan: go run ./04-interfaces/jawaban-saya
// Isi bagian TODO, lalu bandingkan dengan ../latihan/solusi.go.
package main

import "fmt"

func main() {
	fmt.Println("=== Jawaban saya - Modul 04 ===")
	latihan1()
	latihan2()
	latihan3()
	latihan4()
	latihan5()
}

// Latihan 1: interface Shape{Area()}; implementasi Circle & Rectangle;
// fungsi totalArea(shapes []Shape) float64.
func latihan1() {
	fmt.Println("\n-- Latihan 1: Shape & totalArea --")
	// TODO
}

// Latihan 2: describe(i any) memakai type switch (int, string, bool, Shape, default).
func latihan2() {
	fmt.Println("\n-- Latihan 2: type switch --")
	// TODO
}

// Latihan 3: implementasikan fmt.Stringer pada tipe Money (mis. Rp1.000).
func latihan3() {
	fmt.Println("\n-- Latihan 3: Stringer --")
	// TODO
}

// Latihan 4: tipe yang mengimplementasikan io.Writer (hitung total byte),
// lalu pakai dengan fmt.Fprintf.
func latihan4() {
	fmt.Println("\n-- Latihan 4: io.Writer --")
	// TODO
}

// Latihan 5: tunjukkan jebakan typed nil (return error dari *MyError nil).
func latihan5() {
	fmt.Println("\n-- Latihan 5: typed nil --")
	// TODO
}
