// REAL-CASE Modul 36 (sqlc) dengan POSTGRESQL + pgx (target sqlc sebenarnya).
//
// sqlc meng-generate kode Go type-safe dari SQL untuk pgx. File ini meniru
// BENTUK kode hasil generate (interface DBTX, struct Queries, WithTx) terhadap
// PostgreSQL nyata + pool tuning + CopyFrom bulk.
//
// Auto-skip bila POSTGRES_DSN kosong. Jalankan nyata:
//
//	docker compose -f 36-sqlc-advanced-db/real-case/docker-compose.yml up -d
//	POSTGRES_DSN='postgres://postgres:secret@127.0.0.1:5432/appdb' go run ./36-sqlc-advanced-db/real-case
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DBTX = abstraksi yang dipenuhi *pgxpool.Pool DAN pgx.Tx -> query yang sama
// bisa jalan di luar atau di dalam transaksi. Persis pola kode generate sqlc.
type DBTX interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

type Author struct {
	ID   int64
	Nama string
}

type Queries struct{ db DBTX }

func New(db DBTX) *Queries                   { return &Queries{db: db} }
func (q *Queries) WithTx(tx pgx.Tx) *Queries { return &Queries{db: tx} }

// CreateAuthor: type-safe, memakai RETURNING untuk mengambil id (idiom Postgres).
func (q *Queries) CreateAuthor(ctx context.Context, nama string) (int64, error) {
	var id int64
	err := q.db.QueryRow(ctx, `INSERT INTO authors(nama) VALUES($1) RETURNING id`, nama).Scan(&id)
	return id, err
}

func (q *Queries) GetAuthor(ctx context.Context, id int64) (Author, error) {
	var a Author
	err := q.db.QueryRow(ctx, `SELECT id, nama FROM authors WHERE id=$1`, id).Scan(&a.ID, &a.Nama)
	return a, err
}

func main() {
	dsn := os.Getenv("POSTGRES_DSN")
	if dsn == "" {
		fmt.Println("⏭️  DILEWATI: set POSTGRES_DSN untuk versi nyata.")
		fmt.Println("   docker compose -f 36-sqlc-advanced-db/real-case/docker-compose.yml up -d")
		fmt.Println("   POSTGRES_DSN='postgres://postgres:secret@127.0.0.1:5432/appdb' go run ./36-sqlc-advanced-db/real-case")
		return
	}
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		panic(err)
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		panic("gagal konek Postgres: " + err.Error())
	}
	if _, err := pool.Exec(ctx, `CREATE TABLE IF NOT EXISTS authors(id BIGSERIAL PRIMARY KEY, nama TEXT NOT NULL)`); err != nil {
		panic(err)
	}
	_, _ = pool.Exec(ctx, `TRUNCATE authors RESTART IDENTITY`)

	q := New(pool)

	// Pemakaian biasa (di luar transaksi).
	id, err := q.CreateAuthor(ctx, "Rob Pike")
	if err != nil {
		panic(err)
	}
	a, _ := q.GetAuthor(ctx, id)
	fmt.Println("== sqlc-style + Postgres ==")
	fmt.Printf("  dibuat & dibaca: %+v\n", a)

	// TRANSAKSI via WithTx (pgx.Tx) — atomik.
	tx, err := pool.Begin(ctx)
	if err != nil {
		panic(err)
	}
	qtx := q.WithTx(tx)
	_, e1 := qtx.CreateAuthor(ctx, "Ken Thompson")
	_, e2 := qtx.CreateAuthor(ctx, "Robert Griesemer")
	if e1 != nil || e2 != nil {
		tx.Rollback(ctx)
		fmt.Println("  rollback")
	} else {
		tx.Commit(ctx)
		fmt.Println("  commit: 2 author via transaksi")
	}

	// BULK via CopyFrom (COPY protocol) — insert massal tercepat.
	n, err := pool.CopyFrom(ctx, pgx.Identifier{"authors"}, []string{"nama"},
		pgx.CopyFromRows([][]any{{"Alan Turing"}, {"Grace Hopper"}}))
	if err != nil {
		panic(err)
	}
	fmt.Printf("  CopyFrom: %d baris\n", n)

	var total int
	_ = pool.QueryRow(ctx, `SELECT COUNT(*) FROM authors`).Scan(&total)
	fmt.Printf("  total author: %d\n", total)
}
