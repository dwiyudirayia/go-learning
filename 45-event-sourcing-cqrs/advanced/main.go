// Teknik Advanced Modul 45 (event sourcing & CQRS) — contoh runnable + penjelasan detail.
//
// Event store append-only + replay + optimistic concurrency + projection (CQRS).
// Jalankan:
//
//	go run ./45-event-sourcing-cqrs/advanced
package main

import (
	"errors"
	"fmt"
)

// ===========================================================================
// EVENT = fakta yang SUDAH terjadi (immutable, past tense). State bukan disimpan
// langsung, melainkan DITURUNKAN dari urutan event.
//
// 🔍 Analogi: event sourcing itu persis BUKU TABUNGAN BANK — bank tidak
// mencoret-coret angka saldo, melainkan mencatat SETIAP setoran/penarikan
// baris demi baris (append-only). Saldo hanyalah HASIL PENJUMLAHAN semua
// baris (replay); riwayatnya tak pernah hilang dan bisa diaudit kapan pun.
// ===========================================================================
type Event struct {
	Tipe   string
	Jumlah int
}

// Akun = aggregate. Versi dipakai untuk optimistic concurrency control.
type Akun struct {
	saldo int
	versi int
}

// apply memutar SATU event ke state (dipakai saat command & saat replay).
func (a *Akun) apply(e Event) {
	switch e.Tipe {
	case "Deposited":
		a.saldo += e.Jumlah
	case "Withdrawn":
		a.saldo -= e.Jumlah
	}
	a.versi++
}

// ===========================================================================
// EVENT STORE append-only dengan OPTIMISTIC CONCURRENCY
// ===========================================================================
// Append hanya diterima bila versi yang diharapkan cocok dengan versi terkini
// -> mencegah lost update saat dua penulis bersamaan.
//
// 🔍 Analogi: optimistic concurrency itu seperti MENGEDIT DOKUMEN BERSAMA
// dengan nomor revisi — kamu menyerahkan perubahan sambil bilang "ini
// berdasarkan revisi 2". Kalau ternyata dokumen sudah di revisi 3 (orang lain
// menulis duluan), perubahanmu ditolak dan kamu diminta membaca ulang — bukan
// menimpa buta tulisan orang (lost update).
type Store struct{ events []Event }

var ErrKonflik = errors.New("konflik versi (optimistic concurrency)")

func (s *Store) Append(versiDiharapkan int, e Event) error {
	if versiDiharapkan != len(s.events) {
		return ErrKonflik
	}
	s.events = append(s.events, e) // hanya MENAMBAH, tak pernah mengubah/hapus
	return nil
}

// replay membangun ulang aggregate dari NOL dengan memutar semua event.
func (s *Store) replay() Akun {
	var a Akun
	for _, e := range s.events {
		a.apply(e)
	}
	return a
}

func main() {
	store := &Store{}

	// =====================================================================
	// 1. Command -> hasilkan event -> append (append-only)
	// =====================================================================
	fmt.Println("== 1. append events ==")
	_ = store.Append(0, Event{"Deposited", 100})
	_ = store.Append(1, Event{"Withdrawn", 30})
	_ = store.Append(2, Event{"Deposited", 50})
	fmt.Printf("  %d event tersimpan (append-only, tak diubah)\n", len(store.events))

	// =====================================================================
	// 2. Optimistic concurrency: append dgn versi basi -> ditolak
	// =====================================================================
	fmt.Println("== 2. optimistic concurrency ==")
	err := store.Append(1, Event{"Withdrawn", 10}) // versi 1 sudah lampau
	fmt.Println("  append versi basi ditolak:", errors.Is(err, ErrKonflik))

	// =====================================================================
	// 3. REPLAY: rekonstruksi state dari event
	// =====================================================================
	akun := store.replay()
	fmt.Println("== 3. replay ==")
	fmt.Printf("  saldo hasil replay=%d, versi=%d\n", akun.saldo, akun.versi)

	// =====================================================================
	// 4. CQRS PROJECTION: bangun read-model teroptimasi untuk query
	// =====================================================================
	// Sisi tulis (event) dipisah dari sisi baca (projection). Projection ini
	// meringkas riwayat -> read cepat tanpa replay tiap kali.
	//
	// 🔍 Analogi: projection itu seperti PAPAN KLASEMEN LIGA — hasil setiap
	// pertandingan (event) dicatat di arsip, tapi penonton cukup melihat
	// papan ringkasannya (read model), bukan menghitung ulang seluruh musim
	// tiap kali bertanya "siapa juara?".
	proj := struct {
		totalMasuk, totalKeluar, transaksi int
	}{}
	for _, e := range store.events {
		proj.transaksi++
		switch e.Tipe {
		case "Deposited":
			proj.totalMasuk += e.Jumlah
		case "Withdrawn":
			proj.totalKeluar += e.Jumlah
		}
	}
	fmt.Println("== 4. projection (read model) ==")
	fmt.Printf("  masuk=%d keluar=%d transaksi=%d\n", proj.totalMasuk, proj.totalKeluar, proj.transaksi)

	// Catatan: untuk aggregate berumur panjang, simpan SNAPSHOT berkala agar
	// replay tak dari nol. Projection idempoten (dedup by event-id) & eventual
	// consistency (read-model boleh tertinggal sesaat dari write).
}
