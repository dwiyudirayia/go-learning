package main

import (
	"context"
	"database/sql"
)

// OUTBOX PATTERN menyelesaikan masalah "dual write": bagaimana mengubah database
// DAN mengirim event secara ATOMIK? Jawaban: tulis event ke tabel `outbox` dalam
// TRANSAKSI yang sama dengan perubahan bisnis. Sebuah RELAY membaca outbox lalu
// mem-publish event (at-least-once). Tak ada event yang hilang meski app crash
// setelah commit tapi sebelum publish.

// SetupSchema membuat tabel yang dibutuhkan.
func SetupSchema(db *sql.DB) error {
	for _, stmt := range []string{
		`CREATE TABLE IF NOT EXISTS orders (id INTEGER PRIMARY KEY AUTOINCREMENT, item TEXT NOT NULL)`,
		`CREATE TABLE IF NOT EXISTS outbox (id INTEGER PRIMARY KEY AUTOINCREMENT, topic TEXT NOT NULL, payload TEXT NOT NULL, sent INTEGER NOT NULL DEFAULT 0)`,
	} {
		if _, err := db.Exec(stmt); err != nil {
			return err
		}
	}
	return nil
}

// CreateOrder menyimpan order DAN event "order.created" dalam SATU transaksi.
// Keduanya berhasil bersama atau gagal bersama (atomik).
func CreateOrder(db *sql.DB, item, eventPayload string) (int64, error) {
	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	res, err := tx.Exec("INSERT INTO orders(item) VALUES(?)", item)
	if err != nil {
		_ = tx.Rollback()
		return 0, err
	}
	id, _ := res.LastInsertId()

	// Event masuk outbox dalam transaksi yang SAMA.
	if _, err := tx.Exec("INSERT INTO outbox(topic,payload) VALUES(?,?)", "order.created", eventPayload); err != nil {
		_ = tx.Rollback()
		return 0, err
	}
	return id, tx.Commit()
}

// Relay membaca event pending dari outbox lalu mem-publish-nya.
type Relay struct {
	db      *sql.DB
	publish func(topic, payload string) error
}

func NewRelay(db *sql.DB, publish func(topic, payload string) error) *Relay {
	return &Relay{db: db, publish: publish}
}

// ProcessOnce mem-publish semua event pending & menandainya terkirim.
// Mengembalikan jumlah event yang diproses. Idempotent: event yang sudah
// terkirim (sent=1) tak diproses lagi.
func (r *Relay) ProcessOnce(ctx context.Context) (int, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, topic, payload FROM outbox WHERE sent = 0 ORDER BY id")
	if err != nil {
		return 0, err
	}
	type ev struct {
		id             int64
		topic, payload string
	}
	var pending []ev
	for rows.Next() {
		var e ev
		if err := rows.Scan(&e.id, &e.topic, &e.payload); err != nil {
			rows.Close()
			return 0, err
		}
		pending = append(pending, e)
	}
	rows.Close()

	count := 0
	for _, e := range pending {
		if err := r.publish(e.topic, e.payload); err != nil {
			// Gagal publish -> biarkan sent=0, coba lagi di ProcessOnce berikutnya.
			return count, err
		}
		if _, err := r.db.ExecContext(ctx, "UPDATE outbox SET sent = 1 WHERE id = ?", e.id); err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}
