package service

import (
	"errors"
	"testing"

	"go-learning/29-clean-architecture/internal/domain"
)

// fakeRepo = implementasi port PALSU untuk menguji use case TANPA DB.
// Membuktikan: core bisa diuji terisolasi berkat dependency inversion.
type fakeRepo struct {
	notes map[int]domain.Note
	seq   int
}

func newFake() *fakeRepo { return &fakeRepo{notes: map[int]domain.Note{}} }

func (f *fakeRepo) Save(n *domain.Note) error {
	if n.ID == 0 {
		f.seq++
		n.ID = f.seq
	}
	f.notes[n.ID] = *n
	return nil
}
func (f *fakeRepo) FindByID(id int) (*domain.Note, error) {
	n, ok := f.notes[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return &n, nil
}
func (f *fakeRepo) All() ([]domain.Note, error) {
	out := make([]domain.Note, 0, len(f.notes))
	for _, n := range f.notes {
		out = append(out, n)
	}
	return out, nil
}

func TestCreateDanGet(t *testing.T) {
	svc := New(newFake())

	n, err := svc.Create("Judul", "Isi")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if n.ID == 0 {
		t.Error("ID belum di-set")
	}

	got, err := svc.Get(n.ID)
	if err != nil || got.Title != "Judul" {
		t.Errorf("get -> %+v err=%v", got, err)
	}
}

func TestValidasiJudulKosong(t *testing.T) {
	svc := New(newFake())
	if _, err := svc.Create("   ", "isi"); !errors.Is(err, domain.ErrEmptyTitle) {
		t.Errorf("err = %v; want ErrEmptyTitle", err)
	}
}

func TestGetTidakAda(t *testing.T) {
	svc := New(newFake())
	if _, err := svc.Get(999); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("err = %v; want ErrNotFound", err)
	}
}
