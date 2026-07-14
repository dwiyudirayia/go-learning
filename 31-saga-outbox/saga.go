// Modul 31 — Saga & Outbox: transaksi terdistribusi tanpa two-phase commit.
//
// SAGA: rangkaian langkah lokal; bila satu langkah gagal, jalankan aksi
// KOMPENSASI (undo) untuk langkah-langkah yang sudah sukses — dalam urutan
// TERBALIK. Ini cara menjaga konsistensi lintas service tanpa distributed lock.
package main

import (
	"context"
	"fmt"
)

// Step = satu langkah saga. Compensate boleh nil bila langkah tak perlu di-undo.
type Step struct {
	Name       string
	Action     func(ctx context.Context) error
	Compensate func(ctx context.Context) error
}

// 🔍 Analogi besar SAGA: bayangkan MEMESAN PAKET LIBURAN via beberapa vendor terpisah: pesan
// tiket pesawat -> booking hotel -> sewa mobil. Tak ada "satu tombol batal" ajaib yang mencakup
// ketiganya (itu two-phase commit, mahal & rapuh di sistem terdistribusi). Kalau sewa mobil GAGAL,
// saga menjalankan "KOMPENSASI" mundur: batalkan hotel, lalu batalkan pesawat — undo urutan terbalik,
// seperti melepas tumpukan piring dari atas. Hasil akhir: konsisten (semua jadi, atau semua dibatalkan).

// Saga (pola ORKESTRASI): satu koordinator menjalankan langkah berurutan.
type Saga struct {
	steps []Step
	log   []string // jejak eksekusi (untuk demo/test)
}

func NewSaga() *Saga { return &Saga{} }

func (s *Saga) AddStep(step Step) *Saga {
	s.steps = append(s.steps, step)
	return s
}

// Execute menjalankan semua langkah. Bila ada yang gagal, langkah yang sudah
// sukses di-KOMPENSASI dalam urutan terbalik, lalu error dikembalikan.
func (s *Saga) Execute(ctx context.Context) error {
	var completed []Step

	for _, step := range s.steps {
		if err := step.Action(ctx); err != nil {
			s.log = append(s.log, "GAGAL:"+step.Name)
			// Kompensasi mundur: undo yang terakhir sukses lebih dulu.
			for i := len(completed) - 1; i >= 0; i-- {
				c := completed[i]
				if c.Compensate != nil {
					s.log = append(s.log, "KOMPENSASI:"+c.Name)
					if cerr := c.Compensate(ctx); cerr != nil {
						// Kompensasi gagal = butuh intervensi manual (alert!).
						s.log = append(s.log, "KOMPENSASI_GAGAL:"+c.Name)
					}
				}
			}
			return fmt.Errorf("saga gagal di langkah %q: %w", step.Name, err)
		}
		s.log = append(s.log, "OK:"+step.Name)
		completed = append(completed, step)
	}
	return nil
}

// Log mengembalikan jejak eksekusi (mis. ["OK:reserve", "OK:pay", "GAGAL:ship", "KOMPENSASI:pay", "KOMPENSASI:reserve"]).
func (s *Saga) Log() []string { return s.log }
