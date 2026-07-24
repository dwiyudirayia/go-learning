// jmoiron/sqlx — perpanjangan tipis dari database/sql: scan langsung ke struct.
//
// Jalankan: go run ./libraries/sqlx     (SQLite in-memory pure-Go — tanpa database)
// Test:     go test ./libraries/sqlx
//
// 🔍 Analogi besar: `database/sql` bawaan Go itu seperti MENGAMBIL BELANJAAN SATU-SATU dari
// rak lalu menaruhnya sendiri ke tas — kamu harus menulis `rows.Scan(&a, &b, &c, ...)` dan
// mencocokkan urutan kolom secara manual, tiap query. Melelahkan dan gampang salah urutan.
//
// sqlx itu KASIR SWALAYAN yang menata belanjaan ke tasmu otomatis: `db.Get(&produk, ...)`
// langsung mengisi struct. Ia BUKAN ORM — kamu tetap menulis SQL sendiri (kendali penuh,
// tak ada query tersembunyi seperti GORM), sqlx cuma menghapus pekerjaan scan yang berulang.
//
// Posisi di spektrum: `database/sql` (paling mentah) → **sqlx** (SQL + scan otomatis) →
// `sqlc` (SQL + kode di-generate, modul 36) → GORM (ORM penuh, modul 14). sqlx dipilih saat
// kamu ingin SQL apa adanya tapi malas menulis Scan berulang.
package main

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite" // driver database/sql bernama "sqlite" (pure-Go, tanpa cgo)
)

func main() {
	fmt.Println("=== jmoiron/sqlx ===")

	db, err := Buka(":memory:")
	if err != nil {
		fmt.Println("gagal membuka database:", err)
		return
	}
	defer db.Close()

	repo := NewProdukRepo(db)
	demoGetSelect(repo)
	demoNamed(repo)
	demoInDanNil(repo, db)
	demoTransaksi(repo)
}

// ------------------------------------------------------------------
// Model
// ------------------------------------------------------------------

// Produk — perhatikan tag `db`. Inilah yang dipakai sqlx untuk mencocokkan
// KOLOM database dengan FIELD struct.
//
// 🔍 Analogi tag db: seperti LABEL RAK. Tanpa label, sqlx menebak dari nama field
// (huruf besar/kecil diabaikan), jadi kolom "harga" cocok dengan field "Harga". Tapi kolom
// "penulis_id" TAK cocok dengan "PenulisID" — di situ kamu WAJIB memberi tag `db:"penulis_id"`,
// kalau tidak sqlx error "missing destination". Selalu beri tag eksplisit agar tak kaget.
type Produk struct {
	ID       int64  `db:"id"`
	Nama     string `db:"nama"`
	Harga    int    `db:"harga"`
	Stok     int    `db:"stok"`
	Kategori string `db:"kategori"`
}

// ------------------------------------------------------------------
// Membuka & menyiapkan skema
// ------------------------------------------------------------------

const skema = `
CREATE TABLE produk (
	id       INTEGER PRIMARY KEY AUTOINCREMENT,
	nama     TEXT NOT NULL,
	harga    INTEGER NOT NULL,
	stok     INTEGER NOT NULL DEFAULT 0,
	kategori TEXT NOT NULL
);`

// Buka menyambung ke SQLite dan membuat tabelnya.
//
// sqlx.Connect = sql.Open + Ping (memastikan koneksi benar-benar hidup). Nama driver
// "sqlite" dari modernc — sama seperti modul 14 & 21.
func Buka(dsn string) (*sqlx.DB, error) {
	db, err := sqlx.Connect("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("gagal menyambung database: %w", err)
	}
	if _, err := db.Exec(skema); err != nil {
		db.Close()
		return nil, fmt.Errorf("gagal membuat skema: %w", err)
	}
	return db, nil
}

// ------------------------------------------------------------------
// Repository
// ------------------------------------------------------------------

// ErrTidakDitemukan sentinel milik aplikasi.
//
// 🔍 Analogi: sqlx.Get mengembalikan sql.ErrNoRows saat tak ada baris — sama seperti
// database/sql biasa. Terjemahkan ke sentinel milikmu agar lapisan atas tak bergantung
// pada detail paket sql (pola yang sama dipakai modul 14 & 15).
var ErrTidakDitemukan = errors.New("produk tidak ditemukan")

type ProdukRepo struct {
	db *sqlx.DB
}

func NewProdukRepo(db *sqlx.DB) *ProdukRepo {
	return &ProdukRepo{db: db}
}

// ------------------------------------------------------------------
// 1. Get (satu baris) & Select (banyak baris)
// ------------------------------------------------------------------

// 🔍 Analogi Get vs Select: keduanya mengisi struct otomatis, bedanya jumlah.
//   Get(&satu, ...)      = ambil SATU produk. Kosong -> sql.ErrNoRows.
//   Select(&banyak, ...) = ambil BANYAK ke slice. Kosong -> slice kosong, BUKAN error.
// Ini menghapus seluruh boilerplate rows.Next()/rows.Scan()/rows.Err() yang biasa.

// Ambil mengambil satu produk berdasarkan ID.
func (r *ProdukRepo) Ambil(id int64) (Produk, error) {
	var p Produk
	err := r.db.Get(&p, "SELECT * FROM produk WHERE id = ?", id)
	if errors.Is(err, sql.ErrNoRows) {
		return Produk{}, fmt.Errorf("id %d: %w", id, ErrTidakDitemukan)
	}
	if err != nil {
		return Produk{}, fmt.Errorf("gagal mengambil produk %d: %w", id, err)
	}
	return p, nil
}

// PerKategori mengambil banyak produk (Select mengisi slice langsung).
func (r *ProdukRepo) PerKategori(kategori string) ([]Produk, error) {
	var hasil []Produk
	err := r.db.Select(&hasil, "SELECT * FROM produk WHERE kategori = ? ORDER BY harga", kategori)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil kategori %q: %w", kategori, err)
	}
	return hasil, nil
}

// Buat menyisipkan produk dan mengembalikan ID barunya.
func (r *ProdukRepo) Buat(p Produk) (int64, error) {
	res, err := r.db.Exec(
		"INSERT INTO produk (nama, harga, stok, kategori) VALUES (?, ?, ?, ?)",
		p.Nama, p.Harga, p.Stok, p.Kategori)
	if err != nil {
		return 0, fmt.Errorf("gagal membuat produk %q: %w", p.Nama, err)
	}
	return res.LastInsertId()
}

func demoGetSelect(r *ProdukRepo) {
	fmt.Println("\n-- Get & Select --")

	_, _ = r.Buat(Produk{Nama: "Kopi", Harga: 25_000, Stok: 10, Kategori: "minuman"})
	_, _ = r.Buat(Produk{Nama: "Teh", Harga: 15_000, Stok: 5, Kategori: "minuman"})
	id, _ := r.Buat(Produk{Nama: "Roti", Harga: 20_000, Stok: 8, Kategori: "makanan"})

	p, err := r.Ambil(id)
	if err != nil {
		fmt.Println("   error:", err)
		return
	}
	fmt.Printf("   Get  -> %+v\n", p)

	minuman, _ := r.PerKategori("minuman")
	fmt.Printf("   Select 'minuman' -> %d produk\n", len(minuman))

	if _, err := r.Ambil(9999); errors.Is(err, ErrTidakDitemukan) {
		fmt.Println("   Ambil(9999) -> tidak ditemukan (sql.ErrNoRows diterjemahkan)")
	}
}

// ------------------------------------------------------------------
// 2. Named query — parameter berdasarkan NAMA, bukan urutan
// ------------------------------------------------------------------

// 🔍 Analogi named query: query biasa pakai tanda tanya berurutan (?, ?, ?) — kalau ada
// 8 parameter, satu salah urut saja bikin bug senyap. Named query pakai ":nama", dan sqlx
// mencocokkannya dengan field struct berdasar NAMA. Seperti mengisi formulir berlabel vs
// menaruh barang di kotak bernomor tanpa label. Sangat mengurangi bug pada INSERT lebar.

// BuatNamed menyisipkan memakai parameter bernama langsung dari struct.
func (r *ProdukRepo) BuatNamed(p Produk) (int64, error) {
	res, err := r.db.NamedExec(
		`INSERT INTO produk (nama, harga, stok, kategori)
		 VALUES (:nama, :harga, :stok, :kategori)`, p)
	if err != nil {
		return 0, fmt.Errorf("gagal named insert %q: %w", p.Nama, err)
	}
	return res.LastInsertId()
}

func demoNamed(r *ProdukRepo) {
	fmt.Println("\n-- Named query --")

	id, err := r.BuatNamed(Produk{Nama: "Gula", Harga: 12_000, Stok: 20, Kategori: "bahan"})
	if err != nil {
		fmt.Println("   error:", err)
		return
	}
	p, _ := r.Ambil(id)
	fmt.Printf("   named insert -> %+v\n", p)
}

// ------------------------------------------------------------------
// 3. sqlx.In — meratakan slice jadi banyak tanda tanya
// ------------------------------------------------------------------

// 🔍 Analogi sqlx.In: ini menyelesaikan masalah klasik "WHERE id IN (?)". SQL tak bisa
// menerima slice di satu tanda tanya; kamu butuh sebanyak (?,?,?) sesuai jumlah isi.
// Menulisnya manual itu rawan & jelek. sqlx.In "meratakan" slice: ia mengubah satu
// placeholder jadi sebanyak yang diperlukan, dan mengembalikan argumen yang sudah pas.
//
// Jebakan yang WAJIB diingat: hasil sqlx.In memakai "?" — pada driver yang butuh gaya lain
// (mis. $1 di Postgres) kamu harus memanggil db.Rebind() setelahnya. Contoh ini pakai
// SQLite (gaya ?), jadi Rebind tak wajib, tapi tetap dipakai agar polanya benar di mana pun.

// AmbilBanyak mengambil produk untuk sekumpulan ID sekaligus.
func (r *ProdukRepo) AmbilBanyak(ids []int64) ([]Produk, error) {
	if len(ids) == 0 {
		return nil, nil // hindari "WHERE id IN ()" yang tak valid
	}

	query, args, err := sqlx.In("SELECT * FROM produk WHERE id IN (?) ORDER BY id", ids)
	if err != nil {
		return nil, fmt.Errorf("gagal menyusun query IN: %w", err)
	}
	query = r.db.Rebind(query) // sesuaikan gaya placeholder ke driver

	var hasil []Produk
	if err := r.db.Select(&hasil, query, args...); err != nil {
		return nil, fmt.Errorf("gagal mengambil banyak produk: %w", err)
	}
	return hasil, nil
}

func demoInDanNil(r *ProdukRepo, _ *sqlx.DB) {
	fmt.Println("\n-- sqlx.In (WHERE id IN (...)) --")

	got, err := r.AmbilBanyak([]int64{1, 2, 3})
	if err != nil {
		fmt.Println("   error:", err)
		return
	}
	fmt.Printf("   AmbilBanyak([1,2,3]) -> %d produk\n", len(got))

	kosong, _ := r.AmbilBanyak(nil)
	fmt.Printf("   AmbilBanyak(nil)     -> %d produk (tak menembak SQL rusak)\n", len(kosong))
}

// ------------------------------------------------------------------
// 4. Transaksi
// ------------------------------------------------------------------

// KurangiStok mengurangi stok dua produk dalam satu transaksi.
//
// 🔍 Analogi: sama seperti transaksi di modul 14 — Beginx() membuka transaksi, dan
// defer tx.Rollback() adalah JARING PENGAMAN: kalau fungsi keluar lewat jalur error mana
// pun, perubahan dibatalkan. Rollback SETELAH Commit tidak berbahaya (jadi no-op), jadi
// pola "defer Rollback lalu Commit di akhir" itu aman dan idiomatik.
func (r *ProdukRepo) KurangiStok(id int64, jumlah int) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return fmt.Errorf("gagal membuka transaksi: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck // no-op bila sudah Commit

	var stok int
	if err := tx.Get(&stok, "SELECT stok FROM produk WHERE id = ?", id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("id %d: %w", id, ErrTidakDitemukan)
		}
		return fmt.Errorf("gagal membaca stok: %w", err)
	}
	if stok < jumlah {
		return fmt.Errorf("stok produk %d hanya %d, diminta %d", id, stok, jumlah)
	}

	if _, err := tx.Exec("UPDATE produk SET stok = stok - ? WHERE id = ?", jumlah, id); err != nil {
		return fmt.Errorf("gagal mengurangi stok: %w", err)
	}
	return tx.Commit()
}

func demoTransaksi(r *ProdukRepo) {
	fmt.Println("\n-- Transaksi --")

	if err := r.KurangiStok(1, 3); err != nil {
		fmt.Println("   error:", err)
	} else {
		p, _ := r.Ambil(1)
		fmt.Printf("   kurangi stok id 1 sebanyak 3 -> sisa %d\n", p.Stok)
	}

	if err := r.KurangiStok(1, 9999); err != nil {
		fmt.Println("   minta lebih dari stok -> ditolak, transaksi dibatalkan")
	}
}
