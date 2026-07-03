// Package domain = INTI (core) aplikasi: entity + "port" (interface).
// TIDAK bergantung pada apa pun (tidak import HTTP, database, framework).
// Semua lapisan luar bergantung KE SINI, bukan sebaliknya (dependency inversion).
package domain

import "errors"

// Note = entity domain (aturan bisnis murni).
type Note struct {
	ID    int
	Title string
	Body  string
}

var (
	ErrNotFound   = errors.New("note tidak ditemukan")
	ErrEmptyTitle = errors.New("judul wajib diisi")
)

// NoteRepository = PORT (driven): kontrak penyimpanan yang dibutuhkan core.
// Implementasinya (memory/postgres/...) ada di lapisan ADAPTER, bukan di sini.
type NoteRepository interface {
	Save(n *Note) error
	FindByID(id int) (*Note, error)
	All() ([]Note, error)
}
