// Package service = USE CASE (application layer): logika bisnis.
// Bergantung pada PORT (domain.NoteRepository), bukan implementasi konkret ->
// bisa diuji dengan repo palsu & ditukar databasenya tanpa mengubah logika.
package service

import (
	"strings"

	"go-learning/29-clean-architecture/internal/domain"
)

type NoteService struct {
	repo domain.NoteRepository // PORT, bukan tipe konkret
}

func New(repo domain.NoteRepository) *NoteService {
	return &NoteService{repo: repo}
}

func (s *NoteService) Create(title, body string) (*domain.Note, error) {
	if strings.TrimSpace(title) == "" {
		return nil, domain.ErrEmptyTitle // aturan bisnis
	}
	n := &domain.Note{Title: title, Body: body}
	if err := s.repo.Save(n); err != nil {
		return nil, err
	}
	return n, nil
}

func (s *NoteService) Get(id int) (*domain.Note, error) {
	return s.repo.FindByID(id)
}

func (s *NoteService) List() ([]domain.Note, error) {
	return s.repo.All()
}
