package main

import "sync"

// EventStore menyimpan aliran event per-aggregate (append-only, tak pernah di-UPDATE).
// Di produksi: tabel events (aggregate_id, version, type, payload, timestamp) atau
// EventStoreDB/Kafka.
type EventStore struct {
	mu      sync.Mutex
	streams map[string][]Event // aggregateID -> event terurut
	subs    []func(Event)      // pelanggan (untuk memperbarui projeksi/read model)
}

func NewEventStore() *EventStore {
	return &EventStore{streams: make(map[string][]Event)}
}

// Subscribe mendaftarkan handler yang dipanggil untuk SETIAP event baru
// (dipakai read model/projeksi & integrasi — sisi "read" CQRS).
func (s *EventStore) Subscribe(fn func(Event)) {
	s.mu.Lock()
	s.subs = append(s.subs, fn)
	s.mu.Unlock()
}

// Append menambahkan event ke stream & memberi tahu pelanggan.
func (s *EventStore) Append(aggregateID string, events []Event) {
	s.mu.Lock()
	s.streams[aggregateID] = append(s.streams[aggregateID], events...)
	subs := append([]func(Event){}, s.subs...)
	s.mu.Unlock()

	for _, e := range events {
		for _, fn := range subs {
			fn(e)
		}
	}
}

// Load mengembalikan semua event sebuah aggregate (untuk rekonstruksi state).
func (s *EventStore) Load(aggregateID string) []Event {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Event, len(s.streams[aggregateID]))
	copy(out, s.streams[aggregateID])
	return out
}
