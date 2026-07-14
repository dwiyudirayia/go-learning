// Package main untuk modul 05 — Error Handling.
// Jalankan: go run ./05-errors
package main

import (
	"errors"
	"fmt"
)

func main() {
	fmt.Println("=== 05 — Error Handling ===")
	sentinelDanWrap()
	errorsIsAs()
	errorsJoin()
	panicRecover()
}

// ------------------------------------------------------------------
// Sentinel errors (dibandingkan by identity)
// ------------------------------------------------------------------
// 🔍 Analogi besar: di Go, error itu BUKAN alarm kebakaran yang tiba-tiba meledak
// (seperti exception di Java/Python). Error itu NILAI BIASA — seperti struk yang
// dikembalikan kasir. Kamu WAJIB melihat struknya ("if err != nil") sebelum lanjut.
// Eksplisit & tak ada kejutan: jalur sukses dan jalur gagal sama-sama terlihat di kode.

// 🔍 Analogi: sentinel error itu seperti KODE ERROR RESMI yang dipajang (mis. "404").
// Dibuat sekali di satu tempat, lalu kode lain membandingkannya dengan errors.Is —
// seperti mencocokkan "apakah ini benar-benar error 'data tidak ditemukan' yang itu?".
var (
	ErrNotFound          = errors.New("data tidak ditemukan")
	ErrInsufficientFunds = errors.New("saldo tidak mencukupi")
)

// ------------------------------------------------------------------
// 1. Sentinel + wrapping dengan %w melalui beberapa lapisan
// ------------------------------------------------------------------

// Lapis repo: sumber error asli (sentinel).
func repoGetUser(id int) error {
	if id <= 0 {
		return ErrNotFound
	}
	return nil
}

// 🔍 Analogi: membungkus error dengan verb-w (%w) itu seperti menempel STIKER catatan
// di atas struk lama tanpa menutupinya: "gagal di lapisan service, saat load user 5".
// Tiap lapisan menambah konteks, tapi struk asli tetap bisa dibaca lewat errors.Is/As.
// Kalau pakai verb-v (%v), struk aslinya cuma jadi teks — identitasnya HILANG, tak bisa dilacak lagi.

// Lapis service: menambah konteks TAPI tetap membungkus asal dengan %w.
func serviceLoadUser(id int) error {
	if err := repoGetUser(id); err != nil {
		return fmt.Errorf("service: gagal load user %d: %w", id, err)
	}
	return nil
}

func sentinelDanWrap() {
	fmt.Println("\n-- Sentinel & wrapping (%w) --")

	err := serviceLoadUser(-1)
	fmt.Printf("rantai error: %v\n", err)

	// Walau sudah dibungkus 2 lapis konteks, errors.Is tetap menemukan sentinel asal.
	if errors.Is(err, ErrNotFound) {
		fmt.Println("errors.Is menemukan ErrNotFound di dalam rantai ✅")
	}
}

// ------------------------------------------------------------------
// 2. Custom error type + errors.As untuk mengekstrak datanya
// ------------------------------------------------------------------
type ValidationError struct {
	Field string
	Msg   string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validasi %s: %s", e.Field, e.Msg)
}

func validateAge(age int) error {
	if age < 0 {
		return &ValidationError{Field: "age", Msg: "tidak boleh negatif"}
	}
	if age > 150 {
		return &ValidationError{Field: "age", Msg: "tidak masuk akal"}
	}
	return nil
}

func errorsIsAs() {
	fmt.Println("\n-- errors.As (ekstrak tipe konkret) --")

	// Bungkus custom error dengan konteks tambahan.
	err := fmt.Errorf("registrasi gagal: %w", validateAge(-5))
	fmt.Printf("rantai error: %v\n", err)

	// 🔍 Analogi: errors.Is = "apakah di tumpukan struk ini ADA struk merek X?" (cocokkan identitas).
	//            errors.As = "cari struk merek X di tumpukan, lalu SERAHKAN ke saya biar bisa
	//            kubaca detailnya (Field, Msg)". Is untuk mengenali; As untuk mengambil isinya.
	// errors.As menembus rantai & mengisi target dengan *ValidationError.
	var ve *ValidationError
	if errors.As(err, &ve) {
		fmt.Printf("berhasil diekstrak -> Field=%q Msg=%q\n", ve.Field, ve.Msg)
	}
}

// ------------------------------------------------------------------
// 3. errors.Join — gabungkan banyak error (Go 1.20+)
// ------------------------------------------------------------------
// 🔍 Analogi: errors.Join itu seperti FORMULIR yang mengumpulkan SEMUA kesalahan sekaligus,
// bukan berhenti di kesalahan pertama. Seperti petugas yang bilang "nama kosong DAN umur
// kurang" dalam sekali sebut — bukan menyuruhmu bolak-balik memperbaiki satu per satu.
func validateForm(name string, age int) error {
	var errs []error
	if name == "" {
		errs = append(errs, &ValidationError{Field: "name", Msg: "wajib diisi"})
	}
	if age < 17 {
		errs = append(errs, &ValidationError{Field: "age", Msg: "minimal 17"})
	}
	return errors.Join(errs...) // nil jika errs kosong
}

func errorsJoin() {
	fmt.Println("\n-- errors.Join (gabung banyak error) --")

	err := validateForm("", 15)
	if err != nil {
		fmt.Println("form tidak valid:")
		fmt.Println(err) // Join mencetak tiap error di baris terpisah
	}

	if validateForm("Ana", 20) == nil {
		fmt.Println("form valid -> Join mengembalikan nil ✅")
	}
}

// ------------------------------------------------------------------
// 4. panic & recover — ubah panic jadi error di batas API
// ------------------------------------------------------------------
// 🔍 Analogi: panic itu "REM DARURAT" — dipakai hanya untuk kondisi benar-benar rusak
// (bug programmer), bukan error biasa. recover() itu seperti JARING PENGAMAN di dalam defer
// yang menangkap orang jatuh, lalu mengubah kepanikan jadi error rapi di batas API — sehingga
// seluruh program tak ikut roboh. Aturan main Go: error biasa pakai nilai, panic hanya untuk darurat.
func safeDivide(a, b int) (result int, err error) {
	// recover hanya bermakna di dalam defer.
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("safeDivide pulih dari panic: %v", r)
		}
	}()
	result = a / b // b==0 -> panic runtime error
	return result, nil
}

func panicRecover() {
	fmt.Println("\n-- panic & recover --")

	if r, err := safeDivide(10, 2); err == nil {
		fmt.Printf("10 / 2 = %d\n", r)
	}

	if _, err := safeDivide(10, 0); err != nil {
		fmt.Printf("10 / 0 ditangani -> %v\n", err)
	}
	fmt.Println("program tetap jalan (tidak crash) ✅")
}
