package main

import (
	"testing"
	"time"
)

// Test logika Hub secara langsung (tanpa jaringan) — cepat & deterministik.
func TestHubBroadcast(t *testing.T) {
	h := NewHub()
	a := h.Subscribe()
	b := h.Subscribe()

	if h.Count() != 2 {
		t.Fatalf("Count = %d; want 2", h.Count())
	}

	h.Broadcast([]byte("halo"))

	for name, ch := range map[string]chan []byte{"a": a, "b": b} {
		select {
		case msg := <-ch:
			if string(msg) != "halo" {
				t.Errorf("subscriber %s dapat %q; want halo", name, msg)
			}
		case <-time.After(time.Second):
			t.Errorf("subscriber %s tidak menerima broadcast", name)
		}
	}
}

func TestHubUnsubscribe(t *testing.T) {
	h := NewHub()
	ch := h.Subscribe()
	h.Unsubscribe(ch)

	if h.Count() != 0 {
		t.Errorf("Count setelah unsubscribe = %d; want 0", h.Count())
	}
	// Channel harus tertutup.
	if _, open := <-ch; open {
		t.Error("channel seharusnya tertutup setelah Unsubscribe")
	}
}
