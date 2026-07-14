// Modul 14 — Database: database/sql (mentah) + GORM.
// Memakai SQLite pure-Go (tanpa cgo). Jalankan: go run ./14-database
package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/glebarez/sqlite" // driver GORM pure-Go; sekaligus mendaftar
	"gorm.io/gorm"               // driver database/sql bernama "sqlite" (via go-sqlite)
	"gorm.io/gorm/logger"
)

func main() {
	fmt.Println("=== 14 — Database ===")
	rawSQLDemo()
	gormDemo()
	latihanDemo()
}

// tempDB membuat path file SQLite sementara + fungsi cleanup.
func tempDB(name string) (string, func()) {
	path := filepath.Join(os.TempDir(), name)
	_ = os.Remove(path)
	return path, func() { _ = os.Remove(path) }
}

// ------------------------------------------------------------------
// Bagian A: database/sql MENTAH (tanpa ORM)
// ------------------------------------------------------------------
// 🔍 Analogi besar: ada 2 cara bicara dengan database.
//   - database/sql (mentah) = kamu menulis SQL sendiri, seperti MASAK DARI NOL. Kontrol penuh,
//     tapi lebih banyak kerja tangan (Scan tiap kolom, kelola rows, transaksi manual).
//   - GORM (ORM) = "penerjemah" yang mengubah struct Go <-> tabel otomatis, seperti KATERING.
//     Cepat & praktis, tapi ada "sihir" yang perlu dipahami. Modul ini menunjukkan keduanya.
func rawSQLDemo() {
	fmt.Println("\n-- database/sql (mentah) --")

	path, cleanup := tempDB("m14-raw.db")
	defer cleanup()

	db, err := sql.Open("sqlite", path) // "sqlite" = driver modernc
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// DDL
	_, err = db.Exec(`CREATE TABLE users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		email TEXT UNIQUE NOT NULL
	)`)
	if err != nil {
		log.Fatal(err)
	}

	// 🔍 Analogi: placeholder (?) itu KOTAK TERKUNCI untuk data pengguna. Alih-alih menempel
	// teks langsung ke SQL (bahaya! pengguna nakal bisa menyelipkan perintah = "SQL injection"),
	// kita kirim data terpisah lewat (?) — driver memperlakukannya murni sebagai DATA, bukan perintah.
	// Ibarat memberi tamu formulir isian, bukan membiarkannya menulis ulang kontrak.
	// INSERT dengan placeholder (?) -> cegah SQL injection. LastInsertId ambil ID.
	res, _ := db.Exec("INSERT INTO users(name,email) VALUES(?,?)", "Ana", "ana@mail.id")
	id, _ := res.LastInsertId()
	fmt.Printf("insert Ana -> id=%d\n", id)
	_, _ = db.Exec("INSERT INTO users(name,email) VALUES(?,?)", "Budi", "budi@mail.id")

	// QueryRow: ambil satu baris. sql.ErrNoRows bila tak ada.
	var name string
	err = db.QueryRow("SELECT name FROM users WHERE id=?", 1).Scan(&name)
	if errors.Is(err, sql.ErrNoRows) {
		fmt.Println("tidak ada")
	} else {
		fmt.Printf("QueryRow id=1 -> %s\n", name)
	}

	// Query: banyak baris. WAJIB rows.Close() dan cek rows.Err().
	rows, _ := db.Query("SELECT id,name,email FROM users ORDER BY id")
	defer rows.Close()
	fmt.Println("semua user:")
	for rows.Next() {
		var (
			uid          int
			uname, umail string
		)
		_ = rows.Scan(&uid, &uname, &umail)
		fmt.Printf("  #%d %s <%s>\n", uid, uname, umail)
	}

	// 🔍 Analogi: transaksi itu "SEMUA-ATAU-TIDAK SAMA SEKALI", seperti transfer bank: kurangi
	// saldo A DAN tambah saldo B harus sukses bersama. Begin = buka sesi; Commit = "sahkan semua";
	// Rollback = "batalkan semua seolah tak pernah terjadi" bila ada langkah gagal di tengah.
	// Transaksi: Begin -> ... -> Commit/Rollback.
	tx, _ := db.Begin()
	_, err = tx.Exec("UPDATE users SET name=? WHERE id=?", "Ana Updated", 1)
	if err != nil {
		_ = tx.Rollback()
	} else {
		_ = tx.Commit()
	}
	_ = db.QueryRow("SELECT name FROM users WHERE id=1").Scan(&name)
	fmt.Printf("setelah transaksi update -> %s\n", name)
}

// ------------------------------------------------------------------
// Bagian B: GORM (ORM)
// ------------------------------------------------------------------

// Model GORM: gorm.Model menyertakan ID, CreatedAt, UpdatedAt, DeletedAt (soft delete).
type User struct {
	gorm.Model
	Name  string
	Email string `gorm:"uniqueIndex"`
	Posts []Post // relasi has-many
}

type Post struct {
	gorm.Model
	Title      string
	UserID     uint // foreign key ke User
	CategoryID uint // latihan 1: belongs-to Category
}

// Category (latihan 1): satu kategori punya banyak Post.
type Category struct {
	gorm.Model
	Name string
}

func gormDemo() {
	fmt.Println("\n-- GORM (ORM) --")

	path, cleanup := tempDB("m14-gorm.db")
	defer cleanup()

	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // matikan log query agar output bersih
	})
	if err != nil {
		log.Fatal(err)
	}

	// AutoMigrate: buat/ubah tabel sesuai struct.
	if err := db.AutoMigrate(&User{}, &Post{}); err != nil {
		log.Fatal(err)
	}

	// Create (termasuk relasi has-many sekaligus).
	user := User{
		Name:  "Ciko",
		Email: "ciko@mail.id",
		Posts: []Post{{Title: "Belajar GORM"}, {Title: "Belajar Fiber"}},
	}
	db.Create(&user)
	fmt.Printf("create -> user id=%d dengan %d post\n", user.ID, len(user.Posts))

	// Read: First (satu), Find (banyak), Where (filter).
	var found User
	db.First(&found, user.ID)
	fmt.Printf("First -> %s <%s>\n", found.Name, found.Email)

	// Preload: muat relasi.
	var withPosts User
	db.Preload("Posts").First(&withPosts, user.ID)
	fmt.Printf("Preload Posts -> %d post: ", len(withPosts.Posts))
	for _, p := range withPosts.Posts {
		fmt.Printf("%q ", p.Title)
	}
	fmt.Println()

	// Update
	db.Model(&found).Update("Name", "Ciko Updated")
	fmt.Printf("update -> %s\n", found.Name)

	// Count
	var total int64
	db.Model(&User{}).Count(&total)
	fmt.Printf("jumlah user = %d\n", total)

	// 🔍 Analogi: soft delete itu "PINDAH KE TEMPAT SAMPAH", bukan hapus permanen. Barisnya masih
	// ada di disk tapi ditandai DeletedAt, jadi query biasa pura-pura tak melihatnya. Unscoped =
	// "buka tempat sampah, tampilkan semua". Berguna untuk audit / undo. GORM melakukannya otomatis.
	// Delete (soft delete: baris tak hilang, DeletedAt diisi).
	db.Delete(&found)
	var afterDelete int64
	db.Model(&User{}).Count(&afterDelete) // tak terhitung karena soft delete
	fmt.Printf("setelah soft delete, Count user = %d (baris masih ada tapi tersembunyi)\n", afterDelete)

	// Unscoped: lihat termasuk yang soft-deleted.
	var raw int64
	db.Unscoped().Model(&User{}).Count(&raw)
	fmt.Printf("Unscoped Count = %d (baris fisik masih ada)\n", raw)
}

// ------------------------------------------------------------------
// Latihan Modul 14: Category (belongs-to), raw JOIN, hard delete
// ------------------------------------------------------------------
func latihanDemo() {
	fmt.Println("\n-- Latihan: Category, JOIN, hard delete --")

	path, cleanup := tempDB("m14-latihan.db")
	defer cleanup()

	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatal(err)
	}
	if err := db.AutoMigrate(&User{}, &Post{}, &Category{}); err != nil {
		log.Fatal(err)
	}

	// Seed: 1 kategori, 1 user dengan 2 post di kategori itu.
	cat := Category{Name: "Tutorial"}
	db.Create(&cat)
	user := User{Name: "Ana", Email: "ana@mail.id", Posts: []Post{
		{Title: "Post 1", CategoryID: cat.ID},
		{Title: "Post 2", CategoryID: cat.ID},
	}}
	db.Create(&user)

	// Latihan 3: raw database/sql-style query lewat GORM (JOIN + GROUP BY).
	type Row struct {
		Name  string
		Total int
	}
	var rows []Row
	db.Raw(`SELECT users.name AS name, COUNT(posts.id) AS total
	        FROM users LEFT JOIN posts ON posts.user_id = users.id
	        GROUP BY users.id`).Scan(&rows)
	for _, r := range rows {
		fmt.Printf("JOIN -> %s punya %d post\n", r.Name, r.Total)
	}

	// Latihan 4: HARD delete (baris benar-benar hilang) memakai Unscoped.
	db.Unscoped().Delete(&user)
	var sisa int64
	db.Unscoped().Model(&User{}).Count(&sisa)
	fmt.Printf("setelah Unscoped().Delete -> baris fisik user = %d (benar-benar terhapus)\n", sisa)
}
