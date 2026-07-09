// Teknik Advanced Modul 02 (collections) — contoh runnable + penjelasan detail.
//
// Jalankan: go run ./02-collections/advanced
package main

import (
	"fmt"
	"maps"
	"slices"
	"strings"
)

// ===========================================================================
// 1. Slice tricks TANPA alokasi baru (reuse backing array)
// ===========================================================================

// hapusIndex menghapus elemen ke-i. Trik append(s[:i], s[i+1:]...) menggeser
// elemen setelah i ke kiri lalu memotong panjangnya — tanpa slice baru.
// Urutan elemen tetap terjaga.
func hapusIndex[T any](s []T, i int) []T {
	return append(s[:i], s[i+1:]...)
}

// filterInPlace menyaring slice memakai kembali array yang sama.
// out := s[:0] artinya "slice kosong tapi tetap menunjuk backing array s",
// jadi append menimpa data lama alih-alih mengalokasikan yang baru.
func filterInPlace(s []int, keep func(int) bool) []int {
	out := s[:0]
	for _, v := range s {
		if keep(v) {
			out = append(out, v)
		}
	}
	return out
}

// ===========================================================================
// 2. Three-index slice s[low:high:max] — membatasi cap agar aman dari aliasing
// ===========================================================================
// Slice biasa berbagi backing array dengan induknya; append pada anak bisa
// menimpa data induk bila cap masih tersisa. Bentuk s[low:high:max] mematok
// cap=max-low, sehingga append PASTI mengalokasikan array baru => induk aman.
func demoThreeIndex() {
	induk := []int{1, 2, 3, 4, 5}

	bahaya := induk[0:2]        // len=2, cap=5 -> masih bisa timpa induk
	bahaya = append(bahaya, 99) // menimpa induk[2] (3 -> 99)!
	fmt.Println("  aliasing (cap penuh):", induk)

	induk = []int{1, 2, 3, 4, 5}
	aman := induk[0:2:2]    // len=2, cap=2 -> append memaksa array baru
	aman = append(aman, 99) // TIDAK menyentuh induk
	fmt.Println("  three-index (aman) :", induk, "| anak:", aman)
}

// ===========================================================================
// 3. Set idiomatik dengan map[T]struct{}{} (nilai berukuran 0 byte)
// ===========================================================================
// struct{} tidak memakan memori, jadi lebih hemat daripada map[T]bool untuk
// merepresentasikan himpunan keanggotaan.
func uniqueTerurut(kata []string) []string {
	set := make(map[string]struct{}, len(kata))
	for _, k := range kata {
		set[k] = struct{}{}
	}
	hasil := slices.Collect(maps.Keys(set)) // maps.Keys -> iterator, slices.Collect -> slice
	slices.Sort(hasil)
	return hasil
}

func main() {
	fmt.Println("== 1. slice tricks ==")
	s := []int{10, 20, 30, 40, 50}
	fmt.Println("  hapus index 2:", hapusIndex(s, 2)) // buang 30
	genap := filterInPlace([]int{1, 2, 3, 4, 5, 6}, func(n int) bool { return n%2 == 0 })
	fmt.Println("  filter genap in-place:", genap)

	fmt.Println("== 2. three-index slice ==")
	demoThreeIndex()

	fmt.Println("== 3. set + package slices/maps ==")
	fmt.Println("  unik+terurut:", uniqueTerurut([]string{"go", "rust", "go", "c", "rust"}))

	// ---------------------------------------------------------------------
	// 4. clear() builtin (Go 1.21) — kosongkan map/slice di tempat
	// ---------------------------------------------------------------------
	m := map[string]int{"a": 1, "b": 2}
	clear(m) // hapus semua entri tanpa mengganti map (referensi tetap sama)
	fmt.Println("== 4. clear() ==")
	fmt.Println("  panjang map setelah clear:", len(m))

	// ---------------------------------------------------------------------
	// 5. strings.Builder — konkatenasi hemat (hindari += yang O(n^2))
	// ---------------------------------------------------------------------
	var b strings.Builder
	b.Grow(16) // prealokasi buffer bila perkiraan ukuran diketahui
	for i := 0; i < 3; i++ {
		fmt.Fprintf(&b, "baris-%d;", i) // Builder implement io.Writer
	}
	fmt.Println("== 5. strings.Builder ==")
	fmt.Println("  hasil:", b.String())
}
