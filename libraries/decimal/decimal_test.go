package main

import (
	"testing"

	"github.com/shopspring/decimal"
)

// dec pintasan agar test enak dibaca.
func dec(s string) decimal.Decimal { return decimal.RequireFromString(s) }

// Test ini MEMBUKTIKAN kenapa modul ini ada: float64 salah, decimal benar.
func TestFloatSalahDecimalBenar(t *testing.T) {
	// Catatan penting: konstanta 0.3 di sebelah kanan dikonversi ke float64 saat
	// dibandingkan, jadi ini benar-benar membandingkan dua nilai float64.
	if FloatJumlah() == 0.3 {
		t.Error("float64 0.1+0.2 seharusnya TIDAK persis 0.3 — asumsi seluruh file ini runtuh")
	}
	// Sebaliknya, ekspresi KONSTANTA dihitung kompiler dengan presisi tak terbatas.
	// Ini bukan sekadar remeh: ia menjelaskan kenapa contoh float yang keliru kadang
	// tampak "benar" saat dicoba sekilas di playground.
	const konstan = 0.1 + 0.2
	if konstan != 0.3 {
		t.Error("ekspresi konstanta seharusnya tepat 0.3 — dihitung kompiler, bukan CPU")
	}
	if !DecimalJumlah().Equal(dec("0.3")) {
		t.Errorf("decimal 0.1+0.2 = %s, ingin tepat 0.3", DecimalJumlah())
	}
}

func TestAkumulasiGalatMenumpuk(t *testing.T) {
	// 0,01 ditambahkan 1000 kali seharusnya = 10, tapi float64 meleset.
	if FloatAkumulasi(1000) == 10.0 {
		t.Error("float64 seharusnya meleset setelah 1000 penjumlahan")
	}
	if !DecimalAkumulasi(1000).Equal(dec("10")) {
		t.Errorf("DecimalAkumulasi(1000) = %s, ingin tepat 10", DecimalAkumulasi(1000))
	}
	// Berapa pun jumlah pengulangannya, decimal tetap tepat.
	if !DecimalAkumulasi(0).Equal(decimal.Zero) {
		t.Error("DecimalAkumulasi(0) harus nol")
	}
}

func TestDariString(t *testing.T) {
	tests := []struct {
		nama    string
		input   string
		want    string
		wantErr bool
	}{
		{"desimal biasa", "19.99", "19.99", false},
		{"bilangan bulat", "20", "20", false},
		{"negatif", "-5.25", "-5.25", false},
		{"nol berkoma", "0.00", "0", false},
		{"presisi tinggi", "0.000000000000000001", "0.000000000000000001", false},
		{"teks bukan angka", "sembilan ribu", "", true},
		{"kosong", "", "", true},
		{"pakai koma bukan titik", "19,99", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.nama, func(t *testing.T) {
			got, err := DariString(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("DariString(%q) ingin error, dapat %s", tt.input, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("DariString(%q) error tak terduga: %v", tt.input, err)
			}
			if !got.Equal(dec(tt.want)) {
				t.Errorf("DariString(%q) = %s, ingin %s", tt.input, got, tt.want)
			}
		})
	}
}

func TestSubtotal(t *testing.T) {
	if got := Subtotal(dec("19.99"), 3); !got.Equal(dec("59.97")) {
		t.Errorf("Subtotal = %s, ingin 59.97", got)
	}
	if got := Subtotal(dec("19.99"), 0); !got.IsZero() {
		t.Errorf("Subtotal jumlah 0 = %s, ingin 0", got)
	}
}

func TestPersen(t *testing.T) {
	if got := Persen(dec("100"), "11"); !got.Equal(dec("11")) {
		t.Errorf("PPN 11%% dari 100 = %s, ingin 11", got)
	}
	if got := Persen(dec("84.97"), "10"); !got.Equal(dec("8.497")) {
		t.Errorf("diskon 10%% dari 84.97 = %s, ingin 8.497 (belum dibulatkan)", got)
	}
}

// Perbedaan Round vs RoundBank paling terlihat tepat di angka 0,5.
func TestPembulatan(t *testing.T) {
	tests := []struct {
		nilai        string
		wantRound    string
		wantBank     string
		wantTruncate string
	}{
		{"2.5", "3", "2", "2"},     // Round naik, Bank ke genap terdekat (2)
		{"3.5", "4", "4", "3"},     // di sini keduanya sepakat: 4 memang genap
		{"2.4", "2", "2", "2"},     // di bawah setengah: sama saja
		{"2.6", "3", "3", "2"},     // di atas setengah: sama saja
		{"-2.5", "-3", "-2", "-2"}, // negatif: Round menjauh dari nol
	}

	for _, tt := range tests {
		t.Run(tt.nilai, func(t *testing.T) {
			d := dec(tt.nilai)
			if got := BulatkanBiasa(d, 0); !got.Equal(dec(tt.wantRound)) {
				t.Errorf("Round(%s) = %s, ingin %s", tt.nilai, got, tt.wantRound)
			}
			if got := BulatkanBank(d, 0); !got.Equal(dec(tt.wantBank)) {
				t.Errorf("RoundBank(%s) = %s, ingin %s", tt.nilai, got, tt.wantBank)
			}
			if got := Potong(d, 0); !got.Equal(dec(tt.wantTruncate)) {
				t.Errorf("Truncate(%s) = %s, ingin %s", tt.nilai, got, tt.wantTruncate)
			}
		})
	}
}

// Mengunci aturan bisnis: PPN dihitung SETELAH diskon.
func TestHitungStruk(t *testing.T) {
	items := []Item{
		{Nama: "Kopi", Harga: dec("19.99"), Jumlah: 3}, // 59.97
		{Nama: "Roti", Harga: dec("12.50"), Jumlah: 2}, // 25.00
	}
	s := HitungStruk(items, "10", "11")

	cek := []struct {
		nama string
		got  decimal.Decimal
		want string
	}{
		{"subtotal", s.Subtotal, "84.97"},
		{"diskon 10%", s.Diskon, "8.50"}, // 8.497 dibulatkan
		{"PPN 11%", s.PPN, "8.41"},       // 11% dari 76.47
		{"total", s.Total, "84.88"},      // 76.47 + 8.41
	}
	for _, c := range cek {
		if !c.got.Equal(dec(c.want)) {
			t.Errorf("%s = %s, ingin %s", c.nama, c.got, c.want)
		}
	}
}

func TestHitungStrukKosong(t *testing.T) {
	s := HitungStruk(nil, "10", "11")
	if !s.Total.IsZero() || !s.Subtotal.IsZero() {
		t.Errorf("struk kosong harus nol semua, dapat total=%s", s.Total)
	}
}

// Sifat terpenting BagiRata: jumlah seluruh bagian HARUS sama persis dengan total.
func TestBagiRataTidakKehilanganSen(t *testing.T) {
	tests := []struct {
		total string
		orang int
	}{
		{"10000", 3},   // 3333,33... sisa 0,01
		{"100", 3},     // 33,33 sisa 0,01
		{"10", 4},      // 2,50 pas
		{"0.05", 3},    // sangat kecil, tetap harus pas
		{"99.99", 7},   // angka ganjil
		{"1000000", 1}, // satu orang
	}

	for _, tt := range tests {
		t.Run(tt.total+"/"+string(rune('0'+tt.orang)), func(t *testing.T) {
			total := dec(tt.total)
			bagian := BagiRata(total, tt.orang)

			if len(bagian) != tt.orang {
				t.Fatalf("jumlah bagian = %d, ingin %d", len(bagian), tt.orang)
			}
			jumlah := decimal.Zero
			for _, b := range bagian {
				jumlah = jumlah.Add(b)
			}
			if !jumlah.Equal(total) {
				t.Errorf("jumlah bagian = %s, ingin persis %s (ada sen yang hilang!)", jumlah, total)
			}
		})
	}
}

func TestBagiRataOrangTidakValid(t *testing.T) {
	if got := BagiRata(dec("100"), 0); got != nil {
		t.Errorf("BagiRata dengan 0 orang = %v, ingin nil", got)
	}
	if got := BagiRata(dec("100"), -3); got != nil {
		t.Errorf("BagiRata dengan orang negatif = %v, ingin nil", got)
	}
}

// Jebakan klasik: "10.00" dan "10" sama nilainya tapi beda representasi internal.
func TestPerbandinganNilaiBukanTulisan(t *testing.T) {
	a, b := dec("10.00"), dec("10")

	if !SamaNilai(a, b) {
		t.Error("10.00 dan 10 seharusnya sama nilainya")
	}
	// Bukti bahwa representasinya memang berbeda — inilah alasan == tak boleh dipakai.
	if a.String() == b.String() {
		t.Logf("catatan: String() kebetulan sama (%s); Equal tetap cara yang benar", a)
	}
}

func TestCukup(t *testing.T) {
	tests := []struct {
		nama    string
		saldo   string
		tagihan string
		want    bool
	}{
		{"saldo lebih", "100", "75", true},
		{"saldo pas", "100", "100", true},
		{"saldo kurang", "50", "75", false},
		{"beda representasi tapi pas", "100.00", "100", true},
		{"kurang sedikit sekali", "99.99", "100", false},
	}

	for _, tt := range tests {
		t.Run(tt.nama, func(t *testing.T) {
			if got := Cukup(dec(tt.saldo), dec(tt.tagihan)); got != tt.want {
				t.Errorf("Cukup(%s, %s) = %t, ingin %t", tt.saldo, tt.tagihan, got, tt.want)
			}
		})
	}
}

func TestRupiah(t *testing.T) {
	if got := Rupiah(dec("84.5")); got != "Rp84.50" {
		t.Errorf("Rupiah = %q, ingin Rp84.50 (selalu 2 angka desimal)", got)
	}
}

func TestSenKeTeks(t *testing.T) {
	tests := []struct {
		sen  int64
		want string
	}{
		{1999, "Rp19.99"},
		{100, "Rp1.00"},
		{5, "Rp0.05"},
		{0, "Rp0.00"},
		{-250, "-Rp2.50"},
	}
	for _, tt := range tests {
		if got := SenKeTeks(tt.sen); got != tt.want {
			t.Errorf("SenKeTeks(%d) = %q, ingin %q", tt.sen, got, tt.want)
		}
	}
}
