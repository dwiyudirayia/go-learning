// gorm.io/gorm — ORM: memetakan struct Go ke tabel database.
//
// Jalankan: go run ./libraries/gorm     (SQLite di memori — tanpa perlu database)
// Test:     go test ./libraries/gorm
//
// 🔍 Analogi besar: menulis SQL sendiri itu MASAK DARI NOL — kamu mengendalikan tiap
// bumbu, hasilnya persis seperti yang kamu mau, tapi memasak 50 menu memakan waktu.
// ORM itu KATERING: sebut "ambilkan semua buku milik penulis ini beserta penulisnya",
// dan ia menyusun SQL-nya untukmu.
//
// Harga yang dibayar: kamu tak lagi melihat SQL yang benar-benar dijalankan. Itu nyaman
// sampai suatu hari halamanmu jadi lambat dan kamu tak tahu kenapa — biasanya karena
// masalah N+1 (dibahas di bagian 5). Karena itu: pakai ORM untuk 90% operasi
// yang membosankan, dan jangan ragu turun ke SQL mentah untuk query yang rumit.
package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	fmt.Println("=== gorm.io/gorm ===")

	db, err := Buka("file:contoh?mode=memory&cache=shared")
	if err != nil {
		fmt.Println("gagal membuka database:", err)
		return
	}

	demoCRUD(db)
	demoJebakanUpdate(db)
	demoSoftDelete(db)
	demoRelasi(db)
	demoTransaksi(db)
}

// ------------------------------------------------------------------
// 1. Model
// ------------------------------------------------------------------

// Penulis punya banyak Buku.
//
// 🔍 Analogi gorm.Model: ini KOP SURAT BAKU yang disematkan ke tiap tabel — berisi
// ID, CreatedAt, UpdatedAt, dan DeletedAt. Menyematkannya berarti "tabel ini mengikuti
// kebiasaan umum": punya kunci utama angka, dicatat kapan dibuat/diubah, dan mendukung
// penghapusan lunak. Kalau tabelmu tak butuh itu (mis. tabel referensi statis),
// jangan sematkan — jangan ikut-ikutan hanya karena semua contoh memakainya.
type Penulis struct {
	gorm.Model
	Nama  string `gorm:"size:100;not null;index"`
	Email string `gorm:"size:150;uniqueIndex"`
	Buku  []Buku // relasi has-many: diisi hanya bila di-Preload
}

// Buku milik satu Penulis.
//
// 🔍 Analogi PenulisID: ini NOMOR RAK tempat bukunya berasal. Konvensi GORM: nama field
// relasi + "ID" otomatis dikenali sebagai kunci asing. Kalau kolommu bernama lain,
// kamu harus menyebutkannya eksplisit lewat tag — GORM tidak bisa menebak.
type Buku struct {
	gorm.Model
	Judul     string `gorm:"size:200;not null"`
	Tahun     int    `gorm:"index"`
	Harga     int
	PenulisID uint     `gorm:"index"`
	Penulis   *Penulis `gorm:"foreignKey:PenulisID"`
}

// ------------------------------------------------------------------
// 2. Membuka koneksi
// ------------------------------------------------------------------

// Buka menyambung ke SQLite dan menyiapkan tabelnya.
func Buka(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		// Logger disenyapkan agar keluaran contoh tetap rapi.
		// Saat menyelidiki query lambat, ganti ke logger.Info — kamu akan MELIHAT
		// SQL yang sebenarnya dijalankan GORM. Ini cara tercepat menemukan N+1.
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("gagal membuka database: %w", err)
	}

	// 🔍 Analogi AutoMigrate: seperti TUKANG yang membandingkan denah (struct) dengan
	// bangunan yang ada (tabel), lalu menambahkan yang kurang. Ia MENAMBAH kolom & indeks,
	// tapi tak pernah MENGHAPUS kolom — supaya tak ada data yang lenyap tanpa sengaja.
	//
	// Cocok untuk pengembangan & prototipe. Untuk produksi, pakai migrasi bernomor
	// yang bisa ditinjau & di-rollback (lihat modul 21) — perubahan skema adalah hal
	// yang harus kamu kendalikan sepenuhnya, bukan diserahkan pada tebakan.
	if err := db.AutoMigrate(&Penulis{}, &Buku{}); err != nil {
		return nil, fmt.Errorf("gagal menyiapkan tabel: %w", err)
	}
	return db, nil
}

// ------------------------------------------------------------------
// 3. CRUD dasar
// ------------------------------------------------------------------

// ErrTidakDitemukan sentinel milik aplikasi.
//
// 🔍 Analogi: jangan biarkan gorm.ErrRecordNotFound bocor sampai ke lapisan handler.
// Itu seperti membiarkan pelanggan mendengar istilah teknis gudang. Terjemahkan ke
// bahasa aplikasimu sendiri, supaya suatu hari kamu bisa mengganti GORM tanpa
// mengubah seluruh kode di atasnya.
var ErrTidakDitemukan = errors.New("data tidak ditemukan")

// BuatPenulis menyimpan penulis baru. GORM mengisi ID & CreatedAt setelah berhasil.
func BuatPenulis(db *gorm.DB, nama, email string) (*Penulis, error) {
	p := &Penulis{Nama: nama, Email: email}
	if err := db.Create(p).Error; err != nil {
		return nil, fmt.Errorf("gagal membuat penulis %q: %w", nama, err)
	}
	return p, nil
}

// AmbilPenulis mencari berdasarkan ID.
func AmbilPenulis(db *gorm.DB, id uint) (*Penulis, error) {
	var p Penulis
	err := db.First(&p, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("penulis %d: %w", id, ErrTidakDitemukan)
	}
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil penulis %d: %w", id, err)
	}
	return &p, nil
}

// BuatBuku menyimpan buku milik seorang penulis.
func BuatBuku(db *gorm.DB, penulisID uint, judul string, tahun, harga int) (*Buku, error) {
	b := &Buku{Judul: judul, Tahun: tahun, Harga: harga, PenulisID: penulisID}
	if err := db.Create(b).Error; err != nil {
		return nil, fmt.Errorf("gagal membuat buku %q: %w", judul, err)
	}
	return b, nil
}

// CariBuku menyaring berdasarkan tahun terbit minimal, diurutkan dari yang terbaru.
func CariBuku(db *gorm.DB, tahunMin int) ([]Buku, error) {
	var buku []Buku
	err := db.Where("tahun >= ?", tahunMin).
		Order("tahun desc").
		Find(&buku).Error
	if err != nil {
		return nil, fmt.Errorf("gagal mencari buku: %w", err)
	}
	// Catatan: Find TIDAK menganggap "tak ada hasil" sebagai error — ia mengembalikan
	// slice kosong. Yang mengembalikan ErrRecordNotFound hanyalah First/Take/Last.
	return buku, nil
}

func demoCRUD(db *gorm.DB) {
	fmt.Println("\n-- CRUD dasar --")

	p, err := BuatPenulis(db, "Andrea Hirata", "andrea@contoh.id")
	if err != nil {
		fmt.Println("   error:", err)
		return
	}
	fmt.Printf("   penulis dibuat -> ID=%d dibuat pada %s\n",
		p.ID, p.CreatedAt.Format(time.Kitchen))

	if _, err := BuatBuku(db, p.ID, "Laskar Pelangi", 2005, 95_000); err != nil {
		fmt.Println("   error:", err)
		return
	}
	if _, err := BuatBuku(db, p.ID, "Sang Pemimpi", 2006, 88_000); err != nil {
		fmt.Println("   error:", err)
		return
	}

	buku, err := CariBuku(db, 2006)
	if err != nil {
		fmt.Println("   error:", err)
		return
	}
	fmt.Printf("   buku terbit >= 2006: %d\n", len(buku))

	if _, err := AmbilPenulis(db, 9999); errors.Is(err, ErrTidakDitemukan) {
		fmt.Println("   penulis 9999 ->", err)
	}
}

// ------------------------------------------------------------------
// 4. Jebakan update: nilai nol diabaikan
// ------------------------------------------------------------------

// 🔍 Analogi: Updates dengan STRUCT itu seperti petugas yang berasumsi "kolom yang
// kosong berarti tidak ingin diubah". Masuk akal — sampai kamu benar-benar ingin
// mengubah harga menjadi 0 (barang gratis) atau status menjadi false. Permintaanmu
// akan DIABAIKAN diam-diam, karena 0 dan false tak bisa dibedakan dari "tidak diisi".
//
// Ada dua jalan keluar, keduanya membuat maksudmu eksplisit:
//   - Updates dengan MAP: apa pun yang ada di map pasti ditulis.
//   - Select("kolom"): "tulis kolom ini, apa pun isinya".

// UbahHargaSalah memakai struct — harga 0 TIDAK akan tersimpan.
func UbahHargaSalah(db *gorm.DB, id uint, harga int) error {
	return db.Model(&Buku{}).Where("id = ?", id).Updates(Buku{Harga: harga}).Error
}

// UbahHargaBenar memakai map — nilai berapa pun, termasuk 0, pasti tersimpan.
func UbahHargaBenar(db *gorm.DB, id uint, harga int) error {
	err := db.Model(&Buku{}).Where("id = ?", id).
		Updates(map[string]any{"harga": harga}).Error
	if err != nil {
		return fmt.Errorf("gagal mengubah harga buku %d: %w", id, err)
	}
	return nil
}

// AmbilBuku mengambil satu buku (dipakai untuk memeriksa hasil update).
func AmbilBuku(db *gorm.DB, id uint) (*Buku, error) {
	var b Buku
	err := db.First(&b, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("buku %d: %w", id, ErrTidakDitemukan)
	}
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil buku %d: %w", id, err)
	}
	return &b, nil
}

func demoJebakanUpdate(db *gorm.DB) {
	fmt.Println("\n-- Jebakan: update nilai nol --")

	p, err := BuatPenulis(db, "Uji Update", "update@contoh.id")
	if err != nil {
		fmt.Println("   error:", err)
		return
	}
	b, err := BuatBuku(db, p.ID, "Buku Diskon", 2024, 50_000)
	if err != nil {
		fmt.Println("   error:", err)
		return
	}

	_ = UbahHargaSalah(db, b.ID, 0)
	if got, err := AmbilBuku(db, b.ID); err == nil {
		fmt.Printf("   pakai struct -> harga %d (perubahan ke 0 DIABAIKAN)\n", got.Harga)
	}

	_ = UbahHargaBenar(db, b.ID, 0)
	if got, err := AmbilBuku(db, b.ID); err == nil {
		fmt.Printf("   pakai map    -> harga %d (tersimpan)\n", got.Harga)
	}
}

// ------------------------------------------------------------------
// 5. Soft delete
// ------------------------------------------------------------------

// 🔍 Analogi: gorm.Model membawa kolom DeletedAt, dan itu mengubah arti Delete secara
// diam-diam. Data tidak benar-benar dibuang — ia dipindahkan ke TEMPAT SAMPAH: barisnya
// masih ada, cuma diberi cap waktu penghapusan, dan semua query berikutnya otomatis
// berpura-pura ia tak ada.
//
// Bagus untuk pemulihan tak sengaja & jejak audit. Tapi ingat dua hal:
//  1. Data itu MASIH ADA. Kalau pengguna meminta datanya benar-benar dihapus
//     (hak untuk dilupakan), soft delete saja TIDAK cukup secara hukum.
//  2. Indeks unik tetap melihat baris terhapus — email yang "sudah dihapus"
//     masih memblokir pendaftaran ulang dengan email yang sama.

// HapusBuku melakukan penghapusan lunak.
func HapusBuku(db *gorm.DB, id uint) error {
	if err := db.Delete(&Buku{}, id).Error; err != nil {
		return fmt.Errorf("gagal menghapus buku %d: %w", id, err)
	}
	return nil
}

// HitungBuku menghitung buku yang terlihat (yang terhapus tak dihitung).
func HitungBuku(db *gorm.DB) (int64, error) {
	var n int64
	if err := db.Model(&Buku{}).Count(&n).Error; err != nil {
		return 0, fmt.Errorf("gagal menghitung buku: %w", err)
	}
	return n, nil
}

// HitungBukuTermasukTerhapus memakai Unscoped untuk mengintip isi tempat sampah.
func HitungBukuTermasukTerhapus(db *gorm.DB) (int64, error) {
	var n int64
	if err := db.Unscoped().Model(&Buku{}).Count(&n).Error; err != nil {
		return 0, fmt.Errorf("gagal menghitung buku: %w", err)
	}
	return n, nil
}

// HapusPermanen benar-benar membuang barisnya dari tabel.
func HapusPermanen(db *gorm.DB, id uint) error {
	if err := db.Unscoped().Delete(&Buku{}, id).Error; err != nil {
		return fmt.Errorf("gagal menghapus permanen buku %d: %w", id, err)
	}
	return nil
}

func demoSoftDelete(db *gorm.DB) {
	fmt.Println("\n-- Soft delete --")

	p, err := BuatPenulis(db, "Uji Hapus", "hapus@contoh.id")
	if err != nil {
		fmt.Println("   error:", err)
		return
	}
	b, err := BuatBuku(db, p.ID, "Buku Sementara", 2024, 10_000)
	if err != nil {
		fmt.Println("   error:", err)
		return
	}

	sebelum, _ := HitungBuku(db)
	if err := HapusBuku(db, b.ID); err != nil {
		fmt.Println("   error:", err)
		return
	}
	sesudah, _ := HitungBuku(db)
	total, _ := HitungBukuTermasukTerhapus(db)

	fmt.Printf("   terlihat: %d -> %d, tapi di tabel masih ada %d baris\n",
		sebelum, sesudah, total)

	if _, err := AmbilBuku(db, b.ID); errors.Is(err, ErrTidakDitemukan) {
		fmt.Println("   query biasa tak lagi menemukannya")
	}
}

// ------------------------------------------------------------------
// 6. Relasi & masalah N+1
// ------------------------------------------------------------------

// 🔍 Analogi N+1 — INI penyakit paling umum pengguna ORM:
// kamu mengambil 100 penulis (1 query), lalu untuk tiap penulis mengambil bukunya
// (100 query lagi). Total 101 perjalanan ke database. Seperti belanja 100 barang
// dengan bolak-balik ke toko 100 kali, padahal sekali jalan cukup.
//
// Preload adalah DAFTAR BELANJA: GORM mengambil semua penulis, lalu mengambil SEMUA
// buku milik mereka dalam satu query tambahan (WHERE penulis_id IN (...)). Dua query,
// bukan 101. Gejalanya di produksi: halaman yang makin lambat seiring data bertambah.

// AmbilPenulisDenganBuku mengambil penulis beserta bukunya — dua query saja.
func AmbilPenulisDenganBuku(db *gorm.DB, id uint) (*Penulis, error) {
	var p Penulis
	err := db.Preload("Buku").First(&p, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("penulis %d: %w", id, ErrTidakDitemukan)
	}
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil penulis %d: %w", id, err)
	}
	return &p, nil
}

// AmbilSemuaPenulisDenganBuku versi daftar — tetap dua query, berapa pun penulisnya.
func AmbilSemuaPenulisDenganBuku(db *gorm.DB) ([]Penulis, error) {
	var ps []Penulis
	if err := db.Preload("Buku").Find(&ps).Error; err != nil {
		return nil, fmt.Errorf("gagal mengambil daftar penulis: %w", err)
	}
	return ps, nil
}

// AmbilTanpaPreload sengaja TIDAK memuat relasi — untuk membuktikan bedanya.
func AmbilTanpaPreload(db *gorm.DB, id uint) (*Penulis, error) {
	var p Penulis
	if err := db.First(&p, id).Error; err != nil {
		return nil, fmt.Errorf("gagal mengambil penulis %d: %w", id, err)
	}
	return &p, nil // p.Buku akan kosong — bukan berarti penulisnya tak punya buku!
}

func demoRelasi(db *gorm.DB) {
	fmt.Println("\n-- Relasi & Preload --")

	p, err := BuatPenulis(db, "Dee Lestari", "dee@contoh.id")
	if err != nil {
		fmt.Println("   error:", err)
		return
	}
	for _, j := range []string{"Supernova", "Filosofi Kopi", "Perahu Kertas"} {
		if _, err := BuatBuku(db, p.ID, j, 2010, 75_000); err != nil {
			fmt.Println("   error:", err)
			return
		}
	}

	tanpa, err := AmbilTanpaPreload(db, p.ID)
	if err != nil {
		fmt.Println("   error:", err)
		return
	}
	fmt.Printf("   tanpa Preload -> %d buku (kosong & menyesatkan!)\n", len(tanpa.Buku))

	dengan, err := AmbilPenulisDenganBuku(db, p.ID)
	if err != nil {
		fmt.Println("   error:", err)
		return
	}
	fmt.Printf("   dengan Preload -> %d buku\n", len(dengan.Buku))
}

// ------------------------------------------------------------------
// 7. Transaksi
// ------------------------------------------------------------------

// PindahkanBuku memindahkan semua buku dari satu penulis ke penulis lain.
//
// 🔍 Analogi transaksi: seperti TRANSFER BANK. Uang keluar dari rekening A dan masuk
// ke rekening B harus terjadi SEKALIGUS — kalau listrik mati di tengah, jangan sampai
// uang sudah keluar tapi belum masuk. db.Transaction menjamin itu: kalau fungsi di
// dalamnya mengembalikan error (atau panic), SEMUA perubahan dibatalkan seolah tak
// pernah terjadi.
//
// Perhatikan: di dalam blok ini kamu WAJIB memakai 'tx', bukan 'db'. Memakai db akan
// menjalankan query di luar transaksi — dan perubahannya tak akan ikut dibatalkan.
func PindahkanBuku(db *gorm.DB, dariID, keID uint) (int64, error) {
	var dipindah int64

	err := db.Transaction(func(tx *gorm.DB) error {
		// Pastikan penulis tujuan benar-benar ada sebelum apa pun diubah.
		var tujuan Penulis
		if err := tx.First(&tujuan, keID).Error; err != nil {
			return fmt.Errorf("penulis tujuan %d tidak ada: %w", keID, ErrTidakDitemukan)
		}

		hasil := tx.Model(&Buku{}).
			Where("penulis_id = ?", dariID).
			Update("penulis_id", keID)
		if hasil.Error != nil {
			return fmt.Errorf("gagal memindahkan buku: %w", hasil.Error)
		}
		dipindah = hasil.RowsAffected
		return nil
	})
	if err != nil {
		return 0, err
	}
	return dipindah, nil
}

func demoTransaksi(db *gorm.DB) {
	fmt.Println("\n-- Transaksi --")

	a, err := BuatPenulis(db, "Penulis A", "a@contoh.id")
	if err != nil {
		fmt.Println("   error:", err)
		return
	}
	b, err := BuatPenulis(db, "Penulis B", "b@contoh.id")
	if err != nil {
		fmt.Println("   error:", err)
		return
	}
	for _, j := range []string{"Buku X", "Buku Y"} {
		if _, err := BuatBuku(db, a.ID, j, 2020, 60_000); err != nil {
			fmt.Println("   error:", err)
			return
		}
	}

	n, err := PindahkanBuku(db, a.ID, b.ID)
	if err != nil {
		fmt.Println("   error:", err)
		return
	}
	fmt.Printf("   %d buku dipindahkan dari A ke B\n", n)

	if _, err := PindahkanBuku(db, b.ID, 9999); err != nil {
		fmt.Println("   tujuan tak ada -> seluruh transaksi dibatalkan")
	}
}
