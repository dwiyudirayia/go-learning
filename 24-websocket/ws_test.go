package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
)

// Integration test: dua client WebSocket sungguhan lewat httptest.Server.
// Pesan dari client 1 harus diterima client 2 (broadcast).
func TestWebSocketBroadcast(t *testing.T) {
	hub := NewHub()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /ws", hub.handleWS)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws"

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	c1, _, err := websocket.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatalf("dial c1: %v", err)
	}
	defer c1.CloseNow()

	c2, _, err := websocket.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatalf("dial c2: %v", err)
	}
	defer c2.CloseNow()

	// Tunggu kedua client terdaftar di hub.
	deadline := time.Now().Add(2 * time.Second)
	for hub.Count() < 2 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}

	// Client 1 kirim pesan.
	if err := c1.Write(ctx, websocket.MessageText, []byte("halo semua")); err != nil {
		t.Fatalf("write c1: %v", err)
	}

	// Client 2 harus menerimanya (broadcast).
	_, data, err := c2.Read(ctx)
	if err != nil {
		t.Fatalf("read c2: %v", err)
	}
	if string(data) != "halo semua" {
		t.Errorf("c2 terima %q; want 'halo semua'", data)
	}
}
