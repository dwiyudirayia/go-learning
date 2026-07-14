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

	// 🔍 Analogi: variabel itu seperti KOTAK BERLABEL. Labelnya = nama (a, b, c),
	// isinya = nilai. "Tipe" menentukan kotak ini boleh diisi barang jenis apa
	// (angka bulat, angka desimal, teks) — dan Go melarang salah taruh barang.
	var a int = 10 // deklarasi eksplisit: "kotak a khusus angka bulat, isi 10"
	var b = 3.5    // tipe di-infer jadi float64 (Go menebak dari isinya: desimal)
	c := "Go"      // short declaration := (hanya boleh di dalam fungsi)

	// Zero value: variabel tanpa nilai awal punya nilai default.
	// 🔍 Analogi: kotak baru di Go tidak pernah "berisi sampah acak" seperti di
	// beberapa bahasa lain. Kotak kosong sudah otomatis diisi nilai netral —
	// seperti buku tabungan baru yang saldonya pasti 0, bukan angka misterius.
	var kosong int  // 0
	var teks string // "" (string kosong)
	var aktif bool  // false
	var ptr *int    // nil
	fmt.Printf("zero values -> int:%d string:%q bool:%t ptr:%v\n", kosong, teks, aktif, ptr)

	// Go TIDAK mengonversi tipe otomatis. Harus eksplisit.
	// 🔍 Analogi: mencampur int & float64 itu seperti menjumlah "3 apel + 2 liter".
	// Go menolak sampai kamu samakan satuannya dulu: float64(a) = "ubah 3 apel jadi
	// bentuk desimal 3.0 dulu", baru boleh dijumlah. Ribet? Justru mencegah bug diam-diam.
	hasil := float64(a) + b
	fmt.Printf("a=%d b=%.1f c=%q  float64(a)+b=%.1f\n", a, b, c, hasil)
}

// ------------------------------------------------------------------
// 2. Konstanta & iota (enum ala Go)
// ------------------------------------------------------------------
type Hari int

// 🔍 Analogi: iota itu seperti MESIN NOMOR ANTRIAN di bank. Baris pertama dapat
// nomor 0, tiap baris berikutnya nomor +1 otomatis — kamu tak perlu menulis
// 0,1,2,3 sendiri. Kalau nanti ada hari disisipkan, nomor lain ikut bergeser rapi.
const (
	Senin  Hari = iota // 0
	Selasa             // 1
	Rabu               // 2
	Kamis              // 3
	Jumat              // 4
	Sabtu              // 5
	Minggu             // 6
)

// 🔍 Analogi: method String() itu seperti JURU BAHASA. Di dalam, komputer cuma
// paham angka 0..6. Saat kita mau menampilkannya ke manusia, String() menerjemahkan
// angka itu jadi nama hari yang bisa dibaca ("0" -> "Senin"). fmt otomatis memanggilnya.
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
	// 🔍 Analogi: "if n := 42; ..." itu seperti "ambil dulu barangnya (n), baru
	// periksa di tempat". Variabel n hanya hidup di dalam if ini — habis itu dibuang,
	// jadi tak mengotori kode di luar. Rapi, seperti pisau yang langsung dicuci setelah dipakai.
	if n := 42; n%2 == 0 {
		fmt.Printf("%d genap\n", n)
	} else {
		fmt.Printf("%d ganjil\n", n)
	}

	// switch: tidak perlu 'break', tidak fall-through otomatis.
	// 🔍 Analogi: switch di Go itu seperti SATPAM yang mengecek dari atas ke bawah dan
	// BERHENTI di pintu pertama yang cocok — tak lanjut nyelonong ke pintu berikutnya
	// (beda dengan C/Java yang butuh 'break' biar tak "kebablasan").
	nilai := 85
	switch {
	case nilai >= 90:
		fmt.Println("Grade A")
	case nilai >= 80:
		fmt.Println("Grade B")
	default:
		fmt.Println("Grade C")
	}

	// 🔍 Analogi: Go cuma punya SATU alat perulangan, yaitu 'for' — seperti pisau
	// serbaguna. Bahasa lain punya for, while, do-while terpisah; Go menyatukannya.
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
// 🔍 Analogi: multiple return itu seperti mesin pembagi yang sekali proses langsung
// memberi DUA hasil: hasil bagi + sisa (persis pembagian bersusun di SD: 17:5 = 3 sisa 2).
func bagi(a, b int) (int, int) { // multiple return: hasil bagi & sisa
	return a / b, a % b
}

// 🔍 Analogi: variadic (nums ...int) itu seperti KOTAK AMAL — boleh menerima berapa pun
// koin: 0, 1, atau 100. Di dalam fungsi, semua koin itu dianggap satu slice 'nums'.
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
// 🔍 Analogi: defer itu seperti TUMPUKAN PIRING. Tiap 'defer' menaruh tugas di atas
// tumpukan; saat fungsi selesai, tugas diambil dari ATAS dulu (yang terakhir ditaruh,
// pertama dikerjakan = LIFO). Gunanya: menunda "beres-beres" (tutup file, tutup koneksi)
// agar pasti dijalankan walau fungsi keluar lewat jalur mana pun.
func contohDefer() {
	fmt.Println("\n-- defer (LIFO) --")
	defer fmt.Println("3. ini jalan terakhir (defer pertama)")
	defer fmt.Println("2. ini jalan kedua")
	fmt.Println("1. ini jalan lebih dulu")
}
