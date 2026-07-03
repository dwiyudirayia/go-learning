// Package main untuk modul 02 — Collections.
// Jalankan: go run ./02-collections
package main

import (
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"
)

func main() {
	fmt.Println("=== 02 — Collections ===")
	arrayVsSlice()
	appendDanBackingArray()
	copyDanSliceTricks()
	contohMap()
	stringByteRune()
}

// ------------------------------------------------------------------
// 1. Array (value) vs Slice (view)
// ------------------------------------------------------------------
func arrayVsSlice() {
	fmt.Println("\n-- Array vs Slice --")

	// Array di-COPY saat dikirim ke fungsi -> perubahan di dalam tak terlihat di luar.
	arr := [3]int{1, 2, 3}
	ubahArray(arr)
	fmt.Printf("array setelah dikirim ke fungsi: %v (tetap, karena di-copy)\n", arr)

	// Slice berbagi backing array -> perubahan di dalam fungsi TERLIHAT di luar.
	sl := []int{1, 2, 3}
	ubahSlice(sl)
	fmt.Printf("slice setelah dikirim ke fungsi: %v (berubah, karena share backing)\n", sl)

	// len vs cap
	s := make([]int, 2, 5)
	fmt.Printf("make([]int,2,5) -> len=%d cap=%d isi=%v\n", len(s), cap(s), s)
}

func ubahArray(a [3]int) { a[0] = 99 } // a adalah salinan
func ubahSlice(s []int)  { s[0] = 99 } // s menunjuk backing array yang sama

// ------------------------------------------------------------------
// 2. append & jebakan backing array bersama
// ------------------------------------------------------------------
func appendDanBackingArray() {
	fmt.Println("\n-- append & backing array --")

	s := make([]int, 0, 4)
	fmt.Printf("awal   : len=%d cap=%d\n", len(s), cap(s))
	for i := 1; i <= 6; i++ {
		s = append(s, i) // saat cap terlampaui (4 -> 8), Go realokasi array baru
		fmt.Printf("append %d -> len=%d cap=%d\n", i, len(s), cap(s))
	}

	// JEBAKAN: dua slice berbagi backing array.
	induk := []int{10, 20, 30, 40}
	anak := induk[1:3] // view {20,30}, TAPI backing sama dgn induk
	anak[0] = 999      // mengubah anak -> ikut mengubah induk[1]
	fmt.Printf("induk=%v anak=%v (mengubah anak ikut mengubah induk!)\n", induk, anak)

	// Solusi kalau ingin benar-benar terpisah: copy ke slice baru (lihat fungsi berikut).
}

// ------------------------------------------------------------------
// 3. copy & slice tricks (hapus / sisip)
// ------------------------------------------------------------------
func copyDanSliceTricks() {
	fmt.Println("\n-- copy & slice tricks --")

	src := []int{1, 2, 3, 4, 5}

	// copy: bikin salinan independen (tidak berbagi backing array)
	dst := make([]int, len(src))
	copy(dst, src)
	dst[0] = -1
	fmt.Printf("src=%v dst=%v (dst independen berkat copy)\n", src, dst)

	// Hapus elemen index 2 (nilai 3)
	i := 2
	hapus := append(src[:i], src[i+1:]...)
	fmt.Printf("hapus index %d -> %v\n", i, hapus)

	// Sisip 99 di index 1 (pakai salinan biar jelas)
	base := []int{1, 2, 3}
	base = append(base[:1], append([]int{99}, base[1:]...)...)
	fmt.Printf("sisip 99 di index 1 -> %v\n", base)
}

// ------------------------------------------------------------------
// 4. Map: comma-ok, nil map, iterasi acak
// ------------------------------------------------------------------
func contohMap() {
	fmt.Println("\n-- Map --")

	m := map[string]int{"apel": 3, "jeruk": 5}

	// comma-ok: cara benar cek keberadaan key
	if v, ok := m["apel"]; ok {
		fmt.Printf("apel ada, nilai=%d\n", v)
	}
	if _, ok := m["mangga"]; !ok {
		fmt.Println("mangga tidak ada (comma-ok mengembalikan false)")
	}

	m["pisang"] = 7
	delete(m, "jeruk")

	// Iterasi map urutannya ACAK. Untuk output stabil, urutkan key-nya dulu.
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Printf("  %s = %d\n", k, m[k])
	}

	// Map nil: aman dibaca, panic kalau ditulis.
	var nilMap map[string]int
	fmt.Printf("baca nil map: %d (zero value, aman)\n", nilMap["x"])
	// nilMap["x"] = 1  // <- ini akan PANIC: assignment to entry in nil map
}

// ------------------------------------------------------------------
// 5. String = byte (UTF-8), bukan karakter
// ------------------------------------------------------------------
func stringByteRune() {
	fmt.Println("\n-- String, byte, rune --")

	s := "Halo, 世界"
	fmt.Printf("string=%q\n", s)
	fmt.Printf("len(s)=%d byte (BUKAN jumlah karakter)\n", len(s))
	fmt.Printf("jumlah rune=%d (utf8.RuneCountInString)\n", utf8.RuneCountInString(s))

	// range string -> index BYTE + rune
	fmt.Print("range: ")
	for i, r := range s {
		fmt.Printf("[%d:%c] ", i, r)
	}
	fmt.Println()

	// s[i] adalah BYTE (uint8), bukan karakter
	fmt.Printf("s[0]=%d (byte 'H'), tipe %T\n", s[0], s[0])

	// Reverse string yang benar (per rune, bukan per byte) memakai []rune
	fmt.Printf("reverse %q -> %q\n", s, reverseString(s))

	// Menyusun string efisien dengan strings.Builder (hindari += berulang)
	var b strings.Builder
	for i := 0; i < 3; i++ {
		fmt.Fprintf(&b, "bagian%d ", i)
	}
	fmt.Printf("builder -> %q\n", strings.TrimSpace(b.String()))
}

// reverseString membalik string per-rune supaya karakter multi-byte tidak rusak.
func reverseString(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}
