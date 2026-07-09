// Teknik Advanced Modul 07 (concurrency) — contoh runnable + penjelasan detail.
//
// Semua stdlib. Jalankan (dengan detektor race):
//
//	go run -race ./07-concurrency/advanced
package main

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	// =====================================================================
	// 1. sync/atomic bertipe (Go 1.19+) — hitung aman tanpa mutex
	// =====================================================================
	// atomic.Int64 membungkus operasi atomik; lebih aman & ringkas daripada
	// fungsi atomic.AddInt64 lawas. Cocok untuk counter yang di-update banyak
	// goroutine. (Uji dengan -race untuk buktikan bebas data race.)
	var counter atomic.Int64
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			counter.Add(1) // atomik: tak ada race walau 100 goroutine bersamaan
		}()
	}
	wg.Wait()
	fmt.Println("== 1. atomic.Int64 ==")
	fmt.Println("  counter =", counter.Load()) // selalu 100

	// =====================================================================
	// 2. Semaphore via buffered channel — batasi konkurensi
	// =====================================================================
	// Channel berkapasitas N berperan sebagai "tiket": kirim = ambil tiket
	// (blok bila penuh), terima = kembalikan tiket. Maksimal N goroutine
	// berjalan bersamaan. Pola ini membatasi beban (mis. koneksi DB, API call).
	const maxKonkuren = 3
	sem := make(chan struct{}, maxKonkuren)
	var aktif, puncak atomic.Int64
	var wg2 sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg2.Add(1)
		go func() {
			defer wg2.Done()
			sem <- struct{}{}        // ambil tiket (blok jika sudah 3 berjalan)
			defer func() { <-sem }() // kembalikan tiket saat selesai

			n := aktif.Add(1)
			for { // catat puncak konkurensi secara atomik
				p := puncak.Load()
				if n <= p || puncak.CompareAndSwap(p, n) {
					break
				}
			}
			time.Sleep(20 * time.Millisecond) // simulasi kerja
			aktif.Add(-1)
		}()
	}
	wg2.Wait()
	fmt.Println("== 2. semaphore (buffered channel) ==")
	fmt.Printf("  konkurensi puncak = %d (dibatasi maks %d)\n", puncak.Load(), maxKonkuren)

	// =====================================================================
	// 3. select dengan default — operasi channel NON-BLOCKING
	// =====================================================================
	// Tanpa case yang siap, cabang default langsung dieksekusi (tak menunggu).
	// Berguna untuk "coba kirim/terima, kalau tak bisa lakukan hal lain".
	ch := make(chan int, 1)
	ch <- 1 // buffer penuh
	select {
	case ch <- 2:
		fmt.Println("  terkirim")
	default:
		fmt.Println("== 3. select default ==")
		fmt.Println("  channel penuh -> jalankan default (tak memblokir)")
	}

	// =====================================================================
	// 4. context.AfterFunc (Go 1.21) — jalankan callback saat context batal
	// =====================================================================
	// Mendaftarkan fungsi yang otomatis dipanggil ketika ctx dibatalkan/timeout.
	// Praktis untuk cleanup (tutup koneksi, batalkan pekerjaan hilir).
	ctx, cancel := context.WithCancel(context.Background())
	selesai := make(chan struct{})
	stop := context.AfterFunc(ctx, func() {
		fmt.Println("  [AfterFunc] context dibatalkan -> membersihkan resource")
		close(selesai)
	})
	defer stop() // stop() membatalkan pendaftaran bila belum sempat jalan
	fmt.Println("== 4. context.AfterFunc ==")
	cancel()  // memicu callback di atas
	<-selesai // tunggu callback selesai agar output tampil rapi
}
