package main

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// awal adalah "sekarang" yang dipakai seluruh test, supaya hasilnya selalu sama.
var awal = time.Date(2026, 7, 23, 10, 0, 0, 0, time.UTC)

// penerbitUji mengembalikan penerbit dengan jam yang bisa dikendalikan.
func penerbitUji() *Penerbit {
	return NewPenerbit("rahasia-uji-yang-panjang", "go-learning").
		DenganJam(func() time.Time { return awal })
}

// pada mengembalikan salinan penerbit yang "jam dindingnya" sudah dimajukan.
func pada(p *Penerbit, maju time.Duration) *Penerbit {
	return p.DenganJam(func() time.Time { return awal.Add(maju) })
}

func TestBuatDanVerifikasiAkses(t *testing.T) {
	p := penerbitUji()

	tok, err := p.BuatAkses("user-42", "admin")
	if err != nil {
		t.Fatalf("BuatAkses gagal: %v", err)
	}

	k, err := p.Verifikasi(tok, JenisAkses)
	if err != nil {
		t.Fatalf("Verifikasi gagal: %v", err)
	}

	if k.Subject != "user-42" {
		t.Errorf("subjek = %q, ingin user-42", k.Subject)
	}
	if k.Peran != "admin" {
		t.Errorf("peran = %q, ingin admin", k.Peran)
	}
	if k.Issuer != "go-learning" {
		t.Errorf("penerbit = %q, ingin go-learning", k.Issuer)
	}
	if k.Jenis != JenisAkses {
		t.Errorf("jenis = %q, ingin %q", k.Jenis, JenisAkses)
	}
	// Masa berlaku access token = 15 menit sejak diterbitkan.
	if got := k.ExpiresAt.Time; !got.Equal(awal.Add(15 * time.Minute)) {
		t.Errorf("kedaluwarsa = %v, ingin %v", got, awal.Add(15*time.Minute))
	}
}

func TestSubjekKosongDitolak(t *testing.T) {
	if _, err := penerbitUji().BuatAkses("", "admin"); err == nil {
		t.Error("subjek kosong seharusnya ditolak saat penerbitan")
	}
}

// Jam palsu membuat test kedaluwarsa berjalan seketika, tanpa menunggu 15 menit.
func TestKedaluwarsa(t *testing.T) {
	p := penerbitUji()
	tok, err := p.BuatAkses("user-1", "pembaca")
	if err != nil {
		t.Fatalf("BuatAkses gagal: %v", err)
	}

	tests := []struct {
		nama    string
		maju    time.Duration
		wantErr bool
	}{
		{"baru diterbitkan", 0, false},
		{"14 menit kemudian", 14 * time.Minute, false},
		{"tepat di batas masih ditoleransi", 15 * time.Minute, false},
		{"lewat toleransi 5 detik", 15*time.Minute + 6*time.Second, true},
		{"16 menit kemudian", 16 * time.Minute, true},
		{"sehari kemudian", 24 * time.Hour, true},
	}

	for _, tt := range tests {
		t.Run(tt.nama, func(t *testing.T) {
			_, err := pada(p, tt.maju).Verifikasi(tok, JenisAkses)
			if tt.wantErr {
				if err == nil {
					t.Fatal("ingin token ditolak")
				}
				if !errors.Is(err, jwt.ErrTokenExpired) {
					t.Errorf("error = %v, ingin ErrTokenExpired", err)
				}
				return
			}
			if err != nil {
				t.Errorf("token seharusnya masih sah, dapat: %v", err)
			}
		})
	}
}

// Refresh token berumur jauh lebih panjang daripada access token.
func TestRefreshBerumurPanjang(t *testing.T) {
	p := penerbitUji()
	rt, err := p.BuatRefresh("user-42")
	if err != nil {
		t.Fatalf("BuatRefresh gagal: %v", err)
	}

	// Access token sudah lama mati di titik ini, refresh masih hidup.
	if _, err := pada(p, 6*24*time.Hour).Verifikasi(rt, JenisRefresh); err != nil {
		t.Errorf("refresh token seharusnya masih sah setelah 6 hari: %v", err)
	}
	if _, err := pada(p, 8*24*time.Hour).Verifikasi(rt, JenisRefresh); err == nil {
		t.Error("refresh token seharusnya mati setelah 8 hari")
	}
}

// Isi token yang diubah membuat tanda tangan tak cocok lagi.
func TestKlaimYangDiubahDitolak(t *testing.T) {
	p := penerbitUji()
	tok, err := p.BuatAkses("user-42", "pembaca")
	if err != nil {
		t.Fatalf("BuatAkses gagal: %v", err)
	}

	rusak := TukarKlaim(tok)
	if rusak == tok {
		t.Fatal("TukarKlaim tidak mengubah apa pun")
	}
	if _, err := p.Verifikasi(rusak, JenisAkses); err == nil {
		t.Error("token dengan klaim diubah seharusnya ditolak")
	}
}

func TestKunciBerbedaDitolak(t *testing.T) {
	tok, err := penerbitUji().BuatAkses("user-42", "admin")
	if err != nil {
		t.Fatalf("BuatAkses gagal: %v", err)
	}

	lain := NewPenerbit("kunci-yang-sama-sekali-berbeda", "go-learning").
		DenganJam(func() time.Time { return awal })

	_, err = lain.Verifikasi(tok, JenisAkses)
	if err == nil {
		t.Fatal("token dari penerbit lain seharusnya ditolak")
	}
	if !errors.Is(err, jwt.ErrTokenSignatureInvalid) {
		t.Errorf("error = %v, ingin ErrTokenSignatureInvalid", err)
	}
}

// Penerbit yang tak dikenal harus ditolak walau tanda tangannya benar.
func TestPenerbitBerbedaDitolak(t *testing.T) {
	rahasia := "rahasia-yang-sama-persis"
	a := NewPenerbit(rahasia, "layanan-a").DenganJam(func() time.Time { return awal })
	b := NewPenerbit(rahasia, "layanan-b").DenganJam(func() time.Time { return awal })

	tok, err := a.BuatAkses("user-1", "admin")
	if err != nil {
		t.Fatalf("BuatAkses gagal: %v", err)
	}
	if _, err := b.Verifikasi(tok, JenisAkses); err == nil {
		t.Error("token dari penerbit lain seharusnya ditolak walau kuncinya sama")
	}
}

// INI test keamanan terpenting di berkas ini.
func TestSeranganAlgNoneDitolak(t *testing.T) {
	jahat, err := TokenTanpaTandaTangan("user-1", "admin")
	if err != nil {
		t.Fatalf("gagal membuat token none: %v", err)
	}

	if _, err := penerbitUji().Verifikasi(jahat, JenisAkses); err == nil {
		t.Fatal("token beralgoritma 'none' DITERIMA — ini lubang keamanan serius; " +
			"pastikan jwt.WithValidMethods masih terpasang di Verifikasi")
	}
}

// Refresh token tak boleh bisa dipakai sebagai access token, dan sebaliknya.
func TestJenisTokenTidakBolehTertukar(t *testing.T) {
	p := penerbitUji()

	at, err := p.BuatAkses("user-42", "admin")
	if err != nil {
		t.Fatalf("BuatAkses gagal: %v", err)
	}
	rt, err := p.BuatRefresh("user-42")
	if err != nil {
		t.Fatalf("BuatRefresh gagal: %v", err)
	}

	if _, err := p.Verifikasi(rt, JenisAkses); !errors.Is(err, ErrJenisTokenSalah) {
		t.Errorf("refresh sebagai akses: error = %v, ingin ErrJenisTokenSalah", err)
	}
	if _, err := p.Verifikasi(at, JenisRefresh); !errors.Is(err, ErrJenisTokenSalah) {
		t.Errorf("akses sebagai refresh: error = %v, ingin ErrJenisTokenSalah", err)
	}
}

func TestSegarkan(t *testing.T) {
	p := penerbitUji()
	rt, err := p.BuatRefresh("user-42")
	if err != nil {
		t.Fatalf("BuatRefresh gagal: %v", err)
	}

	// Peran diambil ulang saat refresh — jadi perubahan hak akses langsung berlaku
	// pada access token berikutnya, tanpa menunggu refresh token kedaluwarsa.
	baru, err := p.Segarkan(rt, "editor")
	if err != nil {
		t.Fatalf("Segarkan gagal: %v", err)
	}

	k, err := p.Verifikasi(baru, JenisAkses)
	if err != nil {
		t.Fatalf("Verifikasi token baru gagal: %v", err)
	}
	if k.Subject != "user-42" {
		t.Errorf("subjek = %q, ingin user-42", k.Subject)
	}
	if k.Peran != "editor" {
		t.Errorf("peran = %q, ingin editor (peran terbaru)", k.Peran)
	}
}

func TestSegarkanDenganTokenTidakSah(t *testing.T) {
	p := penerbitUji()

	tests := []struct {
		nama  string
		token string
	}{
		{"teks sembarang", "bukan-token"},
		{"kosong", ""},
		{"hanya dua bagian", "aaa.bbb"},
	}
	for _, tt := range tests {
		t.Run(tt.nama, func(t *testing.T) {
			if _, err := p.Segarkan(tt.token, "admin"); err == nil {
				t.Error("ingin error untuk token tidak sah")
			}
		})
	}

	// Refresh token yang sudah kedaluwarsa juga harus ditolak.
	rt, err := p.BuatRefresh("user-1")
	if err != nil {
		t.Fatalf("BuatRefresh gagal: %v", err)
	}
	if _, err := pada(p, 8*24*time.Hour).Segarkan(rt, "admin"); err == nil {
		t.Error("refresh token kedaluwarsa seharusnya ditolak")
	}
}

func TestTokenRusakDitolak(t *testing.T) {
	p := penerbitUji()

	tests := []struct {
		nama  string
		token string
	}{
		{"kosong", ""},
		{"teks biasa", "halo-dunia"},
		{"bagian kurang", "aaa.bbb"},
		{"bagian berlebih", "aaa.bbb.ccc.ddd"},
		{"base64 tidak sah", "!!!.???.###"},
	}

	for _, tt := range tests {
		t.Run(tt.nama, func(t *testing.T) {
			if _, err := p.Verifikasi(tt.token, JenisAkses); err == nil {
				t.Errorf("token %q seharusnya ditolak", tt.token)
			}
		})
	}
}

// Isi token memang bisa dibaca siapa pun — ini bukan bug, ini desain JWT.
func TestKlaimBisaDibacaTanpaKunci(t *testing.T) {
	tok, err := penerbitUji().BuatAkses("user-42", "admin")
	if err != nil {
		t.Fatalf("BuatAkses gagal: %v", err)
	}

	k, err := BacaKlaimTanpaVerifikasi(tok)
	if err != nil {
		t.Fatalf("BacaKlaimTanpaVerifikasi gagal: %v", err)
	}
	if k.Subject != "user-42" || k.Peran != "admin" {
		t.Errorf("klaim = %+v, ingin subjek user-42 & peran admin", k)
	}

	// Token terdiri dari tiga bagian yang dipisah titik: header.klaim.tandatangan
	if got := len(strings.Split(tok, ".")); got != 3 {
		t.Errorf("jumlah bagian token = %d, ingin 3", got)
	}
}

func TestBacaKlaimTokenRusak(t *testing.T) {
	if _, err := BacaKlaimTanpaVerifikasi("bukan.token.jwt"); err == nil {
		t.Error("ingin error untuk token rusak")
	}
}
