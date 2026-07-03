// Command tasks — CLI manajemen tugas (contoh Cobra) untuk modul 11.
//
// Contoh:
//
//	go run ./11-cli-cobra add "belajar cobra"
//	go run ./11-cli-cobra list
//	go run ./11-cli-cobra done 1
//	go run ./11-cli-cobra list --all
//	go run ./11-cli-cobra --help
package main

import (
	"fmt"
	"os"

	"go-learning/11-cli-cobra/cmd"
)

func main() {
	// main tipis: hanya menjalankan Cobra dan menetapkan exit code bila error.
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
