// Solusi latihan Modul 04 — Interface.
// Jalankan: go run ./04-interfaces/latihan
package main

import (
	"fmt"
	"math"
)

func main() {
	fmt.Println("=== Solusi Latihan Modul 04 ===")
	latihan1()
	latihan2()
	latihan3()
	latihan4()
	latihan5()
}

// ------------------------------------------------------------------
// Latihan 1: interface Shape + Circle/Rectangle + totalArea
// ------------------------------------------------------------------
type Shape interface {
	Area() float64
}

type Circle struct{ R float64 }

func (c Circle) Area() float64 { return math.Pi * c.R * c.R }

type Rectangle struct{ W, H float64 }

func (r Rectangle) Area() float64 { return r.W * r.H }

func totalArea(shapes []Shape) float64 {
	var total float64
	for _, s := range shapes {
		total += s.Area()
	}
	return total
}

func latihan1() {
	fmt.Println("\n-- Latihan 1: Shape & totalArea --")
	shapes := []Shape{Circle{R: 1}, Rectangle{W: 2, H: 3}, Circle{R: 2}}
	for _, s := range shapes {
		fmt.Printf("  %-16T area=%.2f\n", s, s.Area())
	}
	fmt.Printf("total area = %.2f\n", totalArea(shapes))
}

// ------------------------------------------------------------------
// Latihan 2: describe(i any) dengan type switch
// ------------------------------------------------------------------
func describe(i any) string {
	switch v := i.(type) {
	case nil:
		return "nil"
	case int:
		return fmt.Sprintf("int: %d", v)
	case string:
		return fmt.Sprintf("string: %q", v)
	case bool:
		return fmt.Sprintf("bool: %t", v)
	case Shape:
		return fmt.Sprintf("Shape (area %.2f)", v.Area())
	default:
		return fmt.Sprintf("tipe lain (%T): %v", v, v)
	}
}

func latihan2() {
	fmt.Println("\n-- Latihan 2: describe (type switch) --")
	for _, v := range []any{100, "halo", true, Rectangle{W: 2, H: 2}, 3.14, nil} {
		fmt.Printf("  %s\n", describe(v))
	}
}

// ------------------------------------------------------------------
// Latihan 3: fmt.Stringer pada Money
// ------------------------------------------------------------------
type Money int

func (m Money) String() string {
	s := fmt.Sprintf("%d", int(m))
	var out []byte
	for i, d := range []byte(s) {
		if i > 0 && (len(s)-i)%3 == 0 {
			out = append(out, '.')
		}
		out = append(out, d)
	}
	return "Rp" + string(out)
}

func latihan3() {
	fmt.Println("\n-- Latihan 3: Stringer Money --")
	fmt.Printf("  %v\n", Money(1000))
	fmt.Printf("  %v\n", Money(2500000))
}

// ------------------------------------------------------------------
// Latihan 4: implement io.Writer sederhana (hitung byte)
// ------------------------------------------------------------------
type countingWriter struct{ total int }

// Write memenuhi io.Writer: (int, error).
func (w *countingWriter) Write(p []byte) (int, error) {
	w.total += len(p)
	return len(p), nil
}

func latihan4() {
	fmt.Println("\n-- Latihan 4: io.Writer --")
	cw := &countingWriter{}
	// Karena cw io.Writer, fmt.Fprintf bisa menulis ke sana.
	fmt.Fprintf(cw, "Halo %s!", "dunia")
	fmt.Fprint(cw, " tambahan")
	fmt.Printf("  total byte ditulis = %d\n", cw.total)
}

// ------------------------------------------------------------------
// Latihan 5: jebakan typed nil
// ------------------------------------------------------------------
type MyError struct{ Msg string }

func (e *MyError) Error() string { return e.Msg }

// SALAH: mengembalikan pointer bertipe -> interface berisi (tipe,*MyError; nilai,nil)
func returnTypedNil() error {
	var e *MyError // nil
	return e
}

// BENAR
func returnRealNil() error { return nil }

func latihan5() {
	fmt.Println("\n-- Latihan 5: typed nil --")
	fmt.Printf("  returnTypedNil() == nil ? %t  <- MENGEJUTKAN\n", returnTypedNil() == nil)
	fmt.Printf("  returnRealNil()  == nil ? %t\n", returnRealNil() == nil)
}
