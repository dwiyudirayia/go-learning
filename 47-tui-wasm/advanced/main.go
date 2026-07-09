// Teknik Advanced Modul 47 (TUI & WASM) — contoh runnable + penjelasan detail.
//
// Arsitektur Elm Bubble Tea (Model/Update/View) diUJI TANPA TTY dengan
// mengirim pesan sintetis langsung ke Update. Jalankan:
//
//	go run ./47-tui-wasm/advanced
package main

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// ===========================================================================
// Arsitektur ELM: state (Model) hanya berubah lewat Update(msg) yang murni,
// dan View() merender state jadi teks. Efek samping diwakili tea.Cmd (async).
// Karena Update murni (input msg -> output model), ia mudah diuji: cukup panggil
// dengan pesan sintetis, tanpa terminal sungguhan.
// ===========================================================================

type model struct {
	nilai int
	log   []string
}

func (m model) Init() tea.Cmd { return nil }

// Update: transisi state berdasarkan pesan. Di sini menangani penekanan tombol.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if k, ok := msg.(tea.KeyMsg); ok {
		switch k.String() {
		case "+":
			m.nilai++
		case "-":
			m.nilai--
		case "r":
			m.nilai = 0
		case "q":
			return m, tea.Quit // Cmd untuk keluar program
		}
		m.log = append(m.log, k.String())
	}
	return m, nil
}

func (m model) View() string {
	return fmt.Sprintf("Nilai: %d  (riwayat tombol: %v)", m.nilai, m.log)
}

// keyMsg membuat pesan penekanan tombol sintetis (persis yang dikirim runtime
// Bubble Tea saat user menekan tombol) — kunci pengujian tanpa TTY.
func keyMsg(s string) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
}

func main() {
	// =====================================================================
	// DRIVE model secara terprogram — inilah pola UNIT TEST untuk TUI.
	// Tidak ada TTY; kita hanya memanggil Update lalu memeriksa View.
	// =====================================================================
	var m tea.Model = model{}
	fmt.Println("== drive Update tanpa TTY ==")
	fmt.Println("  awal   :", m.View())

	for _, tombol := range []string{"+", "+", "+", "-", "r", "+"} {
		m, _ = m.Update(keyMsg(tombol))
		fmt.Printf("  tekan %q -> %s\n", tombol, m.View())
	}

	// Verifikasi (assert) seperti di test sungguhan:
	if got := m.(model).nilai; got != 1 {
		panic(fmt.Sprintf("harusnya 1, dapat %d", got))
	}
	fmt.Println("  ✔ state akhir sesuai ekspektasi")

	// =====================================================================
	// Catatan WASM: modul ini juga menyediakan target WebAssembly. File WASM
	// diberi header build tag agar HANYA dikompilasi untuk js/wasm:
	//   //go:build js && wasm
	// dan ada stub untuk platform lain, sehingga `go build ./...` di host tetap
	// lolos. Interop DOM lewat syscall/js; untuk biner kecil pertimbangkan TinyGo.
}
