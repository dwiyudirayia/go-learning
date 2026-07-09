// Teknik Advanced Modul 26 (profiling) — contoh runnable + penjelasan detail.
//
// Menangkap profil & statistik memori secara programatik. Jalankan:
//
//	go run ./26-profiling/advanced
package main

import (
	"bytes"
	"fmt"
	"runtime"
	"runtime/pprof"
)

// alokasiBanyak sengaja mengalokasikan ~5MB agar terlihat di MemStats & profil.
func alokasiBanyak() [][]byte {
	var data [][]byte
	for i := 0; i < 5000; i++ {
		data = append(data, make([]byte, 1024)) // 1KB x 5000
	}
	return data
}

func main() {
	// =====================================================================
	// 1. runtime.ReadMemStats — ukur pertumbuhan heap sebelum vs sesudah
	// =====================================================================
	var m1, m2 runtime.MemStats
	runtime.GC() // bersihkan dulu agar angka lebih jujur
	runtime.ReadMemStats(&m1)

	data := alokasiBanyak() // tahan referensi -> tak di-GC

	runtime.ReadMemStats(&m2)
	fmt.Println("== 1. MemStats ==")
	fmt.Printf("  HeapAlloc naik ~%d KB\n", (m2.HeapAlloc-m1.HeapAlloc)/1024)
	fmt.Printf("  total alokasi (Mallocs) bertambah: %d objek\n", m2.Mallocs-m1.Mallocs)

	// =====================================================================
	// 2. runtime/pprof — tangkap HEAP profile ke buffer (bisa ke file utk
	//    dianalisis `go tool pprof`)
	// =====================================================================
	var buf bytes.Buffer
	if err := pprof.WriteHeapProfile(&buf); err != nil {
		panic(err)
	}
	fmt.Println("== 2. heap profile ==")
	fmt.Printf("  profil ditangkap: %d byte (analisis: go tool pprof <file>)\n", buf.Len())
	runtime.KeepAlive(data) // pastikan data masih hidup saat profil diambil

	// =====================================================================
	// 3. Info runtime yang relevan untuk tuning
	// =====================================================================
	fmt.Println("== 3. runtime info ==")
	fmt.Printf("  NumCPU=%d  GOMAXPROCS=%d  NumGoroutine=%d\n",
		runtime.NumCPU(), runtime.GOMAXPROCS(0), runtime.NumGoroutine())

	// Peta jalan optimasi (profil DULU, baru optimasi):
	//   - go test -bench . -benchmem   -> allocs/op; bandingkan pakai benchstat
	//   - go tool pprof cpu.prof        -> temukan hot path (flamegraph -http)
	//   - go build -gcflags='-m'        -> escape analysis (heap vs stack)
	//   - GOGC / GOMEMLIMIT             -> atur agresivitas GC vs jejak memori
}
