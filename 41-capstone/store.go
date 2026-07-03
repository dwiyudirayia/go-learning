// Modul 41 — CAPSTONE: URL Shortener.
// Menggabungkan: Fiber (13), database/sql (14), JWT+bcrypt (15,27), Redis
// cache-aside (22), config env (19), graceful shutdown (20), observability (18),
// arsitektur berlapis (10,29). store.go = lapisan data.
package main

import (
	"database/sql"
	"errors"
)

type User struct {
	ID           int64
	Email        string
	PasswordHash string
}

type Link struct {
	ID     int64
	Code   string
	URL    string
	UserID int64
	Clicks int64
}

var ErrNotFound = errors.New("tidak ditemukan")

// setupSchema membuat tabel (di produksi pakai migrasi berversi, Modul 21).
func setupSchema(db *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			email TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL)`,
		`CREATE TABLE IF NOT EXISTS links (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			code TEXT UNIQUE NOT NULL,
			url TEXT NOT NULL,
			user_id INTEGER NOT NULL,
			clicks INTEGER NOT NULL DEFAULT 0)`,
	}
	for _, s := range stmts {
		if _, err := db.Exec(s); err != nil {
			return err
		}
	}
	return nil
}

type Store struct{ db *sql.DB }

func NewStore(db *sql.DB) *Store { return &Store{db: db} }

func (s *Store) CreateUser(email, hash string) (User, error) {
	res, err := s.db.Exec("INSERT INTO users(email,password_hash) VALUES(?,?)", email, hash)
	if err != nil {
		return User{}, err
	}
	id, _ := res.LastInsertId()
	return User{ID: id, Email: email, PasswordHash: hash}, nil
}

func (s *Store) GetUserByEmail(email string) (User, error) {
	var u User
	err := s.db.QueryRow("SELECT id,email,password_hash FROM users WHERE email=?", email).
		Scan(&u.ID, &u.Email, &u.PasswordHash)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrNotFound
	}
	return u, err
}

func (s *Store) CreateLink(code, url string, userID int64) (Link, error) {
	res, err := s.db.Exec("INSERT INTO links(code,url,user_id) VALUES(?,?,?)", code, url, userID)
	if err != nil {
		return Link{}, err
	}
	id, _ := res.LastInsertId()
	return Link{ID: id, Code: code, URL: url, UserID: userID}, nil
}

func (s *Store) GetLinkByCode(code string) (Link, error) {
	var l Link
	err := s.db.QueryRow("SELECT id,code,url,user_id,clicks FROM links WHERE code=?", code).
		Scan(&l.ID, &l.Code, &l.URL, &l.UserID, &l.Clicks)
	if errors.Is(err, sql.ErrNoRows) {
		return Link{}, ErrNotFound
	}
	return l, err
}

func (s *Store) IncrementClicks(code string) error {
	_, err := s.db.Exec("UPDATE links SET clicks=clicks+1 WHERE code=?", code)
	return err
}
