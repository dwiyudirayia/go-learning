// Teknik Advanced Modul 42 (go internals) — contoh runnable + penjelasan detail.
//
// Scheduler GMP, GC, alignment/padding, escape analysis. Jalankan:
//
//	go run ./42-go-internals/advanced
//
// Untuk melihat escape analysis: go build -gcflags='-m' ./42-go-internals/advanced
package main

import (
	"fmt"
	"runtime"
	"unsafe"
)

// ===========================================================================
// 1. ALIGNMENT & PADDING — urutan field mengubah ukuran struct
// ===========================================================================
// Field dirapikan (aligned) ke kelipatan ukurannya; celah kosong = padding.
//
// 🔍 Analogi: seperti MENYUSUN KARDUS DI TRUK — kardus besar (int64) harus
// menempel di garis kelipatan 8; menaruh kardus kecil (bool) di depannya
// menciptakan rongga kosong yang tetap terangkut (padding). Susun dari besar
// ke kecil = muatan lebih padat, truk (struct) lebih ramping.
type Boros struct {
	a bool  // 1 byte, lalu 7 padding agar b rata di 8
	b int64 // 8 byte
	c bool  // 1 byte, lalu 7 padding di akhir
} // = 24 byte

type Hemat struct {
	b int64 // 8
	a bool  // 1
	c bool  // 1, + 6 padding
} // = 16 byte

func main() {
	// =====================================================================
	// 1. alignment
	// =====================================================================
	fmt.Println("== 1. alignment & padding ==")
	fmt.Printf("  Boros=%d byte, Hemat=%d byte (field identik, beda urutan)\n",
		unsafe.Sizeof(Boros{}), unsafe.Sizeof(Hemat{}))
	fmt.Printf("  offset Boros.b=%d (ada padding), Hemat.b=%d\n",
		unsafe.Offsetof(Boros{}.b), unsafe.Offsetof(Hemat{}.b))

	// =====================================================================
	// 2. SCHEDULER GMP — goroutine (G) dijadwal ke thread (M) via processor (P)
	// =====================================================================
	// GOMAXPROCS = jumlah P = paralelisme sejati. Goroutine sangat murah (stack
	// awal ~2KB, tumbuh dinamis) sehingga ribuan goroutine itu lumrah.
	//
	// 🔍 Analogi GMP: bayangkan RUMAH SAKIT — pasien (G/goroutine) jumlahnya
	// ribuan, dokter (M/thread OS) terbatas, dan RUANG PRAKTIK (P/processor)
	// yang menentukan berapa pasien bisa ditangani BERSAMAAN. Dokter berpindah
	// menangani pasien mana pun yang antre di ruangnya; pasien yang menunggu
	// hasil lab (I/O) dipersilakan duduk dulu agar ruang dipakai pasien lain.
	fmt.Println("== 2. scheduler GMP ==")
	fmt.Printf("  NumCPU=%d  GOMAXPROCS(P)=%d  goroutine aktif=%d\n",
		runtime.NumCPU(), runtime.GOMAXPROCS(0), runtime.NumGoroutine())

	// =====================================================================
	// 3. GARBAGE COLLECTOR — statistik & pemicuan manual
	// =====================================================================
	// GC concurrent tri-color mark-sweep. Kendalikan lewat GOGC (frekuensi) &
	// GOMEMLIMIT (batas memori lunak). Amati via runtime.MemStats / GODEBUG=gctrace=1.
	//
	// 🔍 Analogi: GC itu seperti PETUGAS KEBERSIHAN GEDUNG yang menandai
	// barang mana yang masih dipakai penghuni (mark) lalu mengangkut sisanya
	// (sweep) — sambil gedung tetap beroperasi (concurrent). GOGC = seberapa
	// sering ronda; GOMEMLIMIT = "kalau gudang hampir penuh, ronda lebih rajin".
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Println("== 3. GC ==")
	fmt.Printf("  HeapAlloc=%d KB  NumGC=%d  NextGC=%d KB\n",
		m.HeapAlloc/1024, m.NumGC, m.NextGC/1024)
	sebelum := m.NumGC
	runtime.GC() // paksa satu siklus GC
	runtime.ReadMemStats(&m)
	fmt.Printf("  setelah runtime.GC(): NumGC %d -> %d\n", sebelum, m.NumGC)

	// =====================================================================
	// 4. ESCAPE ANALYSIS — stack (murah) vs heap (butuh GC)
	// =====================================================================
	// Compiler memutuskan apakah nilai "lolos" ke heap. Mengembalikan pointer ke
	// variabel lokal memaksanya ke heap. Cek dengan: go build -gcflags='-m'.
	//
	// 🔍 Analogi: stack itu MEJA KERJA pribadi — barang di atasnya otomatis
	// dibereskan begitu kamu selesai (fungsi return), tanpa biaya. Heap itu
	// GUDANG BERSAMA — barang yang masih dibutuhkan orang lain setelah kamu
	// pulang (pointer yang dikembalikan) harus dititip ke gudang, dan petugas
	// kebersihan (GC) yang kelak membereskannya.
	fmt.Println("== 4. escape analysis ==")
	p := buatDiHeap()
	fmt.Printf("  nilai dari pointer yang escape ke heap: %d\n", *p)
	fmt.Println("  (jalankan `go build -gcflags=-m` untuk melihat keputusan escape)")
}

// buatDiHeap: karena mengembalikan pointer ke lokal, x ESCAPE ke heap.
func buatDiHeap() *int {
	x := 42
	return &x
}
