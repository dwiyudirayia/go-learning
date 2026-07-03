// Latihan 4: test untuk internal/store (tanpa menyentuh Cobra).
package store

import (
	"errors"
	"path/filepath"
	"testing"
)

// tempStore membuat store di file sementara (t.TempDir otomatis dibersihkan).
func tempStore(t *testing.T) *Store {
	t.Helper()
	s, err := Load(filepath.Join(t.TempDir(), "tasks.json"))
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	return s
}

func TestAddAndPersist(t *testing.T) {
	s := tempStore(t)
	id, err := s.Add("belajar")
	if err != nil {
		t.Fatalf("add: %v", err)
	}
	if id != 1 {
		t.Errorf("id pertama = %d; want 1", id)
	}

	// Muat ulang dari file yang sama -> data harus persisten.
	s2, _ := Load(s.path)
	if len(s2.Tasks) != 1 || s2.Tasks[0].Text != "belajar" {
		t.Errorf("data tidak persisten: %+v", s2.Tasks)
	}
}

func TestMarkDoneDanNotFound(t *testing.T) {
	s := tempStore(t)
	_, _ = s.Add("a")

	if err := s.MarkDone(1); err != nil {
		t.Fatalf("MarkDone: %v", err)
	}
	if !s.Tasks[0].Done {
		t.Error("task belum ditandai selesai")
	}

	if err := s.MarkDone(999); !errors.Is(err, ErrNotFound) {
		t.Errorf("MarkDone(999) = %v; want ErrNotFound", err)
	}
}

func TestRemove(t *testing.T) {
	s := tempStore(t)
	_, _ = s.Add("a")
	_, _ = s.Add("b")

	if err := s.Remove(1); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	if len(s.Tasks) != 1 || s.Tasks[0].Text != "b" {
		t.Errorf("setelah Remove(1): %+v", s.Tasks)
	}
	if err := s.Remove(1); !errors.Is(err, ErrNotFound) {
		t.Errorf("Remove ulang = %v; want ErrNotFound", err)
	}
}

func TestStats(t *testing.T) {
	s := tempStore(t)
	_, _ = s.Add("a")
	_, _ = s.Add("b")
	_, _ = s.AddWithPriority("c penting", 5)
	_ = s.MarkDone(1)

	total, done, pending := s.Stats()
	if total != 3 || done != 1 || pending != 2 {
		t.Errorf("Stats = (%d,%d,%d); want (3,1,2)", total, done, pending)
	}
}
