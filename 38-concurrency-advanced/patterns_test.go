package main

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestFetchAllSukses(t *testing.T) {
	res, err := FetchAll(context.Background(), []int{1, 2, 3}, func(_ context.Context, id int) (string, error) {
		return string(rune('a' + id)), nil
	})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(res) != 3 {
		t.Errorf("hasil = %d; want 3", len(res))
	}
}

func TestFetchAllErrorMembatalkan(t *testing.T) {
	var started int64
	_, err := FetchAll(context.Background(), []int{1, 2, 3, 4, 5}, func(ctx context.Context, id int) (string, error) {
		atomic.AddInt64(&started, 1)
		if id == 1 {
			return "", errors.New("gagal cepat")
		}
		// Goroutine lain harus terbatalkan lewat ctx.
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(time.Second):
			return "lambat", nil
		}
	})
	if err == nil {
		t.Fatal("mengharapkan error")
	}
}

func TestSingleflightDedupe(t *testing.T) {
	loader := &Loader{}
	var wg sync.WaitGroup
	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			v, err := loader.Get("k", func() (string, error) {
				time.Sleep(10 * time.Millisecond)
				return "value", nil
			})
			if err != nil || v != "value" {
				t.Errorf("get -> %q err=%v", v, err)
			}
		}()
	}
	wg.Wait()
	// 200 panggilan bersamaan key sama -> loader asli dipanggil jauh lebih sedikit.
	if loader.Calls() >= 200 {
		t.Errorf("loader dipanggil %d kali; harusnya jauh < 200 (dedupe)", loader.Calls())
	}
	if loader.Calls() == 0 {
		t.Error("loader tak pernah dipanggil?")
	}
}

func TestRunLimitedMembatasiKonkurensi(t *testing.T) {
	var current, peak int64
	tasks := make([]func(), 10)
	for i := range tasks {
		tasks[i] = func() {
			c := atomic.AddInt64(&current, 1)
			for {
				p := atomic.LoadInt64(&peak)
				if c <= p || atomic.CompareAndSwapInt64(&peak, p, c) {
					break
				}
			}
			time.Sleep(10 * time.Millisecond)
			atomic.AddInt64(&current, -1)
		}
	}
	if err := RunLimited(context.Background(), tasks, 3); err != nil {
		t.Fatalf("run: %v", err)
	}
	if peak > 3 {
		t.Errorf("puncak konkurensi = %d; want <= 3", peak)
	}
}

func TestPoolKorektnes(t *testing.T) {
	// Pastikan buffer yang dipakai ulang tak membawa data lama (Reset bekerja).
	if got := FormatGreeting([]string{"A"}); got != "Halo A" {
		t.Errorf("got %q; want 'Halo A'", got)
	}
	if got := FormatGreeting([]string{"B", "C"}); got != "Halo B, Halo C" {
		t.Errorf("got %q; want 'Halo B, Halo C'", got)
	}
}
