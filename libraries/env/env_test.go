package main

import (
	"errors"
	"slices"
	"testing"
	"time"
)

func TestNilaiDefaultDipakaiSaatKosong(t *testing.T) {
	// Environment kosong -> semua field memakai envDefault.
	k, err := MuatDari(map[string]string{})
	if err != nil {
		t.Fatalf("Muat gagal: %v", err)
	}

	cek := []struct {
		nama string
		got  any
		want any
	}{
		{"nama", k.Nama, "go-learning"},
		{"port", k.Port, 8080},
		{"debug", k.Debug, false},
		{"timeout", k.Timeout, 30 * time.Second},
		{"db.dsn", k.DB.DSN, "file::memory:"},
		{"db.maxconn", k.DB.MaxConn, 10},
	}
	for _, c := range cek {
		if c.got != c.want {
			t.Errorf("%s = %v, ingin %v", c.nama, c.got, c.want)
		}
	}
}

func TestEnvironmentMenimpaDefault(t *testing.T) {
	k, err := MuatDari(map[string]string{
		"APP_NAME": "toko",
		"APP_PORT": "3000",
		"DEBUG":    "true",
	})
	if err != nil {
		t.Fatalf("Muat gagal: %v", err)
	}

	if k.Nama != "toko" {
		t.Errorf("nama = %q, ingin toko", k.Nama)
	}
	if k.Port != 3000 {
		t.Errorf("port = %d, ingin 3000", k.Port)
	}
	if !k.Debug {
		t.Error("debug seharusnya true")
	}
	// Yang tak diset tetap default.
	if k.DB.MaxConn != 10 {
		t.Errorf("db.maxconn = %d, ingin tetap default 10", k.DB.MaxConn)
	}
}

func TestKonversiTipe(t *testing.T) {
	k, err := MuatDari(map[string]string{
		"APP_PORT": "3000",
		"DEBUG":    "true",
		"TIMEOUT":  "1m30s",
		"FITUR":    "a,b,c",
	})
	if err != nil {
		t.Fatalf("Muat gagal: %v", err)
	}

	if k.Port != 3000 {
		t.Errorf("port = %d, ingin 3000", k.Port)
	}
	if k.Timeout != 90*time.Second {
		t.Errorf("timeout = %v, ingin 1m30s", k.Timeout)
	}
	if !slices.Equal(k.Fitur, []string{"a", "b", "c"}) {
		t.Errorf("fitur = %v, ingin [a b c]", k.Fitur)
	}
}

// Struct bersarang memakai awalan (DB_DSN, DB_MAX_CONN).
func TestStructBersarangDenganAwalan(t *testing.T) {
	k, err := MuatDari(map[string]string{
		"DB_DSN":      "postgres://localhost/toko",
		"DB_MAX_CONN": "25",
	})
	if err != nil {
		t.Fatalf("Muat gagal: %v", err)
	}
	if k.DB.DSN != "postgres://localhost/toko" {
		t.Errorf("db.dsn = %q", k.DB.DSN)
	}
	if k.DB.MaxConn != 25 {
		t.Errorf("db.maxconn = %d, ingin 25", k.DB.MaxConn)
	}
}

// Tipe yang tak bisa dikonversi harus menghasilkan error, bukan diam-diam nol.
func TestTipeSalahDitolak(t *testing.T) {
	tests := []struct {
		nama string
		vars map[string]string
	}{
		{"port bukan angka", map[string]string{"APP_PORT": "abc"}},
		{"debug bukan bool", map[string]string{"DEBUG": "mungkin"}},
		{"timeout bukan durasi", map[string]string{"TIMEOUT": "lama sekali"}},
		{"maxconn bukan angka", map[string]string{"DB_MAX_CONN": "banyak"}},
	}
	for _, tt := range tests {
		t.Run(tt.nama, func(t *testing.T) {
			if _, err := MuatDari(tt.vars); err == nil {
				t.Error("ingin error untuk tipe yang tak bisa dikonversi")
			}
		})
	}
}

func TestFieldWajib(t *testing.T) {
	t.Run("ada -> sukses", func(t *testing.T) {
		k, err := MuatRahasia(map[string]string{"SECRET_KEY": "rahasia"})
		if err != nil {
			t.Fatalf("MuatRahasia gagal: %v", err)
		}
		if k.SecretKey != "rahasia" {
			t.Errorf("secret = %q, ingin rahasia", k.SecretKey)
		}
	})

	t.Run("tak ada -> gagal (fail fast)", func(t *testing.T) {
		if _, err := MuatRahasia(map[string]string{}); err == nil {
			t.Error("field notEmpty yang tak diisi seharusnya menggagalkan Parse")
		}
	})

	t.Run("diisi kosong -> gagal (inti notEmpty)", func(t *testing.T) {
		// Inilah beda notEmpty dari required biasa: string kosong DITOLAK.
		// (Dengan tag `required` saja, SECRET_KEY="" justru akan LOLOS.)
		if _, err := MuatRahasia(map[string]string{"SECRET_KEY": ""}); err == nil {
			t.Error("SECRET_KEY kosong seharusnya ditolak oleh notEmpty")
		}
	})
}

// Perilaku halus yang wajib diketahui: nilai environment KOSONG pada field ber-default
// TIDAK menimpa default — ia jatuh kembali ke default. Jadi "APP_NAME=" tetap menghasilkan
// nama default, bukan string kosong.
func TestNilaiKosongJatuhKeDefault(t *testing.T) {
	k, err := MuatDari(map[string]string{"APP_NAME": "", "APP_PORT": ""})
	if err != nil {
		t.Fatalf("Muat gagal: %v", err)
	}
	if k.Nama != "go-learning" {
		t.Errorf("nama = %q, ingin default 'go-learning' (env kosong jatuh ke default)", k.Nama)
	}
	if k.Port != 8080 {
		t.Errorf("port = %d, ingin default 8080", k.Port)
	}
}

func TestValidasiAturanBisnis(t *testing.T) {
	tests := []struct {
		nama      string
		vars      map[string]string
		wantPesan string
	}{
		{"port terlalu besar", map[string]string{"APP_PORT": "99999"}, "APP_PORT"},
		{"port nol", map[string]string{"APP_PORT": "0"}, "APP_PORT"},
		{"maxconn nol", map[string]string{"DB_MAX_CONN": "0"}, "DB_MAX_CONN"},
		{"timeout nol", map[string]string{"TIMEOUT": "0s"}, "TIMEOUT"},
	}
	for _, tt := range tests {
		t.Run(tt.nama, func(t *testing.T) {
			_, err := MuatDari(tt.vars)
			if err == nil {
				t.Fatal("ingin error validasi")
			}
			if !errors.Is(err, ErrKonfigTidakValid) {
				t.Errorf("error = %v, ingin membungkus ErrKonfigTidakValid", err)
			}
			if !contains(err.Error(), tt.wantPesan) {
				t.Errorf("pesan = %q, ingin menyebut %q", err, tt.wantPesan)
			}
		})
	}
}

func TestKonfigurasiSahLolosValidasi(t *testing.T) {
	k, err := MuatDari(map[string]string{
		"APP_NAME":    "toko",
		"APP_PORT":    "8080",
		"DB_MAX_CONN": "20",
		"TIMEOUT":     "15s",
	})
	if err != nil {
		t.Fatalf("konfigurasi sah seharusnya lolos, dapat: %v", err)
	}
	if k.Port != 8080 || k.DB.MaxConn != 20 {
		t.Errorf("nilai tak sesuai: %+v", k)
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
