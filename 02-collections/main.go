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

	// 🔍 Analogi: Array itu seperti KARDUS TELUR dengan jumlah lubang TETAP (misal 3).
	// Slice itu seperti JENDELA/BINGKAI yang mengintip sebagian rak — bisa geser & melebar.
	// Array di-COPY saat dikirim ke fungsi -> perubahan di dalam tak terlihat di luar.
	// 🔍 Analogi: mengirim array ke fungsi = memberi FOTOKOPI dokumen. Orang mencoret
	// fotokopinya, dokumen aslimu tetap bersih.
	arr := [3]int{1, 2, 3}
	ubahArray(arr)
	fmt.Printf("array setelah dikirim ke fungsi: %v (tetap, karena di-copy)\n", arr)

	// Slice berbagi backing array -> perubahan di dalam fungsi TERLIHAT di luar.
	// 🔍 Analogi: mengirim slice = memberi ALAMAT gudang, bukan fotokopi. Orang lain
	// datang ke gudang yang sama dan menggeser barang -> kamu ikut melihat perubahannya.
	sl := []int{1, 2, 3}
	ubahSlice(sl)
	fmt.Printf("slice setelah dikirim ke fungsi: %v (berubah, karena share backing)\n", sl)

	// len vs cap
	// 🔍 Analogi: len = berapa kursi yang SUDAH DIDUDUKI; cap = total kursi yang TERSEDIA
	// di gerbong sebelum harus menyewa gerbong baru. len=2, cap=5 -> 2 terisi, muat 5.
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

	// 🔍 Analogi: append itu seperti menambah tamu ke MEJA. Selama kursi masih ada
	// (cap belum penuh), tamu langsung duduk. Begitu penuh, restoran memindahkan
	// SEMUA tamu ke meja baru yang lebih besar (realokasi) — makanya cap melonjak 4->8.
	s := make([]int, 0, 4)
	fmt.Printf("awal   : len=%d cap=%d\n", len(s), cap(s))
	for i := 1; i <= 6; i++ {
		s = append(s, i) // saat cap terlampaui (4 -> 8), Go realokasi array baru
		fmt.Printf("append %d -> len=%d cap=%d\n", i, len(s), cap(s))
	}

	// JEBAKAN: dua slice berbagi backing array.
	// 🔍 Analogi: induk[1:3] itu seperti MENYOROT sebagian teks yang sama, bukan menyalinnya.
	// Kamu mengedit teks yang disorot -> dokumen aslinya ikut berubah, karena itu teks yang SAMA.
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
	// 🔍 Analogi: copy itu MENGGANDAKAN dokumen ke kertas baru. Setelah itu dua dokumen
	// hidup terpisah — coret salah satu, yang lain aman. Inilah obat jebakan di atas.
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

	// 🔍 Analogi: map itu seperti KAMUS atau BUKU INDEKS — cari lewat "kata" (key)
	// langsung ketemu "artinya" (value), tanpa membaca dari halaman 1. Cepat.
	m := map[string]int{"apel": 3, "jeruk": 5}

	// comma-ok: cara benar cek keberadaan key
	// 🔍 Analogi: 'ok' itu jawaban satpam: "ada/tidak". Tanpa 'ok', map balas 0 untuk key
	// yang tak ada — kamu tak bisa bedakan "nilainya memang 0" vs "key-nya tak ada". 'ok' memperjelas.
	if v, ok := m["apel"]; ok {
		fmt.Printf("apel ada, nilai=%d\n", v)
	}
	if _, ok := m["mangga"]; !ok {
		fmt.Println("mangga tidak ada (comma-ok mengembalikan false)")
	}

	m["pisang"] = 7
	delete(m, "jeruk")

	// Iterasi map urutannya ACAK. Untuk output stabil, urutkan key-nya dulu.
	// 🔍 Analogi: mengambil kunci map itu seperti mengambil kartu dari kantong yang DIKOCOK —
	// urutannya bisa beda tiap kali. Kalau butuh urutan tetap, kumpulkan dulu lalu sortir.
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Printf("  %s = %d\n", k, m[k])
	}

	// Map nil: aman dibaca, panic kalau ditulis.
	// 🔍 Analogi: nil map itu seperti RAK yang belum dibeli. Mengintip rak kosong boleh
	// (dapat nilai netral). Tapi mencoba MENARUH barang di rak yang belum ada -> ambruk (panic).
	var nilMap map[string]int
	fmt.Printf("baca nil map: %d (zero value, aman)\n", nilMap["x"])
	// nilMap["x"] = 1  // <- ini akan PANIC: assignment to entry in nil map
}

// ------------------------------------------------------------------
// 5. String = byte (UTF-8), bukan karakter
// ------------------------------------------------------------------
func stringByteRune() {
	fmt.Println("\n-- String, byte, rune --")

	// 🔍 Analogi: string di Go disimpan sebagai deretan BYTE (UTF-8), bukan deretan huruf.
	// Huruf latin (H, a) = 1 byte, tapi huruf seperti 世 memakan 3 byte. Jadi len() menghitung
	// "berapa banyak amplop", bukan "berapa banyak surat". Satu huruf 世 = 3 amplop.
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
// 🔍 Analogi: rune itu "satu huruf utuh". Membalik per-byte bisa MEMOTONG huruf 3-byte
// jadi sampah. []rune mengumpulkan tiap huruf utuh dulu, baru dibalik — huruf tetap utuh.
func reverseString(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}
