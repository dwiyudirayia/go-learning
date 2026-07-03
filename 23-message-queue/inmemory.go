package main

import (
	"context"
	"sync"
	"sync/atomic"
)

// InMemoryBroker = broker in-process untuk DEMO & TEST (tanpa server eksternal).
// Bisa mensimulasikan koneksi PUTUS/SAMBUNG untuk menguji skema resiliensi.
type InMemoryBroker struct {
	mu          sync.Mutex
	subscribers map[string][]*memSub
	rr          map[string]uint64 // penghitung round-robin per (topic|group)
	connected   atomic.Bool
}

type memSub struct {
	group   string
	handler Handler
}

func NewInMemoryBroker() *InMemoryBroker {
	b := &InMemoryBroker{
		subscribers: make(map[string][]*memSub),
		rr:          make(map[string]uint64),
	}
	b.connected.Store(true)
	return b
}

// SetConnected mensimulasikan gangguan jaringan (false = broker "mati").
func (b *InMemoryBroker) SetConnected(v bool) { b.connected.Store(v) }

func (b *InMemoryBroker) Publish(ctx context.Context, topic string, data []byte) error {
	if !b.connected.Load() {
		return ErrDisconnected // memicu retry di sisi publisher
	}

	b.mu.Lock()
	// Pisahkan: subscriber tanpa grup (broadcast) vs grup bernama (load balance).
	var targets []*memSub
	named := map[string][]*memSub{}
	for _, s := range b.subscribers[topic] {
		if s.group == "" {
			targets = append(targets, s) // semua broadcast dapat pesan
		} else {
			named[s.group] = append(named[s.group], s)
		}
	}
	for g, members := range named {
		key := topic + "|" + g
		idx := b.rr[key] % uint64(len(members))
		b.rr[key]++
		targets = append(targets, members[idx]) // satu anggota grup saja
	}
	b.mu.Unlock()

	// Kirim (sinkron -> deterministik untuk test).
	for _, s := range targets {
		_ = s.handler(ctx, data) // di broker nyata, error -> nack/requeue
	}
	return nil
}

func (b *InMemoryBroker) Subscribe(ctx context.Context, topic, group string, handler Handler) error {
	s := &memSub{group: group, handler: handler}
	b.mu.Lock()
	b.subscribers[topic] = append(b.subscribers[topic], s)
	b.mu.Unlock()

	// Berhenti (unsubscribe) saat ctx dibatalkan.
	go func() {
		<-ctx.Done()
		b.mu.Lock()
		defer b.mu.Unlock()
		list := b.subscribers[topic]
		for i, x := range list {
			if x == s {
				b.subscribers[topic] = append(list[:i], list[i+1:]...)
				break
			}
		}
	}()
	return nil
}

func (b *InMemoryBroker) Close() error { return nil }
