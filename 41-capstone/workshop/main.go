package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "modernc.org/sqlite" // driver SQLite pure-Go (Modul 14) — tanpa CGo
)

func main() {
	// ":memory:" = database sementara di RAM, hilang saat program berhenti.
	// Cocok untuk mencoba-coba. Nanti kita ganti ke file di langkah akhir.
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	store := NewStore(db)
	if err := store.setupSchema(); err != nil {
		log.Fatal(err)
	}

	// 1) Simpan sebuah link.
	link, err := store.CreateLink("go123", "https://go.dev")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Dibuat: id=%d code=%q url=%q\n", link.ID, link.Code, link.URL)

	// 2) Ambil kembali & catat beberapa klik.
	for i := 0; i < 3; i++ {
		_ = store.IncrementClicks("go123")
	}
	got, err := store.GetLinkByCode("go123")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Dibaca : code=%q url=%q clicks=%d\n", got.Code, got.URL, got.Clicks)

	// 3) Coba ambil kode yang tak ada -> harus ErrNotFound.
	_, err = store.GetLinkByCode("tidakada")
	fmt.Printf("Cari 'tidakada' -> error: %v\n", err)
}
