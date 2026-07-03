package main

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func setup(t *testing.T) (*ProductService, *miniredis.Miniredis) {
	t.Helper()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis: %v", err)
	}
	t.Cleanup(mr.Close)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { rdb.Close() })
	return NewProductService(rdb), mr
}

func TestCacheHit(t *testing.T) {
	svc, _ := setup(t)
	ctx := context.Background()

	// Miss pertama.
	if _, fromCache, err := svc.Get(ctx, 1); err != nil || fromCache {
		t.Fatalf("call 1: fromCache=%v err=%v; want fromCache=false", fromCache, err)
	}
	// Hit kedua.
	if _, fromCache, _ := svc.Get(ctx, 1); !fromCache {
		t.Error("call 2 seharusnya cache HIT")
	}
	// DB hanya disentuh SEKALI meski Get dipanggil dua kali.
	if svc.DBHits() != 1 {
		t.Errorf("DBHits = %d; want 1 (cache mencegah akses DB kedua)", svc.DBHits())
	}
}

func TestInvalidation(t *testing.T) {
	svc, _ := setup(t)
	ctx := context.Background()

	_, _, _ = svc.Get(ctx, 1) // isi cache (harga awal 250000)
	if err := svc.UpdatePrice(ctx, 1, 300000); err != nil {
		t.Fatalf("update: %v", err)
	}
	// Setelah invalidate, Get harus miss & mengambil harga baru.
	p, fromCache, _ := svc.Get(ctx, 1)
	if fromCache {
		t.Error("setelah invalidate seharusnya cache MISS")
	}
	if p.Price != 300000 {
		t.Errorf("harga = %d; want 300000 (data segar setelah invalidate)", p.Price)
	}
}

func TestTTLExpiry(t *testing.T) {
	svc, mr := setup(t)
	ctx := context.Background()

	_, _, _ = svc.Get(ctx, 1) // isi cache dgn TTL 30 detik
	// miniredis bisa "memajukan waktu" untuk menguji TTL tanpa menunggu sungguhan.
	mr.FastForward(31 * time.Second)

	if _, fromCache, _ := svc.Get(ctx, 1); fromCache {
		t.Error("setelah TTL lewat seharusnya cache MISS")
	}
}
