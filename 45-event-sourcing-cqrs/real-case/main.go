// REAL-CASE Modul 45 (Event Sourcing & CQRS) dengan DATABASE SUNGGUHAN.
//
// Beda dari folder advanced/ (yang pakai in-memory agar ringkas), versi ini
// menyimpan event ke SQLite pada FILE nyata di disk — persis pola produksi
// (tinggal ganti driver ke Postgres). Yang ditunjukkan:
//
//   - Event store append-only sebagai TABEL (bukan slice).
//   - Optimistic concurrency lewat constraint UNIQUE(aggregate_id, version).
//   - Write & update read-model dalam SATU transaksi (konsisten).
//   - Bukti DURABILITAS: tutup DB, buka lagi dari file, state tetap ada.
//   - Rebuild projection dari nol (read-model itu DERIVATIF, bisa dibuang & dibangun ulang).
//
// Jalankan:
//
//	go run ./45-event-sourcing-cqrs/real-case
package main

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

// ===========================================================================
// SKEMA
// ===========================================================================
// events        : log append-only. UNIQUE(aggregate_id, version) = jantung
//
//	optimistic concurrency — dua penulis tak bisa menaruh event
//	dengan versi sama untuk aggregate yang sama.
//
// account_read  : READ MODEL (CQRS). Diturunkan dari events; dioptimalkan untuk
//
//	query (baca saldo tanpa replay).
const schema = `
CREATE TABLE IF NOT EXISTS events (
	seq          INTEGER PRIMARY KEY AUTOINCREMENT,
	aggregate_id TEXT    NOT NULL,
	version      INTEGER NOT NULL,
	type         TEXT    NOT NULL,
	amount       INTEGER NOT NULL,
	created_at   TEXT    NOT NULL,
	UNIQUE(aggregate_id, version)
);
CREATE TABLE IF NOT EXISTS account_read (
	aggregate_id TEXT PRIMARY KEY,
	balance      INTEGER NOT NULL,
	version      INTEGER NOT NULL
);`

var ErrKonflik = errors.New("konflik versi (optimistic concurrency)")
var ErrSaldoKurang = errors.New("saldo tidak cukup")

// EventStore = repository di atas *sql.DB. Bisa dibungkus *sql.Tx juga.
type EventStore struct{ db *sql.DB }

// versiTerkini membaca versi terakhir aggregate (0 bila belum ada event).
func (s *EventStore) versiTerkini(aggID string) (int, error) {
	var v sql.NullInt64
	err := s.db.QueryRow(`SELECT MAX(version) FROM events WHERE aggregate_id=?`, aggID).Scan(&v)
	if err != nil {
		return 0, err
	}
	if v.Valid {
		return int(v.Int64), nil
	}
	return 0, nil
}

// Append menyimpan SATU event + memperbarui read-model secara ATOMIK.
// expectedVersion = versi yang penulis KIRA sedang berlaku. Bila ada penulis
// lain menyelip (versi tak lagi cocok), INSERT gagal oleh UNIQUE -> ErrKonflik.
func (s *EventStore) Append(aggID string, expectedVersion int, tipe string, amount int) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() // no-op setelah Commit; jaring pengaman bila error/panic

	newVersion := expectedVersion + 1
	_, err = tx.Exec(
		`INSERT INTO events(aggregate_id, version, type, amount, created_at) VALUES(?,?,?,?,?)`,
		aggID, newVersion, tipe, amount, time.Now().UTC().Format(time.RFC3339))
	if err != nil {
		// Pelanggaran UNIQUE(aggregate_id, version) = ada yang menulis versi ini duluan.
		return fmt.Errorf("%w: %v", ErrKonflik, err)
	}

	// Perbarui read-model dalam transaksi yang sama -> event & proyeksi konsisten.
	delta := amount
	if tipe == "Withdrawn" {
		delta = -amount
	}
	_, err = tx.Exec(`
		INSERT INTO account_read(aggregate_id, balance, version) VALUES(?,?,?)
		ON CONFLICT(aggregate_id) DO UPDATE SET balance = balance + ?, version = ?`,
		aggID, delta, newVersion, delta, newVersion)
	if err != nil {
		return err
	}
	return tx.Commit()
}

// Aggregate = state yang DITURUNKAN dari replay event (sumber kebenaran).
type Akun struct {
	Saldo, Versi int
}

// Load merekonstruksi aggregate dengan MEMUTAR ULANG event dari DB.
func (s *EventStore) Load(aggID string) (Akun, error) {
	rows, err := s.db.Query(`SELECT type, amount FROM events WHERE aggregate_id=? ORDER BY version`, aggID)
	if err != nil {
		return Akun{}, err
	}
	defer rows.Close()
	var a Akun
	for rows.Next() {
		var tipe string
		var amount int
		if err := rows.Scan(&tipe, &amount); err != nil {
			return Akun{}, err
		}
		switch tipe { // fungsi "apply": satu event -> perubahan state
		case "Deposited":
			a.Saldo += amount
		case "Withdrawn":
			a.Saldo -= amount
		}
		a.Versi++
	}
	return a, rows.Err()
}

// ---- Command layer: menegakkan ATURAN BISNIS sebelum menghasilkan event ----

func (s *EventStore) Deposit(aggID string, jumlah int) error {
	v, err := s.versiTerkini(aggID)
	if err != nil {
		return err
	}
	return s.Append(aggID, v, "Deposited", jumlah)
}

func (s *EventStore) Withdraw(aggID string, jumlah int) error {
	akun, err := s.Load(aggID) // muat state untuk cek aturan
	if err != nil {
		return err
	}
	if akun.Saldo < jumlah {
		return ErrSaldoKurang // aturan bisnis: tak boleh saldo negatif
	}
	return s.Append(aggID, akun.Versi, "Withdrawn", jumlah)
}

// SaldoDariReadModel membaca saldo langsung dari proyeksi (CEPAT, tanpa replay).
func (s *EventStore) SaldoDariReadModel(aggID string) (int, error) {
	var saldo int
	err := s.db.QueryRow(`SELECT balance FROM account_read WHERE aggregate_id=?`, aggID).Scan(&saldo)
	return saldo, err
}

// RebuildReadModel membuang read-model lalu membangunnya ULANG dari events.
// Membuktikan read-model itu DERIVATIF — aman dihapus; kebenaran ada di events.
func (s *EventStore) RebuildReadModel() error {
	if _, err := s.db.Exec(`DELETE FROM account_read`); err != nil {
		return err
	}
	rows, err := s.db.Query(`SELECT aggregate_id, type, amount, version FROM events ORDER BY seq`)
	if err != nil {
		return err
	}
	defer rows.Close()
	saldo := map[string]int{}
	versi := map[string]int{}
	for rows.Next() {
		var id, tipe string
		var amount, v int
		if err := rows.Scan(&id, &tipe, &amount, &v); err != nil {
			return err
		}
		if tipe == "Deposited" {
			saldo[id] += amount
		} else {
			saldo[id] -= amount
		}
		versi[id] = v
	}
	for id, s2 := range saldo {
		if _, err := s.db.Exec(
			`INSERT INTO account_read(aggregate_id, balance, version) VALUES(?,?,?)`,
			id, s2, versi[id]); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	// File DB nyata di direktori temp (fresh tiap run agar output deterministik).
	// Di produksi: path persisten atau DSN Postgres.
	dbPath := filepath.Join(os.TempDir(), "es_realcase.db")
	_ = os.Remove(dbPath)
	defer os.Remove(dbPath)
	fmt.Println("DB file:", dbPath)

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		panic(err)
	}
	if _, err := db.Exec(schema); err != nil {
		panic(err)
	}
	store := &EventStore{db: db}

	const akun = "acc-100"

	// =====================================================================
	// 1. Command menghasilkan event yang di-PERSIST ke tabel
	// =====================================================================
	fmt.Println("== 1. jalankan command (menulis event ke DB) ==")
	must(store.Deposit(akun, 100))
	must(store.Withdraw(akun, 30))
	must(store.Deposit(akun, 50))
	if err := store.Withdraw(akun, 999); errors.Is(err, ErrSaldoKurang) {
		fmt.Println("  withdraw 999 ditolak aturan bisnis:", err)
	}

	var jumlahEvent int
	_ = db.QueryRow(`SELECT COUNT(*) FROM events WHERE aggregate_id=?`, akun).Scan(&jumlahEvent)
	fmt.Printf("  %d event tersimpan di tabel events\n", jumlahEvent)

	// =====================================================================
	// 2. Optimistic concurrency: dua penulis pakai versi yang sama
	// =====================================================================
	fmt.Println("== 2. optimistic concurrency ==")
	v, _ := store.versiTerkini(akun)               // mis. 3
	must(store.Append(akun, v, "Deposited", 10))   // penulis A: v -> 4 (sukses)
	errB := store.Append(akun, v, "Deposited", 20) // penulis B: masih pakai v -> tabrakan UNIQUE
	fmt.Println("  penulis kedua (versi basi) ditolak:", errors.Is(errB, ErrKonflik))

	// =====================================================================
	// 3. DURABILITAS: tutup koneksi, buka lagi dari FILE -> state tetap ada
	// =====================================================================
	fmt.Println("== 3. durabilitas (tutup & buka ulang dari disk) ==")
	db.Close()
	db2, err := sql.Open("sqlite", dbPath) // koneksi BARU ke file yang sama
	if err != nil {
		panic(err)
	}
	defer db2.Close()
	store2 := &EventStore{db: db2}

	akunReplay, _ := store2.Load(akun) // rekonstruksi dari event di disk
	saldoRead, _ := store2.SaldoDariReadModel(akun)
	fmt.Printf("  replay dari disk : saldo=%d versi=%d\n", akunReplay.Saldo, akunReplay.Versi)
	fmt.Printf("  read-model (cepat): saldo=%d\n", saldoRead)
	fmt.Println("  keduanya cocok?", akunReplay.Saldo == saldoRead)

	// =====================================================================
	// 4. Rebuild read-model dari nol (proyeksi itu bisa dibuang & dibangun ulang)
	// =====================================================================
	fmt.Println("== 4. rebuild projection ==")
	must(store2.RebuildReadModel())
	saldoBaru, _ := store2.SaldoDariReadModel(akun)
	fmt.Printf("  read-model dibangun ulang dari events -> saldo=%d\n", saldoBaru)
	fmt.Println("  (event = sumber kebenaran; read-model = turunan yang bisa diregenerasi)")
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
