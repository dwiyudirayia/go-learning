// Jalankan: go run ./32-resiliency-patterns
// Verifikasi otomatis: go test ./32-resiliency-patterns
package main

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func main() {
	fmt.Println("=== 32 — Resiliency Patterns ===")
	demoCircuitBreaker()
	demoBulkhead()
	demoLock()
}

func demoCircuitBreaker() {
	fmt.Println("\n-- Circuit Breaker --")
	cb := NewCircuitBreaker(3, 100*time.Millisecond)
	flaky := func() error { return errors.New("service down") }

	for i := 1; i <= 5; i++ {
		err := cb.Execute(flaky)
		fmt.Printf("  call %d: state=%s err=%v\n", i, cb.State(), err)
	}
	// Setelah 3 gagal -> OPEN -> call 4 & 5 di-tolak cepat (ErrCircuitOpen).
}

func demoBulkhead() {
	fmt.Println("\n-- Bulkhead (maks 2 bersamaan) --")
	bh := NewBulkhead(2)
	var wg sync.WaitGroup
	var rejected int
	var mu sync.Mutex

	for i := 1; i <= 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			err := bh.TryExecute(func() error {
				time.Sleep(50 * time.Millisecond)
				return nil
			})
			if errors.Is(err, ErrBulkheadFull) {
				mu.Lock()
				rejected++
				mu.Unlock()
			}
		}(i)
	}
	wg.Wait()
	fmt.Printf("  5 request, maks 2 bersamaan -> %d ditolak (fail fast)\n", rejected)
}

func demoLock() {
	fmt.Println("\n-- Distributed Lock (Redis) --")
	mr, _ := miniredis.Run()
	defer mr.Close()
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	lock := NewDistributedLock(rdb)
	ctx := context.Background()

	token, ok, _ := lock.Acquire(ctx, "cron:cleanup", 5*time.Second)
	fmt.Printf("  instance A ambil lock: %t\n", ok)

	_, ok2, _ := lock.Acquire(ctx, "cron:cleanup", 5*time.Second)
	fmt.Printf("  instance B ambil lock: %t (ditolak, A masih pegang)\n", ok2)

	_ = lock.Release(ctx, "cron:cleanup", token)
	_, ok3, _ := lock.Acquire(ctx, "cron:cleanup", 5*time.Second)
	fmt.Printf("  setelah A release, instance B ambil lock: %t\n", ok3)
}
