// Package main untuk modul 07 — Concurrency.
// Jalankan:        go run ./07-concurrency
// Dengan race det: go run -race ./07-concurrency
package main

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	fmt.Println("=== 07 — Concurrency ===")
	goroutineWaitGroup()
	channelDasar()
	selectDanTimeout()
	syncMutexOnceAtomic()
	contextCancellation()
	workerPool()
	pipeline()
}

// ------------------------------------------------------------------
// 1. Goroutine + WaitGroup (agar main menunggu)
// ------------------------------------------------------------------
// 🔍 Analogi besar modul ini: bayangkan kamu KOKI KEPALA di dapur.
//   - goroutine = pegawai tambahan yang kamu suruh kerja SENDIRI, paralel denganmu.
//     Sangat murah — Go sanggup ribuan sekaligus (beda dgn "thread" OS yang berat).
//   - channel  = LOKET/JENDELA penyerahan piring antar pegawai. Cara aman berbagi kerja.
//   - Motto Go: "Jangan berbagi memori lalu dikunci; berbagilah lewat channel."

// 🔍 Analogi: WaitGroup itu seperti ABSENSI. wg.Add(1) = "ada 1 pegawai berangkat",
// wg.Done() = "1 pegawai pulang", wg.Wait() = "koki kepala menunggu di pintu sampai
// SEMUA pegawai pulang". Tanpa ini, main() bisa selesai duluan & pegawai ditinggal.
func goroutineWaitGroup() {
	fmt.Println("\n-- Goroutine + WaitGroup --")

	var wg sync.WaitGroup
	for i := 1; i <= 3; i++ {
		wg.Add(1) // daftarkan 1 pekerjaan sebelum go
		go func(id int) {
			defer wg.Done() // tandai selesai
			// simulasi kerja singkat
			time.Sleep(time.Duration(id) * 10 * time.Millisecond)
		}(i) // kirim i sebagai argumen (aman terhadap loop variable)
	}
	wg.Wait() // blok sampai ketiga goroutine memanggil Done
	fmt.Println("ketiga goroutine selesai (main menunggu via WaitGroup)")
}

// ------------------------------------------------------------------
// 2. Channel: unbuffered, buffered, close+range, arah
// ------------------------------------------------------------------

// generate mengirim 1..n lalu menutup channel. Return channel receive-only.
func generate(n int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out) // pengirim yang menutup
		for i := 1; i <= n; i++ {
			out <- i
		}
	}()
	return out
}

func channelDasar() {
	fmt.Println("\n-- Channel dasar --")

	// Unbuffered: kirim & terima harus bertemu (sinkron).
	// 🔍 Analogi: channel unbuffered itu seperti SERAH-TERIMA barang tangan-ke-tangan.
	// Pengirim menunggu sampai penerima siap menerima (dan sebaliknya) — janjian ketemu.
	ch := make(chan string)
	go func() { ch <- "pesan dari goroutine" }()
	fmt.Println("terima:", <-ch)

	// Buffered: kirim tak blok selama buffer belum penuh.
	// 🔍 Analogi: channel buffered itu LOKER PENITIPAN dengan N kotak. Pengirim boleh
	// menaruh barang & pergi (tak menunggu) selama masih ada kotak kosong. Penuh -> baru menunggu.
	buf := make(chan int, 3)
	buf <- 1
	buf <- 2
	fmt.Printf("buffered len=%d cap=%d\n", len(buf), cap(buf))

	// range over channel sampai ditutup
	fmt.Print("generate(5): ")
	for v := range generate(5) {
		fmt.Printf("%d ", v)
	}
	fmt.Println()

	// comma-ok pada channel tertutup
	done := make(chan int)
	close(done)
	if v, ok := <-done; !ok {
		fmt.Printf("channel tertutup -> v=%d ok=%t\n", v, ok)
	}
}

// ------------------------------------------------------------------
// 3. select + timeout
// ------------------------------------------------------------------
// 🔍 Analogi: select itu seperti PENJAGA yang mengawasi BEBERAPA loket sekaligus dan
// melayani loket MANA PUN yang lebih dulu ada barangnya. Dipadu time.After, ia jadi
// "tunggu hasil, TAPI kalau lebih dari 100ms tak datang, ambil jalan timeout". default =
// "kalau tak ada satu pun loket siap SEKARANG, langsung lewat" (tak menunggu / non-blocking).
func selectDanTimeout() {
	fmt.Println("\n-- select + timeout --")

	cepat := make(chan string)
	go func() {
		time.Sleep(20 * time.Millisecond)
		cepat <- "hasil cepat"
	}()

	select {
	case res := <-cepat:
		fmt.Println("dapat:", res)
	case <-time.After(100 * time.Millisecond):
		fmt.Println("timeout!")
	}

	// default: non-blocking
	kosong := make(chan int)
	select {
	case v := <-kosong:
		fmt.Println("dapat", v)
	default:
		fmt.Println("tidak ada data siap (default, non-blocking)")
	}
}

// ------------------------------------------------------------------
// 4. sync: Mutex, Once, atomic
// ------------------------------------------------------------------

// 🔍 Analogi: Mutex itu KUNCI KAMAR MANDI. Sebelum masuk (mengubah value), pegawai
// mengunci pintu (Lock); yang lain antre. Selesai, buka kunci (Unlock) supaya berikutnya
// masuk. Tanpa kunci, dua pegawai menulis bersamaan -> DATA RACE (hasil kacau & tak terduga).
// Counter aman-konkuren dengan Mutex.
type Counter struct {
	mu    sync.Mutex
	value int
}

func (c *Counter) Inc() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.value++
}

func syncMutexOnceAtomic() {
	fmt.Println("\n-- sync: Mutex, Once, atomic --")

	// Mutex: 100 goroutine menaikkan counter -> hasil harus tepat 100.
	c := &Counter{}
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); c.Inc() }()
	}
	wg.Wait()
	fmt.Printf("Counter (Mutex) = %d (harus 100)\n", c.value)

	// atomic: counter tanpa mutex untuk operasi sederhana.
	// 🔍 Analogi: atomic itu operasi "sekali gerak tak bisa disela" — seperti mesin penghitung
	// klik di pintu yang tiap tekan pasti +1 utuh, tak mungkin setengah-setengah. Untuk hitungan
	// angka sederhana, atomic lebih ringan dari Mutex (tak perlu kunci-buka pintu).
	var atomicCount int64
	var wg2 sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg2.Add(1)
		go func() { defer wg2.Done(); atomic.AddInt64(&atomicCount, 1) }()
	}
	wg2.Wait()
	fmt.Printf("Counter (atomic) = %d (harus 100)\n", atomic.LoadInt64(&atomicCount))

	// Once: inisialisasi tepat sekali walau dipanggil banyak goroutine.
	// 🔍 Analogi: sync.Once itu seperti SAKELAR LAMPU yang cuma menyala sekali walau ditekan
	// beramai-ramai. Cocok untuk "buka koneksi DB / baca config" yang boleh terjadi tepat 1x.
	var once sync.Once
	var wg3 sync.WaitGroup
	var initCount int64
	for i := 0; i < 5; i++ {
		wg3.Add(1)
		go func() {
			defer wg3.Done()
			once.Do(func() { atomic.AddInt64(&initCount, 1) })
		}()
	}
	wg3.Wait()
	fmt.Printf("once.Do dijalankan %d kali (harus 1)\n", initCount)
}

// ------------------------------------------------------------------
// 5. context: cancellation & timeout
// ------------------------------------------------------------------

// 🔍 Analogi: context itu seperti REMOTE PEMBATAL + ALARM WAKTU yang dibagikan ke semua
// pegawai. Kalau pelanggan pergi (request batal) atau waktu habis, satu sinyal ctx.Done()
// menyuruh semua pegawai berhenti serentak — tak ada yang kerja sia-sia. Aturan Go: context
// selalu jadi argumen PERTAMA fungsi, dan diteruskan ke bawah rantai panggilan.
// kerjaLama mensimulasikan pekerjaan yang butuh 'butuh' waktu, tapi menghormati
// pembatalan lewat ctx.
func kerjaLama(ctx context.Context, butuh time.Duration) (string, error) {
	select {
	case <-time.After(butuh):
		return "selesai", nil
	case <-ctx.Done():
		return "", ctx.Err() // context deadline exceeded / canceled
	}
}

func contextCancellation() {
	fmt.Println("\n-- context: timeout --")

	// Kasus 1: pekerjaan cepat (30ms) < timeout (100ms) -> sukses.
	ctx1, cancel1 := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel1()
	if res, err := kerjaLama(ctx1, 30*time.Millisecond); err == nil {
		fmt.Println("kasus cepat ->", res)
	}

	// Kasus 2: pekerjaan lama (200ms) > timeout (50ms) -> dibatalkan.
	ctx2, cancel2 := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel2()
	if _, err := kerjaLama(ctx2, 200*time.Millisecond); err != nil {
		fmt.Println("kasus lambat -> dibatalkan:", err)
	}
}

// ------------------------------------------------------------------
// 6. Worker pool
// ------------------------------------------------------------------
// 🔍 Analogi: worker pool itu seperti 3 KASIR melayani 1 ANTREAN panjang berisi 9 pelanggan.
// Alih-alih membuka 9 kasir (boros), 3 kasir mengambil pelanggan berikutnya begitu selesai.
// 'jobs' = antrean masuk; 'results' = keranjang hasil. Ini pola pengendali beban paling umum.
func workerPool() {
	fmt.Println("\n-- Worker pool (3 worker, 9 job) --")

	const numWorkers, numJobs = 3, 9
	jobs := make(chan int, numJobs)
	results := make(chan int, numJobs)

	var wg sync.WaitGroup
	// Jalankan worker.
	for w := 1; w <= numWorkers; w++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := range jobs { // ambil job sampai channel jobs ditutup
				results <- j * j // kuadratkan
			}
		}(w)
	}

	// Kirim job lalu tutup.
	for j := 1; j <= numJobs; j++ {
		jobs <- j
	}
	close(jobs)

	// Tutup results setelah semua worker selesai (goroutine terpisah agar tak deadlock).
	go func() { wg.Wait(); close(results) }()

	// Kumpulkan hasil (urutan bisa acak) lalu urutkan untuk output stabil.
	var collected []int
	for r := range results {
		collected = append(collected, r)
	}
	sort.Ints(collected)
	fmt.Printf("hasil kuadrat = %v\n", collected)
}

// ------------------------------------------------------------------
// 7. Pipeline: gen -> square -> konsumsi
// ------------------------------------------------------------------
// 🔍 Analogi: pipeline itu seperti LINI PRODUKSI bertahap. Tiap tahap (generate -> square)
// adalah stasiun yang mengambil dari channel masuk, mengolah, lalu mengirim ke channel keluar.
// Barang mengalir tahap demi tahap tanpa menunggu semua selesai — hemat memori & alami.
func square(in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for v := range in {
			out <- v * v
		}
	}()
	return out
}

func pipeline() {
	fmt.Println("\n-- Pipeline (gen -> square) --")
	fmt.Print("hasil: ")
	for v := range square(generate(5)) {
		fmt.Printf("%d ", v)
	}
	fmt.Println()
}
