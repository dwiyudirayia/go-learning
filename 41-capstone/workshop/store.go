package main

import (
	"database/sql"
	"errors"
)

// Link = data inti aplikasi kita: satu tautan pendek.
//
//	Code  = kode pendek unik (mis. "aB3x9")
//	URL   = tujuan asli (mis. "https://go.dev")
//	Clicks= berapa kali diakses
type Link struct {
	ID     int64
	Code   string
	URL    string
	Clicks int64
}

// ErrNotFound: sentinel error khas Go (Modul 5) — dipakai saat data tak ada.
var ErrNotFound = errors.New("link tidak ditemukan")

// Store membungkus koneksi database. Semua akses ke tabel lewat sini
// (lapisan data — Modul 10, 14).
type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

// setupSchema membuat tabel links bila belum ada.
// (Di produksi pakai migrasi berversi — Modul 21. Untuk belajar, ini cukup.)
func (s *Store) setupSchema() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS links (
			id     INTEGER PRIMARY KEY AUTOINCREMENT,
			code   TEXT UNIQUE NOT NULL,
			url    TEXT NOT NULL,
			clicks INTEGER NOT NULL DEFAULT 0
		)`)
	return err
}

// CreateLink menyimpan link baru & mengembalikannya (lengkap dengan ID).
func (s *Store) CreateLink(code, url string) (Link, error) {
	// Placeholder '?' mencegah SQL injection (Modul 14) — JANGAN string concat.
	res, err := s.db.Exec("INSERT INTO links(code, url) VALUES(?, ?)", code, url)
	if err != nil {
		return Link{}, err
	}
	id, _ := res.LastInsertId()
	return Link{ID: id, Code: code, URL: url}, nil
}

// GetLinkByCode mengambil link berdasarkan kode pendeknya.
func (s *Store) GetLinkByCode(code string) (Link, error) {
	var l Link
	err := s.db.
		QueryRow("SELECT id, code, url, clicks FROM links WHERE code = ?", code).
		Scan(&l.ID, &l.Code, &l.URL, &l.Clicks)

	// Bedakan "tidak ada" dari error sungguhan (Modul 5).
	if errors.Is(err, sql.ErrNoRows) {
		return Link{}, ErrNotFound
	}
	return l, err
}

// IncrementClicks menambah 1 penghitung klik (dipanggil saat redirect nanti).
func (s *Store) IncrementClicks(code string) error {
	_, err := s.db.Exec("UPDATE links SET clicks = clicks + 1 WHERE code = ?", code)
	return err
}
