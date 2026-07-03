// Jalankan: go run ./31-saga-outbox
// Verifikasi otomatis: go test ./31-saga-outbox
package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

func main() {
	fmt.Println("=== 31 — Saga & Outbox ===")
	demoSagaSukses()
	demoSagaGagal()
	demoOutbox()
}

// Simulasi state 3 service untuk demo saga.
type world struct{ stokDikurangi, dibayar, dikirim bool }

func demoSagaSukses() {
	fmt.Println("\n-- Saga sukses --")
	w := &world{}
	saga := buildOrderSaga(w, false)
	if err := saga.Execute(context.Background()); err != nil {
		fmt.Println("error:", err)
	}
	fmt.Printf("jejak: %v\n", saga.Log())
	fmt.Printf("state akhir: %+v\n", *w)
}

func demoSagaGagal() {
	fmt.Println("\n-- Saga gagal di 'kirim' -> kompensasi mundur --")
	w := &world{}
	saga := buildOrderSaga(w, true) // langkah kirim akan gagal
	if err := saga.Execute(context.Background()); err != nil {
		fmt.Println("error:", err)
	}
	fmt.Printf("jejak: %v\n", saga.Log())
	fmt.Printf("state akhir: %+v (semua ter-undo)\n", *w)
}

// buildOrderSaga: reserve stok -> bayar -> kirim (dengan kompensasi tiap langkah).
func buildOrderSaga(w *world, gagalKirim bool) *Saga {
	return NewSaga().
		AddStep(Step{
			Name:       "reserve-stok",
			Action:     func(context.Context) error { w.stokDikurangi = true; return nil },
			Compensate: func(context.Context) error { w.stokDikurangi = false; return nil },
		}).
		AddStep(Step{
			Name:       "bayar",
			Action:     func(context.Context) error { w.dibayar = true; return nil },
			Compensate: func(context.Context) error { w.dibayar = false; return nil },
		}).
		AddStep(Step{
			Name: "kirim",
			Action: func(context.Context) error {
				if gagalKirim {
					return errors.New("kurir tidak tersedia")
				}
				w.dikirim = true
				return nil
			},
			Compensate: func(context.Context) error { w.dikirim = false; return nil },
		})
}

func demoOutbox() {
	fmt.Println("\n-- Outbox: DB + event atomik --")
	path := filepath.Join(os.TempDir(), "m31.db")
	_ = os.Remove(path)
	defer os.Remove(path)

	db, err := sql.Open("sqlite", path)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	if err := SetupSchema(db); err != nil {
		log.Fatal(err)
	}

	// Simpan order + event dalam satu transaksi.
	id, _ := CreateOrder(db, "Keyboard", `{"order_id":1,"item":"Keyboard"}`)
	fmt.Printf("order #%d dibuat (order + event tersimpan atomik)\n", id)

	// Relay mem-publish event dari outbox.
	relay := NewRelay(db, func(topic, payload string) error {
		fmt.Printf("  [relay] publish %s: %s\n", topic, payload)
		return nil
	})
	n, _ := relay.ProcessOnce(context.Background())
	fmt.Printf("relay mem-publish %d event\n", n)

	// Jalankan lagi -> 0 (event sudah ditandai terkirim).
	n2, _ := relay.ProcessOnce(context.Background())
	fmt.Printf("relay jalan lagi -> %d event (idempotent)\n", n2)
}
