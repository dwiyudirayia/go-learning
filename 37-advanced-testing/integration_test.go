package main

import (
	"os"
	"testing"
)

// Integration test dengan database SUNGGUHAN via testcontainers.
// Otomatis DI-SKIP kecuali env RUN_INTEGRATION=1 (butuh Docker).
//
// Contoh pola testcontainers (butuh github.com/testcontainers/testcontainers-go):
//
//	ctx := context.Background()
//	pg, _ := postgres.Run(ctx, "postgres:16",
//	    postgres.WithDatabase("test"), postgres.WithUsername("u"), postgres.WithPassword("p"))
//	defer pg.Terminate(ctx)
//	dsn, _ := pg.ConnectionString(ctx)
//	db, _ := sql.Open("pgx", dsn)   // -> uji query terhadap Postgres NYATA, lalu wadah dihapus
//
// Keunggulan: uji terhadap DB asli (bukan mock), otomatis start/stop container,
// bersih tiap run, jalan di CI. Lihat README.
func TestWithRealDatabase(t *testing.T) {
	if os.Getenv("RUN_INTEGRATION") == "" {
		t.Skip("set RUN_INTEGRATION=1 (butuh Docker) untuk menjalankan integration test")
	}
	// (Di sini kamu akan menjalankan testcontainers + query DB nyata.)
	t.Log("integration test akan berjalan di sini dengan DB sungguhan")
}
