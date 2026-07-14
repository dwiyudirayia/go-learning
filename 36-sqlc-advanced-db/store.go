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

// 🔍 Analogi besar sqlc: kamu tulis SQL biasa di file .sql, lalu sqlc meng-GENERATE fungsi Go
// type-safe dari situ. Seperti punya PENERJEMAH yang mengubah query-mu jadi kode Go otomatis —
// beda dari ORM (GORM) yang menyembunyikan SQL, sqlc justru merangkul SQL tapi memberi keamanan
// tipe: salah nama kolom ketahuan saat COMPILE, bukan saat program jalan di produksi. SQL jujur + aman.

// 🔍 Analogi connection pool: buka-tutup koneksi DB itu mahal (seperti antre & basa-basi tiap kali).
// Pool = SEKUMPULAN KONEKSI SIAP PAKAI yang dipinjam-kembalikan, bukan dibuat baru tiap query.
// SetMaxOpenConns = batas maksimal "loket" agar DB tak kewalahan; SetMaxIdleConns = berapa loket
// dibiarkan siaga; ConnMaxLifetime = daur ulang koneksi tua agar tak basi. Wajib disetel di produksi.

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

// 🔍 Analogi execTx: fungsi ini "PEMBUNGKUS TRANSAKSI" — kamu serahkan pekerjaanmu (fn), ia yang
// urus buka transaksi, dan otomatis Commit kalau sukses / Rollback kalau gagal. Seperti kasir yang
// menjamin "kalau salah satu barang gagal di-scan, seluruh keranjang dibatalkan". WithTx(tx) membuat
// query berjalan DI DALAM keranjang transaksi yang sama, bukan langsung ke DB. Menghindari kode
// buka/commit/rollback yang berulang & rawan lupa di tiap operasi multi-langkah.
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
