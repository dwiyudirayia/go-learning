package main

import (
	"errors"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestHashLaluPeriksa(t *testing.T) {
	hash, err := HashKataSandi("rahasia123")
	if err != nil {
		t.Fatalf("HashKataSandi gagal: %v", err)
	}

	// Hash tak boleh sama dengan kata sandi asli, dan harus berbentuk hash bcrypt.
	if hash == "rahasia123" {
		t.Fatal("hash TIDAK boleh sama dengan kata sandi asli")
	}
	if !strings.HasPrefix(hash, "$2") {
		t.Errorf("hash = %q, ingin berawalan $2 (format bcrypt)", hash)
	}

	if err := PeriksaKataSandi(hash, "rahasia123"); err != nil {
		t.Errorf("kata sandi benar seharusnya lolos, dapat: %v", err)
	}
}

func TestKataSandiSalahDitolak(t *testing.T) {
	hash, err := HashKataSandi("benar")
	if err != nil {
		t.Fatalf("HashKataSandi gagal: %v", err)
	}

	tests := []string{"salah", "Benar", "benar ", " benar", "", "benarr"}
	for _, ks := range tests {
		t.Run("coba="+ks, func(t *testing.T) {
			err := PeriksaKataSandi(hash, ks)
			if !errors.Is(err, ErrKataSandiSalah) {
				t.Errorf("PeriksaKataSandi(%q) = %v, ingin ErrKataSandiSalah", ks, err)
			}
		})
	}
}

func TestKataSandiKosongDitolakSaatHash(t *testing.T) {
	if _, err := HashKataSandi(""); !errors.Is(err, ErrKataSandiKosong) {
		t.Errorf("hash kata sandi kosong = %v, ingin ErrKataSandiKosong", err)
	}
}

// Sifat terpenting: kata sandi SAMA menghasilkan hash BERBEDA (karena garam acak).
func TestGaramMembuatHashBerbeda(t *testing.T) {
	const ks = "samasekali"

	hash1, err := HashKataSandi(ks)
	if err != nil {
		t.Fatalf("hash pertama gagal: %v", err)
	}
	hash2, err := HashKataSandi(ks)
	if err != nil {
		t.Fatalf("hash kedua gagal: %v", err)
	}

	if hash1 == hash2 {
		t.Error("dua hash dari kata sandi sama seharusnya BERBEDA — garam sepertinya tak bekerja")
	}
	// Tapi keduanya harus tetap lolos verifikasi.
	if err := PeriksaKataSandi(hash1, ks); err != nil {
		t.Errorf("hash1 gagal diverifikasi: %v", err)
	}
	if err := PeriksaKataSandi(hash2, ks); err != nil {
		t.Errorf("hash2 gagal diverifikasi: %v", err)
	}
}

func TestCostTersimpanDiHash(t *testing.T) {
	tests := []struct {
		nama string
		cost int
		want int
	}{
		{"minimum", bcrypt.MinCost, bcrypt.MinCost},
		{"default", bcrypt.DefaultCost, bcrypt.DefaultCost},
		{"cost 11", 11, 11},
		// Cost di bawah minimum otomatis dinaikkan ke default oleh bcrypt.
		{"di bawah minimum jadi default", 1, bcrypt.DefaultCost},
	}

	for _, tt := range tests {
		t.Run(tt.nama, func(t *testing.T) {
			hash, err := HashDenganCost("uji", tt.cost)
			if err != nil {
				t.Fatalf("HashDenganCost gagal: %v", err)
			}
			got, err := CostDariHash(hash)
			if err != nil {
				t.Fatalf("CostDariHash gagal: %v", err)
			}
			if got != tt.want {
				t.Errorf("cost tersimpan = %d, ingin %d", got, tt.want)
			}
		})
	}
}

func TestCostDariHashRusak(t *testing.T) {
	if _, err := CostDariHash("bukan-hash-bcrypt"); err == nil {
		t.Error("ingin error saat membaca cost dari string yang bukan hash")
	}
}

func TestPerluRehash(t *testing.T) {
	hashLama, err := HashDenganCost("rahasia", bcrypt.MinCost)
	if err != nil {
		t.Fatalf("HashDenganCost gagal: %v", err)
	}

	// Cost lama (4) di bawah target (10) -> perlu rehash.
	perlu, err := PerluRehash(hashLama, bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("PerluRehash gagal: %v", err)
	}
	if !perlu {
		t.Error("hash cost rendah seharusnya perlu di-rehash ke cost lebih tinggi")
	}

	// Hash yang sudah pada cost target tak perlu diganti.
	hashBaru, err := HashDenganCost("rahasia", bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("HashDenganCost gagal: %v", err)
	}
	perlu, err = PerluRehash(hashBaru, bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("PerluRehash gagal: %v", err)
	}
	if perlu {
		t.Error("hash yang sudah pada cost target tidak perlu di-rehash")
	}
}

func TestLoginDanMungkinRehash(t *testing.T) {
	hashLama, err := HashDenganCost("rahasia123", bcrypt.MinCost)
	if err != nil {
		t.Fatalf("HashDenganCost gagal: %v", err)
	}

	t.Run("login benar memicu rehash", func(t *testing.T) {
		hashBaru, err := LoginDanMungkinRehash(hashLama, "rahasia123", 11)
		if err != nil {
			t.Fatalf("LoginDanMungkinRehash gagal: %v", err)
		}
		if hashBaru == "" {
			t.Fatal("ingin hash baru karena cost lama lebih rendah")
		}
		// Hash baru harus bercost 11 DAN tetap cocok dengan kata sandi asli.
		if c, _ := CostDariHash(hashBaru); c != 11 {
			t.Errorf("cost hash baru = %d, ingin 11", c)
		}
		if err := PeriksaKataSandi(hashBaru, "rahasia123"); err != nil {
			t.Errorf("hash baru gagal diverifikasi: %v", err)
		}
	})

	t.Run("kata sandi salah tidak menghasilkan hash", func(t *testing.T) {
		hashBaru, err := LoginDanMungkinRehash(hashLama, "salah", 11)
		if !errors.Is(err, ErrKataSandiSalah) {
			t.Errorf("error = %v, ingin ErrKataSandiSalah", err)
		}
		if hashBaru != "" {
			t.Error("kata sandi salah tidak boleh menghasilkan hash baru")
		}
	})

	t.Run("cost sudah cukup tidak rehash", func(t *testing.T) {
		hashBaru, err := LoginDanMungkinRehash(hashLama, "rahasia123", bcrypt.MinCost)
		if err != nil {
			t.Fatalf("LoginDanMungkinRehash gagal: %v", err)
		}
		if hashBaru != "" {
			t.Error("cost lama sudah memenuhi target -> tidak perlu hash baru")
		}
	})
}

// Batas 72 byte: pas diterima, lewat sedikit ditolak (bukan dipotong diam-diam).
func TestBatas72Byte(t *testing.T) {
	tests := []struct {
		nama    string
		panjang int
		wantErr bool
	}{
		{"pendek", 10, false},
		{"tepat 72 byte", 72, false},
		{"73 byte ditolak", 73, true},
		{"jauh melebihi", 200, true},
	}

	for _, tt := range tests {
		t.Run(tt.nama, func(t *testing.T) {
			ks := strings.Repeat("a", tt.panjang)
			_, err := HashKataSandi(ks)
			if tt.wantErr {
				if !errors.Is(err, ErrKataSandiTerlaluPanjang) {
					t.Errorf("panjang %d: error = %v, ingin ErrKataSandiTerlaluPanjang", tt.panjang, err)
				}
				return
			}
			if err != nil {
				t.Errorf("panjang %d seharusnya diterima, dapat: %v", tt.panjang, err)
			}
		})
	}
}

// Hash yang rusak/dipotong harus ditolak saat verifikasi, bukan malah lolos.
func TestHashRusakDitolak(t *testing.T) {
	tests := []string{
		"",
		"bukan-hash",
		"$2a$10$", // prefix benar tapi terpotong
	}
	for _, h := range tests {
		t.Run("hash="+h, func(t *testing.T) {
			if err := PeriksaKataSandi(h, "apa saja"); err == nil {
				t.Error("hash rusak seharusnya menghasilkan error, bukan lolos")
			}
		})
	}
}
