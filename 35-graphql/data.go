// Modul 35 — GraphQL: satu endpoint, client minta persis field yang dibutuhkan.
package main

import "sync"

type Author struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Book struct {
	ID       int    `json:"id"`
	Title    string `json:"title"`
	AuthorID int    `json:"authorId"`
}

// store in-memory (thread-safe) sebagai sumber data.
type store struct {
	mu      sync.Mutex
	authors map[int]Author
	books   map[int]Book
	seq     int
}

func newStore() *store {
	s := &store{
		authors: map[int]Author{
			1: {ID: 1, Name: "Rob Pike"},
			2: {ID: 2, Name: "Dennis Ritchie"},
		},
		books: map[int]Book{
			1: {ID: 1, Title: "The Go Programming Language", AuthorID: 1},
			2: {ID: 2, Title: "The C Programming Language", AuthorID: 2},
			3: {ID: 3, Title: "The Unix Programming Environment", AuthorID: 1},
		},
		seq: 3,
	}
	return s
}

func (s *store) allAuthors() []Author {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Author, 0, len(s.authors))
	for _, a := range s.authors {
		out = append(out, a)
	}
	return out
}

func (s *store) booksByAuthor(authorID int) []Book {
	s.mu.Lock()
	defer s.mu.Unlock()
	var out []Book
	for _, b := range s.books {
		if b.AuthorID == authorID {
			out = append(out, b)
		}
	}
	return out
}

func (s *store) addBook(title string, authorID int) Book {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.seq++
	b := Book{ID: s.seq, Title: title, AuthorID: authorID}
	s.books[b.ID] = b
	return b
}
