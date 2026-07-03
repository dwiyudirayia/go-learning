// Solusi latihan Modul 03 — Struct, Method, Embedding.
// Jalankan: go run ./03-structs-methods/latihan
package main

import (
	"errors"
	"fmt"
)

func main() {
	fmt.Println("=== Solusi Latihan Modul 03 ===")
	latihan1dan2()
	latihan3()
	latihan4()
	latihan5()
}

// ------------------------------------------------------------------
// Latihan 1 & 2: Rectangle dengan Area/Perimeter (value) + Scale (pointer)
// ------------------------------------------------------------------
type Rectangle struct {
	Width, Height float64
}

// Value receiver: hanya membaca, tidak mengubah -> aman pakai T.
func (r Rectangle) Area() float64      { return r.Width * r.Height }
func (r Rectangle) Perimeter() float64 { return 2 * (r.Width + r.Height) }

// Pointer receiver: MENGUBAH state -> wajib *T agar perubahan bertahan.
func (r *Rectangle) Scale(factor float64) {
	r.Width *= factor
	r.Height *= factor
}

func latihan1dan2() {
	fmt.Println("\n-- Latihan 1 & 2: Rectangle --")
	r := Rectangle{Width: 3, Height: 4}
	fmt.Printf("awal   : %+v area=%.1f perimeter=%.1f\n", r, r.Area(), r.Perimeter())

	r.Scale(2) // Go otomatis pakai (&r).Scale karena r addressable
	fmt.Printf("scale2 : %+v area=%.1f (perubahan bertahan berkat pointer receiver)\n", r, r.Area())
}

// ------------------------------------------------------------------
// Latihan 3: konstruktor NewRectangle dengan validasi
// ------------------------------------------------------------------
func NewRectangle(w, h float64) (*Rectangle, error) {
	if w <= 0 || h <= 0 {
		return nil, fmt.Errorf("dimensi tidak valid: w=%.1f h=%.1f (harus > 0)", w, h)
	}
	return &Rectangle{Width: w, Height: h}, nil
}

func latihan3() {
	fmt.Println("\n-- Latihan 3: konstruktor + validasi --")
	if r, err := NewRectangle(5, 2); err == nil {
		fmt.Printf("valid   -> %+v\n", *r)
	}
	if _, err := NewRectangle(0, 3); err != nil {
		fmt.Println("ditolak ->", err)
	}
}

// ------------------------------------------------------------------
// Latihan 4: Account dengan Deposit & Withdraw (error saldo kurang)
// ------------------------------------------------------------------
var ErrInsufficientFunds = errors.New("saldo tidak mencukupi")

type Account struct {
	Balance int
}

// Pointer receiver karena mengubah Balance.
func (a *Account) Deposit(amount int) error {
	if amount <= 0 {
		return errors.New("jumlah deposit harus > 0")
	}
	a.Balance += amount
	return nil
}

func (a *Account) Withdraw(amount int) error {
	if amount <= 0 {
		return errors.New("jumlah tarik harus > 0")
	}
	if amount > a.Balance {
		return ErrInsufficientFunds
	}
	a.Balance -= amount
	return nil
}

func latihan4() {
	fmt.Println("\n-- Latihan 4: Account --")
	acc := &Account{Balance: 100}
	_ = acc.Deposit(50)
	fmt.Printf("setelah deposit 50 -> saldo=%d\n", acc.Balance)

	if err := acc.Withdraw(30); err == nil {
		fmt.Printf("tarik 30 sukses   -> saldo=%d\n", acc.Balance)
	}
	if err := acc.Withdraw(1000); err != nil {
		fmt.Printf("tarik 1000 gagal  -> %v (saldo tetap %d)\n", err, acc.Balance)
	}
}

// ------------------------------------------------------------------
// Latihan 5: Embedding + promosi method + override
// ------------------------------------------------------------------
type Person struct {
	Name string
}

func (p Person) Greet() string { return "Halo, saya " + p.Name }

type Employee struct {
	Person // embedded -> Name & Greet() dipromosikan
	Salary int
}

// Override Greet() khusus Employee, tetap memanfaatkan versi Person.
func (e Employee) Greet() string {
	return e.Person.Greet() + fmt.Sprintf(" (gaji %d)", e.Salary)
}

func latihan5() {
	fmt.Println("\n-- Latihan 5: Embedding & override --")

	p := Person{Name: "Ana"}
	fmt.Println("Person.Greet()   ->", p.Greet())

	e := Employee{Person: Person{Name: "Budi"}, Salary: 5000}
	fmt.Println("Employee.Greet() ->", e.Greet())        // versi override
	fmt.Println("e.Person.Greet() ->", e.Person.Greet()) // versi asli (promoted)
	fmt.Printf("akses field promoted: e.Name=%q\n", e.Name)
}
