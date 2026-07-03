// Solusi latihan Modul 05 — Error Handling.
// Jalankan: go run ./05-errors/latihan
package main

import (
	"errors"
	"fmt"
)

func main() {
	fmt.Println("=== Solusi Latihan Modul 05 ===")
	latihan1()
	latihan2()
	latihan3()
	latihan4()
	latihan5()
}

// ------------------------------------------------------------------
// Latihan 1: sentinel ErrInsufficientFunds + errors.Is
// ------------------------------------------------------------------
var ErrInsufficientFunds = errors.New("saldo tidak mencukupi")

type Account struct{ Balance int }

func (a *Account) Withdraw(amount int) error {
	if amount > a.Balance {
		return ErrInsufficientFunds
	}
	a.Balance -= amount
	return nil
}

func latihan1() {
	fmt.Println("\n-- Latihan 1: sentinel + errors.Is --")
	acc := &Account{Balance: 100}
	err := acc.Withdraw(500)
	if errors.Is(err, ErrInsufficientFunds) {
		fmt.Println("terdeteksi ErrInsufficientFunds ✅ ->", err)
	}
}

// ------------------------------------------------------------------
// Latihan 2: ValidationError + errors.As
// ------------------------------------------------------------------
type ValidationError struct {
	Field string
	Msg   string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("field %q tidak valid: %s", e.Field, e.Msg)
}

func validateUser(name string) error {
	if name == "" {
		return &ValidationError{Field: "name", Msg: "wajib diisi"}
	}
	return nil
}

func latihan2() {
	fmt.Println("\n-- Latihan 2: custom error + errors.As --")
	err := fmt.Errorf("registrasi gagal: %w", validateUser(""))
	var ve *ValidationError
	if errors.As(err, &ve) {
		fmt.Printf("diekstrak -> Field=%q Msg=%q\n", ve.Field, ve.Msg)
	}
}

// ------------------------------------------------------------------
// Latihan 3: error berlapis repo -> service -> handler dengan %w
// ------------------------------------------------------------------
var ErrNotFound = errors.New("data tidak ditemukan")

func repoLayer(id int) error {
	if id <= 0 {
		return ErrNotFound
	}
	return nil
}

func serviceLayer(id int) error {
	if err := repoLayer(id); err != nil {
		return fmt.Errorf("service.GetUser(%d): %w", id, err)
	}
	return nil
}

func handlerLayer(id int) error {
	if err := serviceLayer(id); err != nil {
		return fmt.Errorf("handler: gagal memproses request: %w", err)
	}
	return nil
}

func latihan3() {
	fmt.Println("\n-- Latihan 3: error berlapis (%w) --")
	err := handlerLayer(-1)
	fmt.Printf("rantai penuh: %v\n", err)
	if errors.Is(err, ErrNotFound) {
		fmt.Println("meski 2x dibungkus, errors.Is tetap menemukan ErrNotFound ✅")
	}
}

// ------------------------------------------------------------------
// Latihan 4: safeDivide dengan recover
// ------------------------------------------------------------------
func safeDivide(a, b int) (result int, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("pulih dari panic: %v", r)
		}
	}()
	return a / b, nil // b==0 -> panic, ditangkap defer di atas
}

func latihan4() {
	fmt.Println("\n-- Latihan 4: recover safeDivide --")
	if r, err := safeDivide(20, 4); err == nil {
		fmt.Printf("20 / 4 = %d\n", r)
	}
	if _, err := safeDivide(1, 0); err != nil {
		fmt.Printf("1 / 0 -> %v (program tetap jalan)\n", err)
	}
}

// ------------------------------------------------------------------
// Latihan 5: errors.Join menggabungkan banyak error
// ------------------------------------------------------------------
func validateForm(name, email string, age int) error {
	var errs []error
	if name == "" {
		errs = append(errs, &ValidationError{Field: "name", Msg: "wajib diisi"})
	}
	if email == "" {
		errs = append(errs, &ValidationError{Field: "email", Msg: "wajib diisi"})
	}
	if age < 17 {
		errs = append(errs, &ValidationError{Field: "age", Msg: "minimal 17"})
	}
	return errors.Join(errs...) // nil jika tidak ada error
}

func latihan5() {
	fmt.Println("\n-- Latihan 5: errors.Join --")
	if err := validateForm("", "", 10); err != nil {
		fmt.Println("form tidak valid (semua field dilaporkan):")
		fmt.Println(err)
	}
}
