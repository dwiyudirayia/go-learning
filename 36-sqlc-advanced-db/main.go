// Jalankan: go run ./36-sqlc-advanced-db
// Regenerate kode dari SQL: cd 36-sqlc-advanced-db && sqlc generate
// Verifikasi otomatis: go test ./36-sqlc-advanced-db
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"go-learning/36-sqlc-advanced-db/internal/sqlcdb"

	_ "modernc.org/sqlite"
)

func main() {
	fmt.Println("=== 36 — sqlc + pool + transaksi ===")

	path := filepath.Join(os.TempDir(), "m36.db")
	_ = os.Remove(path)
	defer os.Remove(path)

	db, err := OpenDB(path)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	if err := applySchema(db); err != nil {
		log.Fatal(err)
	}

	store := NewStore(db)
	ctx := context.Background()

	// Query type-safe hasil sqlc (tanpa string SQL manual, tanpa Scan manual).
	author, _ := store.CreateAuthor(ctx, "Rob Pike")
	fmt.Printf("CreateAuthor -> %+v (tipe %T)\n", author, author)

	book, _ := store.CreateBook(ctx, sqlcdb.CreateBookParams{Title: "The Go PL", AuthorID: author.ID})
	fmt.Printf("CreateBook   -> %+v\n", book)

	// Transaksi: author + buku dalam satu unit atomik.
	a2, err := store.CreateAuthorWithBook(ctx, "Dennis Ritchie", "The C PL")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("transaksi    -> author %q dibuat dengan bukunya (atomik)\n", a2.Name)

	count, _ := store.CountAuthors(ctx)
	authors, _ := store.ListAuthors(ctx)
	fmt.Printf("total author = %d, list = %d nama\n", count, len(authors))
}
