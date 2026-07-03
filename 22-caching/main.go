// Jalankan: go run ./22-caching
//
// Demo ini memakai miniredis (Redis in-memory) agar jalan TANPA server Redis.
// Di produksi, ganti ke: redis.NewClient(&redis.Options{Addr: "localhost:6379"}).
//
// Verifikasi otomatis: go test ./22-caching
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func main() {
	fmt.Println("=== 22 — Caching (Redis) ===")

	// miniredis = Redis palsu in-memory (untuk demo/test).
	mr, err := miniredis.Run()
	if err != nil {
		log.Fatal(err)
	}
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	ctx := context.Background()
	svc := NewProductService(rdb)

	// Panggilan 1: cache miss -> baca DB.
	start := time.Now()
	p, fromCache, _ := svc.Get(ctx, 1)
	fmt.Printf("call 1: %s Rp%d  fromCache=%t  (%.0fms)\n", p.Name, p.Price, fromCache, ms(start))

	// Panggilan 2: cache hit -> cepat, tidak menyentuh DB.
	start = time.Now()
	p, fromCache, _ = svc.Get(ctx, 1)
	fmt.Printf("call 2: %s Rp%d  fromCache=%t  (%.0fms)\n", p.Name, p.Price, fromCache, ms(start))

	fmt.Printf("total akses DB sejauh ini: %d (harusnya 1)\n", svc.DBHits())

	// Update harga -> cache di-invalidate.
	_ = svc.UpdatePrice(ctx, 1, 300000)
	p, fromCache, _ = svc.Get(ctx, 1) // miss lagi -> ambil data baru
	fmt.Printf("setelah update: %s Rp%d fromCache=%t (data segar)\n", p.Name, p.Price, fromCache)

	fmt.Printf("total akses DB: %d\n", svc.DBHits())
}

func ms(start time.Time) float64 { return float64(time.Since(start).Microseconds()) / 1000 }
