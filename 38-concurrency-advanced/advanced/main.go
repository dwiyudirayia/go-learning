// Teknik Advanced Modul 38 (concurrency advanced) — contoh runnable + penjelasan detail.
//
// errgroup (batas + cancel), semaphore berbobot, dan sync.Pool. Jalankan:
//
//	go run -race ./38-concurrency-advanced/advanced
package main

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

func main() {
	// =====================================================================
	// 1. errgroup — parallelisme terbatas + cancel-on-first-error
	// =====================================================================
	// SetLimit(n) membatasi goroutine bersamaan. WithContext membatalkan sisa
	// pekerjaan begitu SATU tugas mengembalikan error; Wait() mengembalikan
	// error pertama itu. Jauh lebih ringkas & aman dari WaitGroup manual.
	fmt.Println("== 1. errgroup (limit 2, cancel on error) ==")
	g, ctx := errgroup.WithContext(context.Background())
	g.SetLimit(2)
	var jalan atomic.Int64
	for i := 1; i <= 5; i++ {
		id := i
		g.Go(func() error {
			jalan.Add(1)
			if id == 3 {
				return errors.New("tugas 3 gagal")
			}
			select {
			case <-ctx.Done(): // dibatalkan karena tugas lain error
				return ctx.Err()
			case <-time.After(20 * time.Millisecond):
				return nil
			}
		})
	}
	err := g.Wait()
	fmt.Printf("  error pertama: %v (tugas sempat dimulai: %d)\n", err, jalan.Load())

	// =====================================================================
	// 2. semaphore berbobot — batasi resource yang tidak seragam
	// =====================================================================
	// Beda dari sekadar "N slot": tiap tugas bisa meminta BOBOT berbeda (mis.
	// job berat butuh 2 unit memori). Total bobot aktif dibatasi kapasitas.
	fmt.Println("== 2. weighted semaphore (kapasitas 3) ==")
	sem := semaphore.NewWeighted(3)
	var wg sync.WaitGroup
	bobot := map[string]int64{"ringan": 1, "berat": 2}
	for nama, w := range bobot {
		wg.Add(1)
		go func(nama string, w int64) {
			defer wg.Done()
			_ = sem.Acquire(context.Background(), w) // blok bila kapasitas kurang
			defer sem.Release(w)
			fmt.Printf("  jalankan tugas %q (bobot %d)\n", nama, w)
			time.Sleep(10 * time.Millisecond)
		}(nama, w)
	}
	wg.Wait()

	// =====================================================================
	// 3. sync.Pool — daur ulang objek panas untuk menekan alokasi/GC
	// =====================================================================
	// Pool menyimpan objek yang bisa dipakai ulang antar-goroutine. Cocok untuk
	// buffer sementara. CATATAN: isi pool bisa di-GC kapan saja -> jangan
	// menyimpan state penting di sana; selalu reset sebelum pakai.
	fmt.Println("== 3. sync.Pool ==")
	var pool = sync.Pool{New: func() any { return make([]byte, 0, 1024) }}
	var reused atomic.Int64
	var wg2 sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg2.Add(1)
		go func() {
			defer wg2.Done()
			buf := pool.Get().([]byte)
			buf = append(buf[:0], "kerja"...) // reset panjang lalu pakai
			_ = buf
			reused.Add(1)
			pool.Put(buf[:0]) // kembalikan untuk dipakai ulang
		}()
	}
	wg2.Wait()
	fmt.Printf("  %d operasi memakai buffer dari pool (alokasi ditekan)\n", reused.Load())
}
