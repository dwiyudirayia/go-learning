// Modul 32 — Resiliency Patterns: circuit breaker, bulkhead, distributed lock.
package main

import (
	"errors"
	"sync"
	"time"
)

// 🔍 Analogi 3 keadaan (perluasan analogi sekring):
//   - CLOSED    = sekring NYAMBUNG, listrik mengalir normal (request diteruskan).
//   - OPEN      = sekring PUTUS setelah korslet berulang; kita berhenti mencoba & langsung tolak
//                 (fail fast) — mencegah "menendang orang yang sudah jatuh" & memberi waktu pulih.
//   - HALF-OPEN = "colok satu alat untuk tes" setelah jeda. Kalau nyala (sukses) -> sambung penuh;
//                 kalau korslet lagi -> putus lagi. Ini mencegah membanjiri service yang belum sembuh.
// Kenapa penting? Tanpa breaker, ribuan request menumpuk menunggu service mati -> efek domino
// yang menjatuhkan seluruh sistem (cascading failure). Breaker memutus rantai kejatuhan itu.

// CIRCUIT BREAKER mencegah aplikasi terus memanggil service yang sedang bermasalah
// (fail fast). Seperti sekring listrik: bila terlalu banyak gagal, "putus" (open)
// sementara agar service downstream sempat pulih, lalu "coba lagi" (half-open).
//
// State:  CLOSED --(gagal >= threshold)--> OPEN --(setelah timeout)--> HALF-OPEN
//
//	HALF-OPEN --(sukses)--> CLOSED   |   HALF-OPEN --(gagal)--> OPEN
type State int

const (
	StateClosed   State = iota // normal, request diteruskan
	StateOpen                  // memutus, request langsung ditolak (fail fast)
	StateHalfOpen              // mencoba 1 request untuk cek pemulihan
)

func (s State) String() string {
	return [...]string{"closed", "open", "half-open"}[s]
}

var ErrCircuitOpen = errors.New("circuit breaker terbuka (fail fast)")

type CircuitBreaker struct {
	mu        sync.Mutex
	state     State
	failures  int
	threshold int           // jumlah gagal berturut sebelum membuka
	timeout   time.Duration // berapa lama tetap terbuka sebelum half-open
	openedAt  time.Time
	now       func() time.Time // di-inject untuk test
}

func NewCircuitBreaker(threshold int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{threshold: threshold, timeout: timeout, now: time.Now}
}

// Execute menjalankan fn dengan proteksi circuit breaker.
func (cb *CircuitBreaker) Execute(fn func() error) error {
	cb.mu.Lock()
	// Bila terbuka: cek apakah sudah waktunya mencoba lagi (half-open).
	if cb.state == StateOpen {
		if cb.now().Sub(cb.openedAt) < cb.timeout {
			cb.mu.Unlock()
			return ErrCircuitOpen // masih dalam masa "putus" -> tolak cepat
		}
		cb.state = StateHalfOpen // beri satu kesempatan mencoba
	}
	cb.mu.Unlock()

	err := fn()

	cb.mu.Lock()
	defer cb.mu.Unlock()
	if err != nil {
		cb.failures++
		if cb.failures >= cb.threshold || cb.state == StateHalfOpen {
			cb.state = StateOpen // buka (atau tetap buka bila trial half-open gagal)
			cb.openedAt = cb.now()
		}
		return err
	}
	// Sukses -> reset ke normal.
	cb.failures = 0
	cb.state = StateClosed
	return nil
}

func (cb *CircuitBreaker) State() State {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}
