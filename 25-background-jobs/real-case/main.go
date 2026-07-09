// REAL-CASE Modul 25 (background jobs) — ANTREAN DURABLE di REDIS.
//
// Beda dari worker-pool in-memory (advanced/) yang HILANG saat proses mati,
// antrean di Redis PERSIST -> job selamat dari restart. Pola "reliable queue":
// BLMOVE memindah job dari daftar pending ke daftar processing secara ATOMIK
// (at-least-once), lalu dihapus dari processing saat sukses (ack). Bila worker
// mati saat memproses, job tetap di processing dan bisa dipulihkan.
//
// Auto-skip bila REDIS_ADDR kosong. Jalankan nyata:
//
//	docker compose -f 25-background-jobs/real-case/docker-compose.yml up -d
//	REDIS_ADDR=127.0.0.1:6379 go run ./25-background-jobs/real-case
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	pendingKey    = "jobs:pending"
	processingKey = "jobs:processing"
)

func main() {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		fmt.Println("⏭️  DILEWATI: set REDIS_ADDR untuk versi nyata.")
		fmt.Println("   docker compose -f 25-background-jobs/real-case/docker-compose.yml up -d")
		fmt.Println("   REDIS_ADDR=127.0.0.1:6379 go run ./25-background-jobs/real-case")
		return
	}
	ctx := context.Background()
	rdb := redis.NewClient(&redis.Options{Addr: addr})
	defer rdb.Close()
	if err := rdb.Ping(ctx).Err(); err != nil {
		panic("gagal konek Redis: " + err.Error())
	}
	rdb.Del(ctx, pendingKey, processingKey) // mulai bersih

	// PRODUCER: masukkan job ke antrean pending (LPUSH). Ini PERSIST di Redis.
	for i := 1; i <= 5; i++ {
		if err := rdb.LPush(ctx, pendingKey, fmt.Sprintf("job-%d", i)).Err(); err != nil {
			panic(err)
		}
	}
	n, _ := rdb.LLen(ctx, pendingKey).Result()
	fmt.Printf("== %d job antre di Redis (durable) ==\n", n)

	// CONSUMER reliable: BLMOVE pending->processing (atomik), proses, lalu ack
	// dengan LREM dari processing. Bila crash sebelum ack, job masih di processing.
	diproses := 0
	for {
		job, err := rdb.BLMove(ctx, pendingKey, processingKey, "RIGHT", "LEFT", time.Second).Result()
		if errors.Is(err, redis.Nil) {
			break // antrean kosong (timeout)
		}
		if err != nil {
			panic(err)
		}
		// ... kerjakan job (idempoten, karena bisa terkirim >1x) ...
		if err := rdb.LRem(ctx, processingKey, 1, job).Err(); err != nil { // ACK
			panic(err)
		}
		diproses++
		fmt.Printf("  diproses & di-ack: %s\n", job)
	}
	sisaProcessing, _ := rdb.LLen(ctx, processingKey).Result()
	fmt.Printf("== selesai: %d job, %d tersangkut di processing (0 = semua bersih) ==\n", diproses, sisaProcessing)
	fmt.Println("  produksi: Asynq (Redis) atau River (Postgres) memberi retry/scheduler/DLQ siap pakai.")
}
