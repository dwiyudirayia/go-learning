// Package main untuk modul 03 — Struct, Method, Pointer vs Value, Embedding.
// Jalankan: go run ./03-structs-methods
package main

import (
	"errors"
	"fmt"
)

func main() {
	fmt.Println("=== 03 — Struct, Method, Embedding ===")
	structDasar()
	valueVsPointerReceiver()
	konstruktorDanError()
	embedding()
}

// ------------------------------------------------------------------
// 1. Struct: literal, zero value, perbandingan, value semantics
// ------------------------------------------------------------------
type User struct {
	ID    int
	Name  string
	Email string
}

func structDasar() {
	fmt.Println("\n-- Struct dasar --")

	u := User{ID: 1, Name: "Ana", Email: "ana@mail.id"} // literal by field
	fmt.Printf("user = %+v\n", u)                       // %+v cetak nama field

	var kosong User // zero value: semua field default
	fmt.Printf("zero value user = %+v (siap pakai tanpa konstruktor)\n", kosong)

	// Struct dibandingkan dengan == jika semua field comparable
	a := User{ID: 1, Name: "Ana", Email: "ana@mail.id"}
	fmt.Printf("u == a ? %t (perbandingan per-field)\n", u == a)

	// Value semantics: struct di-copy saat di-assign
	b := u
	b.Name = "Berubah"
	fmt.Printf("u.Name=%q b.Name=%q (b salinan, tidak memengaruhi u)\n", u.Name, b.Name)
}

// ------------------------------------------------------------------
// 2. Value receiver vs Pointer receiver
// ------------------------------------------------------------------
type Counter struct {
	value int
}

// Value receiver: bekerja pada SALINAN -> perubahan tidak bertahan.
func (c Counter) IncBroken() {
	c.value++ // hanya mengubah salinan lokal
}

// Pointer receiver: bekerja pada struct ASLI -> perubahan bertahan.
func (c *Counter) Inc() {
	c.value++
}

func (c Counter) Value() int { return c.value }

func valueVsPointerReceiver() {
	fmt.Println("\n-- Value vs Pointer receiver --")

	c := Counter{}
	c.IncBroken()
	c.IncBroken()
	fmt.Printf("setelah 2x IncBroken (value receiver): %d (tetap 0!)\n", c.Value())

	c.Inc()
	c.Inc()
	c.Inc()
	// c addressable, jadi Go otomatis pakai (&c).Inc()
	fmt.Printf("setelah 3x Inc (pointer receiver): %d\n", c.Value())
}

// ------------------------------------------------------------------
// 3. Konstruktor NewXxx + validasi mengembalikan error
// ------------------------------------------------------------------
type Rectangle struct {
	Width, Height float64
}

func NewRectangle(w, h float64) (*Rectangle, error) {
	if w <= 0 || h <= 0 {
		return nil, errors.New("width & height harus > 0")
	}
	return &Rectangle{Width: w, Height: h}, nil
}

func (r Rectangle) Area() float64 { return r.Width * r.Height }

// Pointer receiver karena mengubah state.
func (r *Rectangle) Scale(f float64) {
	r.Width *= f
	r.Height *= f
}

func konstruktorDanError() {
	fmt.Println("\n-- Konstruktor & error --")

	r, err := NewRectangle(3, 4)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Printf("rect %+v, area=%.1f\n", *r, r.Area())

	r.Scale(2)
	fmt.Printf("setelah Scale(2): %+v, area=%.1f\n", *r, r.Area())

	if _, err := NewRectangle(-1, 5); err != nil {
		fmt.Println("input tidak valid ditolak ->", err)
	}
}

// ------------------------------------------------------------------
// 4. Embedding (komposisi) + promosi method + override
// ------------------------------------------------------------------
type Person struct {
	Name string
}

func (p Person) Greet() string {
	return "Halo, saya " + p.Name
}

type Employee struct {
	Person // embedded -> Name & Greet() dipromosikan
	Title  string
}

// Override: Employee punya Greet() sendiri yang "menutup" milik Person.
func (e Employee) Greet() string {
	// Masih bisa memanggil versi Person secara eksplisit.
	return e.Person.Greet() + fmt.Sprintf(" (%s)", e.Title)
}

func embedding() {
	fmt.Println("\n-- Embedding --")

	e := Employee{
		Person: Person{Name: "Ciko"},
		Title:  "Engineer",
	}
	fmt.Printf("e.Name (promoted) = %q\n", e.Name) // akses field promoted
	fmt.Printf("e.Greet() (override) = %q\n", e.Greet())
	fmt.Printf("e.Person.Greet() (asli) = %q\n", e.Person.Greet())
}
