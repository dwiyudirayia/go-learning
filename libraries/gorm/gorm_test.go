package main

import (
	"errors"
	"fmt"
	"sync/atomic"
	"testing"

	"gorm.io/gorm"
)

// nomorDB memastikan tiap test memakai database di memori yang BERBEDA.
//
// 🔍 Analogi: kalau semua test memakai nama database yang sama, mereka jadi seperti
// beberapa orang menulis di papan tulis yang sama — test yang satu melihat coretan
// test lainnya. Nama unik = papan tulis sendiri-sendiri.
var nomorDB atomic.Int64

func dbUji(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := fmt.Sprintf("file:uji%d?mode=memory&cache=shared", nomorDB.Add(1))
	db, err := Buka(dsn)
	if err != nil {
		t.Fatalf("Buka gagal: %v", err)
	}
	t.Cleanup(func() {
		sqlDB, err := db.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	})
	return db
}

// penulisUji membuat satu penulis siap pakai.
func penulisUji(t *testing.T, db *gorm.DB, nama, email string) *Penulis {
	t.Helper()
	p, err := BuatPenulis(db, nama, email)
	if err != nil {
		t.Fatalf("BuatPenulis(%q) gagal: %v", nama, err)
	}
	return p
}

func TestBuatPenulisMengisiKolomBaku(t *testing.T) {
	db := dbUji(t)

	p := penulisUji(t, db, "Andrea Hirata", "andrea@contoh.id")

	// GORM mengisi ID & CreatedAt SETELAH Create berhasil.
	if p.ID == 0 {
		t.Error("ID seharusnya terisi setelah Create")
	}
	if p.CreatedAt.IsZero() {
		t.Error("CreatedAt seharusnya terisi otomatis oleh gorm.Model")
	}
	if p.UpdatedAt.IsZero() {
		t.Error("UpdatedAt seharusnya terisi otomatis")
	}
}

func TestAmbilPenulis(t *testing.T) {
	db := dbUji(t)
	dibuat := penulisUji(t, db, "Dee Lestari", "dee@contoh.id")

	got, err := AmbilPenulis(db, dibuat.ID)
	if err != nil {
		t.Fatalf("AmbilPenulis gagal: %v", err)
	}
	if got.Nama != "Dee Lestari" || got.Email != "dee@contoh.id" {
		t.Errorf("penulis = %+v, ingin Dee Lestari", got)
	}
}

// Error GORM harus diterjemahkan ke sentinel milik aplikasi.
func TestAmbilPenulisTidakDitemukan(t *testing.T) {
	db := dbUji(t)

	_, err := AmbilPenulis(db, 9999)
	if err == nil {
		t.Fatal("ingin error untuk ID yang tak ada")
	}
	if !errors.Is(err, ErrTidakDitemukan) {
		t.Errorf("error = %v, ingin membungkus ErrTidakDitemukan", err)
	}
	// gorm.ErrRecordNotFound boleh tetap ada di rantai, tapi pemanggil cukup
	// mengenali sentinel milik aplikasi — itulah gunanya penerjemahan ini.
}

func TestEmailHarusUnik(t *testing.T) {
	db := dbUji(t)
	penulisUji(t, db, "Pertama", "sama@contoh.id")

	if _, err := BuatPenulis(db, "Kedua", "sama@contoh.id"); err == nil {
		t.Error("email kembar seharusnya ditolak oleh uniqueIndex")
	}
}

// Membuktikan klaim di komentar: indeks unik TETAP melihat baris yang di-soft-delete.
func TestSoftDeleteTetapMemblokirIndeksUnik(t *testing.T) {
	db := dbUji(t)
	p := penulisUji(t, db, "Akan Dihapus", "daur-ulang@contoh.id")

	if err := db.Delete(&Penulis{}, p.ID).Error; err != nil {
		t.Fatalf("Delete gagal: %v", err)
	}

	// Barisnya "hilang" dari query biasa...
	if _, err := AmbilPenulis(db, p.ID); !errors.Is(err, ErrTidakDitemukan) {
		t.Errorf("setelah dihapus, AmbilPenulis = %v, ingin ErrTidakDitemukan", err)
	}
	// ...tapi emailnya MASIH memblokir pendaftaran ulang. Inilah kejutan soft delete.
	if _, err := BuatPenulis(db, "Pengguna Baru", "daur-ulang@contoh.id"); err == nil {
		t.Error("email milik baris terhapus seharusnya masih memblokir — " +
			"kalau ini lolos, komentar jebakan di main.go perlu diperbarui")
	}
}

func TestCariBuku(t *testing.T) {
	db := dbUji(t)
	p := penulisUji(t, db, "Penulis", "cari@contoh.id")

	for _, b := range []struct {
		judul string
		tahun int
	}{
		{"Buku 2005", 2005},
		{"Buku 2006", 2006},
		{"Buku 2010", 2010},
		{"Buku 2020", 2020},
	} {
		if _, err := BuatBuku(db, p.ID, b.judul, b.tahun, 50_000); err != nil {
			t.Fatalf("BuatBuku gagal: %v", err)
		}
	}

	tests := []struct {
		nama     string
		tahunMin int
		wantN    int
	}{
		{"semua", 2000, 4},
		{"sejak 2006", 2006, 3},
		{"sejak 2010", 2010, 2},
		{"tidak ada yang cocok", 2030, 0},
	}

	for _, tt := range tests {
		t.Run(tt.nama, func(t *testing.T) {
			buku, err := CariBuku(db, tt.tahunMin)
			// Find TIDAK menganggap hasil kosong sebagai error — ini beda penting
			// dari First/Take/Last.
			if err != nil {
				t.Fatalf("CariBuku error: %v", err)
			}
			if len(buku) != tt.wantN {
				t.Fatalf("dapat %d buku, ingin %d", len(buku), tt.wantN)
			}
			// Urutan harus menurun berdasarkan tahun.
			for i := 1; i < len(buku); i++ {
				if buku[i-1].Tahun < buku[i].Tahun {
					t.Errorf("urutan salah: %d sebelum %d", buku[i-1].Tahun, buku[i].Tahun)
				}
			}
		})
	}
}

// INI jebakan GORM yang paling sering menggigit: nilai nol diabaikan Updates(struct).
func TestUpdateNilaiNolDiabaikanOlehStruct(t *testing.T) {
	db := dbUji(t)
	p := penulisUji(t, db, "Uji Update", "update@contoh.id")

	b, err := BuatBuku(db, p.ID, "Buku Diskon", 2024, 50_000)
	if err != nil {
		t.Fatalf("BuatBuku gagal: %v", err)
	}

	// Versi struct: perubahan ke 0 hilang diam-diam.
	if err := UbahHargaSalah(db, b.ID, 0); err != nil {
		t.Fatalf("UbahHargaSalah error: %v", err)
	}
	got, err := AmbilBuku(db, b.ID)
	if err != nil {
		t.Fatalf("AmbilBuku gagal: %v", err)
	}
	if got.Harga != 50_000 {
		t.Errorf("harga = %d, ingin tetap 50000 — kalau berubah, jebakannya sudah "+
			"tidak berlaku dan komentar di main.go perlu diperbarui", got.Harga)
	}

	// Versi map: nilai 0 benar-benar tersimpan.
	if err := UbahHargaBenar(db, b.ID, 0); err != nil {
		t.Fatalf("UbahHargaBenar error: %v", err)
	}
	got, err = AmbilBuku(db, b.ID)
	if err != nil {
		t.Fatalf("AmbilBuku gagal: %v", err)
	}
	if got.Harga != 0 {
		t.Errorf("harga = %d, ingin 0 (map selalu menulis apa adanya)", got.Harga)
	}
}

// Nilai bukan-nol tetap tersimpan lewat kedua cara.
func TestUpdateNilaiBukanNolBerhasilDenganStruct(t *testing.T) {
	db := dbUji(t)
	p := penulisUji(t, db, "Uji", "nonzero@contoh.id")

	b, err := BuatBuku(db, p.ID, "Buku", 2024, 50_000)
	if err != nil {
		t.Fatalf("BuatBuku gagal: %v", err)
	}

	if err := UbahHargaSalah(db, b.ID, 75_000); err != nil {
		t.Fatalf("UbahHargaSalah error: %v", err)
	}
	got, err := AmbilBuku(db, b.ID)
	if err != nil {
		t.Fatalf("AmbilBuku gagal: %v", err)
	}
	if got.Harga != 75_000 {
		t.Errorf("harga = %d, ingin 75000", got.Harga)
	}
}

func TestSoftDelete(t *testing.T) {
	db := dbUji(t)
	p := penulisUji(t, db, "Uji Hapus", "hapus@contoh.id")

	b, err := BuatBuku(db, p.ID, "Buku Sementara", 2024, 10_000)
	if err != nil {
		t.Fatalf("BuatBuku gagal: %v", err)
	}

	if err := HapusBuku(db, b.ID); err != nil {
		t.Fatalf("HapusBuku gagal: %v", err)
	}

	terlihat, err := HitungBuku(db)
	if err != nil {
		t.Fatalf("HitungBuku gagal: %v", err)
	}
	if terlihat != 0 {
		t.Errorf("buku terlihat = %d, ingin 0", terlihat)
	}

	// Barisnya MASIH ADA di tabel — cuma disembunyikan.
	total, err := HitungBukuTermasukTerhapus(db)
	if err != nil {
		t.Fatalf("HitungBukuTermasukTerhapus gagal: %v", err)
	}
	if total != 1 {
		t.Errorf("total baris = %d, ingin 1 (soft delete tidak membuang baris)", total)
	}

	if _, err := AmbilBuku(db, b.ID); !errors.Is(err, ErrTidakDitemukan) {
		t.Errorf("AmbilBuku setelah soft delete = %v, ingin ErrTidakDitemukan", err)
	}
}

func TestHapusPermanen(t *testing.T) {
	db := dbUji(t)
	p := penulisUji(t, db, "Uji", "permanen@contoh.id")

	b, err := BuatBuku(db, p.ID, "Buku", 2024, 10_000)
	if err != nil {
		t.Fatalf("BuatBuku gagal: %v", err)
	}

	if err := HapusPermanen(db, b.ID); err != nil {
		t.Fatalf("HapusPermanen gagal: %v", err)
	}

	total, err := HitungBukuTermasukTerhapus(db)
	if err != nil {
		t.Fatalf("HitungBukuTermasukTerhapus gagal: %v", err)
	}
	if total != 0 {
		t.Errorf("total baris = %d, ingin 0 — Unscoped().Delete harus benar-benar membuang", total)
	}
}

// Tanpa Preload, slice relasi kosong — dan itu MENYESATKAN, bukan berarti tak punya buku.
func TestPreloadMemuatRelasi(t *testing.T) {
	db := dbUji(t)
	p := penulisUji(t, db, "Dee Lestari", "preload@contoh.id")

	judul := []string{"Supernova", "Filosofi Kopi", "Perahu Kertas"}
	for _, j := range judul {
		if _, err := BuatBuku(db, p.ID, j, 2010, 75_000); err != nil {
			t.Fatalf("BuatBuku gagal: %v", err)
		}
	}

	tanpa, err := AmbilTanpaPreload(db, p.ID)
	if err != nil {
		t.Fatalf("AmbilTanpaPreload gagal: %v", err)
	}
	if len(tanpa.Buku) != 0 {
		t.Errorf("tanpa Preload dapat %d buku, ingin 0 (relasi tidak dimuat)", len(tanpa.Buku))
	}

	dengan, err := AmbilPenulisDenganBuku(db, p.ID)
	if err != nil {
		t.Fatalf("AmbilPenulisDenganBuku gagal: %v", err)
	}
	if len(dengan.Buku) != len(judul) {
		t.Errorf("dengan Preload dapat %d buku, ingin %d", len(dengan.Buku), len(judul))
	}
}

func TestPreloadPenulisTidakDitemukan(t *testing.T) {
	db := dbUji(t)
	if _, err := AmbilPenulisDenganBuku(db, 9999); !errors.Is(err, ErrTidakDitemukan) {
		t.Errorf("error = %v, ingin ErrTidakDitemukan", err)
	}
}

// Preload pada DAFTAR tetap dua query, berapa pun jumlah penulisnya — bukan N+1.
func TestPreloadPadaDaftar(t *testing.T) {
	db := dbUji(t)

	for i := 1; i <= 3; i++ {
		p := penulisUji(t, db, fmt.Sprintf("Penulis %d", i), fmt.Sprintf("p%d@contoh.id", i))
		for j := 1; j <= i; j++ { // penulis ke-i punya i buku
			if _, err := BuatBuku(db, p.ID, fmt.Sprintf("Buku %d-%d", i, j), 2020, 50_000); err != nil {
				t.Fatalf("BuatBuku gagal: %v", err)
			}
		}
	}

	ps, err := AmbilSemuaPenulisDenganBuku(db)
	if err != nil {
		t.Fatalf("AmbilSemuaPenulisDenganBuku gagal: %v", err)
	}
	if len(ps) != 3 {
		t.Fatalf("dapat %d penulis, ingin 3", len(ps))
	}
	for i, p := range ps {
		if len(p.Buku) != i+1 {
			t.Errorf("penulis %s punya %d buku, ingin %d", p.Nama, len(p.Buku), i+1)
		}
	}
}

func TestTransaksiSukses(t *testing.T) {
	db := dbUji(t)
	a := penulisUji(t, db, "Penulis A", "a@contoh.id")
	b := penulisUji(t, db, "Penulis B", "b@contoh.id")

	for _, j := range []string{"Buku X", "Buku Y"} {
		if _, err := BuatBuku(db, a.ID, j, 2020, 60_000); err != nil {
			t.Fatalf("BuatBuku gagal: %v", err)
		}
	}

	n, err := PindahkanBuku(db, a.ID, b.ID)
	if err != nil {
		t.Fatalf("PindahkanBuku gagal: %v", err)
	}
	if n != 2 {
		t.Errorf("dipindahkan %d buku, ingin 2", n)
	}

	// Semua buku kini milik B.
	pa, err := AmbilPenulisDenganBuku(db, a.ID)
	if err != nil {
		t.Fatalf("AmbilPenulisDenganBuku gagal: %v", err)
	}
	if len(pa.Buku) != 0 {
		t.Errorf("penulis A masih punya %d buku, ingin 0", len(pa.Buku))
	}

	pb, err := AmbilPenulisDenganBuku(db, b.ID)
	if err != nil {
		t.Fatalf("AmbilPenulisDenganBuku gagal: %v", err)
	}
	if len(pb.Buku) != 2 {
		t.Errorf("penulis B punya %d buku, ingin 2", len(pb.Buku))
	}
}

// Inti transaksi: kalau ada yang gagal, TIDAK ADA perubahan yang tersisa.
func TestTransaksiDibatalkanSaatGagal(t *testing.T) {
	db := dbUji(t)
	a := penulisUji(t, db, "Penulis A", "rollback-a@contoh.id")

	for _, j := range []string{"Buku X", "Buku Y"} {
		if _, err := BuatBuku(db, a.ID, j, 2020, 60_000); err != nil {
			t.Fatalf("BuatBuku gagal: %v", err)
		}
	}

	// Tujuan tak ada -> transaksi harus gagal seluruhnya.
	n, err := PindahkanBuku(db, a.ID, 9999)
	if err == nil {
		t.Fatal("ingin error karena penulis tujuan tak ada")
	}
	if !errors.Is(err, ErrTidakDitemukan) {
		t.Errorf("error = %v, ingin membungkus ErrTidakDitemukan", err)
	}
	if n != 0 {
		t.Errorf("dipindahkan = %d, ingin 0", n)
	}

	// Bukti rollback: buku A harus utuh, tidak berpindah ke mana pun.
	pa, err := AmbilPenulisDenganBuku(db, a.ID)
	if err != nil {
		t.Fatalf("AmbilPenulisDenganBuku gagal: %v", err)
	}
	if len(pa.Buku) != 2 {
		t.Errorf("penulis A punya %d buku setelah transaksi gagal, ingin tetap 2", len(pa.Buku))
	}
}

func TestPindahkanBukuTanpaBukuApaPun(t *testing.T) {
	db := dbUji(t)
	a := penulisUji(t, db, "Kosong", "kosong@contoh.id")
	b := penulisUji(t, db, "Tujuan", "tujuan@contoh.id")

	n, err := PindahkanBuku(db, a.ID, b.ID)
	if err != nil {
		t.Fatalf("PindahkanBuku gagal: %v", err)
	}
	if n != 0 {
		t.Errorf("dipindahkan = %d, ingin 0", n)
	}
}

func TestBukaDSNTidakValid(t *testing.T) {
	// Direktori yang tak ada -> SQLite tak bisa membuat berkasnya.
	if _, err := Buka("/direktori/yang/pasti/tidak/ada/db.sqlite"); err == nil {
		t.Error("ingin error untuk DSN yang tak bisa dibuka")
	}
}
