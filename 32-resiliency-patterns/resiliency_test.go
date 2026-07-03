package main

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestCircuitBreaker(t *testing.T) {
	cb := NewCircuitBreaker(3, 100*time.Millisecond)
	// Clock yang bisa dikontrol untuk menguji timeout.
	now := time.Now()
	cb.now = func() time.Time { return now }

	fail := func() error { return errors.New("gagal") }
	callCount := 0
	tracked := func() error { callCount++; return fail() }

	// 3 kegagalan -> circuit OPEN.
	for i := 0; i < 3; i++ {
		_ = cb.Execute(tracked)
	}
	if cb.State() != StateOpen {
		t.Fatalf("state = %s; want open", cb.State())
	}

	// Saat OPEN, fn TIDAK dipanggil (fail fast).
	before := callCount
	if err := cb.Execute(tracked); !errors.Is(err, ErrCircuitOpen) {
		t.Errorf("err = %v; want ErrCircuitOpen", err)
	}
	if callCount != before {
		t.Error("fn tidak boleh dipanggil saat circuit OPEN")
	}

	// Majukan waktu melewati timeout -> half-open -> sukses -> CLOSED.
	now = now.Add(200 * time.Millisecond)
	if err := cb.Execute(func() error { return nil }); err != nil {
		t.Errorf("half-open sukses harusnya nil: %v", err)
	}
	if cb.State() != StateClosed {
		t.Errorf("state setelah sukses = %s; want closed", cb.State())
	}
}

func TestBulkheadTolakSaatPenuh(t *testing.T) {
	bh := NewBulkhead(2)

	// Isi 2 slot dengan operasi yang menahan.
	block := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = bh.Execute(context.Background(), func() error { <-block; return nil })
		}()
	}
	// Tunggu keduanya masuk.
	deadline := time.Now().Add(time.Second)
	for bh.InFlight() < 2 && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}

	// Slot penuh -> TryExecute ditolak.
	if err := bh.TryExecute(func() error { return nil }); !errors.Is(err, ErrBulkheadFull) {
		t.Errorf("err = %v; want ErrBulkheadFull", err)
	}

	close(block)
	wg.Wait()

	// Setelah kosong, boleh lagi.
	if err := bh.TryExecute(func() error { return nil }); err != nil {
		t.Errorf("setelah kosong harusnya lolos: %v", err)
	}
}

func TestDistributedLock(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis: %v", err)
	}
	defer mr.Close()
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	lock := NewDistributedLock(rdb)
	ctx := context.Background()

	tokenA, okA, _ := lock.Acquire(ctx, "job", time.Minute)
	if !okA {
		t.Fatal("A gagal ambil lock")
	}
	// Instance kedua ditolak.
	if _, okB, _ := lock.Acquire(ctx, "job", time.Minute); okB {
		t.Error("B seharusnya ditolak (A masih pegang)")
	}
	// Release oleh A, lalu B berhasil.
	if err := lock.Release(ctx, "job", tokenA); err != nil {
		t.Fatalf("release: %v", err)
	}
	if _, okB2, _ := lock.Acquire(ctx, "job", time.Minute); !okB2 {
		t.Error("B seharusnya berhasil setelah A release")
	}
}

func TestLockReleaseTokenSalah(t *testing.T) {
	mr, _ := miniredis.Run()
	defer mr.Close()
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()
	lock := NewDistributedLock(rdb)
	ctx := context.Background()

	_, _, _ = lock.Acquire(ctx, "job", time.Minute)
	// Release dengan token salah TIDAK boleh melepas lock orang lain.
	_ = lock.Release(ctx, "job", "token-salah")
	if _, ok, _ := lock.Acquire(ctx, "job", time.Minute); ok {
		t.Error("lock tak boleh terlepas oleh token salah")
	}
}
