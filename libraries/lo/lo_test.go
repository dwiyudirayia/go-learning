package main

import (
	"slices"
	"testing"
)

func TestNamaProduk(t *testing.T) {
	got := NamaProduk(dataContoh())
	want := []string{"Kopi Arabika", "Teh Melati", "Roti Gandum", "Keju Cheddar", "Gelas Keramik"}
	if !slices.Equal(got, want) {
		t.Errorf("NamaProduk = %v, ingin %v", got, want)
	}

	// Map pada slice kosong menghasilkan slice kosong, bukan nil-panic.
	if got := NamaProduk(nil); len(got) != 0 {
		t.Errorf("NamaProduk(nil) = %v, ingin kosong", got)
	}
}

func TestProdukTersedia(t *testing.T) {
	got := NamaProduk(ProdukTersedia(dataContoh()))
	want := []string{"Kopi Arabika", "Roti Gandum", "Keju Cheddar", "Gelas Keramik"}
	if !slices.Equal(got, want) {
		t.Errorf("ProdukTersedia = %v, ingin %v (Teh Melati stok 0 harus tersaring)", got, want)
	}
}

func TestTotalNilaiStokSamaDenganVersiManual(t *testing.T) {
	ps := dataContoh()
	// 85000*12 + 35000*0 + 25000*7 + 120000*3 + 45000*20 = 2_455_000
	const want = 2_455_000

	if got := TotalNilaiStok(ps); got != want {
		t.Errorf("TotalNilaiStok = %d, ingin %d", got, want)
	}
	// Inti pelajarannya: lo.Reduce dan for-loop harus memberi jawaban identik.
	if got := TotalNilaiStokManual(ps); got != want {
		t.Errorf("TotalNilaiStokManual = %d, ingin %d", got, want)
	}
}

func TestJumlahHabis(t *testing.T) {
	if got := JumlahHabis(dataContoh()); got != 1 {
		t.Errorf("JumlahHabis = %d, ingin 1", got)
	}
	if got := JumlahHabis(nil); got != 0 {
		t.Errorf("JumlahHabis(nil) = %d, ingin 0", got)
	}
}

func TestKelompokPerKategori(t *testing.T) {
	grup := KelompokPerKategori(dataContoh())

	if len(grup) != 3 {
		t.Fatalf("ingin 3 kategori, dapat %d", len(grup))
	}
	if got := len(grup["minuman"]); got != 2 {
		t.Errorf("kategori minuman berisi %d produk, ingin 2", got)
	}
	if got := len(grup["peralatan"]); got != 1 {
		t.Errorf("kategori peralatan berisi %d produk, ingin 1", got)
	}
	// Kategori yang tak ada mengembalikan slice nil — aman di-range, bukan panic.
	if got := grup["elektronik"]; len(got) != 0 {
		t.Errorf("kategori tak dikenal harus kosong, dapat %v", got)
	}
}

func TestIndeksPerID(t *testing.T) {
	idx := IndeksPerID(dataContoh())
	if idx[3].Nama != "Roti Gandum" {
		t.Errorf("idx[3].Nama = %q, ingin Roti Gandum", idx[3].Nama)
	}
	// ID tak dikenal mengembalikan zero value Produk, BUKAN error.
	// Ini jebakan KeyBy: selalu cek dengan bentuk comma-ok bila ID bisa tak ada.
	if p, ada := idx[99]; ada || p.Nama != "" {
		t.Errorf("ID tak dikenal seharusnya zero value & ada=false, dapat %+v", p)
	}
}

// Membuktikan jebakan KeyBy: kunci kembar membuat data sebelumnya tertimpa diam-diam.
func TestKeyByMenimpaKunciKembar(t *testing.T) {
	kembar := []Produk{
		{ID: 1, Nama: "Pertama"},
		{ID: 1, Nama: "Kedua"},
	}
	idx := IndeksPerID(kembar)

	if len(idx) != 1 {
		t.Fatalf("ingin 1 entri (tertimpa), dapat %d", len(idx))
	}
	if idx[1].Nama != "Kedua" {
		t.Errorf("idx[1].Nama = %q, ingin Kedua (yang belakangan menimpa)", idx[1].Nama)
	}
}

func TestPetaNamaHarga(t *testing.T) {
	peta := PetaNamaHarga(dataContoh())
	if got := peta["Teh Melati"]; got != 35_000 {
		t.Errorf("harga Teh Melati = %d, ingin 35000", got)
	}
}

func TestKategoriUnikMenjagaUrutanKemunculan(t *testing.T) {
	got := KategoriUnik(dataContoh())
	want := []string{"minuman", "makanan", "peralatan"}
	if !slices.Equal(got, want) {
		t.Errorf("KategoriUnik = %v, ingin %v (urutan kemunculan pertama)", got, want)
	}

	// Versi stdlib memberi isi yang sama tapi TERURUT — beda perilaku yang perlu disadari.
	gotStd := KategoriUnikStdlib(dataContoh())
	wantStd := []string{"makanan", "minuman", "peralatan"}
	if !slices.Equal(gotStd, wantStd) {
		t.Errorf("KategoriUnikStdlib = %v, ingin %v (terurut)", gotStd, wantStd)
	}
}

func TestSatuPerKategori(t *testing.T) {
	got := NamaProduk(SatuPerKategori(dataContoh()))
	want := []string{"Kopi Arabika", "Roti Gandum", "Gelas Keramik"}
	if !slices.Equal(got, want) {
		t.Errorf("SatuPerKategori = %v, ingin %v", got, want)
	}
}

func TestHalaman(t *testing.T) {
	tests := []struct {
		nama         string
		ukuran       int
		wantPotong   int
		wantTerakhir int
	}{
		{"pas habis dibagi", 5, 1, 5},
		{"sisa di potongan terakhir", 2, 3, 1},
		{"ukuran lebih besar dari data", 10, 1, 5},
		{"satu per potongan", 1, 5, 1},
	}

	for _, tt := range tests {
		t.Run(tt.nama, func(t *testing.T) {
			hal := Halaman(dataContoh(), tt.ukuran)
			if len(hal) != tt.wantPotong {
				t.Fatalf("jumlah potongan = %d, ingin %d", len(hal), tt.wantPotong)
			}
			if got := len(hal[len(hal)-1]); got != tt.wantTerakhir {
				t.Errorf("potongan terakhir berisi %d, ingin %d", got, tt.wantTerakhir)
			}
		})
	}
}

func TestCariPertama(t *testing.T) {
	p, ok := CariPertama(dataContoh(), "makanan")
	if !ok {
		t.Fatal("ingin ketemu, dapat ok=false")
	}
	if p.Nama != "Roti Gandum" {
		t.Errorf("produk = %q, ingin Roti Gandum (yang pertama cocok)", p.Nama)
	}

	if _, ok := CariPertama(dataContoh(), "elektronik"); ok {
		t.Error("kategori tak ada seharusnya ok=false")
	}
}

func TestTermahal(t *testing.T) {
	if got := Termahal(dataContoh()).Nama; got != "Keju Cheddar" {
		t.Errorf("Termahal = %q, ingin Keju Cheddar", got)
	}
}

func TestAdaKategori(t *testing.T) {
	if !AdaKategori(dataContoh(), "peralatan") {
		t.Error("peralatan seharusnya ada")
	}
	if AdaKategori(dataContoh(), "elektronik") {
		t.Error("elektronik seharusnya tidak ada")
	}
}

func TestDiskonOpsionalMembedakanNilDanNol(t *testing.T) {
	tests := []struct {
		nama    string
		persen  int
		wantNil bool
		wantVal int
	}{
		{"diskon wajar", 15, false, 15},
		{"nol berarti tidak ada diskon", 0, true, 0},
		{"negatif dianggap tidak ada", -5, true, 0},
		{"diskon penuh", 100, false, 100},
	}

	for _, tt := range tests {
		t.Run(tt.nama, func(t *testing.T) {
			p := DiskonOpsional(tt.persen)
			if (p == nil) != tt.wantNil {
				t.Fatalf("nil? %t, ingin %t", p == nil, tt.wantNil)
			}
			// FromPtr membaca pointer nil dengan aman (jadi 0), tanpa panic.
			if got := BacaDiskon(p); got != tt.wantVal {
				t.Errorf("BacaDiskon = %d, ingin %d", got, tt.wantVal)
			}
		})
	}
}

func TestLabelStok(t *testing.T) {
	if got := LabelStok(3); got != "tersedia" {
		t.Errorf("LabelStok(3) = %q, ingin tersedia", got)
	}
	if got := LabelStok(0); got != "habis" {
		t.Errorf("LabelStok(0) = %q, ingin habis", got)
	}
}
