// Solusi latihan Modul 02.
// Jalankan: go run ./02-collections/latihan
package main

import (
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"
)

func main() {
	fmt.Println("=== Solusi Latihan Modul 02 ===")
	latihan1()
	latihan2()
	latihan3()
	latihan4()
	latihan5()
}

// ------------------------------------------------------------------
// Latihan 1: dari 1..10, ambil hanya yang genap.
// ------------------------------------------------------------------
// Idiom "filter": bikin slice tujuan kosong dgn cap secukupnya lalu append.
// make([]int, 0, len(src)) -> len 0, tapi cap sudah dialokasi agar append
// tidak berkali-kali realokasi.
func latihan1() {
	fmt.Println("\n-- Latihan 1: filter genap --")

	angka := make([]int, 0, 10)
	for i := 1; i <= 10; i++ {
		angka = append(angka, i)
	}

	genap := make([]int, 0, len(angka))
	for _, n := range angka {
		if n%2 == 0 {
			genap = append(genap, n)
		}
	}
	fmt.Printf("asal  = %v\n", angka)
	fmt.Printf("genap = %v\n", genap)
}

// ------------------------------------------------------------------
// Latihan 2: buktikan jebakan backing array.
// ------------------------------------------------------------------
// b := a[:2] TIDAK menyalin — b memakai backing array yang sama dengan a.
// Jadi mengubah b[0] otomatis mengubah a[0].
func latihan2() {
	fmt.Println("\n-- Latihan 2: jebakan backing array --")

	a := []int{10, 20, 30, 40}
	b := a[:2] // view {10,20}, backing sama dgn a
	b[0] = 999
	fmt.Printf("a=%v b=%v -> mengubah b[0] ikut mengubah a[0]!\n", a, b)

	// Cara aman jika ingin terpisah: salin dulu.
	c := make([]int, 2)
	copy(c, a[:2])
	c[0] = -1
	fmt.Printf("a=%v c=%v -> c independen (pakai copy)\n", a, c)
}

// ------------------------------------------------------------------
// Latihan 3: frekuensi kata pakai map[string]int.
// ------------------------------------------------------------------
// strings.Fields memecah string atas spasi (menangani spasi ganda).
// Map nil TIDAK boleh ditulis, jadi kita pakai literal {} (bukan var m map...).
func latihan3() {
	fmt.Println("\n-- Latihan 3: frekuensi kata --")

	kalimat := "kucing makan ikan dan kucing tidur lalu ikan berenang kucing"
	freq := map[string]int{}
	for _, kata := range strings.Fields(kalimat) {
		freq[kata]++ // key belum ada -> zero value 0, lalu +1
	}

	// Map iterasinya acak; urutkan key agar output stabil.
	keys := make([]string, 0, len(freq))
	for k := range freq {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Printf("  %-8s : %d\n", k, freq[k])
	}
}

// ------------------------------------------------------------------
// Latihan 4: byte vs rune.
// ------------------------------------------------------------------
func latihan4() {
	fmt.Println("\n-- Latihan 4: byte vs rune --")

	s := "naïve café 世界"
	fmt.Printf("string      = %q\n", s)
	fmt.Printf("len (byte)  = %d\n", len(s))
	fmt.Printf("rune count  = %d (utf8.RuneCountInString)\n", utf8.RuneCountInString(s))
	fmt.Printf("[]rune len  = %d (cara lain menghitung rune)\n", len([]rune(s)))
	// Catatan: 'ï', 'é', '世', '界' butuh >1 byte, itulah kenapa byte > rune.
}

// ------------------------------------------------------------------
// Latihan 5: reverse string yang aman terhadap multi-byte rune.
// ------------------------------------------------------------------
// SALAH kalau membalik per byte: karakter multi-byte akan rusak jadi byte tak valid.
// BENAR: konversi ke []rune dulu, balik, lalu konversi balik ke string.
func reverse(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}

func latihan5() {
	fmt.Println("\n-- Latihan 5: reverse string --")

	for _, s := range []string{"Halo", "café 世界", "naïve"} {
		rev := reverse(s)
		// Verifikasi hasil masih UTF-8 valid & panjang rune sama.
		fmt.Printf("%-10q -> %-10q  (valid UTF-8: %t)\n", s, rev, utf8.ValidString(rev))
	}
}
