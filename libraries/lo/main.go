// samber/lo — helper generik untuk slice & map (Map, Filter, Reduce, GroupBy, ...).
//
// Jalankan: go run ./libraries/lo
// Test:     go test ./libraries/lo
//
// 🔍 Analogi besar: Go sengaja tidak menyediakan "Map/Filter/Reduce" bawaan, jadi kamu
// menulis for-loop untuk hampir semua hal. Itu seperti MEMOTONG SAYUR DENGAN PISAU: selalu
// bisa, jelas terlihat apa yang terjadi, tapi untuk 20 wortel jadi melelahkan. samber/lo itu
// FOOD PROCESSOR — sekali tekan, matang. Bahayanya sama seperti food processor sungguhan:
// dipakai untuk hal yang salah, hasilnya malah bubur. Baca bagian "Kapan JANGAN pakai lo"
// di bawah — untuk banyak kasus, for-loop biasa tetap lebih jelas dan lebih cepat.
package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/samber/lo"
)

// Produk dipakai sebagai data contoh di seluruh file.
type Produk struct {
	ID       int
	Nama     string
	Kategori string
	Harga    int
	Stok     int
}

func dataContoh() []Produk {
	return []Produk{
		{ID: 1, Nama: "Kopi Arabika", Kategori: "minuman", Harga: 85_000, Stok: 12},
		{ID: 2, Nama: "Teh Melati", Kategori: "minuman", Harga: 35_000, Stok: 0},
		{ID: 3, Nama: "Roti Gandum", Kategori: "makanan", Harga: 25_000, Stok: 7},
		{ID: 4, Nama: "Keju Cheddar", Kategori: "makanan", Harga: 120_000, Stok: 3},
		{ID: 5, Nama: "Gelas Keramik", Kategori: "peralatan", Harga: 45_000, Stok: 20},
	}
}

func main() {
	fmt.Println("=== samber/lo ===")
	demoMapFilterReduce()
	demoPengelompokan()
	demoUniqChunk()
	demoPencarian()
	demoPointer()
	demoKapanJangan()
}

// ------------------------------------------------------------------
// 1. Map, Filter, Reduce — tiga tiang utama
// ------------------------------------------------------------------

// 🔍 Analogi tiga tiang:
//   Map    = MESIN CETAK LABEL — tiap barang masuk, keluar jadi bentuk lain (jumlah tetap).
//   Filter = SARINGAN PASIR    — hanya yang lolos syarat yang tersisa (jumlah menyusut).
//   Reduce = MESIN KASIR       — seluruh barang diringkas jadi SATU angka (total).

// NamaProduk mengubah []Produk jadi []string. Perhatikan lambdanya menerima (item, index).
func NamaProduk(ps []Produk) []string {
	return lo.Map(ps, func(p Produk, _ int) string {
		return p.Nama
	})
}

// ProdukTersedia menyaring yang stoknya masih ada.
func ProdukTersedia(ps []Produk) []Produk {
	return lo.Filter(ps, func(p Produk, _ int) bool {
		return p.Stok > 0
	})
}

// TotalNilaiStok meringkas seluruh produk jadi satu angka: harga x stok.
func TotalNilaiStok(ps []Produk) int {
	return lo.Reduce(ps, func(total int, p Produk, _ int) int {
		return total + p.Harga*p.Stok
	}, 0)
}

// JumlahHabis menghitung berapa produk yang stoknya nol tanpa membuat slice baru.
func JumlahHabis(ps []Produk) int {
	return lo.CountBy(ps, func(p Produk) bool { return p.Stok == 0 })
}

func demoMapFilterReduce() {
	ps := dataContoh()
	fmt.Println("\n-- Map / Filter / Reduce --")
	fmt.Println("   nama       :", NamaProduk(ps))
	fmt.Println("   tersedia   :", NamaProduk(ProdukTersedia(ps)))
	fmt.Printf("   nilai stok : Rp%d\n", TotalNilaiStok(ps))
	fmt.Println("   stok habis :", JumlahHabis(ps), "produk")
}

// ------------------------------------------------------------------
// 2. GroupBy & KeyBy — dua cara mengubah slice jadi map
// ------------------------------------------------------------------

// 🔍 Analogi: bayangkan setumpuk kartu nama.
//   GroupBy = menyusunnya ke LACI PER KOTA — satu laci berisi BANYAK kartu.
//   KeyBy   = menyusunnya ke KOTAK PER NOMOR ID — satu kotak berisi TEPAT SATU kartu.
// Jebakan KeyBy: kalau kuncinya ternyata kembar, yang belakangan MENIMPA yang sebelumnya
// tanpa peringatan apa pun. Pakai KeyBy hanya bila kuncinya dijamin unik (mis. primary key).

// KelompokPerKategori: satu kategori -> banyak produk.
func KelompokPerKategori(ps []Produk) map[string][]Produk {
	return lo.GroupBy(ps, func(p Produk) string { return p.Kategori })
}

// IndeksPerID: satu ID -> satu produk (untuk pencarian O(1)).
func IndeksPerID(ps []Produk) map[int]Produk {
	return lo.KeyBy(ps, func(p Produk) int { return p.ID })
}

// PetaNamaHarga memakai Associate untuk menentukan kunci DAN nilai sekaligus.
func PetaNamaHarga(ps []Produk) map[string]int {
	return lo.Associate(ps, func(p Produk) (string, int) {
		return p.Nama, p.Harga
	})
}

func demoPengelompokan() {
	ps := dataContoh()
	fmt.Println("\n-- GroupBy / KeyBy / Associate --")

	// Kunci map diurutkan dulu supaya keluaran konsisten tiap kali dijalankan.
	// (Ingat: urutan iterasi map di Go sengaja diacak.)
	grup := KelompokPerKategori(ps)
	kategori := make([]string, 0, len(grup))
	for k := range grup {
		kategori = append(kategori, k)
	}
	slices.Sort(kategori)
	for _, k := range kategori {
		fmt.Printf("   %-10s -> %s\n", k, strings.Join(NamaProduk(grup[k]), ", "))
	}

	fmt.Println("   cari ID 3   ->", IndeksPerID(ps)[3].Nama)
	fmt.Println("   harga Teh   ->", PetaNamaHarga(ps)["Teh Melati"])
}

// ------------------------------------------------------------------
// 3. Uniq & Chunk
// ------------------------------------------------------------------

// KategoriUnik membuang duplikat sambil MEMPERTAHANKAN urutan kemunculan pertama.
//
// 🔍 Analogi: seperti daftar hadir yang mencoret nama kembar — tapi urutan datangnya
// tetap terjaga. Ini beda dengan trik map+loop yang urutannya jadi acak.
func KategoriUnik(ps []Produk) []string {
	return lo.Uniq(lo.Map(ps, func(p Produk, _ int) string { return p.Kategori }))
}

// SatuPerKategori mengambil satu wakil dari tiap kategori (yang pertama ditemui).
func SatuPerKategori(ps []Produk) []Produk {
	return lo.UniqBy(ps, func(p Produk) string { return p.Kategori })
}

// Halaman memecah daftar jadi potongan berukuran n.
//
// 🔍 Analogi: seperti membagi 500 lembar undangan ke amplop-amplop berisi 50 lembar.
// Amplop terakhir boleh kurang dari 50. Berguna untuk pagination atau batch insert
// ke database (kirim 1000 baris sekaligus, bukan satu per satu).
func Halaman(ps []Produk, n int) [][]Produk {
	return lo.Chunk(ps, n)
}

func demoUniqChunk() {
	ps := dataContoh()
	fmt.Println("\n-- Uniq / UniqBy / Chunk --")
	fmt.Println("   kategori unik  :", KategoriUnik(ps))
	fmt.Println("   wakil kategori :", NamaProduk(SatuPerKategori(ps)))
	for i, hal := range Halaman(ps, 2) {
		fmt.Printf("   halaman %d      : %v\n", i+1, NamaProduk(hal))
	}
}

// ------------------------------------------------------------------
// 4. Pencarian
// ------------------------------------------------------------------

// CariTermurahDiKategori mencari produk pertama yang cocok syarat.
// Find mengembalikan (nilai, ketemu) — perhatikan lambdanya TANPA index.
func CariPertama(ps []Produk, kategori string) (Produk, bool) {
	return lo.Find(ps, func(p Produk) bool { return p.Kategori == kategori })
}

// Termahal memakai MaxBy dengan pembanding eksplisit.
func Termahal(ps []Produk) Produk {
	return lo.MaxBy(ps, func(a, b Produk) bool { return a.Harga > b.Harga })
}

// AdaKategori memakai Contains pada slice string.
func AdaKategori(ps []Produk, kategori string) bool {
	return lo.Contains(KategoriUnik(ps), kategori)
}

func demoPencarian() {
	ps := dataContoh()
	fmt.Println("\n-- Find / MaxBy / Contains --")

	if p, ok := CariPertama(ps, "makanan"); ok {
		fmt.Println("   makanan pertama :", p.Nama)
	}
	if _, ok := CariPertama(ps, "elektronik"); !ok {
		fmt.Println("   kategori elektronik tidak ada (ok=false, bukan panic)")
	}
	fmt.Println("   termahal        :", Termahal(ps).Nama)
	fmt.Println("   ada 'peralatan'?", AdaKategori(ps, "peralatan"))
}

// ------------------------------------------------------------------
// 5. Helper pointer & ternary
// ------------------------------------------------------------------

// 🔍 Analogi ToPtr: di Go kamu tak bisa menulis &"halo" — nilai literal tak punya alamat.
// Itu menyebalkan saat mengisi struct JSON yang membedakan "kosong" dan "tidak dikirim".
// ToPtr itu seperti MENEMPELKAN LABEL ALAMAT pada barang yang tadinya tak punya alamat.

// DiskonOpsional mengembalikan pointer: nil berarti "tidak ada diskon",
// berbeda dari 0 yang berarti "diskon nol persen".
func DiskonOpsional(persen int) *int {
	if persen <= 0 {
		return nil
	}
	return lo.ToPtr(persen)
}

// BacaDiskon membaca pointer dengan aman — nil jadi 0, tanpa risiko panic.
func BacaDiskon(p *int) int {
	return lo.FromPtr(p)
}

// LabelStok memakai Ternary sebagai pengganti if-else satu baris.
//
// Peringatan penting: Ternary MENGEVALUASI KEDUA cabang lebih dulu (karena keduanya
// argumen fungsi biasa). Jadi jangan pakai bila salah satu cabang mahal atau bisa panic —
// mis. lo.Ternary(p != nil, p.Nama, "") tetap akan panic saat p nil. Untuk kasus itu,
// pakai if biasa.
func LabelStok(stok int) string {
	return lo.Ternary(stok > 0, "tersedia", "habis")
}

func demoPointer() {
	fmt.Println("\n-- ToPtr / FromPtr / Ternary --")
	fmt.Printf("   diskon 15%%   -> %d\n", BacaDiskon(DiskonOpsional(15)))
	fmt.Printf("   tanpa diskon -> nil? %t, dibaca aman jadi %d\n",
		DiskonOpsional(0) == nil, BacaDiskon(DiskonOpsional(0)))
	fmt.Println("   label stok 0 :", LabelStok(0))
}

// ------------------------------------------------------------------
// 6. Kapan JANGAN pakai lo
// ------------------------------------------------------------------

// 🔍 Analogi: memakai lo untuk satu loop sederhana itu seperti menyalakan food processor
// untuk memotong SATU wortel — lebih repot mencuci alatnya daripada memotong manual.
//
// Tiga alasan tetap memilih for-loop biasa:
//  1. lo.Map/Filter SELALU mengalokasikan slice baru. Rantai Map->Filter->Map membuat
//     3 slice sekali jalan; for-loop tunggal cukup satu.
//  2. Tidak ada early return. Di dalam lambda kamu tak bisa "break" atau "return err" ke
//     fungsi luar — padahal itu pola paling umum di Go.
//  3. Banyak kebutuhan sudah dijawab stdlib sejak Go 1.21: paket slices & maps.
//     Kalau bisa pakai stdlib, pakai stdlib — nol dependensi.

// TotalNilaiStokManual: versi for-loop dari TotalNilaiStok.
// Sama benar, nol alokasi tambahan, dan bisa dibaca siapa pun tanpa mengenal lo.
func TotalNilaiStokManual(ps []Produk) int {
	total := 0
	for _, p := range ps {
		total += p.Harga * p.Stok
	}
	return total
}

// KategoriUnikStdlib: versi stdlib (Go 1.21+) tanpa dependensi apa pun.
// Bedanya dengan lo.Uniq: hasilnya diurutkan, bukan mengikuti urutan kemunculan.
func KategoriUnikStdlib(ps []Produk) []string {
	set := make(map[string]struct{}, len(ps))
	out := make([]string, 0, len(ps))
	for _, p := range ps {
		if _, ada := set[p.Kategori]; ada {
			continue
		}
		set[p.Kategori] = struct{}{}
		out = append(out, p.Kategori)
	}
	slices.Sort(out)
	return out
}

func demoKapanJangan() {
	ps := dataContoh()
	fmt.Println("\n-- Kapan JANGAN pakai lo --")
	fmt.Printf("   lo.Reduce    = %d\n", TotalNilaiStok(ps))
	fmt.Printf("   for-loop     = %d  (hasil sama, tanpa dependensi)\n", TotalNilaiStokManual(ps))
	fmt.Println("   uniq stdlib  =", KategoriUnikStdlib(ps), "(terurut)")
	fmt.Println("   uniq lo      =", KategoriUnik(ps), "(urutan kemunculan)")
	fmt.Println("\n   Pedoman: pakai lo saat rantai transformasinya panjang & jelas;")
	fmt.Println("   pakai for-loop saat butuh early return, error handling, atau performa.")
}
