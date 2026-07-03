// Jalankan (butuh terminal sungguhan): go run ./47-tui-wasm
// Verifikasi otomatis (menguji logika Update tanpa terminal): go test ./47-tui-wasm
package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Program Bubble Tea menguasai terminal (raw mode). Di lingkungan tanpa TTY
	// ia akan mengembalikan error — itu wajar; jalankan di terminal sungguhan.
	if _, err := tea.NewProgram(initialModel()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, "TUI butuh terminal interaktif:", err)
		os.Exit(1)
	}
}
