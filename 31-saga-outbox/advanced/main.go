// Teknik Advanced Modul 31 (saga & outbox) — contoh runnable + penjelasan detail.
//
// (1) Transactional Outbox di SQLite: perubahan bisnis + event dipublikasikan
// atomik. (2) Saga dengan compensating transaction. Jalankan:
//
//	go run ./31-saga-outbox/advanced
package main

import (
	"database/sql"
	"errors"
	"fmt"

	_ "modernc.org/sqlite"
)

func main() {
	outboxDemo()
	fmt.Println()
	sagaDemo()
}

// ===========================================================================
// 1. TRANSACTIONAL OUTBOX
// ===========================================================================
// Masalah: menyimpan order ke DB lalu publish event ke broker adalah DUA sistem
// -> bisa tak konsisten (DB sukses, publish gagal = event hilang).
// Solusi: tulis event ke tabel "outbox" DALAM TRANSAKSI YANG SAMA dengan order.
// Karena atomik, event pasti ada bila order ada. Relay membaca outbox lalu
// mem-publish ke broker & menandainya terkirim (idempoten via id).
//
// 🔍 Analogi: outbox itu seperti menulis surat dan MENARUHNYA DI KOTAK SURAT
// RUMAH dalam satu gerakan dengan mencatat di buku harian — dua-duanya pasti
// terjadi bersama. Tukang pos (relay) lewat berkala mengambil isi kotak dan
// mengirimkannya; kalau tukang pos telat, surat tetap aman menunggu, tak
// pernah hilang.
func outboxDemo() {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		panic(err)
	}
	defer db.Close()
	exec(db, `CREATE TABLE orders(id TEXT PRIMARY KEY, total INT)`)
	exec(db, `CREATE TABLE outbox(id INTEGER PRIMARY KEY, topik TEXT, payload TEXT, terkirim INT DEFAULT 0)`)

	// --- Satu transaksi: order + event outbox sekaligus (atomik) ---
	tx, _ := db.Begin()
	_, err = tx.Exec(`INSERT INTO orders(id,total) VALUES(?,?)`, "ord-1", 250)
	if err == nil {
		_, err = tx.Exec(`INSERT INTO outbox(topik,payload) VALUES(?,?)`, "order.created", `{"id":"ord-1"}`)
	}
	if err != nil {
		tx.Rollback() // bila salah satu gagal, KEDUANYA batal
		panic(err)
	}
	tx.Commit()
	fmt.Println("== 1. transactional outbox ==")
	fmt.Println("  order + event tersimpan atomik dalam 1 transaksi")

	// --- Relay/poller: publish event yang belum terkirim, lalu tandai ---
	rows, _ := db.Query(`SELECT id, topik, payload FROM outbox WHERE terkirim=0`)
	var ids []int
	for rows.Next() {
		var id int
		var topik, payload string
		_ = rows.Scan(&id, &topik, &payload)
		fmt.Printf("  [relay] publish ke broker: %s %s\n", topik, payload)
		ids = append(ids, id)
	}
	rows.Close()
	for _, id := range ids {
		exec(db, `UPDATE outbox SET terkirim=1 WHERE id=?`, id) // idempoten: tak dikirim 2x
	}
	fmt.Println("  event ditandai terkirim (tak akan dipublish ulang)")
}

// ===========================================================================
// 2. SAGA dengan COMPENSATING TRANSACTION
// ===========================================================================
// Transaksi terdistribusi lintas service tak bisa 2PC. Saga = rangkaian langkah
// lokal; bila satu GAGAL, jalankan "undo" (kompensasi) untuk langkah yang sudah
// sukses, dalam urutan TERBALIK. Mencapai konsistensi eventual tanpa lock global.
//
// 🔍 Analogi: saga itu seperti MERENCANAKAN LIBURAN — pesan hotel, pesan
// pesawat, sewa mobil. Kalau sewa mobil gagal, kamu tak bisa "membatalkan
// semuanya sekaligus dengan satu tombol" (tak ada 2PC lintas perusahaan);
// yang bisa: telepon balik MUNDUR satu-satu — batalkan pesawat, lalu batalkan
// hotel (compensating transaction).
type langkah struct {
	nama string
	maju func() error
	undo func()
}

func sagaDemo() {
	fmt.Println("== 2. saga + compensation ==")
	langkahList := []langkah{
		{"reserve-stok", func() error { fmt.Println("  reserve stok OK"); return nil },
			func() { fmt.Println("  [undo] lepas reservasi stok") }},
		{"charge-bayar", func() error { fmt.Println("  charge pembayaran OK"); return nil },
			func() { fmt.Println("  [undo] refund pembayaran") }},
		{"kirim-notif", func() error { return errors.New("service notifikasi mati") },
			func() { fmt.Println("  [undo] batalkan notifikasi") }},
	}

	var selesai []langkah
	var gagal error
	for _, l := range langkahList {
		if err := l.maju(); err != nil {
			gagal = fmt.Errorf("langkah %q gagal: %w", l.nama, err)
			break
		}
		selesai = append(selesai, l)
	}

	if gagal != nil {
		fmt.Println("  ! ", gagal)
		fmt.Println("  menjalankan kompensasi (urutan terbalik):")
		for i := len(selesai) - 1; i >= 0; i-- { // undo mundur
			selesai[i].undo()
		}
	}
}

func exec(db *sql.DB, q string, args ...any) {
	if _, err := db.Exec(q, args...); err != nil {
		panic(err)
	}
}
