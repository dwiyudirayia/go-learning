// Teknik Advanced Modul 22 (caching) — contoh runnable + penjelasan detail.
//
// Cache-aside di Redis (miniredis in-memory) + singleflight untuk mencegah
// cache stampede + TTL berjitter. Jalankan:
//
//	go run ./22-caching/advanced
package main

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"
)

func main() {
	// miniredis = server Redis in-memory untuk demo/test tanpa infra nyata.
	mr, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	defer mr.Close()
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	ctx := context.Background()

	var dbHits atomic.Int64 // menghitung berapa kali "DB" benar-benar disentuh

	// loadDariDB meniru query mahal (lambat). Kita ingin ini dipanggil
	// SESEDIKIT mungkin berkat cache + singleflight.
	loadDariDB := func(key string) string {
		dbHits.Add(1)
		time.Sleep(30 * time.Millisecond) // simulasi latensi query
		return "nilai:" + key
	}

	var sf singleflight.Group // mengecilkan panggilan duplikat concurrent -> 1

	// =====================================================================
	// Pola CACHE-ASIDE: cek cache -> miss -> load DB -> isi cache.
	// singleflight membungkus bagian miss agar 100 request untuk key yang sama
	// TIDAK membanjiri DB (thundering herd) — hanya 1 yang query, sisanya nebeng.
	// =====================================================================
	getCacheAside := func(key string) (string, error) {
		if v, err := rdb.Get(ctx, key).Result(); err == nil {
			return v, nil // HIT
		} else if !errors.Is(err, redis.Nil) {
			return "", err // error asli (bukan sekadar "tak ada")
		}
		// MISS -> lewati singleflight
		v, err, _ := sf.Do(key, func() (any, error) {
			data := loadDariDB(key)
			// TTL + jitter: hindari banyak key kedaluwarsa SEREMPAK.
			ttl := 5*time.Minute + time.Duration(rand.Intn(30))*time.Second
			if err := rdb.Set(ctx, key, data, ttl).Err(); err != nil {
				return nil, err
			}
			return data, nil
		})
		if err != nil {
			return "", err
		}
		return v.(string), nil
	}

	// --- 50 goroutine menyerbu key yang SAMA secara bersamaan ---
	const n = 50
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = getCacheAside("user:1")
		}()
	}
	wg.Wait()

	fmt.Println("== cache-aside + singleflight ==")
	fmt.Printf("  %d request konkuren untuk key sama -> DB disentuh %d kali\n", n, dbHits.Load())

	// Request berikutnya: sudah ada di cache -> DB tidak disentuh lagi.
	_, _ = getCacheAside("user:1")
	fmt.Printf("  request lagi (cache hit) -> total DB hit tetap %d\n", dbHits.Load())

	ttl := rdb.TTL(ctx, "user:1").Val()
	fmt.Printf("  TTL key user:1 ~ %v (base 5m + jitter)\n", ttl.Truncate(time.Second))
}
