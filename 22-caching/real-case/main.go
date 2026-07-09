// REAL-CASE Modul 22 (caching) dengan REDIS SUNGGUHAN (bukan miniredis).
//
// Pola cache-aside memakai server Redis nyata. Client (go-redis) SAMA persis
// dengan yang dipakai versi advanced/ (miniredis) — cukup ganti Addr.
//
// Auto-skip bila REDIS_ADDR kosong. Jalankan nyata:
//
//	docker compose -f 22-caching/real-case/docker-compose.yml up -d
//	REDIS_ADDR=127.0.0.1:6379 go run ./22-caching/real-case
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

func main() {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		fmt.Println("⏭️  DILEWATI: set REDIS_ADDR untuk versi nyata.")
		fmt.Println("   docker compose -f 22-caching/real-case/docker-compose.yml up -d")
		fmt.Println("   REDIS_ADDR=127.0.0.1:6379 go run ./22-caching/real-case")
		return
	}
	ctx := context.Background()
	rdb := redis.NewClient(&redis.Options{Addr: addr})
	defer rdb.Close()
	if err := rdb.Ping(ctx).Err(); err != nil {
		panic("gagal konek Redis: " + err.Error())
	}

	var dbHits int
	loadDariDB := func(key string) string { // query mahal yang ingin dihindari
		dbHits++
		time.Sleep(20 * time.Millisecond)
		return "nilai:" + key
	}

	// CACHE-ASIDE: cek Redis -> miss -> DB -> isi Redis (TTL).
	get := func(key string) (string, error) {
		v, err := rdb.Get(ctx, key).Result()
		if err == nil {
			return v, nil // HIT
		}
		if !errors.Is(err, redis.Nil) {
			return "", err
		}
		data := loadDariDB(key) // MISS
		if err := rdb.Set(ctx, key, data, 5*time.Minute).Err(); err != nil {
			return "", err
		}
		return data, nil
	}

	_ = rdb.Del(ctx, "user:1").Err() // mulai bersih
	v1, _ := get("user:1")           // miss -> DB
	v2, _ := get("user:1")           // hit -> cache
	fmt.Println("== cache-aside dgn Redis nyata ==")
	fmt.Printf("  nilai=%q, DB disentuh %d kali (get kedua dari cache)\n", v1, dbHits)
	fmt.Println("  konsisten?", v1 == v2)
	ttl := rdb.TTL(ctx, "user:1").Val()
	fmt.Printf("  TTL user:1 ~ %v\n", ttl.Truncate(time.Second))
}
