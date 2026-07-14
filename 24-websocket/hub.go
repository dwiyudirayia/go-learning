// Modul 24 — Real-time: WebSocket & Server-Sent Events (SSE).
// hub.go berisi logika broadcast MURNI (tanpa jaringan) -> mudah di-test.
package main

import "sync"

// 🔍 Analogi besar: WebSocket itu "TELEPON yang tetap tersambung" antara browser & server —
// beda dari HTTP biasa yang seperti kirim SURAT sekali balas lalu putus. Karena saluran terus
// terbuka, server bisa MENDORONG pesan kapan saja (chat masuk, notifikasi) tanpa browser bertanya dulu.
//
// 🔍 Analogi Hub: Hub itu OPERATOR PENYIARAN. Tiap browser yang tersambung = satu pendengar yang
// "Subscribe" (dapat channel sendiri). Saat ada pesan, Hub "Broadcast" ke semua pendengar sekaligus.
// RWMutex = kunci khusus: banyak boleh MEMBACA daftar bersamaan, tapi MENGUBAH daftar harus giliran.
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

// 🔍 Analogi: pola "select-default" di sini itu prinsip "PERTUNJUKAN HARUS LANJUT". Kalau satu
// pendengar terlalu lambat (kotak pesannya penuh), penyiar TIDAK berhenti menunggunya — ia
// lewati & lanjut ke pendengar lain. Mencegah satu klien lemot menyandera seluruh siaran.
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
