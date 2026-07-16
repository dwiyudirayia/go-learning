// Teknik Advanced Modul 04 (interfaces) — contoh runnable + penjelasan detail.
//
// Jalankan: go run ./04-interfaces/advanced
package main

import "fmt"

// ===========================================================================
// 1. TYPED-NIL TRAP — jebakan interface klasik di Go
// ===========================================================================
// Interface value menyimpan DUA hal: (tipe, nilai). Ia == nil HANYA jika
// keduanya nil. Bila kita masukkan pointer nil BERTIPE, tipenya tidak nil,
// jadi interface != nil meski nilainya nil. Ini bikin "if err != nil" bocor
// walau errornya "tidak ada".
//
// 🔍 Analogi: interface itu seperti AMPLOP BERLABEL. "nil sejati" = tidak ada
// amplop sama sekali. Typed nil = amplopnya ADA dan berlabel (*myErr), cuma
// isinya kosong — satpam (if err != nil) tetap melihat "ada amplop!" dan
// membunyikan alarm padahal isinya tak ada.

type myErr struct{}

func (*myErr) Error() string { return "boom" }

// SALAH: mengembalikan *myErr konkret. Saat gagal=false, e tetap *myErr(nil),
// dan begitu di-assign ke error, interface-nya != nil.
func bikinErrorSalah(gagal bool) error {
	var e *myErr
	if gagal {
		e = &myErr{}
	}
	return e // <- typed nil: SELALU != nil
}

// BENAR: kembalikan literal nil untuk interface saat memang tak ada error.
func bikinErrorBenar(gagal bool) error {
	if gagal {
		return &myErr{}
	}
	return nil // <- interface nil sejati
}

// ===========================================================================
// 2. Sealed interface — hanya paket ini yang boleh mengimplementasikan
// ===========================================================================
// Menaruh method UNEXPORTED (huruf kecil) di interface membuat paket lain tak
// bisa memenuhinya (mereka tak bisa memanggil/mendeklarasikan method itu).
// Berguna untuk enum tipe tertutup / sum type ala Go.
//
// 🔍 Analogi: sealed interface itu seperti KLUB DENGAN KARTU ANGGOTA yang
// hanya dicetak di dalam gedung (method unexported) — orang luar tak mungkin
// membuat kartunya sendiri, jadi daftar anggota sepenuhnya kamu kendalikan.

type Bentuk interface {
	Luas() float64
	sealed() // unexported -> "menyegel" interface
}

type Persegi struct{ Sisi float64 }

func (p Persegi) Luas() float64 { return p.Sisi * p.Sisi }
func (Persegi) sealed()         {}

type Lingkaran struct{ R float64 }

func (l Lingkaran) Luas() float64 { return 3.14159 * l.R * l.R }
func (Lingkaran) sealed()         {}

// ===========================================================================
// 3. Type switch multi-tipe + comma-ok assertion
// ===========================================================================
// Satu case bisa menangani beberapa tipe sekaligus. case nil menangani nilai
// nil. Assertion bentuk v, ok := x.(T) aman (tak panic) — pakai ok untuk cek.
//
// 🔍 Analogi: type switch itu seperti MEJA SORTIR PAKET — tiap paket (any)
// dicek isinya lalu diarahkan ke jalur yang tepat. Comma-ok assertion seperti
// membuka paket dengan hati-hati: kalau isinya bukan yang diduga, kamu dapat
// jawaban "bukan" (ok=false), bukan paketnya meledak (panic).
func jelaskan(x any) string {
	switch v := x.(type) {
	case nil:
		return "kosong (nil)"
	case int, int64: // dua tipe dalam satu case -> v bertipe any di sini
		return fmt.Sprintf("bilangan bulat: %v", v)
	case string:
		return fmt.Sprintf("teks panjang %d: %q", len(v), v)
	case Bentuk:
		return fmt.Sprintf("bentuk berluas %.2f", v.Luas())
	default:
		return fmt.Sprintf("tipe tak dikenal: %T", v)
	}
}

func main() {
	fmt.Println("== 1. typed-nil trap ==")
	fmt.Printf("  bikinErrorSalah(false) != nil ? %v  <- JEBAKAN!\n", bikinErrorSalah(false) != nil)
	fmt.Printf("  bikinErrorBenar(false) != nil ? %v  <- benar\n", bikinErrorBenar(false) != nil)

	fmt.Println("== 2. sealed interface ==")
	for _, b := range []Bentuk{Persegi{3}, Lingkaran{2}} {
		fmt.Printf("  %T luas=%.2f\n", b, b.Luas())
	}

	fmt.Println("== 3. type switch + comma-ok ==")
	for _, x := range []any{42, "halo", Persegi{5}, 3.14, nil} {
		fmt.Println("  -", jelaskan(x))
	}

	// comma-ok: aman walau tipe salah (tidak panic)
	var x any = "bukan angka"
	if n, ok := x.(int); ok {
		fmt.Println("  angka:", n)
	} else {
		fmt.Println("  assertion gagal dgn aman (ok=false), bukan panic")
	}
}
