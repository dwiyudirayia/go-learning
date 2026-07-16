// Teknik Advanced Modul 21 (migrations) — contoh runnable + penjelasan detail.
//
// Fokus pada KONSEP lanjutan migrasi (idempotensi, versioning, strategi
// expand/contract untuk zero-downtime) memakai database/sql murni di SQLite
// in-memory, agar robust & self-contained. Di modul utama, golang-migrate
// mengotomasikan mekanisme ini. Jalankan:
//
//	go run ./21-migrations/advanced
package main

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

// migrasi = satu perubahan skema berversi, dengan Up (maju) & Down (rollback).
type migrasi struct {
	versi int
	nama  string
	up    string
}

var migrasiList = []migrasi{
	{1, "buat tabel users", `CREATE TABLE users(id INTEGER PRIMARY KEY, nama TEXT NOT NULL)`},
	// EXPAND: tambah kolom BARU yang NULLABLE. Aman untuk kode versi lama yang
	// belum tahu kolom ini (mereka mengabaikannya) -> zero-downtime.
	{2, "expand: tambah kolom email (nullable)", `ALTER TABLE users ADD COLUMN email TEXT`},
	{3, "buat index email", `CREATE INDEX idx_users_email ON users(email)`},
}

func main() {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// =====================================================================
	// 1. Tabel pelacak versi — inti IDEMPOTENSI
	// =====================================================================
	// Menyimpan versi terakhir yang sudah diterapkan. Menjalankan migrasi dua
	// kali TIDAK mengulang yang sudah ada (persis perilaku golang-migrate).
	//
	// 🔍 Analogi: schema_version itu seperti BUKU VAKSINASI — tercatat vaksin
	// apa saja yang SUDAH disuntikkan (versi terapan), jadi datang ke klinik
	// dua kali tak membuatmu disuntik dobel; petugas cukup melirik bukunya.
	mustExec(db, `CREATE TABLE IF NOT EXISTS schema_version(versi INTEGER NOT NULL)`)

	fmt.Println("== migrasi pertama ==")
	terapkan(db)
	fmt.Println("== migrasi kedua (idempoten — tak ada yang dijalankan lagi) ==")
	terapkan(db)

	// =====================================================================
	// 2. EXPAND/CONTRACT — backfill data setelah kolom baru ditambah
	// =====================================================================
	// Setelah expand (kolom email nullable), isi data lama (backfill) sebelum
	// kelak membuatnya NOT NULL / menghapus kolom lama (fase "contract").
	//
	// 🔍 Analogi: expand/contract itu seperti MENGGANTI JEMBATAN TANPA MENUTUP
	// JALAN — bangun jembatan baru di sampingnya (expand: kolom baru nullable),
	// pindahkan lalu lintas bertahap (backfill), baru bongkar jembatan lama
	// (contract). Merobohkan langsung = seluruh jalan mati (downtime).
	mustExec(db, `INSERT INTO users(nama) VALUES('Budi'),('Ani')`)
	res, _ := db.Exec(`UPDATE users SET email = lower(nama) || '@contoh.com' WHERE email IS NULL`)
	n, _ := res.RowsAffected()
	fmt.Printf("== backfill: %d baris diisi email ==\n", n)

	rows, _ := db.Query(`SELECT nama, email FROM users ORDER BY id`)
	defer rows.Close()
	for rows.Next() {
		var nama, email string
		_ = rows.Scan(&nama, &email)
		fmt.Printf("  %s -> %s\n", nama, email)
	}

	// Catatan dirty-state: bila satu migrasi gagal di tengah, golang-migrate
	// menandai versi "dirty"; pulihkan dengan `migrate force <versi>` setelah
	// memperbaiki manual, baru lanjut. Nomor migrasi harus monoton & TAK diedit
	// setelah rilis.
}

func versiSekarang(db *sql.DB) int {
	var v sql.NullInt64
	_ = db.QueryRow(`SELECT MAX(versi) FROM schema_version`).Scan(&v)
	if v.Valid {
		return int(v.Int64)
	}
	return 0
}

func terapkan(db *sql.DB) {
	sekarang := versiSekarang(db)
	dijalankan := 0
	for _, m := range migrasiList {
		if m.versi <= sekarang {
			continue // sudah diterapkan -> lewati (idempoten)
		}
		mustExec(db, m.up)
		mustExec(db, `INSERT INTO schema_version(versi) VALUES(?)`, m.versi)
		fmt.Printf("  v%d applied: %s\n", m.versi, m.nama)
		dijalankan++
	}
	if dijalankan == 0 {
		fmt.Println("  (skema sudah paling mutakhir)")
	}
}

func mustExec(db *sql.DB, q string, args ...any) {
	if _, err := db.Exec(q, args...); err != nil {
		panic(fmt.Sprintf("exec %q: %v", q, err))
	}
}
