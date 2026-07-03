package main

import (
	"context"
	"io"
	"log/slog"
	"net"
	"net/http"
	"testing"
	"time"
)

// TestGracefulShutdown membuktikan: request yang SUDAH masuk tetap selesai
// meski shutdown dipicu di tengah jalan.
func TestGracefulShutdown(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	// Port 0 = OS pilih port bebas otomatis.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	base := "http://" + ln.Addr().String()

	srv := newServer("", logger)
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error, 1)
	go func() { done <- runWithGracefulShutdown(ctx, srv, ln, logger) }()

	time.Sleep(50 * time.Millisecond) // beri waktu server siap

	// Mulai request lambat di background.
	respCode := make(chan int, 1)
	go func() {
		resp, err := http.Get(base + "/slow")
		if err != nil {
			respCode <- -1
			return
		}
		defer resp.Body.Close()
		_, _ = io.ReadAll(resp.Body)
		respCode <- resp.StatusCode
	}()

	time.Sleep(100 * time.Millisecond) // pastikan request /slow sudah in-flight
	cancel()                           // <-- picu shutdown saat request masih berjalan

	// Request in-flight HARUS tetap selesai dengan 200.
	select {
	case code := <-respCode:
		if code != http.StatusOK {
			t.Errorf("request in-flight = %d; want 200 (harus tetap selesai)", code)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timeout menunggu request in-flight")
	}

	// runWithGracefulShutdown harus kembali tanpa error.
	select {
	case err := <-done:
		if err != nil {
			t.Errorf("shutdown mengembalikan error: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timeout menunggu shutdown selesai")
	}

	// Request BARU setelah shutdown harus gagal (server sudah berhenti).
	client := http.Client{Timeout: 500 * time.Millisecond}
	if _, err := client.Get(base + "/"); err == nil {
		t.Error("request setelah shutdown seharusnya gagal")
	}
}
