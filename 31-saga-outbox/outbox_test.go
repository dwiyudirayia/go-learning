package main

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

func newDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	if err := SetupSchema(db); err != nil {
		t.Fatalf("schema: %v", err)
	}
	return db
}

func TestOutboxAtomik(t *testing.T) {
	db := newDB(t)

	if _, err := CreateOrder(db, "Mouse", `{"id":1}`); err != nil {
		t.Fatalf("create: %v", err)
	}

	// order & event harus sama-sama ada (atomik).
	var orders, events int
	db.QueryRow("SELECT COUNT(*) FROM orders").Scan(&orders)
	db.QueryRow("SELECT COUNT(*) FROM outbox").Scan(&events)
	if orders != 1 || events != 1 {
		t.Errorf("orders=%d events=%d; want 1,1", orders, events)
	}
}

func TestRelayPublishDanIdempotent(t *testing.T) {
	db := newDB(t)
	_, _ = CreateOrder(db, "A", `{"id":1}`)
	_, _ = CreateOrder(db, "B", `{"id":2}`)

	var published []string
	relay := NewRelay(db, func(topic, payload string) error {
		published = append(published, payload)
		return nil
	})

	n, err := relay.ProcessOnce(context.Background())
	if err != nil {
		t.Fatalf("relay: %v", err)
	}
	if n != 2 || len(published) != 2 {
		t.Errorf("published = %d; want 2", n)
	}

	// Jalan lagi -> tidak ada yang baru (sudah ditandai sent).
	n2, _ := relay.ProcessOnce(context.Background())
	if n2 != 0 {
		t.Errorf("relay kedua = %d; want 0 (idempotent)", n2)
	}
}

func TestRelayGagalPublishTidakMenandaiSent(t *testing.T) {
	db := newDB(t)
	_, _ = CreateOrder(db, "A", `{"id":1}`)

	// Publish selalu gagal.
	relay := NewRelay(db, func(topic, payload string) error {
		return context.DeadlineExceeded
	})
	if _, err := relay.ProcessOnce(context.Background()); err == nil {
		t.Error("mengharapkan error publish")
	}

	// Event harus tetap pending (sent=0) agar bisa dicoba ulang.
	var pending int
	db.QueryRow("SELECT COUNT(*) FROM outbox WHERE sent = 0").Scan(&pending)
	if pending != 1 {
		t.Errorf("pending = %d; want 1 (event tak boleh hilang saat publish gagal)", pending)
	}
}
