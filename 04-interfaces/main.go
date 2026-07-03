// Package main untuk modul 04 — Interface.
// Jalankan: go run ./04-interfaces
package main

import (
	"fmt"
	"math"
	"strings"
)

func main() {
	fmt.Println("=== 04 — Interface ===")
	interfaceDasar()
	typeAssertionDanSwitch()
	stringerDanWriter()
	jebakanTypedNil()
}

// ------------------------------------------------------------------
// 1. Interface & implicit satisfaction (duck typing)
// ------------------------------------------------------------------
type Shape interface {
	Area() float64
	Perimeter() float64
}

type Circle struct{ R float64 }

// Circle otomatis jadi Shape karena punya kedua method — tanpa "implements".
func (c Circle) Area() float64      { return math.Pi * c.R * c.R }
func (c Circle) Perimeter() float64 { return 2 * math.Pi * c.R }

type Rectangle struct{ W, H float64 }

func (r Rectangle) Area() float64      { return r.W * r.H }
func (r Rectangle) Perimeter() float64 { return 2 * (r.W + r.H) }

// "Accept interfaces": fungsi bekerja untuk tipe APA PUN yang memenuhi Shape.
func totalArea(shapes []Shape) float64 {
	var t float64
	for _, s := range shapes {
		t += s.Area()
	}
	return t
}

func interfaceDasar() {
	fmt.Println("\n-- Interface dasar --")

	shapes := []Shape{
		Circle{R: 2},
		Rectangle{W: 3, H: 4},
	}
	for _, s := range shapes {
		// %T mencetak tipe dinamis di balik interface.
		fmt.Printf("%-20T area=%.2f perimeter=%.2f\n", s, s.Area(), s.Perimeter())
	}
	fmt.Printf("total area = %.2f\n", totalArea(shapes))
}

// ------------------------------------------------------------------
// 2. Type assertion & type switch
// ------------------------------------------------------------------
func describe(i any) string {
	// type switch: cabang berdasarkan tipe dinamis.
	switch v := i.(type) {
	case nil:
		return "nil"
	case int:
		return fmt.Sprintf("int: %d", v)
	case string:
		return fmt.Sprintf("string(%d huruf): %q", len(v), v)
	case bool:
		return fmt.Sprintf("bool: %t", v)
	case Shape:
		return fmt.Sprintf("Shape dgn area %.2f", v.Area())
	default:
		return fmt.Sprintf("tipe lain (%T): %v", v, v)
	}
}

func typeAssertionDanSwitch() {
	fmt.Println("\n-- Type assertion & switch --")

	// comma-ok assertion: aman, tidak panic kalau tipe salah.
	var i any = "halo"
	if s, ok := i.(string); ok {
		fmt.Printf("assertion sukses: %q (huruf pertama %c)\n", s, s[0])
	}
	if _, ok := i.(int); !ok {
		fmt.Println("assertion ke int gagal (ok=false, tidak panic)")
	}

	for _, val := range []any{42, "Go", true, Circle{R: 1}, 3.14} {
		fmt.Printf("  describe -> %s\n", describe(val))
	}
}

// ------------------------------------------------------------------
// 3. Interface bawaan: fmt.Stringer & io.Writer
// ------------------------------------------------------------------

// Money mengimplementasikan fmt.Stringer -> otomatis dipakai oleh %v/Println.
type Money int // dalam rupiah

func (m Money) String() string {
	// Format ribuan sederhana: 1000000 -> Rp1.000.000
	s := fmt.Sprintf("%d", int(m))
	var out []byte
	for i, digit := range []byte(s) {
		if i > 0 && (len(s)-i)%3 == 0 {
			out = append(out, '.')
		}
		out = append(out, digit)
	}
	return "Rp" + string(out)
}

// countingWriter mengimplementasikan io.Writer: Write([]byte)(int,error).
// Tugasnya cuma menghitung total byte yang lewat.
type countingWriter struct{ total int }

func (w *countingWriter) Write(p []byte) (int, error) {
	w.total += len(p)
	return len(p), nil
}

func stringerDanWriter() {
	fmt.Println("\n-- Stringer & Writer --")

	harga := Money(1500000)
	fmt.Printf("harga = %v (String() dipanggil otomatis)\n", harga)

	// Karena countingWriter adalah io.Writer, fmt.Fprintf bisa menulis ke sana.
	cw := &countingWriter{}
	fmt.Fprintf(cw, "Halo %s, umur %d\n", "Ana", 20)
	fmt.Fprint(cw, strings.Repeat("x", 5))
	fmt.Printf("total byte ditulis ke countingWriter = %d\n", cw.total)
}

// ------------------------------------------------------------------
// 4. Jebakan "typed nil" — sumber bug error handling
// ------------------------------------------------------------------
type MyError struct{ Msg string }

func (e *MyError) Error() string { return e.Msg }

// SALAH: mengembalikan *MyError bertipe. Walau nilainya nil, interface-nya
// berisi (tipe=*MyError, nilai=nil) -> TIDAK sama dengan nil.
func bikinErrorSalah() error {
	var e *MyError // nil
	return e       // dibungkus ke interface error -> jadi "typed nil"
}

// BENAR: kembalikan nil literal saat tidak ada error.
func bikinErrorBenar() error {
	return nil
}

func jebakanTypedNil() {
	fmt.Println("\n-- Jebakan typed nil --")

	errSalah := bikinErrorSalah()
	fmt.Printf("bikinErrorSalah() == nil ? %t  <- MENGEJUTKAN (typed nil)\n", errSalah == nil)

	errBenar := bikinErrorBenar()
	fmt.Printf("bikinErrorBenar() == nil ? %t\n", errBenar == nil)

	fmt.Println("Pelajaran: jangan return pointer bertipe konkret sebagai error;")
	fmt.Println("kembalikan nil literal, atau assign ke variabel error langsung.")
}
