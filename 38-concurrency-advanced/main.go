// Jalankan:        go run ./38-concurrency-advanced
// Dengan race det: go run -race ./38-concurrency-advanced
// Verifikasi:      go test -race ./38-concurrency-advanced
package main

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	fmt.Println("=== 38 — Concurrency Advanced ===")

	// errgroup
	res, err := FetchAll(context.Background(), []int{1, 2, 3}, func(_ context.Context, id int) (string, error) {
		return fmt.Sprintf("data-%d", id), nil
	})
	fmt.Printf("errgroup FetchAll -> %v err=%v\n", res, err)

	// errgroup dengan error -> membatalkan sisanya
	_, err = FetchAll(context.Background(), []int{1, 2, 3}, func(_ context.Context, id int) (string, error) {
		if id == 2 {
			return "", errors.New("id 2 gagal")
		}
		time.Sleep(50 * time.Millisecond)
		return "ok", nil
	})
	fmt.Printf("errgroup dengan error -> %v\n", err)

	// singleflight: 100 goroutine, key sama -> loader 1x
	loader := &Loader{}
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = loader.Get("user:1", func() (string, error) {
				time.Sleep(20 * time.Millisecond) // simulasi query lambat
				return "Ana", nil
			})
		}()
	}
	wg.Wait()
	fmt.Printf("singleflight: 100 request key sama -> loader dipanggil %d kali\n", loader.Calls())

	// semaphore: batasi 2 bersamaan
	var peak, current int64
	tasks := make([]func(), 6)
	for i := range tasks {
		tasks[i] = func() {
			c := atomic.AddInt64(&current, 1)
			for {
				p := atomic.LoadInt64(&peak)
				if c <= p || atomic.CompareAndSwapInt64(&peak, p, c) {
					break
				}
			}
			time.Sleep(20 * time.Millisecond)
			atomic.AddInt64(&current, -1)
		}
	}
	_ = RunLimited(context.Background(), tasks, 2)
	fmt.Printf("semaphore: 6 task, maks 2 -> puncak bersamaan = %d\n", peak)

	// sync.Pool
	fmt.Printf("sync.Pool: %s\n", FormatGreeting([]string{"Ana", "Budi"}))
}
