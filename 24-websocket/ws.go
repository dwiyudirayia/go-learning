package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/coder/websocket"
)

// handleWS meng-upgrade koneksi HTTP menjadi WebSocket (dua arah, persisten).
// Setiap pesan yang dikirim satu client disiarkan ke semua client (chat sederhana).
func (h *Hub) handleWS(w http.ResponseWriter, r *http.Request) {
	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{InsecureSkipVerify: true})
	if err != nil {
		return
	}
	defer c.CloseNow()

	ch := h.Subscribe()
	defer h.Unsubscribe(ch)

	ctx := r.Context()

	// Goroutine penulis: kirim pesan hasil broadcast ke client ini.
	go func() {
		for msg := range ch {
			if err := c.Write(ctx, websocket.MessageText, msg); err != nil {
				return
			}
		}
	}()

	// Loop pembaca: baca pesan dari client, lalu siarkan ke semua.
	for {
		_, data, err := c.Read(ctx)
		if err != nil {
			return // koneksi ditutup
		}
		h.Broadcast(data)
	}
}

// handleSSE = Server-Sent Events: streaming SATU arah (server -> client) di atas
// HTTP biasa. Lebih sederhana dari WebSocket; cocok untuk notifikasi/feed.
func (h *Hub) handleSSE(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming tidak didukung", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ch := h.Subscribe()
	defer h.Unsubscribe(ch)

	for {
		select {
		case msg := <-ch:
			// Format SSE: "data: <isi>\n\n".
			fmt.Fprintf(w, "data: %s\n\n", msg)
			flusher.Flush()
		case <-r.Context().Done():
			return // client menutup koneksi
		case <-time.After(30 * time.Second):
			// keep-alive comment agar koneksi tak dianggap mati.
			fmt.Fprint(w, ": ping\n\n")
			flusher.Flush()
		}
	}
}
