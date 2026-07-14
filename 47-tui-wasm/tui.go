// Modul 47 — TUI (Bubble Tea) & WebAssembly.
//
// 🔍 Analogi besar TUI = aplikasi berbasis TEKS di terminal (seperti htop, vim, git rebase -i) —
// antarmuka interaktif tanpa mouse/grafik. Bubble Tea memakai pola ELM dengan 3 bagian:
//   - Model = FOTO keadaan aplikasi saat ini (di sini: angka counter). Satu-satunya sumber kebenaran.
//   - Update = ATURAN PERUBAHAN: "diberi keadaan lama + kejadian (tekan tombol) -> hasilkan keadaan baru".
//     Fungsi ini MURNI (tak menyentuh layar/dunia luar) -> gampang diuji: cukup cek "tekan + jadi count naik".
//   - View = JURU GAMBAR: mengubah keadaan (Model) jadi teks yang tampil. Hanya menggambar, tak mengubah apa pun.
// Alur berputar: kejadian -> Update -> Model baru -> View menggambar ulang. Rapi & mudah dinalar.

// Bubble Tea memakai arsitektur ELM: Model (state) + Update (transisi) + View
// (render). Update MURNI (state lama + pesan -> state baru) sehingga mudah
// di-test tanpa terminal.
package main

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// model = seluruh state aplikasi TUI (di sini: sebuah counter).
type model struct {
	count    int
	quitting bool
}

func initialModel() model { return model{} }

// Init: perintah awal (tak ada di sini).
func (m model) Init() tea.Cmd { return nil }

// Update: menerima PESAN (keyboard, dll), mengembalikan state baru.
// Fungsi ini MURNI & deterministik -> gampang di-unit-test.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "+", "k":
			m.count++
		case "down", "-", "j":
			if m.count > 0 {
				m.count--
			}
		case "r":
			m.count = 0
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		}
	}
	return m, nil
}

// View: render state menjadi string yang tampil di terminal.
func (m model) View() string {
	if m.quitting {
		return "Sampai jumpa!\n"
	}
	return fmt.Sprintf(
		"Counter: %d\n\n[+/k] naik  [-/j] turun  [r] reset  [q] keluar\n",
		m.count,
	)
}
