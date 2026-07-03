// Modul 24 — Real-time: WebSocket & Server-Sent Events (SSE).
// hub.go berisi logika broadcast MURNI (tanpa jaringan) -> mudah di-test.
package main

import "sync"

// Hub mengelola daftar subscriber dan menyiarkan pesan ke semuanya.
// Ini pola "pub/sub in-process" — jantung fitur chat/notifikasi realtime.
type Hub struct {
	mu          sync.RWMutex
	subscribers map[chan []byte]struct{}
}

func NewHub() *Hub {
	return &Hub{subscribers: make(map[chan []byte]struct{})}
}

// Subscribe mendaftarkan subscriber baru & mengembalikan channel untuk menerima pesan.
func (h *Hub) Subscribe() chan []byte {
	ch := make(chan []byte, 16) // buffer agar broadcast tak memblok
	h.mu.Lock()
	h.subscribers[ch] = struct{}{}
	h.mu.Unlock()
	return ch
}

// Unsubscribe menghapus subscriber (mis. saat koneksi WebSocket putus).
func (h *Hub) Unsubscribe(ch chan []byte) {
	h.mu.Lock()
	if _, ok := h.subscribers[ch]; ok {
		delete(h.subscribers, ch)
		close(ch)
	}
	h.mu.Unlock()
}

// Broadcast mengirim pesan ke SEMUA subscriber. Non-blocking: subscriber yang
// lambat (buffer penuh) dilewati agar tidak menahan yang lain.
func (h *Hub) Broadcast(msg []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for ch := range h.subscribers {
		select {
		case ch <- msg:
		default: // subscriber lambat -> lewati pesan ini
		}
	}
}

func (h *Hub) Count() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.subscribers)
}
