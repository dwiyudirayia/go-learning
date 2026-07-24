// google/go-cmp — membandingkan nilai secara mendalam, dengan LAPORAN PERBEDAAN yang jelas.
//
// Jalankan: go run ./libraries/gocmp
// Test:     go test ./libraries/gocmp    (go-cmp memang dirancang untuk test)
//
// 🔍 Analogi besar: reflect.DeepEqual bawaan itu seperti WASIT yang cuma bisa bilang "SALAH"
// tanpa menjelaskan di mana. Saat dua struct besar berbeda, kamu dapat "false" lalu harus
// menyipitkan mata membandingkan dua dump raksasa sendiri.
//
// go-cmp itu wasit yang menunjuk PERSIS bidak mana yang salah: "field Harga: 100 vs 90".
// cmp.Diff mengembalikan teks perbedaan siap tempel ke pesan test. Inilah kenapa ia jadi
// standar de-facto untuk membandingkan nilai di test Go (dipakai luas, termasuk oleh Google).
//
// go-cmp vs yang lain:
//   - reflect.DeepEqual : true/false saja, tanpa penjelasan. Cukup untuk cek cepat.
//   - testify assert.Equal : punya diff, enak dipakai. go-cmp lebih kuat untuk struct rumit
//     (opsi mengabaikan field, membandingkan unexported, toleransi angka mengambang).
//   - go-cmp : paling ekspresif untuk perbandingan struktur kompleks di test.
//
// PENTING: go-cmp untuk TEST, bukan untuk logika produksi. Ia sengaja PANIC bila tak tahu
// cara membandingkan sesuatu (mis. field unexported tanpa opsi) — supaya kesalahan test
// ketahuan, bukan lolos diam-diam. Jangan pakai cmp.Equal di jalur kode produksi.
package main

import (
	"fmt"
	"strings"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

// Produk & Pesanan — data bersarang untuk memperagakan perbandingan mendalam.
type Produk struct {
	ID    int
	Nama  string
	Harga int
	Tag   []string
}

type Pesanan struct {
	Kode    string
	Item    []Produk
	Catatan string
}

func main() {
	fmt.Println("=== google/go-cmp ===")
	demoDiff()
	demoIgnore()
	demoUnexported()
	demoToleransi()
}

// ------------------------------------------------------------------
// 1. Diff dasar
// ------------------------------------------------------------------

// SamaPersis mengembalikan true bila kedua nilai identik secara mendalam.
func SamaPersis(a, b any) bool {
	return cmp.Equal(a, b)
}

// Perbedaan mengembalikan laporan perbedaan (kosong bila sama).
//
// 🔍 Analogi format diff: baris berawalan "-" adalah yang DIHARAPKAN (want), "+" yang
// DIDAPAT (got) — sama seperti diff git. Konvensi go-cmp: cmp.Diff(want, got).
func Perbedaan(want, got any) string {
	return cmp.Diff(want, got)
}

func demoDiff() {
	fmt.Println("\n-- Diff dasar --")

	a := Produk{ID: 1, Nama: "Kopi", Harga: 100, Tag: []string{"panas", "pahit"}}
	b := Produk{ID: 1, Nama: "Kopi", Harga: 90, Tag: []string{"panas", "manis"}}

	fmt.Printf("   sama? %t\n", SamaPersis(a, b))
	fmt.Println("   perbedaan:")
	fmt.Print(indent(Perbedaan(a, b)))
}

// ------------------------------------------------------------------
// 2. Mengabaikan field tertentu
// ------------------------------------------------------------------

// 🔍 Analogi cmpopts.IgnoreFields: seperti berkata pada wasit "abaikan nomor punggung,
// bandingkan permainannya saja". Sangat berguna untuk field yang TAK RELEVAN bagi test:
// ID yang di-generate database, timestamp CreatedAt, atau field acak. Tanpa ini, test
// jadi rapuh — gagal hanya karena ID kebetulan beda.

// SamaAbaikanID membandingkan produk TANPA mempedulikan field ID.
func SamaAbaikanID(a, b Produk) bool {
	return cmp.Equal(a, b, cmpopts.IgnoreFields(Produk{}, "ID"))
}

func demoIgnore() {
	fmt.Println("\n-- Mengabaikan field --")

	a := Produk{ID: 1, Nama: "Kopi", Harga: 100}
	b := Produk{ID: 999, Nama: "Kopi", Harga: 100} // ID beda, sisanya sama

	fmt.Printf("   sama persis?      %t\n", SamaPersis(a, b))
	fmt.Printf("   sama (abaikan ID)? %t\n", SamaAbaikanID(a, b))
}

// ------------------------------------------------------------------
// 3. Membandingkan field unexported
// ------------------------------------------------------------------

// dompet punya field kecil (unexported) — go-cmp menolak membandingkannya TANPA izin.
type dompet struct {
	pemilik string
	saldo   int
}

// NewDompet konstruktor (karena field-nya unexported dari luar paket).
func NewDompet(pemilik string, saldo int) dompet {
	return dompet{pemilik: pemilik, saldo: saldo}
}

// 🔍 Analogi cmpopts.IgnoreUnexported / cmp.AllowUnexported: field unexported itu seperti
// ISI DOMPET PRIBADI. Secara bawaan go-cmp menolak mengintipnya (PANIC) demi keamanan —
// kamu harus MEMBERI IZIN eksplisit. AllowUnexported("ya, bandingkan isi dompet ini").

// DompetSama membandingkan dua dompet termasuk field privatnya.
func DompetSama(a, b dompet) bool {
	return cmp.Equal(a, b, cmp.AllowUnexported(dompet{}))
}

func demoUnexported() {
	fmt.Println("\n-- Field unexported --")

	a := NewDompet("Ana", 100)
	b := NewDompet("Ana", 100)

	fmt.Printf("   dua dompet identik sama? %t\n", DompetSama(a, b))
	fmt.Printf("   (tanpa AllowUnexported, cmp.Equal akan PANIC — sengaja, demi keamanan)\n")
}

// ------------------------------------------------------------------
// 4. Toleransi angka mengambang
// ------------------------------------------------------------------

// 🔍 Analogi: membandingkan float dengan "==" itu jebakan (lihat libraries/decimal) —
// 0.1+0.2 tak persis 0.3. Untuk hasil perhitungan mengambang, kamu ingin "cukup dekat",
// bukan "persis sama". cmpopts.EquateApprox memberi toleransi: seperti wasit yang
// memaklumi selisih sepersejuta.

// HampirSama membandingkan dua float dengan toleransi kecil.
func HampirSama(a, b float64) bool {
	// toleransi relatif 0, absolut 1e-9
	return cmp.Equal(a, b, cmpopts.EquateApprox(0, 1e-9))
}

func demoToleransi() {
	fmt.Println("\n-- Toleransi float --")

	x := 0.1 + 0.2 // bukan persis 0.3
	fmt.Printf("   0.1+0.2 == 0.3 (operator)?     %t\n", x == 0.3)
	fmt.Printf("   0.1+0.2 ~= 0.3 (EquateApprox)? %t\n", HampirSama(x, 0.3))
}

// indent memberi indentasi pada blok diff agar rapi di keluaran demo.
func indent(s string) string {
	if s == "" {
		return "     (tidak ada perbedaan)\n"
	}
	baris := strings.Split(strings.TrimRight(s, "\n"), "\n")
	for i, b := range baris {
		baris[i] = "     " + b
	}
	return strings.Join(baris, "\n") + "\n"
}
