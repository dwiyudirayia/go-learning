// Tempat latihanmu untuk Modul 03.
// Jalankan: go run ./03-structs-methods/jawaban-saya
// Isi bagian TODO, lalu bandingkan dengan ../latihan/solusi.go.
package main

import "fmt"

func main() {
	fmt.Println("=== Jawaban saya - Modul 03 ===")
	latihan1dan2()
	latihan3()
	latihan4()
	latihan5()
}

// Latihan 1 & 2: Rectangle dengan Area()/Perimeter() (value receiver)
// dan Scale(factor) (pointer receiver) yang perubahannya bertahan.
// TODO: definisikan `type Rectangle struct { Width, Height float64 }` + method-nya.
func latihan1dan2() {
	fmt.Println("\n-- Latihan 1 & 2: Rectangle --")
	// TODO
}

// Latihan 3: konstruktor NewRectangle(w,h) (*Rectangle, error) yang menolak <= 0.
func latihan3() {
	fmt.Println("\n-- Latihan 3: konstruktor --")
	// TODO
}

// Latihan 4: Account{Balance int} dengan Deposit & Withdraw (error saldo kurang).
func latihan4() {
	fmt.Println("\n-- Latihan 4: Account --")
	// TODO
}

// Latihan 5: Employee embed Person; panggil Greet() (promoted) lalu override.
func latihan5() {
	fmt.Println("\n-- Latihan 5: embedding --")
	// TODO
}
