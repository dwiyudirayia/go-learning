// REAL-CASE Modul 27 (security) — RATE LIMITER TERDISTRIBUSI di REDIS.
//
// Rate limiter in-memory (advanced/) hanya berlaku per-proses; di belakang load
// balancer dengan banyak pod, tiap pod menghitung sendiri -> batas bocor. Di
// produksi, hitungan disimpan di REDIS agar batas berlaku LINTAS pod.
//
// Ini implementasi fixed-window: INCR kunci per (identitas, jendela waktu),
// set EXPIRE pada increment pertama. Atomik & dibagikan semua instance.
//
// Auto-skip bila REDIS_ADDR kosong. Jalankan nyata:
//
//	docker compose -f 27-security/real-case/docker-compose.yml up -d
//	REDIS_ADDR=127.0.0.1:6379 go run ./27-security/real-case
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

func main() {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		fmt.Println("⏭️  DILEWATI: set REDIS_ADDR untuk versi nyata.")
		fmt.Println("   docker compose -f 27-security/real-case/docker-compose.yml up -d")
		fmt.Println("   REDIS_ADDR=127.0.0.1:6379 go run ./27-security/real-case")
		return
	}
	ctx := context.Background()
	rdb := redis.NewClient(&redis.Options{Addr: addr})
	defer rdb.Close()
	if err := rdb.Ping(ctx).Err(); err != nil {
		panic("gagal konek Redis: " + err.Error())
	}

	const limit = 5
	const window = time.Second

	// allow: true bila jumlah request dalam jendela <= limit. Hitungan di Redis
	// -> berlaku untuk SEMUA pod yang berbagi Redis ini.
	allow := func(identitas string) (bool, int64) {
		bucket := time.Now().Unix() // jendela per-detik
		key := fmt.Sprintf("rl:%s:%d", identitas, bucket)
		n, err := rdb.Incr(ctx, key).Result()
		if err != nil {
			panic(err)
		}
		if n == 1 {
			rdb.Expire(ctx, key, window+time.Second) // set TTL sekali (auto-bersih)
		}
		return n <= limit, n
	}

	fmt.Printf("== rate limit terdistribusi (Redis), limit %d/detik ==\n", limit)
	for i := 1; i <= 8; i++ { // 8 request beruntun untuk IP yang sama
		ok, n := allow("ip-1.2.3.4")
		fmt.Printf("  request %d -> hit ke-%d, diizinkan=%v\n", i, n, ok)
	}
	fmt.Println("  (di produksi, semua pod berbagi hitungan ini -> batas tak bocor)")
}
