package main

import (
	"strings"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
)

func validatorUji(t *testing.T) *validator.Validate {
	t.Helper()
	v, err := NewValidator()
	if err != nil {
		t.Fatalf("NewValidator gagal: %v", err)
	}
	return v
}

// punyaField mencari apakah suatu field muncul di daftar kesalahan.
func punyaField(errs []KesalahanField, field string) (KesalahanField, bool) {
	for _, e := range errs {
		if e.Field == field {
			return e, true
		}
	}
	return KesalahanField{}, false
}

func TestDataBenarLolos(t *testing.T) {
	v := validatorUji(t)
	if errs := Validasi(v, contohValid()); len(errs) != 0 {
		t.Errorf("data yang benar seharusnya lolos, dapat: %+v", errs)
	}
}

// Tiap kasus merusak SATU field saja, lalu memastikan field itulah yang dilaporkan.
func TestSatuFieldRusak(t *testing.T) {
	v := validatorUji(t)

	tests := []struct {
		nama      string
		rusak     func(*Pendaftaran)
		wantField string
		wantPesan string // potongan pesan yang harus muncul
	}{
		{"nama kosong", func(p *Pendaftaran) { p.Nama = "" }, "nama", "wajib diisi"},
		{"nama terlalu pendek", func(p *Pendaftaran) { p.Nama = "Ab" }, "nama", "minimal 3 karakter"},
		{"email salah bentuk", func(p *Pendaftaran) { p.Email = "bukan-email" }, "email", "alamat email"},
		{"umur di bawah batas", func(p *Pendaftaran) { p.Umur = 12 }, "umur", "minimal 17"},
		{"umur di atas batas", func(p *Pendaftaran) { p.Umur = 150 }, "umur", "maksimal 99"},
		{"telepon bukan nomor Indonesia", func(p *Pendaftaran) { p.Telepon = "12345" }, "telepon", "nomor Indonesia"},
		{"peran di luar daftar", func(p *Pendaftaran) { p.Peran = "raja" }, "peran", "salah satu dari"},
		{"kata sandi pendek", func(p *Pendaftaran) { p.Kata = "pendek" }, "kata_sandi", "minimal 8 karakter"},
		{"konfirmasi tak cocok", func(p *Pendaftaran) { p.KonfirKata = "beda" }, "konfirmasi_kata_sandi", "harus sama dengan"},
		{"situs bukan url", func(p *Pendaftaran) { p.Situs = "bukan-url" }, "situs", "URL"},
		{"kode pos bukan 5 digit", func(p *Pendaftaran) { p.Alamat.KodePos = "123" }, "kode_pos", "tepat 5 karakter"},
		{"provinsi di luar daftar", func(p *Pendaftaran) { p.Alamat.Provinsi = "Papua" }, "provinsi", "salah satu dari"},
	}

	for _, tt := range tests {
		t.Run(tt.nama, func(t *testing.T) {
			p := contohValid()
			tt.rusak(&p)

			errs := Validasi(v, p)
			e, ada := punyaField(errs, tt.wantField)
			if !ada {
				t.Fatalf("ingin kesalahan di field %q, dapat: %+v", tt.wantField, errs)
			}
			if !strings.Contains(e.Pesan, tt.wantPesan) {
				t.Errorf("pesan = %q, ingin mengandung %q", e.Pesan, tt.wantPesan)
			}
		})
	}
}

// Nama field di pesan harus memakai nama JSON, bukan nama field Go.
func TestNamaFieldMemakaiTagJSON(t *testing.T) {
	v := validatorUji(t)

	p := contohValid()
	p.KonfirKata = "tidak cocok"

	errs := Validasi(v, p)
	if _, ada := punyaField(errs, "konfirmasi_kata_sandi"); !ada {
		t.Errorf("ingin nama JSON 'konfirmasi_kata_sandi', dapat: %+v", errs)
	}
	if _, ada := punyaField(errs, "KonfirKata"); ada {
		t.Error("nama field Go seharusnya tidak bocor ke pesan untuk pengguna")
	}
}

// omitempty: kolom opsional boleh kosong, tapi kalau diisi harus benar.
func TestOmitemptyPadaSitus(t *testing.T) {
	v := validatorUji(t)

	p := contohValid()
	p.Situs = "" // kosong -> harus lolos
	if _, ada := punyaField(Validasi(v, p), "situs"); ada {
		t.Error("situs kosong seharusnya lolos karena omitempty")
	}

	p.Situs = "bukan-url" // diisi tapi salah -> harus ditolak
	if _, ada := punyaField(Validasi(v, p), "situs"); !ada {
		t.Error("situs yang diisi salah bentuk seharusnya ditolak")
	}
}

// dive memeriksa SETIAP isi slice, bukan cuma slice-nya.
func TestDiveMemeriksaIsiSlice(t *testing.T) {
	v := validatorUji(t)

	tests := []struct {
		nama    string
		minat   []string
		wantErr bool
	}{
		{"semua sah", []string{"golang", "musik"}, false},
		{"satu isi terlalu pendek", []string{"golang", "go"}, true},
		{"ada isi kosong", []string{"golang", ""}, true},
		{"slice kosong ditolak", []string{}, true},
		{"slice nil ditolak", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.nama, func(t *testing.T) {
			p := contohValid()
			p.Minat = tt.minat

			errs := Validasi(v, p)
			ada := len(errs) > 0
			if ada != tt.wantErr {
				t.Errorf("ada kesalahan = %t, ingin %t (%+v)", ada, tt.wantErr, errs)
			}
		})
	}
}

// Struct bersarang yang kosong harus ikut dilaporkan, bukan dilewati diam-diam.
func TestAlamatKosongDitolak(t *testing.T) {
	v := validatorUji(t)

	p := contohValid()
	p.Alamat = Alamat{}

	if errs := Validasi(v, p); len(errs) == 0 {
		t.Error("alamat kosong seharusnya menghasilkan kesalahan")
	}
}

func TestValidasiTeleponID(t *testing.T) {
	v := validatorUji(t)

	tests := []struct {
		nomor string
		want  bool
	}{
		{"081234567890", true},     // bentuk paling umum
		{"+6281234567890", true},   // awalan +62
		{"6281234567890", true},    // awalan 62 tanpa plus
		{"0812345678", true},       // batas bawah 10 digit
		{"0812", false},            // terlalu pendek
		{"081234567890123", false}, // terlalu panjang
		{"12345678901", false},     // tak diawali 08
		{"0812-3456-7890", false},  // ada tanda hubung
		{"08123abc7890", false},    // ada huruf
		{"", false},                // kosong
	}

	for _, tt := range tests {
		t.Run(tt.nomor, func(t *testing.T) {
			err := v.Var(tt.nomor, "notelp_id")
			if got := err == nil; got != tt.want {
				t.Errorf("nomor %q sah = %t, ingin %t", tt.nomor, got, tt.want)
			}
		})
	}
}

func TestValidasiPemesananLintasField(t *testing.T) {
	v := validatorUji(t)
	mulai := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		nama    string
		p       Pemesanan
		wantErr bool
	}{
		{
			"menginap 3 hari",
			Pemesanan{Kode: "BK-1", Mulai: mulai, Selesai: mulai.Add(72 * time.Hour), Tamu: 2},
			false,
		},
		{
			"selesai sebelum mulai",
			Pemesanan{Kode: "BK-2", Mulai: mulai, Selesai: mulai.Add(-24 * time.Hour), Tamu: 2},
			true,
		},
		{
			"selesai sama dengan mulai",
			Pemesanan{Kode: "BK-3", Mulai: mulai, Selesai: mulai, Tamu: 2},
			true,
		},
		{
			"lebih dari 30 hari",
			Pemesanan{Kode: "BK-4", Mulai: mulai, Selesai: mulai.Add(60 * 24 * time.Hour), Tamu: 2},
			true,
		},
		{
			"tepat 30 hari masih boleh",
			Pemesanan{Kode: "BK-5", Mulai: mulai, Selesai: mulai.Add(30 * 24 * time.Hour), Tamu: 2},
			false,
		},
		{
			"tamu melebihi kapasitas",
			Pemesanan{Kode: "BK-6", Mulai: mulai, Selesai: mulai.Add(24 * time.Hour), Tamu: 50},
			true,
		},
		{
			"tanpa kode",
			Pemesanan{Mulai: mulai, Selesai: mulai.Add(24 * time.Hour), Tamu: 1},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.nama, func(t *testing.T) {
			errs := Validasi(v, tt.p)
			if ada := len(errs) > 0; ada != tt.wantErr {
				t.Errorf("ada kesalahan = %t, ingin %t (%+v)", ada, tt.wantErr, errs)
			}
		})
	}
}

// Jebakan utama: "required" tak bisa membedakan 0 dari "tidak dikirim".
func TestRequiredTidakBisaMembedakanNol(t *testing.T) {
	v := validatorUji(t)

	nol := 0
	d := Diskon{PersenSalah: 0, PersenBenar: &nol}
	errs := Validasi(v, d)

	// int biasa bernilai 0 DITOLAK, walaupun 0% itu diskon yang sah.
	if _, ada := punyaField(errs, "persen_salah"); !ada {
		t.Error("int bernilai 0 seharusnya ditolak oleh required — inilah jebakannya")
	}
	// Pointer ke 0 LOLOS, karena yang diperiksa adalah pointernya (bukan nil).
	if _, ada := punyaField(errs, "persen_benar"); ada {
		t.Error("pointer ke 0 seharusnya lolos: nil berarti 'tak dikirim', 0 berarti 'nol'")
	}

	// Pointer nil memang harus ditolak.
	if _, ada := punyaField(Validasi(v, Diskon{PersenSalah: 5}), "persen_benar"); !ada {
		t.Error("pointer nil seharusnya ditolak oleh required")
	}
}

// Mengirim yang bukan struct adalah kesalahan programmer — jangan sampai panic.
func TestValidasiBukanStruct(t *testing.T) {
	v := validatorUji(t)

	errs := Validasi(v, "ini string, bukan struct")
	if len(errs) != 1 || !strings.Contains(errs[0].Pesan, "bukan struct") {
		t.Errorf("ingin satu pesan 'bukan struct', dapat: %+v", errs)
	}
}

func TestValidasiMengembalikanNilSaatBersih(t *testing.T) {
	v := validatorUji(t)
	if got := Validasi(v, contohValid()); got != nil {
		t.Errorf("ingin nil untuk data bersih, dapat %+v", got)
	}
}
