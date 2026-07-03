// Jalankan: go run ./21-migrations
// Verifikasi otomatis: go test ./21-migrations
package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite" // driver database/sql bernama "sqlite" (pure-Go)
)

func main() {
	fmt.Println("=== 21 — Database Migrations (golang-migrate) ===")

	path := filepath.Join(os.TempDir(), "m21.db")
	_ = os.Remove(path)
	defer os.Remove(path)

	db, err := sql.Open("sqlite", path)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Terapkan semua migrasi.
	if err := Up(db); err != nil {
		log.Fatal(err)
	}
	ver, _, _ := Version(db)
	fmt.Printf("setelah Up  -> versi: %d\n", ver)

	// Jalankan lagi -> idempotent (ErrNoChange ditangani).
	if err := Up(db); err != nil {
		log.Fatal(err)
	}
	fmt.Println("jalankan Up lagi -> tidak ada perubahan (idempotent)")

	// Pakai skema hasil migrasi (kolom 'active' dari migrasi 003).
	_, _ = db.Exec("INSERT INTO users(name,email,active) VALUES(?,?,?)", "Ana", "ana@mail.id", 1)
	var name string
	var active int
	_ = db.QueryRow("SELECT name, active FROM users WHERE id=1").Scan(&name, &active)
	fmt.Printf("query -> user=%s active=%d\n", name, active)

	// Rollback satu langkah (versi 3 -> 2).
	if err := Down(db); err != nil {
		log.Fatal(err)
	}
	ver, _, _ = Version(db)
	fmt.Printf("setelah Down -> versi: %d\n", ver)
}
