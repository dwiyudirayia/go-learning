// Jalankan: go run ./45-event-sourcing-cqrs
// Verifikasi otomatis: go test ./45-event-sourcing-cqrs
package main

import "fmt"

func main() {
	fmt.Println("=== 45 — Event Sourcing & CQRS ===")

	store := NewEventStore()
	svc := NewBankService(store)

	// Read model (CQRS): berlangganan event untuk memelihara saldo.
	proj := NewBalanceProjection()
	store.Subscribe(proj.On)

	// COMMAND (sisi write) -> menghasilkan event.
	_ = svc.OpenAccount("acc-1", "Ana")
	_ = svc.Deposit("acc-1", 150)
	_ = svc.Withdraw("acc-1", 50)
	_ = svc.Deposit("acc-1", 20)

	// Coba tarik melebihi saldo -> ditolak (tak ada event tersimpan).
	if err := svc.Withdraw("acc-1", 1000); err != nil {
		fmt.Printf("tarik 1000 ditolak: %v\n", err)
	}

	// Baca state dari EVENT (sumber kebenaran) vs dari READ MODEL (proyeksi).
	acc := svc.GetAccount("acc-1")
	fmt.Printf("dari event  -> owner=%s saldo=%d (setelah %d event)\n", acc.Owner, acc.Balance, acc.Version)
	fmt.Printf("dari read model -> saldo=%d\n", proj.Balance("acc-1"))

	// Audit trail: seluruh sejarah tersimpan.
	fmt.Println("\nriwayat event (audit):")
	for i, e := range store.Load("acc-1") {
		fmt.Printf("  %d. %T %+v\n", i+1, e, e)
	}
}
