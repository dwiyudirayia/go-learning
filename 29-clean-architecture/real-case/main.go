// REAL-CASE Modul 29 (clean architecture) — adapter repository POSTGRESQL.
//
// Inti clean architecture: DOMAIN & USE CASE tak berubah saat infra berganti.
// Versi advanced/ memakai repo in-memory; versi ini menukarnya dengan adapter
// PostgreSQL. Perhatikan: kode domain (Order, OrderRepo, PlaceOrder) IDENTIK —
// hanya adapter di lapisan luar yang beda. Itulah manfaat Dependency Inversion.
//
// Auto-skip bila POSTGRES_DSN kosong. Jalankan nyata:
//
//	docker compose -f 29-clean-architecture/real-case/docker-compose.yml up -d
//	POSTGRES_DSN='postgres://postgres:secret@127.0.0.1:5432/appdb' go run ./29-clean-architecture/real-case
package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ===========================================================================
// DOMAIN (inti) — tak tahu apa pun soal Postgres/pgx. SAMA dgn versi advanced/.
// ===========================================================================

// Order = entitas domain.
type Order struct {
	ID    string
	Total int
}

var ErrInvalid = errors.New("order tidak valid")

// OrderRepo = PORT: kebutuhan penyimpanan yang DIDEFINISIKAN domain. Adapter
// mana pun (Postgres, in-memory, dll) harus memenuhi kontrak ini.
type OrderRepo interface {
	Save(ctx context.Context, o Order) error
	Get(ctx context.Context, id string) (Order, bool, error)
}

// PlaceOrder = USE CASE: aturan aplikasi, bergantung pada PORT (bukan Postgres).
type PlaceOrder struct{ repo OrderRepo }

// Exec menjalankan aturan bisnis lalu menyimpan lewat port. Tak ada SQL di sini.
func (uc PlaceOrder) Exec(ctx context.Context, id string, total int) error {
	if id == "" || total <= 0 {
		return fmt.Errorf("%w: id/total", ErrInvalid)
	}
	return uc.repo.Save(ctx, Order{ID: id, Total: total})
}

// ===========================================================================
// ADAPTER (luar) — implementasi OrderRepo memakai PostgreSQL/pgx.
// ===========================================================================

// PostgresOrderRepo = adapter konkret. Menyimpan *pgxpool.Pool, BUKAN bagian domain.
type PostgresOrderRepo struct{ pool *pgxpool.Pool }

// migrate memastikan tabel ada (di produksi pakai golang-migrate, lihat modul 21).
func (r *PostgresOrderRepo) migrate(ctx context.Context) error {
	_, err := r.pool.Exec(ctx, `CREATE TABLE IF NOT EXISTS orders(id TEXT PRIMARY KEY, total INT NOT NULL)`)
	return err
}

// Save menerjemahkan entitas domain -> baris SQL (INSERT/UPSERT).
func (r *PostgresOrderRepo) Save(ctx context.Context, o Order) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO orders(id,total) VALUES($1,$2)
		 ON CONFLICT(id) DO UPDATE SET total=EXCLUDED.total`,
		o.ID, o.Total)
	return err
}

// Get menerjemahkan baris SQL -> entitas domain. pgx.ErrNoRows -> (…, false).
func (r *PostgresOrderRepo) Get(ctx context.Context, id string) (Order, bool, error) {
	var o Order
	err := r.pool.QueryRow(ctx, `SELECT id, total FROM orders WHERE id=$1`, id).Scan(&o.ID, &o.Total)
	if errors.Is(err, pgx.ErrNoRows) {
		return Order{}, false, nil
	}
	if err != nil {
		return Order{}, false, err
	}
	return o, true, nil
}

func main() {
	dsn := os.Getenv("POSTGRES_DSN")
	if dsn == "" {
		fmt.Println("⏭️  DILEWATI: set POSTGRES_DSN untuk versi nyata.")
		fmt.Println("   docker compose -f 29-clean-architecture/real-case/docker-compose.yml up -d")
		fmt.Println("   POSTGRES_DSN='postgres://postgres:secret@127.0.0.1:5432/appdb' go run ./29-clean-architecture/real-case")
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

	// COMPOSITION ROOT (di main): rakit adapter Postgres -> use case.
	// Andai mau ganti ke in-memory/MySQL, cukup ubah BARIS INI. Domain & use
	// case tak tersentuh.
	repo := &PostgresOrderRepo{pool: pool}
	if err := repo.migrate(ctx); err != nil {
		panic(err)
	}
	uc := PlaceOrder{repo: repo}

	fmt.Println("== use case memakai adapter Postgres ==")
	if err := uc.Exec(ctx, "ord-1", 150); err != nil {
		fmt.Println("  error:", err)
	} else {
		fmt.Println("  order tersimpan ke Postgres")
	}
	if err := uc.Exec(ctx, "ord-2", 0); err != nil { // langgar aturan domain
		fmt.Println("  ditolak aturan domain:", err)
	}
	if o, ada, _ := repo.Get(ctx, "ord-1"); ada {
		fmt.Printf("  dibaca dari Postgres: %+v\n", o)
	}
	fmt.Println("  => domain/use case TAK berubah dari versi in-memory; hanya adapter.")
}
