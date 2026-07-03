package main

import (
	"errors"
	"testing"
)

func setup() (*BankService, *BalanceProjection) {
	store := NewEventStore()
	proj := NewBalanceProjection()
	store.Subscribe(proj.On)
	return NewBankService(store), proj
}

func TestRebuildStateDariEvent(t *testing.T) {
	svc, _ := setup()
	_ = svc.OpenAccount("a1", "Ana")
	_ = svc.Deposit("a1", 100)
	_ = svc.Withdraw("a1", 30)
	_ = svc.Deposit("a1", 50)

	acc := svc.GetAccount("a1") // direkonstruksi dari event
	if acc.Balance != 120 {
		t.Errorf("saldo = %d; want 120", acc.Balance)
	}
	if acc.Owner != "Ana" {
		t.Errorf("owner = %q; want Ana", acc.Owner)
	}
	if acc.Version != 4 { // 4 event
		t.Errorf("version = %d; want 4", acc.Version)
	}
}

func TestSaldoTidakCukupDitolak(t *testing.T) {
	svc, _ := setup()
	_ = svc.OpenAccount("a1", "Ana")
	_ = svc.Deposit("a1", 50)

	err := svc.Withdraw("a1", 100)
	if !errors.Is(err, ErrInsufficient) {
		t.Errorf("err = %v; want ErrInsufficient", err)
	}
	// State tak berubah (tak ada event ditulis).
	if svc.GetAccount("a1").Balance != 50 {
		t.Errorf("saldo berubah padahal withdraw ditolak")
	}
}

func TestReadModelSinkronDenganWrite(t *testing.T) {
	svc, proj := setup()
	_ = svc.OpenAccount("a1", "Ana")
	_ = svc.Deposit("a1", 200)
	_ = svc.Withdraw("a1", 75)

	// Read model (proyeksi) harus cocok dengan state hasil replay event.
	if proj.Balance("a1") != 125 {
		t.Errorf("read model = %d; want 125", proj.Balance("a1"))
	}
	if svc.GetAccount("a1").Balance != proj.Balance("a1") {
		t.Error("read model & write model tidak konsisten")
	}
}

func TestTidakBisaOperasiSebelumDibuka(t *testing.T) {
	svc, _ := setup()
	if err := svc.Deposit("baru", 100); !errors.Is(err, ErrNotOpen) {
		t.Errorf("deposit sebelum buka = %v; want ErrNotOpen", err)
	}
}
