package main

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestProcessesAllJobs(t *testing.T) {
	q := NewQueue(4, 100, time.Millisecond)
	defer q.Stop()

	var done int64
	for i := 0; i < 20; i++ {
		q.Enqueue(Job{ID: string(rune('a' + i)), Handler: func() error {
			atomic.AddInt64(&done, 1)
			return nil
		}})
	}
	q.Wait()
	if done != 20 {
		t.Errorf("job diproses = %d; want 20", done)
	}
}

func TestRetrySuccess(t *testing.T) {
	q := NewQueue(1, 10, time.Millisecond)
	defer q.Stop()

	var tries int64
	q.Enqueue(Job{ID: "x", MaxRetries: 5, Handler: func() error {
		if atomic.AddInt64(&tries, 1) < 3 {
			return errors.New("gagal sementara")
		}
		return nil // sukses di percobaan ke-3
	}})
	q.Wait()

	if tries != 3 {
		t.Errorf("percobaan = %d; want 3 (2 gagal + 1 sukses)", tries)
	}
	if q.Failed() != 0 {
		t.Errorf("failed = %d; want 0 (akhirnya sukses)", q.Failed())
	}
}

func TestRetryExhausted(t *testing.T) {
	q := NewQueue(1, 10, time.Millisecond)
	defer q.Stop()

	var tries int64
	q.Enqueue(Job{ID: "selalu-gagal", MaxRetries: 2, Handler: func() error {
		atomic.AddInt64(&tries, 1)
		return errors.New("gagal terus")
	}})
	q.Wait()

	if tries != 3 { // 1 awal + 2 retry
		t.Errorf("percobaan = %d; want 3", tries)
	}
	if q.Failed() != 1 {
		t.Errorf("failed = %d; want 1 (dead letter)", q.Failed())
	}
}

func TestIdempotency(t *testing.T) {
	q := NewQueue(1, 10, time.Millisecond)
	defer q.Stop()

	var hits int64
	handler := func() error { atomic.AddInt64(&hits, 1); return nil }

	q.Enqueue(Job{ID: "sama", Handler: handler})
	q.Wait() // pastikan yang pertama selesai (tercatat processed)
	q.Enqueue(Job{ID: "sama", Handler: handler})
	q.Wait()

	if hits != 1 {
		t.Errorf("handler dipanggil %d kali; want 1 (idempotensi: ID sama tak diproses ulang)", hits)
	}
}

func TestScheduler(t *testing.T) {
	var ticks int64
	ctx, cancel := context.WithTimeout(context.Background(), 125*time.Millisecond)
	defer cancel()

	NewScheduler(50*time.Millisecond, func() { atomic.AddInt64(&ticks, 1) }).Run(ctx)

	// Dalam ~125ms dengan interval 50ms -> sekitar 2 tick.
	if ticks < 1 {
		t.Errorf("ticks = %d; want >= 1", ticks)
	}
}
