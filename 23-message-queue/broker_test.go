package main

import (
	"context"
	"io"
	"log/slog"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestInMemoryPubSub(t *testing.T) {
	b := NewInMemoryBroker()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	got := make(chan string, 1)
	_ = b.Subscribe(ctx, "topik", "", func(_ context.Context, data []byte) error {
		got <- string(data)
		return nil
	})

	if err := b.Publish(ctx, "topik", []byte("halo")); err != nil {
		t.Fatalf("publish: %v", err)
	}
	select {
	case m := <-got:
		if m != "halo" {
			t.Errorf("terima %q; want halo", m)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout")
	}
}

func TestInMemoryQueueGroupLoadBalancing(t *testing.T) {
	b := NewInMemoryBroker()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var total int64
	for i := 0; i < 3; i++ { // 3 worker dalam grup sama
		_ = b.Subscribe(ctx, "jobs", "workers", func(_ context.Context, _ []byte) error {
			atomic.AddInt64(&total, 1)
			return nil
		})
	}
	for i := 0; i < 30; i++ {
		_ = b.Publish(ctx, "jobs", []byte("x"))
	}
	// Tiap pesan tepat SEKALI (load balancing), bukan 30*3.
	if total != 30 {
		t.Errorf("total diproses = %d; want 30", total)
	}
}

// Inti modul: publish tahan gangguan koneksi (retry sampai broker pulih).
func TestPublishRetryReconnect(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil)) // senyap
	b := NewInMemoryBroker()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	got := make(chan string, 1)
	_ = b.Subscribe(ctx, "penting", "", func(_ context.Context, data []byte) error {
		got <- string(data)
		return nil
	})

	b.SetConnected(false) // broker "mati"

	done := make(chan error, 1)
	go func() { done <- PublishWithRetry(ctx, b, "penting", []byte("data"), logger) }()

	// Saat mati, pesan belum sampai.
	select {
	case <-got:
		t.Fatal("pesan tidak boleh terkirim saat broker mati")
	case <-time.After(150 * time.Millisecond):
	}

	b.SetConnected(true) // broker pulih -> retry berikutnya sukses

	select {
	case m := <-got:
		if m != "data" {
			t.Errorf("terima %q; want data", m)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("pesan tidak sampai setelah broker pulih")
	}
	if err := <-done; err != nil {
		t.Errorf("PublishWithRetry: %v", err)
	}
}

func TestNATSBroker(t *testing.T) {
	ns, err := StartEmbeddedNATS()
	if err != nil {
		t.Fatalf("nats: %v", err)
	}
	defer ns.Shutdown()

	b, err := NewNATSBroker(ns.ClientURL(), nil)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer b.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(10)
	var total int64
	for i := 0; i < 2; i++ {
		_ = b.Subscribe(ctx, "ev", "grp", func(_ context.Context, _ []byte) error {
			atomic.AddInt64(&total, 1)
			wg.Done()
			return nil
		})
	}
	time.Sleep(100 * time.Millisecond)

	for i := 0; i < 10; i++ {
		if err := b.Publish(ctx, "ev", []byte("m")); err != nil {
			t.Fatalf("publish: %v", err)
		}
	}

	doneCh := make(chan struct{})
	go func() { wg.Wait(); close(doneCh) }()
	select {
	case <-doneCh:
	case <-time.After(3 * time.Second):
		t.Fatalf("timeout; diproses %d dari 10", atomic.LoadInt64(&total))
	}
	if total != 10 {
		t.Errorf("total = %d; want 10 (queue group)", total)
	}
}
