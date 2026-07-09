// REAL-CASE Modul 14 (database) dengan POSTGRESQL (pgx/pgxpool).
//
// SQLite (advanced/) bagus untuk lokal/embedded, tapi backend produksi umumnya
// PostgreSQL: konkurensi tinggi, tipe kaya (JSONB, array), replikasi. Ini
// memakai driver pgx + connection pool.
//
// Auto-skip bila POSTGRES_DSN kosong. Jalankan nyata:
//
//	docker compose -f 14-database/real-case/docker-compose.yml up -d
//	POSTGRES_DSN='postgres://postgres:secret@127.0.0.1:5432/appdb' go run ./14-database/real-case
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	dsn := os.Getenv("POSTGRES_DSN")
	if dsn == "" {
		fmt.Println("⏭️  DILEWATI: set POSTGRES_DSN untuk versi nyata.")
		fmt.Println("   docker compose -f 14-database/real-case/docker-compose.yml up -d")
		fmt.Println("   POSTGRES_DSN='postgres://postgres:secret@127.0.0.1:5432/appdb' go run ./14-database/real-case")
		return
	}
	ctx := context.Background()

	// =====================================================================
	// Connection POOL (pgxpool) — tuning penting untuk produksi.
	// =====================================================================
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		panic(err)
	}
	cfg.MaxConns = 10
	cfg.MinConns = 2
	cfg.MaxConnLifetime = time.Hour
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		panic(err)
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		panic("gagal konek Postgres: " + err.Error())
	}

	// Postgres pakai placeholder $1,$2 (bukan ?). Tipe kaya: SERIAL, TIMESTAMPTZ.
	_, err = pool.Exec(ctx, `CREATE TABLE IF NOT EXISTS artikel(
		id SERIAL PRIMARY KEY,
		judul TEXT NOT NULL,
		tags TEXT[] NOT NULL DEFAULT '{}',
		dibuat TIMESTAMPTZ NOT NULL DEFAULT now()
	)`)
	if err != nil {
		panic(err)
	}
	_, _ = pool.Exec(ctx, `TRUNCATE artikel`)

	// =====================================================================
	// TRANSAKSI — beberapa insert atomik (pola pgx).
	// =====================================================================
	tx, err := pool.Begin(ctx)
	if err != nil {
		panic(err)
	}
	// Array Postgres langsung dari []string.
	_, err = tx.Exec(ctx, `INSERT INTO artikel(judul, tags) VALUES($1, $2)`, "Belajar Go", []string{"go", "backend"})
	if err == nil {
		_, err = tx.Exec(ctx, `INSERT INTO artikel(judul, tags) VALUES($1, $2)`, "Tips SQL", []string{"sql"})
	}
	if err != nil {
		tx.Rollback(ctx)
		panic(err)
	}
	if err := tx.Commit(ctx); err != nil {
		panic(err)
	}

	// =====================================================================
	// Query + scan (termasuk kolom array TEXT[] -> []string).
	// =====================================================================
	rows, err := pool.Query(ctx, `SELECT judul, tags FROM artikel ORDER BY id`)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	fmt.Println("== PostgreSQL via pgx ==")
	for rows.Next() {
		var judul string
		var tags []string
		if err := rows.Scan(&judul, &tags); err != nil {
			panic(err)
		}
		fmt.Printf("  %-12s tags=%v\n", judul, tags)
	}
	if err := rows.Err(); err != nil {
		panic(err)
	}

	// =====================================================================
	// BULK insert via CopyFrom — protokol COPY Postgres (jauh lebih cepat
	// daripada INSERT satu-satu untuk data besar).
	// =====================================================================
	n, err := pool.CopyFrom(ctx,
		pgx.Identifier{"artikel"},
		[]string{"judul", "tags"},
		pgx.CopyFromRows([][]any{
			{"Bulk A", []string{"go"}},
			{"Bulk B", []string{"db", "perf"}},
		}))
	if err != nil {
		panic(err)
	}
	fmt.Printf("  CopyFrom (COPY protocol) menyisipkan %d baris sekaligus\n", n)
}
