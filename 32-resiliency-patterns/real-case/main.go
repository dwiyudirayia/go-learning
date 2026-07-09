// REAL-CASE Modul 32 (resiliency) — DISTRIBUTED LOCK di REDIS.
//
// Bulkhead/circuit-breaker (advanced/) bekerja per-proses. Untuk koordinasi
// LINTAS instance (mis. "hanya satu pod boleh menjalankan job ini"), butuh lock
// terdistribusi. Pola dasar: SET key value NX PX ttl (atomik, dgn kedaluwarsa
// agar tak deadlock bila pemegang mati). Rilis pakai skrip Lua yang mengecek
// TOKEN (fencing) -> hanya pemilik yang boleh melepas.
//
// Auto-skip bila REDIS_ADDR kosong. Jalankan nyata:
//
//	docker compose -f 32-resiliency-patterns/real-case/docker-compose.yml up -d
//	REDIS_ADDR=127.0.0.1:6379 go run ./32-resiliency-patterns/real-case
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

// Rilis aman: DEL hanya bila nilainya masih token kita (cegah menghapus lock
// milik proses lain yang sudah mengambil alih setelah TTL kita habis).
var releaseScript = redis.NewScript(`
if redis.call("get", KEYS[1]) == ARGV[1] then
	return redis.call("del", KEYS[1])
else
	return 0
end`)

func acquire(ctx context.Context, rdb *redis.Client, key, token string, ttl time.Duration) bool {
	ok, err := rdb.SetNX(ctx, key, token, ttl).Result()
	if err != nil {
		panic(err)
	}
	return ok
}

func main() {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		fmt.Println("⏭️  DILEWATI: set REDIS_ADDR untuk versi nyata.")
		fmt.Println("   docker compose -f 32-resiliency-patterns/real-case/docker-compose.yml up -d")
		fmt.Println("   REDIS_ADDR=127.0.0.1:6379 go run ./32-resiliency-patterns/real-case")
		return
	}
	ctx := context.Background()
	rdb := redis.NewClient(&redis.Options{Addr: addr})
	defer rdb.Close()
	if err := rdb.Ping(ctx).Err(); err != nil {
		panic("gagal konek Redis: " + err.Error())
	}

	const lockKey = "lock:laporan-harian"
	rdb.Del(ctx, lockKey)

	fmt.Println("== distributed lock (Redis) ==")
	// Worker A mengambil lock.
	gotA := acquire(ctx, rdb, lockKey, "worker-A", 10*time.Second)
	fmt.Println("  worker A ambil lock:", gotA)

	// Worker B mencoba saat lock masih dipegang A -> gagal (fail-fast).
	gotB := acquire(ctx, rdb, lockKey, "worker-B", 10*time.Second)
	fmt.Println("  worker B ambil lock (saat A pegang):", gotB)

	// A selesai -> rilis aman (hanya bila token cocok).
	n, _ := releaseScript.Run(ctx, rdb, []string{lockKey}, "worker-A").Int()
	fmt.Println("  worker A rilis lock (terhapus?):", n == 1)

	// Sekarang B bisa ambil.
	gotB2 := acquire(ctx, rdb, lockKey, "worker-B", 10*time.Second)
	fmt.Println("  worker B ambil lock (setelah A rilis):", gotB2)
	releaseScript.Run(ctx, rdb, []string{lockKey}, "worker-B")

	fmt.Println("  catatan: Redlock (multi-node) & fencing token untuk correctness ketat.")
}
