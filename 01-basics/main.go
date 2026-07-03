// Package main untuk modul 01 — Dasar Go & Idiom.
// Jalankan: go run ./01-basics
package main

import "fmt"

func main() {
	fmt.Println("=== 01 — Dasar Go & Idiom ===")

	variabelDanTipe()
	konstantaDanIota()
	kontrolAlur()
	contohFungsi()
	contohDefer()
}

// ------------------------------------------------------------------
// 1. Variabel, zero value, dan konversi tipe
// ------------------------------------------------------------------
func variabelDanTipe() {
	fmt.Println("\n-- Variabel & Tipe --")

	var a int = 10 // deklarasi eksplisit
	var b = 3.5    // tipe di-infer jadi float64
	c := "Go"      // short declaration (hanya boleh di dalam fungsi)

	// Zero value: variabel tanpa nilai awal punya nilai default.
	var kosong int  // 0
	var teks string // "" (string kosong)
	var aktif bool  // false
	var ptr *int    // nil
	fmt.Printf("zero values -> int:%d string:%q bool:%t ptr:%v\n", kosong, teks, aktif, ptr)

	// Go TIDAK mengonversi tipe otomatis. Harus eksplisit.
	hasil := float64(a) + b
	fmt.Printf("a=%d b=%.1f c=%q  float64(a)+b=%.1f\n", a, b, c, hasil)
}

// ------------------------------------------------------------------
// 2. Konstanta & iota (enum ala Go)
// ------------------------------------------------------------------
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

func (h Hari) String() string {
	return [...]string{"Senin", "Selasa", "Rabu", "Kamis", "Jumat", "Sabtu", "Minggu"}[h]
}

func konstantaDanIota() {
	fmt.Println("\n-- Konstanta & iota --")
	const Pi = 3.14159
	fmt.Printf("Pi=%.5f\n", Pi)
	fmt.Printf("Hari ke-0=%v, ke-4=%v (nilai int: %d)\n", Senin, Jumat, Jumat)
}

// ------------------------------------------------------------------
// 3. Kontrol alur: if dengan statement, switch tanpa break, satu jenis for
// ------------------------------------------------------------------
func kontrolAlur() {
	fmt.Println("\n-- Kontrol Alur --")

	// if boleh punya statement inisialisasi sebelum kondisi.
	if n := 42; n%2 == 0 {
		fmt.Printf("%d genap\n", n)
	} else {
		fmt.Printf("%d ganjil\n", n)
	}

	// switch: tidak perlu 'break', tidak fall-through otomatis.
	nilai := 85
	switch {
	case nilai >= 90:
		fmt.Println("Grade A")
	case nilai >= 80:
		fmt.Println("Grade B")
	default:
		fmt.Println("Grade C")
	}

	// Hanya ada 'for'. Ini bentuk perulangan biasa...
	jumlah := 0
	for i := 1; i <= 5; i++ {
		jumlah += i
	}
	// ...ini gaya 'while'...
	for jumlah > 10 {
		jumlah -= 3
	}
	fmt.Printf("jumlah akhir=%d\n", jumlah)
}

// ------------------------------------------------------------------
// 4. Fungsi: multiple return, named return, variadic
// ------------------------------------------------------------------
func bagi(a, b int) (int, int) { // multiple return: hasil bagi & sisa
	return a / b, a % b
}

func minMax(nums ...int) (min, max int) { // variadic + named return
	if len(nums) == 0 {
		return 0, 0
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
	return // named return: 'min' dan 'max' otomatis dikembalikan
}

func contohFungsi() {
	fmt.Println("\n-- Fungsi --")
	q, r := bagi(17, 5)
	fmt.Printf("17 / 5 = %d sisa %d\n", q, r)

	mn, mx := minMax(4, 9, 1, 7, 3)
	fmt.Printf("min=%d max=%d\n", mn, mx)
}

// ------------------------------------------------------------------
// 5. defer: dijadwalkan saat fungsi keluar, urutan LIFO
// ------------------------------------------------------------------
func contohDefer() {
	fmt.Println("\n-- defer (LIFO) --")
	defer fmt.Println("3. ini jalan terakhir (defer pertama)")
	defer fmt.Println("2. ini jalan kedua")
	fmt.Println("1. ini jalan lebih dulu")
}
