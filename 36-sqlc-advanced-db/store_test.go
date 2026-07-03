package main

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"go-learning/36-sqlc-advanced-db/internal/sqlcdb"

	_ "modernc.org/sqlite"
)

func newStore(t *testing.T) *Store {
	t.Helper()
	db, err := OpenDB(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	if err := applySchema(db); err != nil {
		t.Fatalf("schema: %v", err)
	}
	return NewStore(db)
}

func TestTypeSafeQueries(t *testing.T) {
	s := newStore(t)
	ctx := context.Background()

	a, err := s.CreateAuthor(ctx, "Ana")
	if err != nil {
		t.Fatalf("create author: %v", err)
	}
	// Hasil sudah bertipe sqlcdb.Author (bukan interface{}), tanpa Scan manual.
	if a.ID == 0 || a.Name != "Ana" {
		t.Errorf("author = %+v", a)
	}

	got, err := s.GetAuthor(ctx, a.ID)
	if err != nil || got.Name != "Ana" {
		t.Errorf("get -> %+v err=%v", got, err)
	}
}

func TestTransaksiSukses(t *testing.T) {
	s := newStore(t)
	ctx := context.Background()

	if _, err := s.CreateAuthorWithBook(ctx, "Budi", "Buku Budi"); err != nil {
		t.Fatalf("tx: %v", err)
	}
	// Author & buku harus ada.
	count, _ := s.CountAuthors(ctx)
	if count != 1 {
		t.Errorf("author = %d; want 1", count)
	}
	authors, _ := s.ListAuthors(ctx)
	books, _ := s.ListBooksByAuthor(ctx, authors[0].ID)
	if len(books) != 1 {
		t.Errorf("buku = %d; want 1", len(books))
	}
}

func TestTransaksiRollback(t *testing.T) {
	s := newStore(t)
	ctx := context.Background()

	// Transaksi yang gagal di tengah -> semua di-rollback.
	err := s.execTx(ctx, func(q *sqlcdb.Queries) error {
		if _, err := q.CreateAuthor(ctx, "AkanBatal"); err != nil {
			return err
		}
		return errors.New("gagal sengaja setelah insert")
	})
	if err == nil {
		t.Fatal("mengharapkan error")
	}

	// Author TIDAK boleh tersimpan (rollback).
	count, _ := s.CountAuthors(ctx)
	if count != 0 {
		t.Errorf("author = %d; want 0 (harus di-rollback)", count)
	}
}
