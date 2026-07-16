// Teknik Advanced Modul 14 (database) — contoh runnable + penjelasan detail.
//
// Memakai SQLite in-memory (pure-Go, tanpa cgo). Jalankan:
//
//	go run ./14-database/advanced
package main

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"strings"

	_ "modernc.org/sqlite" // daftarkan driver "sqlite" (blank import)
)

// ===========================================================================
// 1. Custom type: Tags disimpan sebagai CSV via driver.Valuer + sql.Scanner
// ===========================================================================
// Valuer  : cara tipe kita DITULIS ke DB (Go []string -> string "a,b,c").
// Scanner : cara tipe kita DIBACA dari DB (string -> []string).
// Pola ini membungkus konversi kolom↔tipe domain di satu tempat.
//
// 🔍 Analogi: Valuer/Scanner itu seperti PETUGAS GUDANG KHUSUS — barang
// (Tags) selalu dilipat dengan cara yang sama saat masuk rak (Value) dan
// dibuka kembali dengan cara yang sama saat diambil (Scan). Tak ada lagi
// tiap orang melipat semaunya di sana-sini.

type Tags []string

func (t Tags) Value() (driver.Value, error) { // dipanggil saat menyimpan
	return strings.Join(t, ","), nil
}

func (t *Tags) Scan(src any) error { // dipanggil saat membaca
	if src == nil {
		*t = nil
		return nil
	}
	s, ok := src.(string)
	if !ok {
		return fmt.Errorf("Tags.Scan: tipe tak terduga %T", src)
	}
	if s == "" {
		*t = Tags{}
	} else {
		*t = strings.Split(s, ",")
	}
	return nil
}

func main() {
	ctx := context.Background()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// =====================================================================
	// 2. Connection pool tuning
	// =====================================================================
	// Batas ini mencegah kehabisan koneksi & koneksi basi di produksi.
	// (Untuk :memory: nilainya simbolis, tapi API-nya sama seperti Postgres.)
	//
	// 🔍 Analogi: connection pool itu seperti ARMADA TAKSI PANGKALAN —
	// MaxOpenConns = jumlah taksi maksimal yang boleh beroperasi,
	// MaxIdleConns = berapa yang standby di pangkalan (tak dipulangkan),
	// ConnMaxLifetime = usia pensiun taksi agar tak ada mobil tua mogok
	// (koneksi basi) di tengah jalan.
	db.SetMaxOpenConns(10)   // maksimum koneksi terbuka bersamaan
	db.SetMaxIdleConns(5)    // koneksi menganggur yang dipertahankan
	db.SetConnMaxLifetime(0) // 0 = tak dibatasi umur
	_ = db.PingContext(ctx)  // verifikasi koneksi hidup

	mustExec(ctx, db, `CREATE TABLE artikel(
		id INTEGER PRIMARY KEY,
		judul TEXT NOT NULL,
		penulis TEXT,        -- boleh NULL
		tags TEXT NOT NULL
	)`)

	// =====================================================================
	// 3. Transaksi aman: pola defer Rollback (no-op setelah Commit)
	// =====================================================================
	// Rollback yang dipanggil setelah Commit tidak berefek, jadi defer aman:
	// bila terjadi error/panic di tengah, perubahan otomatis dibatalkan.
	//
	// 🔍 Analogi: transaksi itu seperti MENULIS DI DRAF — semua perubahan
	// masuk draf dulu; Commit = menekan "publish", Rollback = membuang draf.
	// defer Rollback = kebiasaan "kalau saya lupa/gagal publish, buang
	// otomatis drafnya" — dan membuang draf yang sudah terpublish memang
	// tak berefek apa-apa.
	if err := simpanArtikel(ctx, db, "Belajar Go", "Budi", Tags{"go", "backend"}); err != nil {
		panic(err)
	}
	// Penulis NULL -> pakai sql.NullString saat menyimpan.
	if err := simpanArtikelAnonim(ctx, db, "Tips SQL", Tags{"sql"}); err != nil {
		panic(err)
	}

	// =====================================================================
	// 4. Membaca: sql.NullString untuk kolom NULL + custom Scanner (Tags)
	// =====================================================================
	// SELALU pakai QueryContext agar query ikut batal saat context dibatalkan.
	//
	// 🔍 Analogi: sql.NullString itu seperti KOLOM FORMULIR OPSIONAL dengan
	// centang "diisi/tidak" (.Valid) — beda antara "sengaja kosong" (NULL)
	// dan "berisi string kosong". Membaca NULL langsung ke string biasa =
	// memaksa membaca kolom yang memang tak pernah diisi.
	rows, err := db.QueryContext(ctx, `SELECT judul, penulis, tags FROM artikel ORDER BY id`)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	fmt.Println("== hasil ==")
	for rows.Next() {
		var judul string
		var penulis sql.NullString // menangani NULL tanpa panik
		var tags Tags              // Scan() custom kita dipanggil otomatis
		if err := rows.Scan(&judul, &penulis, &tags); err != nil {
			panic(err)
		}
		nama := "(anonim)"
		if penulis.Valid { // .Valid=false berarti NULL
			nama = penulis.String
		}
		fmt.Printf("  %-12s oleh %-8s tags=%v\n", judul, nama, tags)
	}
	if err := rows.Err(); err != nil { // WAJIB cek error setelah loop
		panic(err)
	}
}

func simpanArtikel(ctx context.Context, db *sql.DB, judul, penulis string, tags Tags) (err error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback() // batalkan bila ada error
		}
	}()
	_, err = tx.ExecContext(ctx,
		`INSERT INTO artikel(judul, penulis, tags) VALUES(?,?,?)`,
		judul, penulis, tags) // tags memakai Value() otomatis
	if err != nil {
		return err
	}
	return tx.Commit()
}

func simpanArtikelAnonim(ctx context.Context, db *sql.DB, judul string, tags Tags) error {
	_, err := db.ExecContext(ctx,
		`INSERT INTO artikel(judul, penulis, tags) VALUES(?,?,?)`,
		judul, sql.NullString{}, tags) // NullString kosong -> NULL
	return err
}

func mustExec(ctx context.Context, db *sql.DB, q string, args ...any) {
	if _, err := db.ExecContext(ctx, q, args...); err != nil {
		panic(errors.New("exec gagal: " + err.Error()))
	}
}
