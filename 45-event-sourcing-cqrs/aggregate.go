// Modul 45 — Event Sourcing & CQRS.
// Lanjutan Modul 31 (saga/outbox). Domain: rekening bank.
package main

import "errors"

// 🔍 Analogi besar EVENT SOURCING: seperti BUKU REKENING/mutasi bank vs cuma tahu saldo akhir.
// Cara biasa menyimpan "Saldo = 100" (kehilangan riwayat). Event sourcing menyimpan tiap TRANSAKSI:
// "buka rekening, setor 150, tarik 50". Saldo (state) dihitung dengan MENJUMLAH ulang seluruh mutasi.
// Untung: riwayat lengkap (audit), bisa "time-travel" ke saldo tanggal lalu, & tiap perubahan jelas
// sebabnya. Perhatikan event ditulis LAMPAU (Opened, Deposited) — fakta yang sudah terjadi, tak bisa diubah.
//
// 🔍 Analogi CQRS: memisahkan MESIN TULIS dari MESIN BACA. Command (Open/Deposit/Withdraw) = sisi
// TULIS: memvalidasi aturan lalu menghasilkan event — tak langsung mengubah state. Nanti sisi BACA
// menyusun "tampilan saldo" dari event. Seperti dapur (tulis pesanan) terpisah dari etalase (baca menu).

// EVENT SOURCING: alih-alih menyimpan STATE saat ini (Balance=100), kita simpan
// urutan PERISTIWA (dibuka, +150, -50). State direkonstruksi dengan MEMUTAR ULANG
// event. Keuntungan: audit lengkap, bisa "time-travel", debug mudah.

// Event = fakta yang telah terjadi (past tense). Interface penanda.
type Event interface{ isEvent() }

type AccountOpened struct {
	AccountID string
	Owner     string
}
type MoneyDeposited struct {
	AccountID string
	Amount    int
}
type MoneyWithdrawn struct {
	AccountID string
	Amount    int
}

func (AccountOpened) isEvent()  {}
func (MoneyDeposited) isEvent() {}
func (MoneyWithdrawn) isEvent() {}

// 🔍 Analogi aggregate: Account itu "REKENING sebagai satu kesatuan yang menjaga aturannya sendiri"
// (mis. tak boleh tarik melebihi saldo). Ia dijaga konsisten sebagai satu unit. 'Version' = nomor
// urut mutasi terakhir; dipakai untuk deteksi bentrok saat dua orang mengubah rekening bersamaan.
// Account = AGGREGATE: state-nya diturunkan dari event.
type Account struct {
	ID      string
	Owner   string
	Balance int
	Version int // jumlah event yang sudah diterapkan
	opened  bool
}

// apply MENGUBAH state sesuai satu event (dipakai saat replay & saat commit).
func (a *Account) apply(e Event) {
	switch ev := e.(type) {
	case AccountOpened:
		a.ID, a.Owner, a.opened = ev.AccountID, ev.Owner, true
	case MoneyDeposited:
		a.Balance += ev.Amount
	case MoneyWithdrawn:
		a.Balance -= ev.Amount
	}
	a.Version++
}

// NewAccountFromEvents merekonstruksi state dengan memutar ulang event.
func NewAccountFromEvents(events []Event) *Account {
	a := &Account{}
	for _, e := range events {
		a.apply(e)
	}
	return a
}

var (
	ErrNotOpen       = errors.New("rekening belum dibuka")
	ErrAlreadyOpen   = errors.New("rekening sudah dibuka")
	ErrInsufficient  = errors.New("saldo tidak cukup")
	ErrInvalidAmount = errors.New("jumlah harus > 0")
)

// --- COMMAND handlers: memvalidasi lalu MENGHASILKAN event baru (tak mengubah
//     state langsung). Ini sisi "write" dari CQRS. ---

func (a *Account) Open(owner string) ([]Event, error) {
	if a.opened {
		return nil, ErrAlreadyOpen
	}
	return []Event{AccountOpened{AccountID: a.ID, Owner: owner}}, nil
}

func (a *Account) Deposit(amount int) ([]Event, error) {
	if !a.opened {
		return nil, ErrNotOpen
	}
	if amount <= 0 {
		return nil, ErrInvalidAmount
	}
	return []Event{MoneyDeposited{AccountID: a.ID, Amount: amount}}, nil
}

func (a *Account) Withdraw(amount int) ([]Event, error) {
	if !a.opened {
		return nil, ErrNotOpen
	}
	if amount <= 0 {
		return nil, ErrInvalidAmount
	}
	if amount > a.Balance {
		return nil, ErrInsufficient
	}
	return []Event{MoneyWithdrawn{AccountID: a.ID, Amount: amount}}, nil
}
