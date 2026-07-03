package main

import (
	"database/sql"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

func newDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestUp(t *testing.T) {
	db := newDB(t)

	if err := Up(db); err != nil {
		t.Fatalf("Up: %v", err)
	}
	v, dirty, err := Version(db)
	if err != nil {
		t.Fatalf("Version: %v", err)
	}
	if v != 3 || dirty {
		t.Errorf("versi = %d dirty=%t; want 3, false", v, dirty)
	}
	// Kolom 'active' (migrasi 003) harus ada.
	if _, err := db.Exec("INSERT INTO users(name,email,active) VALUES('a','a@x',1)"); err != nil {
		t.Errorf("insert dengan kolom active gagal: %v", err)
	}
}

func TestIdempotent(t *testing.T) {
	db := newDB(t)
	if err := Up(db); err != nil {
		t.Fatal(err)
	}
	// Up kedua tidak boleh error (ErrNoChange ditangani).
	if err := Up(db); err != nil {
		t.Errorf("Up kedua = %v; want nil (idempotent)", err)
	}
}

func TestRollback(t *testing.T) {
	db := newDB(t)
	if err := Up(db); err != nil {
		t.Fatal(err)
	}
	if err := Down(db); err != nil {
		t.Fatalf("Down: %v", err)
	}
	if v, _, _ := Version(db); v != 2 {
		t.Errorf("versi setelah rollback = %d; want 2", v)
	}
	// Setelah rollback 003, kolom 'active' harus hilang.
	if _, err := db.Exec("INSERT INTO users(name,email,active) VALUES('a','a@x',1)"); err == nil {
		t.Error("kolom active harusnya sudah hilang setelah rollback")
	}
}
