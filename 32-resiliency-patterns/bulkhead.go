package main

import (
	"context"
	"errors"
)

// BULKHEAD membatasi jumlah operasi bersamaan (seperti sekat kedap air di kapal:
// bila satu bagian bocor, tak menenggelamkan seluruh kapal). Mencegah satu
// dependency lambat menghabiskan semua goroutine/koneksi aplikasi.
//
// Implementasi: semaphore berbasis channel berkapasitas N.
type Bulkhead struct {
	sem chan struct{}
}

var ErrBulkheadFull = errors.New("bulkhead penuh (kapasitas habis)")

func NewBulkhead(maxConcurrent int) *Bulkhead {
	return &Bulkhead{sem: make(chan struct{}, maxConcurrent)}
}

// Execute menunggu slot tersedia (atau ctx dibatalkan), lalu menjalankan fn.
func (b *Bulkhead) Execute(ctx context.Context, fn func() error) error {
	select {
	case b.sem <- struct{}{}: // ambil slot
		defer func() { <-b.sem }() // lepas slot
		return fn()
	case <-ctx.Done():
		return ctx.Err()
	}
}

// TryExecute langsung menolak (ErrBulkheadFull) bila kapasitas penuh — tanpa menunggu.
func (b *Bulkhead) TryExecute(fn func() error) error {
	select {
	case b.sem <- struct{}{}:
		defer func() { <-b.sem }()
		return fn()
	default:
		return ErrBulkheadFull
	}
}

// InFlight mengembalikan jumlah operasi yang sedang berjalan.
func (b *Bulkhead) InFlight() int { return len(b.sem) }
