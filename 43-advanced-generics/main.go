// Jalankan: go run ./43-advanced-generics
// Verifikasi otomatis: go test ./43-advanced-generics
package main

import (
	"fmt"
	"time"
)

func main() {
	fmt.Println("=== 43 — Advanced Generics ===")

	// Set generik.
	fmt.Println("\n-- Set[T] --")
	a := NewSet(1, 2, 3, 4)
	b := NewSet(3, 4, 5, 6)
	fmt.Printf("A ∪ B len=%d, A ∩ B len=%d\n", a.Union(b).Len(), a.Intersect(b).Len())
	fmt.Printf("A punya 3? %t, punya 9? %t\n", a.Has(3), a.Has(9))

	// Iterator (lazy, range-over-func) + rangkaian Filter/Map.
	fmt.Println("\n-- Iterator (iter.Seq) --")
	genap := Filter(Count(10), func(n int) bool { return n%2 == 0 })
	kuadrat := Map(genap, func(n int) int { return n * n })
	fmt.Printf("kuadrat dari genap 0..9 = %v\n", Collect(kuadrat))

	// Lazy: berhenti lebih awal dengan break -> sisanya tak dihitung.
	fmt.Print("ambil 3 pertama: ")
	count := 0
	for n := range Count(1000000) {
		fmt.Printf("%d ", n)
		count++
		if count == 3 {
			break // iterator berhenti; 999.997 sisanya tak pernah dieksekusi
		}
	}
	fmt.Println()

	// Functional options.
	fmt.Println("\n-- Functional Options --")
	s1 := NewServer() // semua default
	s2 := NewServer(WithPort(9090), WithTLS(), WithTimeout(5*time.Second))
	fmt.Printf("default : %+v\n", *s1)
	fmt.Printf("custom  : %+v\n", *s2)
}
