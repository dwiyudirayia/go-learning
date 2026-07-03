package main

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// Karena Update MURNI, kita bisa menguji state machine TUI tanpa terminal:
// beri pesan keyboard, periksa state hasilnya.

func key(s string) tea.KeyMsg {
	switch s {
	case "up":
		return tea.KeyMsg{Type: tea.KeyUp}
	case "down":
		return tea.KeyMsg{Type: tea.KeyDown}
	default: // rune tunggal seperti "+", "-", "r", "q"
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
	}
}

// apply menerapkan urutan tombol ke model dan mengembalikan model akhir.
func apply(m model, keys ...string) model {
	var tm tea.Model = m
	for _, k := range keys {
		tm, _ = tm.Update(key(k))
	}
	return tm.(model)
}

func TestCounterNaikTurun(t *testing.T) {
	m := apply(initialModel(), "+", "+", "+", "-") // 0 ->3 ->2
	if m.count != 2 {
		t.Errorf("count = %d; want 2", m.count)
	}
}

func TestTidakBisaNegatif(t *testing.T) {
	m := apply(initialModel(), "-", "-") // sudah 0, turun ditahan
	if m.count != 0 {
		t.Errorf("count = %d; want 0 (tak boleh negatif)", m.count)
	}
}

func TestReset(t *testing.T) {
	m := apply(initialModel(), "+", "+", "+", "r")
	if m.count != 0 {
		t.Errorf("count setelah reset = %d; want 0", m.count)
	}
}

func TestQuitMengubahView(t *testing.T) {
	m := apply(initialModel(), "q")
	if !m.quitting {
		t.Error("q harus set quitting=true")
	}
	if !strings.Contains(m.View(), "Sampai jumpa") {
		t.Errorf("view setelah quit = %q", m.View())
	}
}

func TestViewMenampilkanCount(t *testing.T) {
	m := apply(initialModel(), "+", "+")
	if !strings.Contains(m.View(), "Counter: 2") {
		t.Errorf("view = %q; harus memuat 'Counter: 2'", m.View())
	}
}
