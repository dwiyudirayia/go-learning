// Teknik Advanced Modul 36 (sqlc / db lanjutan) — contoh runnable + penjelasan detail.
//
// Meniru POLA kode yang DIHASILKAN sqlc (Queries type-safe + WithTx) memakai
// database/sql + SQLite agar self-contained. Jalankan:
//
//	go run ./36-sqlc-advanced-db/advanced
package main

import (
	"context"
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

// ===========================================================================
// DBTX = abstraksi minimal yang dipakai Queries. Baik *sql.DB maupun *sql.Tx
// memenuhinya -> query yang sama bisa jalan di luar ATAU di dalam transaksi.
// (Persis pola yang di-generate sqlc.)
// ===========================================================================
type DBTX interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

type Author struct {
	ID   int64
	Nama string
}

// Queries membungkus DBTX. Method-methodnya TYPE-SAFE: parameter & hasil sudah
// bertipe konkret, salah kolom/tipe ketahuan saat COMPILE (beda dari string SQL
// mentah yang baru meledak saat runtime).
type Queries struct{ db DBTX }

func New(db DBTX) *Queries { return &Queries{db: db} }

// WithTx mengembalikan Queries baru yang terikat ke transaksi -> jalankan
// beberapa query secara ATOMIK.
func (q *Queries) WithTx(tx *sql.Tx) *Queries { return &Queries{db: tx} }

func (q *Queries) CreateAuthor(ctx context.Context, nama string) (int64, error) {
	res, err := q.db.ExecContext(ctx, `INSERT INTO authors(nama) VALUES(?)`, nama)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (q *Queries) GetAuthor(ctx context.Context, id int64) (Author, error) {
	var a Author
	err := q.db.QueryRowContext(ctx, `SELECT id, nama FROM authors WHERE id=?`, id).Scan(&a.ID, &a.Nama)
	return a, err
}

func main() {
	ctx := context.Background()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// Catatan pool (untuk Postgres via pgxpool): atur MaxConns, health check,
	// dan pakai CopyFrom untuk bulk insert (protokol COPY = tercepat).
	db.SetMaxOpenConns(10)

	if _, err := db.ExecContext(ctx, `CREATE TABLE authors(id INTEGER PRIMARY KEY, nama TEXT NOT NULL)`); err != nil {
		panic(err)
	}

	q := New(db)

	// --- Pemakaian biasa (di luar transaksi) ---
	id, _ := q.CreateAuthor(ctx, "Rob Pike")
	a, _ := q.GetAuthor(ctx, id)
	fmt.Println("== query type-safe ==")
	fmt.Printf("  dibuat & dibaca: %+v\n", a)

	// =====================================================================
	// TRANSAKSI via WithTx — beberapa insert ATOMIK (semua atau tak sama sekali)
	// =====================================================================
	fmt.Println("== transaksi WithTx (atomik) ==")
	tx, _ := db.BeginTx(ctx, nil)
	qtx := q.WithTx(tx) // Queries yang sama, tapi lewat transaksi
	_, err1 := qtx.CreateAuthor(ctx, "Ken Thompson")
	_, err2 := qtx.CreateAuthor(ctx, "Robert Griesemer")
	if err1 != nil || err2 != nil {
		tx.Rollback()
		fmt.Println("  rollback (ada yang gagal)")
	} else {
		tx.Commit()
		fmt.Println("  commit: 2 author tersimpan bersama")
	}

	var jumlah int
	_ = db.QueryRowContext(ctx, `SELECT COUNT(*) FROM authors`).Scan(&jumlah)
	fmt.Printf("  total author di DB: %d\n", jumlah)
}
