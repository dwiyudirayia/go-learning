// Jalankan: go run ./42-go-internals
// Lihat escape analysis: go build -gcflags='-m' ./42-go-internals
// Verifikasi otomatis: go test ./42-go-internals
package main

import (
	"fmt"
	"runtime"
	"sync"
)

func main() {
	fmt.Println("=== 42 — Go Internals ===")
	demoScheduler()
	demoGC()
	demoAlignment()
}

// ------------------------------------------------------------------
// 1. Scheduler — model GMP (Goroutine, Machine/thread OS, Processor)
// ------------------------------------------------------------------
func demoScheduler() {
	fmt.Println("\n-- Scheduler (GMP) --")
	// 🔍 Analogi GMP: bayangkan DAPUR restoran. G (goroutine) = daftar PESANAN yang menumpuk
	// (bisa ribuan, murah ditulis). M (thread OS) = KOKI sungguhan (mahal, jumlahnya terbatas).
	// P (processor) = STASIUN MASAK/kompor; hanya koki yang memegang stasiun yang boleh memasak.
	// Scheduler Go = manajer yang menugaskan pesanan (G) ke koki (M) di stasiun (P) secara efisien.
	// Karena "pesanan" jauh lebih murah dari "koki", Go sanggup ribuan goroutine di segelintir thread.
	// GOMAXPROCS = jumlah stasiun = berapa banyak yang benar-benar memasak BERSAMAAN (paralel).
	// G = goroutine (ringan), M = thread OS, P = processor logis (konteks).
	// GOMAXPROCS = jumlah P = maks goroutine yang benar-benar PARALEL.
	fmt.Printf("NumCPU        = %d\n", runtime.NumCPU())
	fmt.Printf("GOMAXPROCS    = %d (goroutine paralel maksimum)\n", runtime.GOMAXPROCS(0))
	fmt.Printf("goroutine now = %d\n", runtime.NumGoroutine())

	// Spawn banyak goroutine ringan (murah — ~2KB stack awal, tumbuh dinamis).
	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() { defer wg.Done() }()
	}
	fmt.Printf("goroutine saat 1000 dibuat = %d (ribuan goroutine itu normal)\n", runtime.NumGoroutine())
	wg.Wait()
}

// ------------------------------------------------------------------
// 2. Garbage Collector
// ------------------------------------------------------------------
// 🔍 Analogi GC (Garbage Collector): seperti PETUGAS KEBERSIHAN otomatis yang keliling membuang
// gelas & piring yang sudah tak dipakai (memori yang tak lagi terpakai) — jadi kamu tak perlu
// mengelola sendiri seperti di C (free() manual, rawan bocor/double-free). Harganya: petugas butuh
// sedikit waktu CPU untuk keliling. GOGC = seberapa sering ia menyapu: sering (hemat memori, boros CPU)
// vs jarang (boros memori, hemat CPU). Trade-off yang bisa disetel sesuai kebutuhan aplikasi.
func demoGC() {
	fmt.Println("\n-- Garbage Collector --")
	var m runtime.MemStats

	runtime.ReadMemStats(&m)
	before := m.NumGC

	// Alokasikan lalu buang banyak memori.
	for i := 0; i < 10; i++ {
		_ = make([]byte, 1<<20) // 1 MB tiap iterasi
	}
	runtime.GC() // paksa GC (biasanya otomatis)

	runtime.ReadMemStats(&m)
	fmt.Printf("total alokasi seumur program = %d MB\n", m.TotalAlloc/(1<<20))
	fmt.Printf("jumlah siklus GC             = %d (naik %d setelah GC dipaksa)\n", m.NumGC, m.NumGC-before)
	fmt.Printf("heap saat ini                = %d KB\n", m.HeapAlloc/1024)
	fmt.Println("Catatan: setel GOGC (default 100) untuk tukar CPU vs memori.")
}

// ------------------------------------------------------------------
// 3. Memory alignment
// ------------------------------------------------------------------
// 🔍 Analogi alignment/padding: memori itu seperti RAK dengan slot berukuran tetap (mis. 8 byte).
// Field harus ditaruh di posisi "rapi" (kelipatan ukurannya), jadi CPU mengambilnya sekali comot.
// Kalau urutan field buruk (bool, int64, bool), Go menyelipkan "ganjalan kosong" (padding) agar rapi —
// boros tempat. Mengurutkan field dari BESAR ke KECIL memampatkan ganjalan itu. Seperti menyusun
// koper: barang besar dulu, celahnya diisi barang kecil — muat lebih banyak di ruang sama.
func demoAlignment() {
	fmt.Println("\n-- Memory Alignment --")
	fmt.Printf("BadStruct  {bool,int64,bool} = %d byte (padding boros)\n", badSize())
	fmt.Printf("GoodStruct {int64,bool,bool} = %d byte (urutan rapi)\n", goodSize())
	fmt.Printf("Hemat %d byte hanya dengan mengurutkan field (besar->kecil).\n", badSize()-goodSize())
}
