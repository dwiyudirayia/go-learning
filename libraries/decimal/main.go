// shopspring/decimal — hitung uang tanpa kehilangan satu sen pun.
//
// Jalankan: go run ./libraries/decimal
// Test:     go test ./libraries/decimal
//
// 🔍 Analogi besar: float64 itu seperti PENGGARIS PLASTIK yang cuma punya garis untuk
// pecahan 1/2, 1/4, 1/8, 1/16... Kamu bisa mengukur 0,5 cm dengan sempurna, tapi 0,1 cm?
// Tidak ada garisnya — kamu terpaksa memilih garis TERDEKAT. Selisih sepersejuta itu tak
// terasa saat mengukur meja, tapi kalau dikalikan sejuta transaksi, uang perusahaan bocor.
//
// decimal.Decimal itu penggaris yang garisnya berbasis SEPULUH (0,1 / 0,01 / 0,001) —
// persis seperti cara manusia dan akuntan menghitung uang. Harganya: lebih lambat dan
// lebih boros memori daripada float64. Untuk uang, itu harga yang murah.
package main

import (
	"fmt"

	"github.com/shopspring/decimal"
)

func main() {
	fmt.Println("=== shopspring/decimal ===")
	demoJebakanFloat()
	demoDasar()
	demoPembulatan()
	demoKasusNyata()
	demoPerbandingan()
	demoAlternatif()
}

// ------------------------------------------------------------------
// 1. Bukti nyata: kenapa float64 tidak boleh dipakai untuk uang
// ------------------------------------------------------------------

// FloatJumlah adalah versi float64 dari 0,1 + 0,2.
// Hasilnya BUKAN 0,3 — melainkan 0.30000000000000004.
//
// 🔍 Analogi (kejutan khas Go!): perhatikan nilainya sengaja dimasukkan ke VARIABEL dulu.
// Kalau ditulis "return 0.1 + 0.2" langsung, hasilnya justru tepat 0,3 — karena itu
// ekspresi KONSTANTA, dan kompiler Go menghitung konstanta dengan presisi tak terbatas
// (seperti guru matematika yang mengerjakan soal dengan pecahan, bukan dengan kalkulator).
// Galat baru lahir saat angka itu dituang ke wadah float64 yang berukuran terbatas.
// Di aplikasi nyata, uang SELALU datang dari variabel (input pengguna, database),
// jadi galat itu selalu terjadi.
func FloatJumlah() float64 {
	a, b := 0.1, 0.2
	return a + b
}

// DecimalJumlah versi decimal dari perhitungan yang sama. Hasilnya persis 0,3.
func DecimalJumlah() decimal.Decimal {
	a := decimal.RequireFromString("0.1")
	b := decimal.RequireFromString("0.2")
	return a.Add(b)
}

// FloatAkumulasi menjumlahkan 0,01 sebanyak n kali dengan float64.
//
// 🔍 Analogi: ini seperti menimbang gula 100 kali, masing-masing 10 gram. Timbangan yang
// meleset 0,0001 gram tiap penimbangan tampak sempurna sekali-dua kali — tapi setelah
// 100 kali, selisihnya jadi nyata. Beginilah galat float64 menumpuk di sistem keuangan.
func FloatAkumulasi(n int) float64 {
	var total float64
	for i := 0; i < n; i++ {
		total += 0.01
	}
	return total
}

// DecimalAkumulasi versi decimal — tetap tepat berapa pun n-nya.
func DecimalAkumulasi(n int) decimal.Decimal {
	total := decimal.Zero
	satuan := decimal.RequireFromString("0.01")
	for i := 0; i < n; i++ {
		total = total.Add(satuan)
	}
	return total
}

func demoJebakanFloat() {
	fmt.Println("\n-- Jebakan float64 --")
	fmt.Printf("   float64  : 0.1 + 0.2 = %.17f\n", FloatJumlah())
	fmt.Printf("   decimal  : 0.1 + 0.2 = %s\n", DecimalJumlah())
	fmt.Printf("   float64  : 0.01 x 1000 = %.17f\n", FloatAkumulasi(1000))
	fmt.Printf("   decimal  : 0.01 x 1000 = %s\n", DecimalAkumulasi(1000))
}

// ------------------------------------------------------------------
// 2. Membuat & menghitung
// ------------------------------------------------------------------

// 🔍 Analogi cara membuat Decimal — ada tiga pintu masuk, dan salah pintu = galat masuk:
//
//	NewFromString("19.99")  = MENYALIN ANGKA DARI STRUK. Tepat. INI yang dipakai untuk
//	                         input pengguna, kolom database, dan payload JSON.
//	NewFromInt(20)          = angka bulat, selalu aman.
//	NewFromFloat(19.99)     = MEMFOTO angka yang sudah buram. Galat float64 sudah terjadi
//	                         SEBELUM decimal menerimanya — decimal tak bisa memperbaikinya.
//
// Aturannya: jangan pernah biarkan uang menyentuh float64, bahkan sedetik pun.

// DariString membuat Decimal dari teks — jalur yang benar untuk uang.
func DariString(s string) (decimal.Decimal, error) {
	d, err := decimal.NewFromString(s)
	if err != nil {
		return decimal.Zero, fmt.Errorf("nilai uang %q tidak valid: %w", s, err)
	}
	return d, nil
}

// Subtotal menghitung harga x jumlah.
func Subtotal(harga decimal.Decimal, jumlah int) decimal.Decimal {
	return harga.Mul(decimal.NewFromInt(int64(jumlah)))
}

// Persen menghitung n persen dari sebuah nilai (mis. PPN 11%).
func Persen(nilai decimal.Decimal, persen string) decimal.Decimal {
	p := decimal.RequireFromString(persen)
	return nilai.Mul(p).Div(decimal.NewFromInt(100))
}

func demoDasar() {
	fmt.Println("\n-- Membuat & menghitung --")

	harga, err := DariString("19.99")
	if err != nil {
		fmt.Println("   error:", err)
		return
	}
	fmt.Println("   harga satuan :", harga)
	fmt.Println("   x 3          :", Subtotal(harga, 3))
	fmt.Println("   PPN 11%      :", Persen(Subtotal(harga, 3), "11"))

	if _, err := DariString("sembilan ribu"); err != nil {
		fmt.Println("   input kotor  ->", err)
	}
}

// ------------------------------------------------------------------
// 3. Pembulatan — di sinilah uang benar-benar hilang atau tercipta
// ------------------------------------------------------------------

// 🔍 Analogi Round vs RoundBank:
//
//	Round     = pembulatan sekolah dasar: 0,5 SELALU naik. Masalahnya, kalau kamu
//	            membulatkan jutaan transaksi, "selalu naik" membuat totalnya menggelembung
//	            sedikit demi sedikit — bank menyebutnya rounding bias.
//	RoundBank = pembulatan akuntan (banker's rounding): 0,5 dibulatkan ke angka GENAP
//	            terdekat. 2,5 -> 2 dan 3,5 -> 4. Karena kadang naik kadang turun,
//	            galatnya saling meniadakan dalam jangka panjang.
//
// Untuk laporan keuangan, RoundBank biasanya yang diminta auditor.

// BulatkanBiasa membulatkan ke n angka di belakang koma (0,5 selalu naik).
func BulatkanBiasa(d decimal.Decimal, tempat int32) decimal.Decimal {
	return d.Round(tempat)
}

// BulatkanBank memakai banker's rounding.
func BulatkanBank(d decimal.Decimal, tempat int32) decimal.Decimal {
	return d.RoundBank(tempat)
}

// Potong memotong tanpa membulatkan (truncate) — dipakai bila aturan bisnis bilang
// "sisa di bawah satu rupiah tidak dihitung".
func Potong(d decimal.Decimal, tempat int32) decimal.Decimal {
	return d.Truncate(tempat)
}

func demoPembulatan() {
	fmt.Println("\n-- Pembulatan --")
	fmt.Printf("   %-8s %-10s %-10s %s\n", "nilai", "Round", "RoundBank", "Truncate")
	for _, s := range []string{"2.5", "3.5", "2.4", "2.6", "-2.5"} {
		d := decimal.RequireFromString(s)
		fmt.Printf("   %-8s %-10s %-10s %s\n",
			s, BulatkanBiasa(d, 0), BulatkanBank(d, 0), Potong(d, 0))
	}
}

// ------------------------------------------------------------------
// 4. Kasus nyata: struk belanja & bagi rata
// ------------------------------------------------------------------

// Item satu baris pada struk.
type Item struct {
	Nama   string
	Harga  decimal.Decimal
	Jumlah int
}

// Struk hasil perhitungan lengkap.
type Struk struct {
	Subtotal decimal.Decimal
	Diskon   decimal.Decimal
	PPN      decimal.Decimal
	Total    decimal.Decimal
}

// HitungStruk menjumlahkan item, memotong diskon, lalu menambahkan PPN.
//
// Urutan operasi ini adalah keputusan BISNIS, bukan teknis: PPN dihitung setelah diskon.
// Menukar urutannya mengubah total yang dibayar pelanggan — pastikan mengikuti aturan
// yang berlaku, dan tuliskan di test agar tak berubah diam-diam.
func HitungStruk(items []Item, persenDiskon, persenPPN string) Struk {
	subtotal := decimal.Zero
	for _, it := range items {
		subtotal = subtotal.Add(Subtotal(it.Harga, it.Jumlah))
	}

	diskon := Persen(subtotal, persenDiskon).RoundBank(2)
	setelahDiskon := subtotal.Sub(diskon)
	ppn := Persen(setelahDiskon, persenPPN).RoundBank(2)

	return Struk{
		Subtotal: subtotal,
		Diskon:   diskon,
		PPN:      ppn,
		Total:    setelahDiskon.Add(ppn),
	}
}

// BagiRata membagi tagihan ke n orang TANPA kehilangan satu sen pun.
//
// 🔍 Analogi: membagi kue seharga Rp10.000 untuk 3 orang. 10000/3 = 3333,33... Kalau tiap
// orang membayar 3333, total terkumpul cuma 9999 — ada Rp1 menguap. Fungsi ini memastikan
// sisa itu dibebankan ke orang pertama, sehingga jumlahnya SELALU pas. Di sistem keuangan
// sungguhan, "uang yang menguap" adalah bug yang membuat pembukuan tak seimbang.
func BagiRata(total decimal.Decimal, orang int) []decimal.Decimal {
	if orang <= 0 {
		return nil
	}
	n := decimal.NewFromInt(int64(orang))
	dasar := total.Div(n).Truncate(2) // dibulatkan ke bawah dulu

	bagian := make([]decimal.Decimal, orang)
	for i := range bagian {
		bagian[i] = dasar
	}
	// Sisa akibat pembulatan ke bawah dibebankan ke bagian pertama.
	sisa := total.Sub(dasar.Mul(n))
	bagian[0] = bagian[0].Add(sisa)
	return bagian
}

// Rupiah memformat Decimal jadi teks 2 angka di belakang koma.
func Rupiah(d decimal.Decimal) string {
	return "Rp" + d.StringFixed(2)
}

func demoKasusNyata() {
	fmt.Println("\n-- Struk belanja --")

	items := []Item{
		{Nama: "Kopi", Harga: decimal.RequireFromString("19.99"), Jumlah: 3},
		{Nama: "Roti", Harga: decimal.RequireFromString("12.50"), Jumlah: 2},
	}
	s := HitungStruk(items, "10", "11")
	fmt.Println("   subtotal :", Rupiah(s.Subtotal))
	fmt.Println("   diskon   :", Rupiah(s.Diskon))
	fmt.Println("   PPN      :", Rupiah(s.PPN))
	fmt.Println("   TOTAL    :", Rupiah(s.Total))

	fmt.Println("\n-- Bagi rata Rp10.000 untuk 3 orang --")
	bagian := BagiRata(decimal.RequireFromString("10000"), 3)
	jumlah := decimal.Zero
	for i, b := range bagian {
		fmt.Printf("   orang %d : %s\n", i+1, Rupiah(b))
		jumlah = jumlah.Add(b)
	}
	fmt.Println("   jumlah  :", Rupiah(jumlah), "<- pas, tak ada sen yang hilang")
}

// ------------------------------------------------------------------
// 5. Perbandingan — JANGAN pakai == pada Decimal
// ------------------------------------------------------------------

// 🔍 Analogi: Decimal itu struct berisi angka besar + posisi koma. "10.00" dan "10"
// bernilai SAMA tapi tersimpan berbeda di dalam (posisi komanya beda) — seperti dua orang
// menuliskan tinggi badan yang sama sebagai "170 cm" dan "1,70 m". Operator == memeriksa
// TULISANNYA, sedangkan .Equal() memeriksa NILAINYA. Untuk uang, selalu pakai .Equal/.Cmp.

// SamaNilai membandingkan dua Decimal secara benar.
func SamaNilai(a, b decimal.Decimal) bool {
	return a.Equal(b)
}

// Cukup mengecek apakah saldo mencukupi tagihan.
func Cukup(saldo, tagihan decimal.Decimal) bool {
	return saldo.Cmp(tagihan) >= 0 // -1 kurang, 0 pas, 1 lebih
}

func demoPerbandingan() {
	fmt.Println("\n-- Perbandingan --")
	a := decimal.RequireFromString("10.00")
	b := decimal.RequireFromString("10")

	fmt.Printf("   \"10.00\".Equal(\"10\") = %t  <- benar\n", SamaNilai(a, b))
	fmt.Printf("   saldo 50 cukup untuk 75? %t\n",
		Cukup(decimal.NewFromInt(50), decimal.NewFromInt(75)))
	fmt.Printf("   saldo 100 cukup untuk 100? %t\n",
		Cukup(decimal.NewFromInt(100), decimal.NewFromInt(100)))
}

// ------------------------------------------------------------------
// 6. Alternatif tanpa dependensi: simpan sebagai bilangan bulat
// ------------------------------------------------------------------

// 🔍 Analogi: alih-alih memakai penggaris desimal khusus, kamu bisa MENGUBAH SATUANNYA —
// hitung semuanya dalam sen, bukan rupiah. Rp19,99 disimpan sebagai 1999. Semua operasi
// jadi aritmetika bilangan bulat yang eksak, cepat, dan tanpa dependensi.
//
// Kelemahannya: pembagian tetap perlu perhatian khusus, kamu harus disiplin mengingat
// satuannya di SELURUH kode, dan formatnya harus dikonversi manual saat ditampilkan.
// Cocok untuk sistem sederhana dengan mata uang tunggal; decimal lebih nyaman bila
// kamu berurusan dengan banyak mata uang, pajak berlapis, atau kurs.

// SenKeTeks memformat jumlah dalam sen jadi teks rupiah.
func SenKeTeks(sen int64) string {
	neg := ""
	if sen < 0 {
		neg, sen = "-", -sen
	}
	return fmt.Sprintf("%sRp%d.%02d", neg, sen/100, sen%100)
}

func demoAlternatif() {
	fmt.Println("\n-- Alternatif: simpan dalam sen (int64) --")
	const hargaSen int64 = 1999
	fmt.Printf("   1 x %s\n", SenKeTeks(hargaSen))
	fmt.Printf("   3 x %s\n", SenKeTeks(hargaSen*3))
	fmt.Println("\n   Pedoman: pakai int64-sen untuk sistem sederhana satu mata uang;")
	fmt.Println("   pakai decimal saat ada pajak berlapis, kurs, atau pembagian rumit.")
}
