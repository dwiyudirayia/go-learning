package main

import (
	"strings"
	"testing"
)

func TestVersi(t *testing.T) {
	out, err := Jalankan(NewPenyimpanan(), "versi")
	if err != nil {
		t.Fatalf("error tak terduga: %v", err)
	}
	if !strings.Contains(out, "v1.0.0") {
		t.Errorf("keluaran = %q, ingin mengandung v1.0.0", out)
	}
}

func TestTambahLaluDaftar(t *testing.T) {
	p := NewPenyimpanan()

	if _, err := Jalankan(p, "tugas", "tambah", "beli susu"); err != nil {
		t.Fatalf("tambah gagal: %v", err)
	}
	if _, err := Jalankan(p, "tugas", "tambah", "cuci baju", "-p", "tinggi"); err != nil {
		t.Fatalf("tambah gagal: %v", err)
	}

	out, err := Jalankan(p, "tugas", "daftar")
	if err != nil {
		t.Fatalf("daftar gagal: %v", err)
	}
	for _, ingin := range []string{"beli susu", "cuci baju", "sedang", "tinggi"} {
		if !strings.Contains(out, ingin) {
			t.Errorf("keluaran %q tidak mengandung %q", out, ingin)
		}
	}
}

// Nilai bawaan flag harus dipakai bila pengguna tak menyebutkannya.
func TestPrioritasBawaan(t *testing.T) {
	p := NewPenyimpanan()

	if _, err := Jalankan(p, "tugas", "tambah", "tanpa prioritas"); err != nil {
		t.Fatalf("tambah gagal: %v", err)
	}
	if got := p.Daftar()[0].Prioritas; got != "sedang" {
		t.Errorf("prioritas = %q, ingin sedang (nilai bawaan)", got)
	}
}

// INILAH alasan NewRootCmd membangun ulang tiap kali: flag tidak boleh bocor antar
// pemanggilan. Kalau perintah dibuat sekali sebagai variabel global, "--format json"
// di baris pertama akan menular ke baris kedua.
func TestFlagTidakBocorAntarPemanggilan(t *testing.T) {
	p := NewPenyimpanan()
	if _, err := Jalankan(p, "tugas", "tambah", "satu"); err != nil {
		t.Fatalf("tambah gagal: %v", err)
	}

	jsonOut, err := Jalankan(p, "--format", "json", "tugas", "daftar")
	if err != nil {
		t.Fatalf("daftar json gagal: %v", err)
	}
	if !strings.HasPrefix(strings.TrimSpace(jsonOut), "[{") {
		t.Errorf("format json = %q, ingin diawali [{", jsonOut)
	}

	teksOut, err := Jalankan(p, "tugas", "daftar")
	if err != nil {
		t.Fatalf("daftar teks gagal: %v", err)
	}
	if strings.HasPrefix(strings.TrimSpace(teksOut), "[{") {
		t.Errorf("format json bocor ke pemanggilan berikutnya: %q", teksOut)
	}
	if !strings.Contains(teksOut, "1. satu") {
		t.Errorf("format teks = %q, ingin baris bernomor", teksOut)
	}
}

func TestAliasDaftar(t *testing.T) {
	p := NewPenyimpanan()
	if _, err := Jalankan(p, "tugas", "tambah", "pakai alias"); err != nil {
		t.Fatalf("tambah gagal: %v", err)
	}

	for _, alias := range []string{"daftar", "ls", "list"} {
		t.Run(alias, func(t *testing.T) {
			out, err := Jalankan(p, "tugas", alias)
			if err != nil {
				t.Fatalf("%s gagal: %v", alias, err)
			}
			if !strings.Contains(out, "pakai alias") {
				t.Errorf("alias %q memberi keluaran berbeda: %q", alias, out)
			}
		})
	}
}

func TestSelesai(t *testing.T) {
	p := NewPenyimpanan()
	if _, err := Jalankan(p, "tugas", "tambah", "kerjakan ini"); err != nil {
		t.Fatalf("tambah gagal: %v", err)
	}

	if _, err := Jalankan(p, "tugas", "selesai", "1"); err != nil {
		t.Fatalf("selesai gagal: %v", err)
	}
	if !p.Daftar()[0].Selesai {
		t.Error("tugas seharusnya bertanda selesai")
	}

	out, err := Jalankan(p, "tugas", "daftar")
	if err != nil {
		t.Fatalf("daftar gagal: %v", err)
	}
	if !strings.Contains(out, "[x]") {
		t.Errorf("keluaran = %q, ingin menandai tugas selesai dengan [x]", out)
	}
}

func TestFlagBelum(t *testing.T) {
	p := NewPenyimpanan()
	for _, judul := range []string{"satu", "dua"} {
		if _, err := Jalankan(p, "tugas", "tambah", judul); err != nil {
			t.Fatalf("tambah gagal: %v", err)
		}
	}
	if _, err := Jalankan(p, "tugas", "selesai", "1"); err != nil {
		t.Fatalf("selesai gagal: %v", err)
	}

	out, err := Jalankan(p, "tugas", "daftar", "--belum")
	if err != nil {
		t.Fatalf("daftar gagal: %v", err)
	}
	if strings.Contains(out, "satu") {
		t.Errorf("tugas selesai seharusnya disembunyikan: %q", out)
	}
	if !strings.Contains(out, "dua") {
		t.Errorf("tugas belum selesai seharusnya tampil: %q", out)
	}
}

// RunE harus mengembalikan error supaya kode keluar proses jadi bukan-nol.
func TestPerintahYangGagalMengembalikanError(t *testing.T) {
	tests := []struct {
		nama      string
		args      []string
		wantPesan string
	}{
		{"judul kosong", []string{"tugas", "tambah", "   "}, "tidak boleh kosong"},
		{"prioritas tak dikenal", []string{"tugas", "tambah", "x", "-p", "darurat"}, "tidak dikenal"},
		{"id bukan angka", []string{"tugas", "selesai", "abc"}, "harus berupa angka"},
		{"id tak ada", []string{"tugas", "selesai", "99"}, "tidak ditemukan"},
		{"argumen kurang", []string{"tugas", "tambah"}, "arg"},
		// Catatan: karena root punya sub-perintah, argumen berlebih ditafsirkan cobra
		// sebagai nama perintah — pesannya "unknown command", bukan soal jumlah argumen.
		{"argumen berlebih", []string{"versi", "berlebih"}, "unknown command"},
		{"perintah tak dikenal", []string{"terbang"}, "unknown command"},
		{"argumen berlebih pada daftar", []string{"tugas", "daftar", "berlebih"}, "unknown command"},
	}

	for _, tt := range tests {
		t.Run(tt.nama, func(t *testing.T) {
			_, err := Jalankan(NewPenyimpanan(), tt.args...)
			if err == nil {
				t.Fatal("ingin error, dapat nil")
			}
			if !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(tt.wantPesan)) {
				t.Errorf("error = %q, ingin mengandung %q", err, tt.wantPesan)
			}
		})
	}
}

func TestDaftarKosong(t *testing.T) {
	out, err := Jalankan(NewPenyimpanan(), "tugas", "daftar")
	if err != nil {
		t.Fatalf("error tak terduga: %v", err)
	}
	if !strings.Contains(out, "belum ada tugas") {
		t.Errorf("keluaran = %q, ingin pesan daftar kosong yang ramah", out)
	}
}

// Bantuan dibuat otomatis oleh cobra — pastikan sub-perintah benar-benar terdaftar.
func TestBantuanMenyebutSubPerintah(t *testing.T) {
	out, err := Jalankan(NewPenyimpanan(), "--help")
	if err != nil {
		t.Fatalf("help gagal: %v", err)
	}
	for _, ingin := range []string{"tugas", "versi", "--format"} {
		if !strings.Contains(out, ingin) {
			t.Errorf("bantuan tidak menyebut %q:\n%s", ingin, out)
		}
	}
}

func TestPenyimpananMengurutkanBerdasarkanID(t *testing.T) {
	p := NewPenyimpanan()
	for _, j := range []string{"a", "b", "c", "d", "e"} {
		p.Tambah(j, "sedang")
	}

	got := p.Daftar()
	for i := range got {
		if got[i].ID != i+1 {
			t.Fatalf("urutan kacau di indeks %d: %+v", i, got)
		}
	}
}
