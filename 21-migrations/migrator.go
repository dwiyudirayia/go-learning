// Modul 21 — Migrations dengan library golang-migrate.
// Memakai driver database SQLite pure-Go (tanpa cgo) + source iofs (embed).
package main

import (
	"database/sql"
	"embed"
	"errors"

	"github.com/golang-migrate/migrate/v4"
	sqlitemigrate "github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

// File .sql ditanam ke binary. golang-migrate membaca versi dari nama file
// dengan format: {versi}_{judul}.{up|down}.sql  (mis. 001_create_users.up.sql).
//
//go:embed migrations/*.sql
var migrationsFS embed.FS

// newMigrator merangkai *migrate.Migrate dari koneksi *sql.DB yang sudah ada.
// Catatan: JANGAN memanggil m.Close() bila kamu masih ingin memakai db —
// Close() ikut menutup koneksi database di baliknya.
func newMigrator(db *sql.DB) (*migrate.Migrate, error) {
	// Source: file .sql dari embed.FS.
	src, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return nil, err
	}
	// Database driver: bungkus *sql.DB (SQLite pure-Go).
	driver, err := sqlitemigrate.WithInstance(db, &sqlitemigrate.Config{})
	if err != nil {
		return nil, err
	}
	return migrate.NewWithInstance("iofs", src, "sqlite", driver)
}

// Up menerapkan SEMUA migrasi yang belum dijalankan.
func Up(db *sql.DB) error {
	m, err := newMigrator(db)
	if err != nil {
		return err
	}
	// ErrNoChange = tidak ada migrasi baru -> itu bukan error (idempotent).
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}

// Down membatalkan SATU migrasi terakhir (rollback satu langkah).
func Down(db *sql.DB) error {
	m, err := newMigrator(db)
	if err != nil {
		return err
	}
	if err := m.Steps(-1); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}

// Version mengembalikan versi migrasi saat ini (0 bila belum ada) dan status
// 'dirty' (true bila sebuah migrasi gagal di tengah -> perlu diperbaiki manual).
func Version(db *sql.DB) (version uint, dirty bool, err error) {
	m, err := newMigrator(db)
	if err != nil {
		return 0, false, err
	}
	v, dirty, err := m.Version()
	if errors.Is(err, migrate.ErrNilVersion) {
		return 0, false, nil // belum ada migrasi yang diterapkan
	}
	return v, dirty, err
}
