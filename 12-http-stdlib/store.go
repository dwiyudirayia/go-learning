package main

import (
	"errors"
	"sync"
)

// Book adalah resource REST kita.
type Book struct {
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Author string `json:"author"`
}

var ErrNotFound = errors.New("buku tidak ditemukan")

// BookStore: penyimpanan in-memory yang aman-konkuren (RWMutex).
// Di produksi ini akan diganti implementasi database (lihat Modul 14).
type BookStore struct {
	mu     sync.RWMutex
	books  map[int]Book
	nextID int
}

func NewBookStore() *BookStore {
	return &BookStore{books: make(map[int]Book), nextID: 1}
}

func (s *BookStore) List() []Book {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Book, 0, len(s.books))
	for _, b := range s.books {
		out = append(out, b)
	}
	return out
}

func (s *BookStore) Get(id int) (Book, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	b, ok := s.books[id]
	if !ok {
		return Book{}, ErrNotFound
	}
	return b, nil
}

func (s *BookStore) Create(b Book) Book {
	s.mu.Lock()
	defer s.mu.Unlock()
	b.ID = s.nextID
	s.nextID++
	s.books[b.ID] = b
	return b
}

func (s *BookStore) Update(id int, b Book) (Book, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.books[id]; !ok {
		return Book{}, ErrNotFound
	}
	b.ID = id
	s.books[id] = b
	return b, nil
}

func (s *BookStore) Delete(id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.books[id]; !ok {
		return ErrNotFound
	}
	delete(s.books, id)
	return nil
}
