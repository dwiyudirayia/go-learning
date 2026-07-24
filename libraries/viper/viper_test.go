package main

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestNilaiBawaanDipakaiSaatTakAdaSumberLain(t *testing.T) {
	k, err := MuatKonfig(NewViper())
	if err != nil {
		t.Fatalf("MuatKonfig gagal: %v", err)
	}

	if k.Aplikasi != "go-learning" {
		t.Errorf("aplikasi = %q, ingin go-learning", k.Aplikasi)
	}
	if k.Server.Port != 8080 {
		t.Errorf("port = %d, ingin 8080", k.Server.Port)
	}
	if k.Timeout != 30*time.Second {
		t.Errorf("timeout = %v, ingin 30s", k.Timeout)
	}
	if k.Debug {
		t.Error("debug seharusnya false secara bawaan")
	}
}

func TestMuatDariYAML(t *testing.T) {
	k, err := MuatDariYAML(contohYAML)
	if err != nil {
		t.Fatalf("MuatDariYAML gagal: %v", err)
	}

	cek := []struct {
		nama string
		got  any
		want any
	}{
		{"aplikasi", k.Aplikasi, "toko-online"},
		{"debug", k.Debug, true},
		{"host", k.Server.Host, "127.0.0.1"},
		{"port", k.Server.Port, 3000},
		{"dsn", k.DB.DSN, "postgres://localhost/toko"},
		{"maks_koneksi", k.DB.MaksKoneksi, 25},
		{"timeout", k.Timeout, 15 * time.Second},
		{"jumlah fitur", len(k.Fitur), 2},
		{"alamat", k.Alamat(), "127.0.0.1:3000"},
	}
	for _, c := range cek {
		if c.got != c.want {
			t.Errorf("%s = %v, ingin %v", c.nama, c.got, c.want)
		}
	}
}

// Berkas hanya menimpa kunci yang DISEBUTKAN; sisanya tetap memakai nilai bawaan.
func TestBerkasHanyaMenimpaSebagian(t *testing.T) {
	k, err := MuatDariYAML("server:\n  port: 3000\n")
	if err != nil {
		t.Fatalf("MuatDariYAML gagal: %v", err)
	}

	if k.Server.Port != 3000 {
		t.Errorf("port = %d, ingin 3000 (dari berkas)", k.Server.Port)
	}
	if k.Server.Host != "0.0.0.0" {
		t.Errorf("host = %q, ingin 0.0.0.0 (nilai bawaan tetap berlaku)", k.Server.Host)
	}
	if k.Aplikasi != "go-learning" {
		t.Errorf("aplikasi = %q, ingin nilai bawaan", k.Aplikasi)
	}
}

// Inti aturan main: environment MENGALAHKAN berkas, berkas mengalahkan nilai bawaan.
func TestUrutanKemenanganLapisan(t *testing.T) {
	t.Setenv("APP_SERVER_PORT", "9999")
	t.Setenv("APP_APLIKASI", "dari-env")

	v := NewViper()
	v.SetConfigType("yaml")
	if err := v.ReadConfig(strings.NewReader(contohYAML)); err != nil {
		t.Fatalf("ReadConfig gagal: %v", err)
	}

	k, err := MuatKonfig(v)
	if err != nil {
		t.Fatalf("MuatKonfig gagal: %v", err)
	}

	// env (9999) mengalahkan berkas (3000) dan bawaan (8080)
	if k.Server.Port != 9999 {
		t.Errorf("port = %d, ingin 9999 dari environment", k.Server.Port)
	}
	if k.Aplikasi != "dari-env" {
		t.Errorf("aplikasi = %q, ingin dari-env", k.Aplikasi)
	}
	// kunci yang tak disebut environment tetap diambil dari berkas
	if k.Server.Host != "127.0.0.1" {
		t.Errorf("host = %q, ingin 127.0.0.1 dari berkas", k.Server.Host)
	}
	if k.DB.MaksKoneksi != 25 {
		t.Errorf("maks_koneksi = %d, ingin 25 dari berkas", k.DB.MaksKoneksi)
	}
}

// Membuktikan SetEnvKeyReplacer benar-benar dibutuhkan untuk kunci bertingkat.
func TestEnvBertingkatButuhGarisBawah(t *testing.T) {
	t.Setenv("APP_DB_MAKS_KONEKSI", "77")

	k, err := MuatKonfig(NewViper())
	if err != nil {
		t.Fatalf("MuatKonfig gagal: %v", err)
	}
	if k.DB.MaksKoneksi != 77 {
		t.Errorf("maks_koneksi = %d, ingin 77 — SetEnvKeyReplacer mungkin hilang", k.DB.MaksKoneksi)
	}
}

// Awalan melindungi dari variabel lingkungan milik program lain.
func TestAwalanMelindungiDariVariabelAsing(t *testing.T) {
	// Tanpa awalan APP_, variabel ini tak boleh berpengaruh.
	t.Setenv("SERVER_PORT", "1234")

	k, err := MuatKonfig(NewViper())
	if err != nil {
		t.Fatalf("MuatKonfig gagal: %v", err)
	}
	if k.Server.Port != 8080 {
		t.Errorf("port = %d, ingin tetap 8080 — variabel tanpa awalan tak boleh terbaca", k.Server.Port)
	}
}

func TestValidasiMenolakKonfigurasiRusak(t *testing.T) {
	tests := []struct {
		nama      string
		yaml      string
		wantPesan string
	}{
		{"port di luar jangkauan", "server:\n  port: 99999\n", "di luar jangkauan"},
		{"port nol", "server:\n  port: 0\n", "di luar jangkauan"},
		{"aplikasi kosong", "aplikasi: \"\"\n", "aplikasi tidak boleh kosong"},
		{"dsn kosong", "db:\n  dsn: \"\"\n", "db.dsn wajib diisi"},
		{"koneksi nol", "db:\n  maks_koneksi: 0\n", "minimal 1"},
		{"timeout nol", "timeout: 0s\n", "lebih besar dari nol"},
	}

	for _, tt := range tests {
		t.Run(tt.nama, func(t *testing.T) {
			_, err := MuatDariYAML(tt.yaml)
			if err == nil {
				t.Fatal("ingin error validasi, dapat nil")
			}
			if !errors.Is(err, ErrKonfigTidakValid) {
				t.Errorf("error = %v, ingin membungkus ErrKonfigTidakValid", err)
			}
			if !strings.Contains(err.Error(), tt.wantPesan) {
				t.Errorf("pesan = %q, ingin mengandung %q", err, tt.wantPesan)
			}
		})
	}
}

// Semua masalah dilaporkan sekaligus, bukan satu per satu.
func TestValidasiMelaporkanSemuaMasalah(t *testing.T) {
	rusak := "aplikasi: \"\"\nserver:\n  port: 99999\ndb:\n  dsn: \"\"\n  maks_koneksi: 0\ntimeout: 0s\n"

	_, err := MuatDariYAML(rusak)
	if err == nil {
		t.Fatal("ingin error")
	}
	for _, potong := range []string{"aplikasi", "server.port", "db.dsn", "db.maks_koneksi", "timeout"} {
		if !strings.Contains(err.Error(), potong) {
			t.Errorf("pesan tidak menyebut %q:\n%s", potong, err)
		}
	}
}

func TestYAMLRusakDitolak(t *testing.T) {
	if _, err := MuatDariYAML("aplikasi: [belum ditutup\n"); err == nil {
		t.Error("YAML rusak seharusnya menghasilkan error")
	}
}

// Jebakan: tag json diabaikan. Viper jatuh ke pencocokan NAMA FIELD, sehingga
// sebagian kunci kebetulan terisi dan sebagian lain diam-diam nol — tanpa error apa pun.
func TestTagJSONDiabaikan(t *testing.T) {
	k, err := MuatDenganTagSalah("aplikasi: toko-online\nmaks_koneksi: 25\n")
	if err != nil {
		t.Fatalf("Unmarshal justru error: %v", err)
	}

	// Terisi bukan karena tag json dihormati, melainkan karena nama field "Aplikasi"
	// kebetulan sama dengan nama kunci "aplikasi" (perbandingan tak peka huruf).
	if k.Aplikasi != "toko-online" {
		t.Errorf("Aplikasi = %q, ingin toko-online (cocok lewat nama field)", k.Aplikasi)
	}

	// INI bukti bahwa tag json benar-benar diabaikan: "maks_koneksi" tak cocok dengan
	// nama field "MaksKoneksi", jadi nilainya nol — dan tak ada peringatan sama sekali.
	if k.MaksKoneksi != 0 {
		t.Errorf("MaksKoneksi = %d, ingin 0 — kalau ini terisi, viper sudah mendukung "+
			"tag json dan komentar jebakan di main.go perlu diperbarui", k.MaksKoneksi)
	}

	// Pembanding: dengan tag mapstructure yang benar, kunci apa pun terbaca.
	benar, err := MuatDariYAML("aplikasi: toko-online\ndb:\n  maks_koneksi: 25\n")
	if err != nil {
		t.Fatalf("MuatDariYAML gagal: %v", err)
	}
	if benar.Aplikasi != "toko-online" || benar.DB.MaksKoneksi != 25 {
		t.Errorf("dengan mapstructure: aplikasi=%q maks_koneksi=%d, ingin toko-online & 25",
			benar.Aplikasi, benar.DB.MaksKoneksi)
	}
}

// Dua instance viper harus benar-benar terpisah (bukan state global).
func TestInstanceViperTerisolasi(t *testing.T) {
	a := NewViper()
	a.Set("server.port", 1111)

	b := NewViper()
	kb, err := MuatKonfig(b)
	if err != nil {
		t.Fatalf("MuatKonfig gagal: %v", err)
	}
	if kb.Server.Port != 8080 {
		t.Errorf("port instance kedua = %d, ingin 8080 — state bocor antar instance", kb.Server.Port)
	}
}
