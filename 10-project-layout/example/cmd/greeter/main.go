// Command greeter adalah entry point (binary). Isinya TIPIS: hanya baca flag
// lalu delegasikan ke package internal. Pola cmd/ -> internal/.
// Jalankan: go run ./10-project-layout/example/cmd/greeter -name Ana -shout
package main

import (
	"flag"
	"fmt"

	"go-learning/10-project-layout/example/internal/greeting"
)

func main() {
	// flag: parser argumen CLI bawaan stdlib (Modul 11 pakai Cobra untuk yg lebih kompleks).
	name := flag.String("name", "", "nama yang disapa")
	shout := flag.Bool("shout", false, "sapa dengan huruf kapital")
	flag.Parse()

	fmt.Println(greeting.Greet(*name, *shout))
}
