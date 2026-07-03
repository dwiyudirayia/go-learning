// Tempat latihanmu untuk Modul 09.
// Jalankan: go run ./09-stdlib/jawaban-saya
// Isi bagian TODO, lalu bandingkan dengan ../latihan/solusi.go.
package main

import "fmt"

func main() {
	fmt.Println("=== Jawaban saya - Modul 09 ===")
	latihan1()
	latihan2()
	latihan3()
	latihan4()
	latihan5()
}

// Latihan 1: struct Book{Title, Author string; Year int; Tags []string} dengan
// tag JSON; marshal & unmarshal.
func latihan1() {
	fmt.Println("\n-- Latihan 1: JSON Book --")
	// TODO
}

// Latihan 2: parse "2026-07-01", tambah 90 hari, format "02 Jan 2006".
func latihan2() {
	fmt.Println("\n-- Latihan 2: time --")
	// TODO: gunakan time.Parse dan t.AddDate(0, 0, 90).
}

// Latihan 3: wordCount(r io.Reader) map[string]int yang membaca dari Reader apa pun.
func latihan3() {
	fmt.Println("\n-- Latihan 3: wordCount(io.Reader) --")
	// TODO
}

// Latihan 4: tulis & baca kembali file JSON sementara berisi []Book.
func latihan4() {
	fmt.Println("\n-- Latihan 4: file JSON --")
	// TODO: pakai os.CreateTemp, json.MarshalIndent, os.WriteFile, os.ReadFile.
}

// Latihan 5: handler HTTP kecil yang mengembalikan JSON, uji dengan httptest.
func latihan5() {
	fmt.Println("\n-- Latihan 5: handler + httptest --")
	// TODO
}
