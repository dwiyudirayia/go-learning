// Jalankan: go run ./37-advanced-testing
// Test biasa:  go test ./37-advanced-testing
// Fuzzing:     go test -fuzz FuzzReverse -fuzztime 15s ./37-advanced-testing
package main

import "fmt"

func main() {
	fmt.Println("=== 37 — Advanced Testing ===")
	lo, hi, _ := ParseRange("3-7")
	fmt.Printf("ParseRange(\"3-7\") = %d..%d\n", lo, hi)
	fmt.Printf("Reverse(\"café 世界\") = %q\n", Reverse("café 世界"))
	fmt.Println("\nJalankan fuzzing: go test -fuzz FuzzReverse -fuzztime 15s ./37-advanced-testing")
}
