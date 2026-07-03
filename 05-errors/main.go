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

	// errors.As menembus rantai & mengisi target dengan *ValidationError.
	var ve *ValidationError
	if errors.As(err, &ve) {
		fmt.Printf("berhasil diekstrak -> Field=%q Msg=%q\n", ve.Field, ve.Msg)
	}
}

// ------------------------------------------------------------------
// 3. errors.Join — gabungkan banyak error (Go 1.20+)
// ------------------------------------------------------------------
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
