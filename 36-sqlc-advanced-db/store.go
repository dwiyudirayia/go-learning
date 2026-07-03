// Modul 36 — sqlc (SQL type-safe) + connection pool + transaksi.
package main

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"strings"
	"time"

	"go-learning/36-sqlc-advanced-db/internal/sqlcdb"
)

//go:embed db/schema.sql
var schemaSQL string

// OpenDB membuka SQLite + MENYETEL CONNECTION POOL.
func OpenDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	// Tuning pool (penting di produksi, terutama PostgreSQL/MySQL):
	db.SetMaxOpenConns(25)                 // maks koneksi terbuka bersamaan
	db.SetMaxIdleConns(5)                  // koneksi idle yang disimpan (hemat handshake)
	db.SetConnMaxLifetime(5 * time.Minute) // daur ulang koneksi lama
	db.SetConnMaxIdleTime(time.Minute)     // tutup koneksi idle terlalu lama

	if err := db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

// applySchema menjalankan DDL dari schema.sql (dipisah per statement).
func applySchema(db *sql.DB) error {
	for _, stmt := range strings.Split(schemaSQL, ";") {
		if strings.TrimSpace(stmt) == "" {
			continue
		}
		if _, err := db.Exec(stmt); err != nil {
			return err
		}
	}
	return nil
}

// Store membungkus Queries hasil sqlc + kemampuan transaksi.
type Store struct {
	*sqlcdb.Queries
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{Queries: sqlcdb.New(db), db: db}
}

// execTx menjalankan fn dalam SATU transaksi. Commit bila sukses, Rollback bila
// error atau panic — sehingga perubahan bersifat atomik.
func (s *Store) execTx(ctx context.Context, fn func(*sqlcdb.Queries) error) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	q := s.Queries.WithTx(tx) // Queries yang beroperasi di dalam tx
	if err := fn(q); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx err: %v, rollback err: %v", err, rbErr)
		}
		return err
	}
	return tx.Commit()
}

// CreateAuthorWithBook membuat author + buku pertamanya secara ATOMIK.
func (s *Store) CreateAuthorWithBook(ctx context.Context, authorName, bookTitle string) (sqlcdb.Author, error) {
	var author sqlcdb.Author
	err := s.execTx(ctx, func(q *sqlcdb.Queries) error {
		a, err := q.CreateAuthor(ctx, authorName)
		if err != nil {
			return err
		}
		author = a
		_, err = q.CreateBook(ctx, sqlcdb.CreateBookParams{Title: bookTitle, AuthorID: a.ID})
		return err
	})
	return author, err
}
