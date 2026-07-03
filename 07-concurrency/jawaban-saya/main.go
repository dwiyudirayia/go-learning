// Tempat latihanmu untuk Modul 07.
// Jalankan:        go run ./07-concurrency/jawaban-saya
// Dengan race det: go run -race ./07-concurrency/jawaban-saya
// Isi bagian TODO, lalu bandingkan dengan ../latihan/solusi.go.
package main

import "fmt"

func main() {
	fmt.Println("=== Jawaban saya - Modul 07 ===")
	latihan1()
	latihan2()
	latihan3()
	latihan4()
	latihan5()
}

// Latihan 1: 5 goroutine mencetak ID-nya; pakai sync.WaitGroup agar main menunggu.
func latihan1() {
	fmt.Println("\n-- Latihan 1: goroutine + WaitGroup --")
	// TODO
}

// Latihan 2: generator -> <-chan int berisi 1..n, konsumsi dengan range.
func latihan2() {
	fmt.Println("\n-- Latihan 2: generator channel --")
	// TODO
}

// Latihan 3: Counter aman-konkuren pakai sync.Mutex; naikkan dari 100 goroutine.
// Hasil harus tepat 100. Uji dengan `go run -race`.
func latihan3() {
	fmt.Println("\n-- Latihan 3: Counter + Mutex --")
	// TODO
}

// Latihan 4: worker pool 3 worker memproses 9 job (kuadratkan), kumpulkan hasil.
func latihan4() {
	fmt.Println("\n-- Latihan 4: worker pool --")
	// TODO
}

// Latihan 5: fetchWithTimeout(d) pakai context.WithTimeout + select.
func latihan5() {
	fmt.Println("\n-- Latihan 5: context timeout --")
	// TODO
}
