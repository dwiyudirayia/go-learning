// REAL-CASE Modul 21 (migrations) — golang-migrate + POSTGRESQL + migrasi ter-embed.
//
// Versi advanced/ menyimulasikan konsep dgn SQLite. Versi ini memakai
// golang-migrate sungguhan terhadap PostgreSQL, dengan file migrasi .sql yang
// DISEMATKAN ke binary (embed.FS) -> satu artefak deploy, migrasi ikut terbawa.
//
// Auto-skip bila POSTGRES_DSN kosong. Jalankan nyata (WAJIB sslmode=disable lokal):
//
//	docker compose -f 21-migrations/real-case/docker-compose.yml up -d
//	POSTGRES_DSN='postgres://postgres:secret@127.0.0.1:5432/appdb?sslmode=disable' \
//	  go run ./21-migrations/real-case
package main

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/lib/pq" // driver database/sql bernama "postgres" (dipakai golang-migrate)
)

// migrationsFS menyematkan seluruh file .sql di folder migrations/ ke binary.
// Konvensi nama file: {versi}_{judul}.{up|down}.sql.
//
//go:embed migrations/*.sql
var migrationsFS embed.FS

// newMigrator merakit objek *migrate.Migrate dari: SOURCE (file ter-embed) +
// DATABASE (Postgres). Ini "mesin" yang menjalankan Up/Down/Version.
//
// Param:
//   - db : koneksi *sql.DB ke Postgres (driver lib/pq).
//
// Return migrator siap pakai atau error bila setup gagal.
func newMigrator(db *sql.DB) (*migrate.Migrate, error) {
	// SOURCE: baca migrasi dari embed.FS (subfolder "migrations").
	src, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return nil, err
	}
	// DATABASE: bungkus koneksi Postgres yang sudah ada (bukan buka DSN baru).
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return nil, err
	}
	return migrate.NewWithInstance("iofs", src, "postgres", driver)
}

func main() {
	dsn := os.Getenv("POSTGRES_DSN")
	if dsn == "" {
		fmt.Println("⏭️  DILEWATI: set POSTGRES_DSN (dgn ?sslmode=disable) untuk versi nyata.")
		fmt.Println("   docker compose -f 21-migrations/real-case/docker-compose.yml up -d")
		fmt.Println("   POSTGRES_DSN='postgres://postgres:secret@127.0.0.1:5432/appdb?sslmode=disable' \\")
		fmt.Println("     go run ./21-migrations/real-case")
		return
	}

	// lib/pq mendaftarkan driver bernama "postgres".
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		panic("gagal konek Postgres: " + err.Error())
	}

	m, err := newMigrator(db)
	if err != nil {
		panic(err)
	}

	// =====================================================================
	// 1. UP: terapkan semua migrasi yang belum dijalankan.
	//    migrate.ErrNoChange = sudah paling mutakhir (bukan error sungguhan).
	// =====================================================================
	fmt.Println("== migrate UP ==")
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		panic(err)
	}
	printVersion(m) // harus versi 2 (dua migrasi diterapkan)

	// Idempoten: Up lagi tak melakukan apa-apa.
	fmt.Println("== migrate UP lagi (idempoten) ==")
	if err := m.Up(); errors.Is(err, migrate.ErrNoChange) {
		fmt.Println("  tidak ada migrasi baru (ErrNoChange)")
	}

	// =====================================================================
	// 2. DOWN satu langkah: Steps(-1) me-rollback migrasi terakhir.
	// =====================================================================
	fmt.Println("== rollback 1 langkah (Steps(-1)) ==")
	if err := m.Steps(-1); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		panic(err)
	}
	printVersion(m) // kembali ke versi 1

	fmt.Println("== produksi: jalankan migrasi di CI/CD (job terpisah), bukan saat boot app ==")
}

// printVersion menampilkan versi skema saat ini + status dirty.
// "dirty" = migrasi gagal di tengah -> perlu `force` setelah perbaikan manual.
func printVersion(m *migrate.Migrate) {
	v, dirty, err := m.Version()
	if errors.Is(err, migrate.ErrNilVersion) {
		fmt.Println("  versi: (belum ada migrasi)")
		return
	}
	if err != nil {
		panic(err)
	}
	fmt.Printf("  versi skema=%d dirty=%v\n", v, dirty)
}
