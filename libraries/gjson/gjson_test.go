package main

import (
	"errors"
	"slices"
	"testing"

	"github.com/tidwall/gjson"
)

func TestBacaNilai(t *testing.T) {
	if got := NamaPelanggan(dataPesanan); got != "Ana Pratiwi" {
		t.Errorf("nama = %q, ingin Ana Pratiwi", got)
	}
	if got := KotaPelanggan(dataPesanan); got != "Bandung" {
		t.Errorf("kota = %q, ingin Bandung", got)
	}
	if StatusLunas(dataPesanan) {
		t.Error("lunas seharusnya false")
	}
}

// Path yang tak ada menghasilkan zero value, BUKAN panic.
func TestPathTakAdaMengembalikanZeroValue(t *testing.T) {
	if got := gjson.Get(dataPesanan, "tidak.ada.sama.sekali").String(); got != "" {
		t.Errorf("path tak ada = %q, ingin string kosong", got)
	}
	if got := gjson.Get(dataPesanan, "tidak.ada").Int(); got != 0 {
		t.Errorf("path tak ada = %d, ingin 0", got)
	}
}

func TestIndeksArray(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"item.0.nama", "Kopi Arabika"},
		{"item.1.nama", "Teh Melati"},
		{"item.2.nama", "Gula Aren"},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if got := gjson.Get(dataPesanan, tt.path).String(); got != tt.want {
				t.Errorf("%s = %q, ingin %q", tt.path, got, tt.want)
			}
		})
	}
}

func TestQueryArray(t *testing.T) {
	if got := JumlahItem(dataPesanan); got != 3 {
		t.Errorf("jumlah item = %d, ingin 3", got)
	}

	nama := SemuaNamaItem(dataPesanan)
	want := []string{"Kopi Arabika", "Teh Melati", "Gula Aren"}
	if !slices.Equal(nama, want) {
		t.Errorf("nama item = %v, ingin %v", nama, want)
	}
}

func TestTotalHarga(t *testing.T) {
	// 85000*2 + 35000*1 + 15000*3 = 170000 + 35000 + 45000 = 250000
	if got := TotalHarga(dataPesanan); got != 250_000 {
		t.Errorf("total = %d, ingin 250000", got)
	}
}

func TestItemMahalFilter(t *testing.T) {
	tests := []struct {
		nama   string
		ambang int
		want   []string
	}{
		{"di atas 30rb", 30_000, []string{"Kopi Arabika", "Teh Melati"}},
		{"di atas 50rb", 50_000, []string{"Kopi Arabika"}},
		{"di atas 100rb (kosong)", 100_000, nil},
		{"di atas 0 (semua)", 0, []string{"Kopi Arabika", "Teh Melati", "Gula Aren"}},
	}
	for _, tt := range tests {
		t.Run(tt.nama, func(t *testing.T) {
			got := ItemMahal(dataPesanan, tt.ambang)
			if !slices.Equal(got, tt.want) {
				t.Errorf("ItemMahal(%d) = %v, ingin %v", tt.ambang, got, tt.want)
			}
		})
	}
}

// Membedakan "tidak ada" / "null" / "false" — inti pemakaian Exists().
func TestExistsMembedakanTigaKeadaan(t *testing.T) {
	// kupon bernilai null -> ada, tapi PunyaKupon harus false.
	if PunyaKupon(dataPesanan) {
		t.Error("kupon null seharusnya dianggap 'tidak punya'")
	}

	// Field yang benar-benar ada.
	if !PunyaField(dataPesanan, "pelanggan.email") {
		t.Error("pelanggan.email seharusnya ada")
	}
	// Field yang tak ada sama sekali.
	if PunyaField(dataPesanan, "pelanggan.telepon") {
		t.Error("pelanggan.telepon seharusnya tidak ada")
	}
	// 'lunas' ada tapi bernilai false -> tetap Exists()=true.
	if !PunyaField(dataPesanan, "lunas") {
		t.Error("lunas ada (walau bernilai false)")
	}
}

func TestTandaiLunasImutabel(t *testing.T) {
	baru, err := TandaiLunas(dataPesanan)
	if err != nil {
		t.Fatalf("TandaiLunas gagal: %v", err)
	}

	if !gjson.Get(baru, "lunas").Bool() {
		t.Error("hasil seharusnya lunas=true")
	}
	// Sifat imutabel: string ASLI tidak boleh ikut berubah.
	if gjson.Get(dataPesanan, "lunas").Bool() {
		t.Error("data asli seharusnya tetap lunas=false — sjson harus imutabel")
	}
}

func TestPasangKupon(t *testing.T) {
	baru, err := PasangKupon(dataPesanan, "HEMAT10")
	if err != nil {
		t.Fatalf("PasangKupon gagal: %v", err)
	}
	if got := gjson.Get(baru, "kupon").String(); got != "HEMAT10" {
		t.Errorf("kupon = %q, ingin HEMAT10", got)
	}
	// Sekarang PunyaKupon harus true pada hasil baru.
	if !PunyaKupon(baru) {
		t.Error("setelah dipasang, kupon seharusnya dianggap ada")
	}
}

// sjson membuat path bersarang yang belum ada sama sekali.
func TestTambahCatatanMembuatPathBaru(t *testing.T) {
	baru, err := TambahCatatan(dataPesanan, "kirim pagi")
	if err != nil {
		t.Fatalf("TambahCatatan gagal: %v", err)
	}
	if got := gjson.Get(baru, "meta.catatan").String(); got != "kirim pagi" {
		t.Errorf("meta.catatan = %q, ingin 'kirim pagi'", got)
	}
	// Asli tak punya meta sama sekali.
	if gjson.Get(dataPesanan, "meta").Exists() {
		t.Error("data asli seharusnya tak punya field meta")
	}
}

func TestHapusField(t *testing.T) {
	baru, err := HapusField(dataPesanan, "pelanggan.email")
	if err != nil {
		t.Fatalf("HapusField gagal: %v", err)
	}
	if gjson.Get(baru, "pelanggan.email").Exists() {
		t.Error("email seharusnya sudah terhapus")
	}
	// Field lain di objek yang sama tetap utuh.
	if gjson.Get(baru, "pelanggan.nama").String() != "Ana Pratiwi" {
		t.Error("nama seharusnya tetap ada setelah email dihapus")
	}
}

// JEBAKAN: gjson.Get TIDAK memvalidasi. AmbilAman harus menolak JSON rusak.
func TestAmbilAmanMenolakJSONRusak(t *testing.T) {
	tests := []struct {
		nama    string
		json    string
		wantErr bool
	}{
		{"valid", dataPesanan, false},
		{"kurung tak tertutup", "{ini rusak", true},
		{"kosong", "", true},
		{"teks biasa", "halo dunia", true},
		{"array valid", `[1,2,3]`, false},
	}
	for _, tt := range tests {
		t.Run(tt.nama, func(t *testing.T) {
			_, err := AmbilAman(tt.json, "pelanggan.nama")
			if tt.wantErr {
				if !errors.Is(err, ErrJSONTidakValid) {
					t.Errorf("error = %v, ingin ErrJSONTidakValid", err)
				}
				return
			}
			if err != nil {
				t.Errorf("JSON valid seharusnya tanpa error, dapat: %v", err)
			}
		})
	}
}

func TestForEachBisaBerhentiLebihAwal(t *testing.T) {
	// Membuktikan return false menghentikan iterasi (ambil hanya item pertama).
	var terkumpul []string
	gjson.Get(dataPesanan, "item").ForEach(func(_, item gjson.Result) bool {
		terkumpul = append(terkumpul, item.Get("nama").String())
		return false // berhenti setelah yang pertama
	})
	if len(terkumpul) != 1 {
		t.Errorf("mengumpulkan %d item, ingin 1 (ForEach harus berhenti saat return false)", len(terkumpul))
	}
}
