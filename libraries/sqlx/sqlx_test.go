package main

import (
	"errors"
	"testing"

	"github.com/jmoiron/sqlx"
)

// repoUji membuat database SQLite in-memory baru + repo, otomatis ditutup saat selesai.
//
// 🔍 Analogi: tiap test butuh database bersih sendiri. ":memory:" membuat SQLite di RAM
// yang lenyap saat koneksi ditutup — sempurna untuk test: cepat, terisolasi, tanpa berkas.
func repoUji(t *testing.T) *ProdukRepo {
	t.Helper()
	db, err := Buka(":memory:")
	if err != nil {
		t.Fatalf("Buka gagal: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return NewProdukRepo(db)
}

func seed(t *testing.T, r *ProdukRepo) {
	t.Helper()
	produk := []Produk{
		{Nama: "Kopi", Harga: 25_000, Stok: 10, Kategori: "minuman"},
		{Nama: "Teh", Harga: 15_000, Stok: 5, Kategori: "minuman"},
		{Nama: "Roti", Harga: 20_000, Stok: 8, Kategori: "makanan"},
	}
	for _, p := range produk {
		if _, err := r.Buat(p); err != nil {
			t.Fatalf("seed gagal: %v", err)
		}
	}
}

func TestBuatDanAmbil(t *testing.T) {
	r := repoUji(t)

	id, err := r.Buat(Produk{Nama: "Gula", Harga: 12_000, Stok: 20, Kategori: "bahan"})
	if err != nil {
		t.Fatalf("Buat gagal: %v", err)
	}

	p, err := r.Ambil(id)
	if err != nil {
		t.Fatalf("Ambil gagal: %v", err)
	}
	if p.Nama != "Gula" || p.Harga != 12_000 || p.Stok != 20 || p.Kategori != "bahan" {
		t.Errorf("produk = %+v, tidak sesuai yang disimpan", p)
	}
}

// sql.ErrNoRows harus diterjemahkan ke sentinel aplikasi.
func TestAmbilTidakDitemukan(t *testing.T) {
	r := repoUji(t)

	_, err := r.Ambil(9999)
	if !errors.Is(err, ErrTidakDitemukan) {
		t.Errorf("error = %v, ingin ErrTidakDitemukan", err)
	}
}

func TestSelectPerKategori(t *testing.T) {
	r := repoUji(t)
	seed(t, r)

	tests := []struct {
		kategori string
		wantN    int
	}{
		{"minuman", 2},
		{"makanan", 1},
		{"tidak-ada", 0}, // Select mengembalikan slice kosong, BUKAN error
	}

	for _, tt := range tests {
		t.Run(tt.kategori, func(t *testing.T) {
			got, err := r.PerKategori(tt.kategori)
			if err != nil {
				t.Fatalf("PerKategori error: %v", err)
			}
			if len(got) != tt.wantN {
				t.Errorf("dapat %d produk, ingin %d", len(got), tt.wantN)
			}
		})
	}
}

// Select mengurutkan berdasarkan harga — pastikan ORDER BY benar-benar berlaku.
func TestSelectTerurut(t *testing.T) {
	r := repoUji(t)
	seed(t, r)

	got, err := r.PerKategori("minuman")
	if err != nil {
		t.Fatalf("PerKategori error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("dapat %d, ingin 2", len(got))
	}
	// Teh (15rb) harus sebelum Kopi (25rb).
	if got[0].Harga > got[1].Harga {
		t.Errorf("urutan salah: %d sebelum %d", got[0].Harga, got[1].Harga)
	}
}

// Named insert harus menghasilkan baris yang identik dengan insert biasa.
func TestNamedInsert(t *testing.T) {
	r := repoUji(t)

	id, err := r.BuatNamed(Produk{Nama: "Susu", Harga: 18_000, Stok: 12, Kategori: "minuman"})
	if err != nil {
		t.Fatalf("BuatNamed gagal: %v", err)
	}

	p, err := r.Ambil(id)
	if err != nil {
		t.Fatalf("Ambil gagal: %v", err)
	}
	if p.Nama != "Susu" || p.Stok != 12 {
		t.Errorf("produk named = %+v, tidak sesuai", p)
	}
}

func TestAmbilBanyak(t *testing.T) {
	r := repoUji(t)
	seed(t, r) // id 1,2,3

	tests := []struct {
		nama  string
		ids   []int64
		wantN int
	}{
		{"tiga id ada semua", []int64{1, 2, 3}, 3},
		{"sebagian ada", []int64{1, 999}, 1},
		{"tak ada yang cocok", []int64{998, 999}, 0},
		{"satu id", []int64{2}, 1},
		{"slice nil", nil, 0},
		{"slice kosong", []int64{}, 0},
		{"id duplikat", []int64{1, 1, 1}, 1},
	}

	for _, tt := range tests {
		t.Run(tt.nama, func(t *testing.T) {
			got, err := r.AmbilBanyak(tt.ids)
			if err != nil {
				t.Fatalf("AmbilBanyak error: %v", err)
			}
			if len(got) != tt.wantN {
				t.Errorf("AmbilBanyak(%v) = %d produk, ingin %d", tt.ids, len(got), tt.wantN)
			}
		})
	}
}

// Membuktikan sqlx.In benar-benar meratakan slice jadi banyak placeholder.
func TestInMeratakanPlaceholder(t *testing.T) {
	query, args, err := sqlx.In("SELECT * FROM produk WHERE id IN (?)", []int64{1, 2, 3})
	if err != nil {
		t.Fatalf("sqlx.In error: %v", err)
	}
	// Satu "?" harus menjadi tiga.
	wantPlaceholder := "SELECT * FROM produk WHERE id IN (?, ?, ?)"
	if query != wantPlaceholder {
		t.Errorf("query = %q, ingin %q", query, wantPlaceholder)
	}
	if len(args) != 3 {
		t.Errorf("jumlah args = %d, ingin 3", len(args))
	}
}

func TestKurangiStok(t *testing.T) {
	r := repoUji(t)
	seed(t, r)

	// id 1 = Kopi, stok 10.
	if err := r.KurangiStok(1, 4); err != nil {
		t.Fatalf("KurangiStok gagal: %v", err)
	}
	p, err := r.Ambil(1)
	if err != nil {
		t.Fatalf("Ambil gagal: %v", err)
	}
	if p.Stok != 6 {
		t.Errorf("stok = %d, ingin 6", p.Stok)
	}
}

// Transaksi harus DIBATALKAN saat stok tak mencukupi — stok tetap utuh.
func TestKurangiStokTidakCukupMembatalkan(t *testing.T) {
	r := repoUji(t)
	seed(t, r)

	err := r.KurangiStok(1, 9999)
	if err == nil {
		t.Fatal("ingin error karena stok tak cukup")
	}

	// Bukti rollback: stok Kopi harus tetap 10.
	p, err := r.Ambil(1)
	if err != nil {
		t.Fatalf("Ambil gagal: %v", err)
	}
	if p.Stok != 10 {
		t.Errorf("stok = %d setelah transaksi gagal, ingin tetap 10 (rollback)", p.Stok)
	}
}

func TestKurangiStokProdukTakAda(t *testing.T) {
	r := repoUji(t)
	if err := r.KurangiStok(9999, 1); !errors.Is(err, ErrTidakDitemukan) {
		t.Errorf("error = %v, ingin ErrTidakDitemukan", err)
	}
}

func TestBukaSkemaGagalPadaDSNRusak(t *testing.T) {
	// Driver yang tak terdaftar -> Connect gagal.
	if _, err := sqlx.Connect("driver-tak-ada", ":memory:"); err == nil {
		t.Error("ingin error untuk driver yang tak dikenal")
	}
}
