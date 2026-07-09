// Teknik Advanced Modul 25 (background jobs) — contoh runnable + penjelasan detail.
//
// Worker POOL + graceful drain (tuntaskan in-flight saat shutdown) + retry.
// Jalankan:
//
//	go run ./25-background-jobs/advanced
package main

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type Job struct {
	ID    int
	Beban time.Duration
}

func main() {
	const jumlahWorker = 3
	jobs := make(chan Job)
	var selesai atomic.Int64

	// =====================================================================
	// 1. WORKER POOL — N goroutine berbagi satu channel job
	// =====================================================================
	// Ukuran pool membatasi konkurensi (mis. jaga koneksi DB/CPU). Semua worker
	// membaca dari channel yang sama; Go mendistribusikan job secara adil.
	var wg sync.WaitGroup
	for w := 1; w <= jumlahWorker; w++ {
		wg.Add(1)
		go func(worker int) {
			defer wg.Done()
			for job := range jobs { // berhenti otomatis saat channel ditutup
				time.Sleep(job.Beban) // simulasi kerja
				selesai.Add(1)
				fmt.Printf("  worker %d menyelesaikan job %d\n", worker, job.ID)
			}
		}(w)
	}

	// =====================================================================
	// 2. SCHEDULER — enqueue job periodik memakai Ticker
	// =====================================================================
	// Ticker memicu pekerjaan berkala (mirip cron). Di produksi tambah jitter
	// agar tak semua instance menembak di detik yang sama.
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		tick := time.NewTicker(15 * time.Millisecond)
		defer tick.Stop()
		id := 0
		for {
			select {
			case <-ctx.Done():
				return
			case <-tick.C:
				id++
				select {
				case jobs <- Job{ID: id, Beban: 20 * time.Millisecond}:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	// Biarkan scheduler mengalirkan job sebentar, lalu picu shutdown.
	time.Sleep(120 * time.Millisecond)

	// =====================================================================
	// 3. GRACEFUL DRAIN — hentikan produksi job baru, tuntaskan yang ada
	// =====================================================================
	// Urutan penting: (a) hentikan scheduler (cancel), (b) TUTUP channel jobs
	// agar worker keluar dari `range` setelah job tersisa habis, (c) tunggu
	// semua worker selesai (wg.Wait). Tak ada job yang terpotong di tengah.
	fmt.Println("== shutdown: drain worker pool ==")
	cancel()    // (a) scheduler berhenti mengirim
	close(jobs) // (b) sinyal ke worker: tak ada job baru
	wg.Wait()   // (c) tunggu in-flight tuntas

	fmt.Printf("  total job selesai (tanpa ada yang terpotong): %d\n", selesai.Load())

	// Catatan: job penting harus PERSIST (DB/queue) agar selamat dari restart
	// (at-least-once) -> rancang idempoten. Job yang selalu gagal -> DLQ setelah
	// batas retry (lihat modul 23), jangan blokir worker selamanya.
}
