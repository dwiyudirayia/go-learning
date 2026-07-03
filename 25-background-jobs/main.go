// Jalankan: go run ./25-background-jobs
// Verifikasi otomatis: go test ./25-background-jobs
package main

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"
)

func main() {
	fmt.Println("=== 25 — Background Jobs & Scheduler ===")

	q := NewQueue(3, 100, 10*time.Millisecond)

	// Job normal.
	for i := 1; i <= 3; i++ {
		id := fmt.Sprintf("email-%d", i)
		q.Enqueue(Job{ID: id, MaxRetries: 2, Handler: func() error {
			fmt.Printf("  kirim %s\n", id)
			return nil
		}})
	}

	// Job "flaky": gagal 2x lalu sukses (menguji retry).
	var tries int64
	q.Enqueue(Job{ID: "flaky", MaxRetries: 5, Handler: func() error {
		n := atomic.AddInt64(&tries, 1)
		if n < 3 {
			fmt.Printf("  flaky gagal (percobaan %d)\n", n)
			return errors.New("sementara gagal")
		}
		fmt.Printf("  flaky sukses di percobaan %d\n", n)
		return nil
	}})

	q.Wait()
	fmt.Printf("selesai: sukses=%d gagal=%d totalPanggilanHandler=%d\n",
		q.ProcessedCount(), q.Failed(), q.HandlerHits())
	q.Stop()

	// Scheduler: jalankan task tiap 50ms selama ~160ms (harusnya ~3 kali).
	fmt.Println("scheduler berjalan 3 tick...")
	var ticks int64
	ctx, cancel := context.WithTimeout(context.Background(), 160*time.Millisecond)
	defer cancel()
	NewScheduler(50*time.Millisecond, func() {
		fmt.Printf("  tick #%d\n", atomic.AddInt64(&ticks, 1))
	}).Run(ctx)
	fmt.Printf("total tick: %d\n", ticks)
}
