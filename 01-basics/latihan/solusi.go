// Solusi latihan Modul 01.
// Jalankan: go run ./01-basics/latihan
//
// File ini sengaja dipisah di package sendiri (folder latihan/) supaya tidak
// bentrok dengan main.go di 01-basics — di Go, satu folder = satu package.
package main

import (
	"errors"
	"fmt"
	"math"
)

func main() {
	fmt.Println("=== Solusi Latihan Modul 01 ===")
	latihan1()
	latihan2()
	latihan3()
	latihan4()
}

// ------------------------------------------------------------------
// Latihan 1: celsiusToFahrenheit + cetak untuk 0, 37, 100
// ------------------------------------------------------------------
// Idiom: parameter & return bertipe eksplisit. Tidak ada konversi implisit,
// jadi angka 9.0/5.0 ditulis sebagai float64 agar tidak jadi pembagian integer.
func celsiusToFahrenheit(c float64) float64 {
	return c*9.0/5.0 + 32
}

func latihan1() {
	fmt.Println("\n-- Latihan 1: Celsius -> Fahrenheit --")
	// range atas literal slice: cara ringkas iterasi beberapa nilai tetap.
	for _, c := range []float64{0, 37, 100} {
		fmt.Printf("%6.1f°C = %6.1f°F\n", c, celsiusToFahrenheit(c))
	}
	// CATATAN idiom: kalau ditulis c*9/5, dan c bertipe float64, tetap float64.
	// Tapi kalau bekerja dgn int, "9/5" jadi 1 (integer division) — sumber bug klasik.
}

// ------------------------------------------------------------------
// Latihan 2: enum hari memakai iota
// ------------------------------------------------------------------
// Idiom enum Go: definisikan tipe baru (Hari) + konstanta ber-iota + method String().
// iota mulai 0 di tiap blok const dan naik 1 tiap baris.
type Hari int

const (
	Senin  Hari = iota // 0
	Selasa             // 1
	Rabu               // 2
	Kamis              // 3
	Jumat              // 4
	Sabtu              // 5
	Minggu             // 6
)

// String() membuat Hari otomatis tercetak sebagai nama saat dipakai %v/Println.
// Ini implementasi interface fmt.Stringer — pola sangat umum di Go.
func (h Hari) String() string {
	nama := [...]string{"Senin", "Selasa", "Rabu", "Kamis", "Jumat", "Sabtu", "Minggu"}
	if h < Senin || h > Minggu {
		return fmt.Sprintf("Hari(%d)", int(h)) // defensif untuk nilai di luar range
	}
	return nama[h]
}

func latihan2() {
	fmt.Println("\n-- Latihan 2: enum Hari (iota) --")
	for h := Senin; h <= Minggu; h++ {
		fmt.Printf("%d = %v\n", int(h), h) // %v memanggil String()
	}
	weekend := Sabtu
	fmt.Printf("Contoh: hari libur = %v (nilai int %d)\n", weekend, int(weekend))
}

// ------------------------------------------------------------------
// Latihan 3: minMax variadic + named return
// ------------------------------------------------------------------
// Idiom: variadic (...int) menerima 0..n argumen sebagai slice.
// Named return (min, max) sudah terdeklarasi, jadi cukup `return` di akhir.
// Karena len 0 tidak punya jawaban valid, kita kembalikan error — idiom Go.
func minMax(nums ...int) (min, max int, err error) {
	if len(nums) == 0 {
		return 0, 0, errors.New("minMax: butuh minimal 1 angka")
	}
	min, max = nums[0], nums[0]
	for _, n := range nums[1:] {
		if n < min {
			min = n
		}
		if n > max {
			max = n
		}
	}
	return // mengembalikan min, max, err(=nil)
}

func latihan3() {
	fmt.Println("\n-- Latihan 3: minMax (variadic + named return) --")
	if mn, mx, err := minMax(4, 9, 1, 7, 3, -2); err == nil {
		fmt.Printf("min=%d max=%d\n", mn, mx)
	}
	// Uji jalur error: idiom `if val, err := f(); err != nil { ... }`
	if _, _, err := minMax(); err != nil {
		fmt.Println("kasus kosong ->", err)
	}
}

// ------------------------------------------------------------------
// Latihan 4: defer mencetak "selesai" setelah fungsi berjalan
// ------------------------------------------------------------------
// defer dievaluasi argumennya SEKARANG tapi dieksekusi saat fungsi keluar.
// Berguna untuk cleanup (tutup file, unlock) — dijamin jalan walau ada return awal.
func latihan4() {
	fmt.Println("\n-- Latihan 4: defer --")
	defer fmt.Println(">> selesai (dicetak saat latihan4 keluar)")

	fmt.Println("mulai proses...")
	fmt.Printf("akar 144 = %.0f\n", math.Sqrt(144))
	fmt.Println("proses inti selesai, fungsi akan return sebentar lagi")
	// Baris defer di atas baru tercetak SETELAH baris ini.
}
