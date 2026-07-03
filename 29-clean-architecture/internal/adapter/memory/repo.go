// Package memory = ADAPTER (driven): implementasi NoteRepository di memori.
// Bisa diganti postgres/mongo tanpa menyentuh domain/service.
package memory

import (
	"sync"

	"go-learning/29-clean-architecture/internal/domain"
)

type Repo struct {
	mu   sync.Mutex
	data map[int]domain.Note
	seq  int
}

func New() *Repo { return &Repo{data: make(map[int]domain.Note)} }

// Pastikan (saat kompilasi) Repo memenuhi port domain.NoteRepository.
var _ domain.NoteRepository = (*Repo)(nil)

func (r *Repo) Save(n *domain.Note) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if n.ID == 0 {
		r.seq++
		n.ID = r.seq
	}
	r.data[n.ID] = *n
	return nil
}

func (r *Repo) FindByID(id int) (*domain.Note, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	n, ok := r.data[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return &n, nil
}

func (r *Repo) All() ([]domain.Note, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]domain.Note, 0, len(r.data))
	for _, n := range r.data {
		out = append(out, n)
	}
	return out, nil
}
