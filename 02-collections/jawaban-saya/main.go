// Tempat latihanmu untuk Modul 02.
// Jalankan: go run ./02-collections/jawaban-saya
// Isi bagian TODO, lalu bandingkan dengan ../latihan/solusi.go.
package main

import "fmt"

func main() {
	fmt.Println("=== Jawaban saya - Modul 02 ===")
	latihan1()
	latihan2()
	latihan3()
	latihan4()
	latihan5()
}

// Latihan 1: dari slice 1..10, buat slice baru berisi hanya bilangan genap.
func latihan1() {
	fmt.Println("\n-- Latihan 1: filter genap --")
	slices := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	var evenNumbers []int
	for _, n := range slices {
		if n%2 == 0 {
			evenNumbers = append(evenNumbers, n)
		}
	}
	fmt.Println("Bilangan genap:", evenNumbers)
}

// Latihan 2: buktikan jebakan backing array (b := a[:2]; ubah b[0]; cetak a).
func latihan2() {
	fmt.Println("\n-- Latihan 2: backing array --")
	// TODO
}

// Latihan 3: hitung frekuensi tiap kata dari sebuah kalimat pakai map[string]int.
func latihan3() {
	fmt.Println("\n-- Latihan 3: frekuensi kata --")
	// TODO: pakai strings.Fields untuk memecah kalimat.
}

// Latihan 4: hitung jumlah rune vs jumlah byte dari "naïve café 世界".
func latihan4() {
	fmt.Println("\n-- Latihan 4: byte vs rune --")
	// TODO: bandingkan len(s) dengan utf8.RuneCountInString(s).
}

// Latihan 5: balik string dengan benar (aman untuk rune multi-byte).
func latihan5() {
	fmt.Println("\n-- Latihan 5: reverse string --")
	// TODO: konversi ke []rune dulu sebelum membalik.
}
