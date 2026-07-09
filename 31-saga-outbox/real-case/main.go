// REAL-CASE Modul 31 (transactional outbox) — POSTGRESQL + KAFKA.
//
// Ini bentuk produksi dari outbox: perubahan bisnis (orders) + event (outbox)
// ditulis dalam SATU transaksi Postgres (atomik). RELAY terpisah mem-poll tabel
// outbox lalu mem-publish ke Kafka & menandai terkirim. Menghindari masalah
// "DB commit tapi publish gagal" tanpa distributed transaction.
//
// (Di produksi, relay sering digantikan CDC seperti Debezium yang men-tail WAL
// Postgres langsung ke Kafka.)
//
// Auto-skip bila POSTGRES_DSN atau KAFKA_BROKERS kosong. Jalankan nyata:
//
//	docker compose -f 31-saga-outbox/real-case/docker-compose.yml up -d
//	POSTGRES_DSN='postgres://postgres:secret@127.0.0.1:5432/appdb' \
//	KAFKA_BROKERS=127.0.0.1:9092 go run ./31-saga-outbox/real-case
package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/segmentio/kafka-go"
)

func main() {
	dsn := os.Getenv("POSTGRES_DSN")
	brokersEnv := os.Getenv("KAFKA_BROKERS")
	if dsn == "" || brokersEnv == "" {
		fmt.Println("⏭️  DILEWATI: set POSTGRES_DSN & KAFKA_BROKERS untuk versi nyata.")
		fmt.Println("   docker compose -f 31-saga-outbox/real-case/docker-compose.yml up -d")
		fmt.Println("   POSTGRES_DSN='postgres://postgres:secret@127.0.0.1:5432/appdb' \\")
		fmt.Println("   KAFKA_BROKERS=127.0.0.1:9092 go run ./31-saga-outbox/real-case")
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
	mustExec(ctx, pool, `CREATE TABLE IF NOT EXISTS orders(id TEXT PRIMARY KEY, total INT NOT NULL)`)
	mustExec(ctx, pool, `CREATE TABLE IF NOT EXISTS outbox(
		id BIGSERIAL PRIMARY KEY, topik TEXT NOT NULL, payload TEXT NOT NULL, terkirim BOOL NOT NULL DEFAULT false)`)
	mustExec(ctx, pool, `TRUNCATE orders`)
	mustExec(ctx, pool, `TRUNCATE outbox RESTART IDENTITY`)

	// =====================================================================
	// 1. TRANSAKSI ATOMIK: order + event outbox bersama.
	// =====================================================================
	tx, err := pool.Begin(ctx)
	if err != nil {
		panic(err)
	}
	_, err = tx.Exec(ctx, `INSERT INTO orders(id,total) VALUES($1,$2)`, "ord-1", 250)
	if err == nil {
		_, err = tx.Exec(ctx, `INSERT INTO outbox(topik,payload) VALUES($1,$2)`, "order.created", `{"id":"ord-1"}`)
	}
	if err != nil {
		tx.Rollback(ctx)
		panic(err)
	}
	if err := tx.Commit(ctx); err != nil {
		panic(err)
	}
	fmt.Println("== 1. order + event outbox ditulis atomik (Postgres) ==")

	// =====================================================================
	// 2. RELAY: poll outbox belum terkirim -> publish ke Kafka -> tandai.
	// =====================================================================
	brokers := strings.Split(brokersEnv, ",")
	w := &kafka.Writer{Addr: kafka.TCP(brokers...), Topic: "order.created", AllowAutoTopicCreation: true}
	defer w.Close()

	rows, err := pool.Query(ctx, `SELECT id, topik, payload FROM outbox WHERE terkirim=false ORDER BY id`)
	if err != nil {
		panic(err)
	}
	type ob struct {
		id             int64
		topik, payload string
	}
	var batch []ob
	for rows.Next() {
		var o ob
		if err := rows.Scan(&o.id, &o.topik, &o.payload); err != nil {
			panic(err)
		}
		batch = append(batch, o)
	}
	rows.Close()

	fmt.Println("== 2. relay -> Kafka ==")
	for _, o := range batch {
		wc, cancel := context.WithTimeout(ctx, 10*time.Second)
		err := w.WriteMessages(wc, kafka.Message{Value: []byte(o.payload)})
		cancel()
		if err != nil {
			fmt.Println("  publish gagal (akan dicoba lagi di siklus berikutnya):", err)
			continue // JANGAN tandai terkirim -> retry aman (at-least-once)
		}
		mustExec(ctx, pool, `UPDATE outbox SET terkirim=true WHERE id=$1`, o.id)
		fmt.Printf("  publish %s -> Kafka, ditandai terkirim\n", o.topik)
	}
	fmt.Println("  (produksi: Debezium CDC men-tail WAL Postgres -> Kafka, tanpa polling)")
}

func mustExec(ctx context.Context, pool *pgxpool.Pool, sql string, args ...any) {
	if _, err := pool.Exec(ctx, sql, args...); err != nil {
		panic(err)
	}
}
