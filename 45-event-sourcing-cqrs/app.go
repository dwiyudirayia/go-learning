package main

import "sync"

// ------------------------------------------------------------------
// SISI WRITE (Command) — CQRS
// ------------------------------------------------------------------

// BankService memproses COMMAND: muat event -> rekonstruksi aggregate ->
// jalankan command -> simpan event baru.
type BankService struct {
	store *EventStore
}

func NewBankService(store *EventStore) *BankService {
	return &BankService{store: store}
}

func (s *BankService) load(id string) *Account {
	acc := NewAccountFromEvents(s.store.Load(id))
	acc.ID = id // pastikan ID terisi meski stream kosong
	return acc
}

func (s *BankService) OpenAccount(id, owner string) error {
	events, err := s.load(id).Open(owner)
	if err != nil {
		return err
	}
	s.store.Append(id, events)
	return nil
}

func (s *BankService) Deposit(id string, amount int) error {
	events, err := s.load(id).Deposit(amount)
	if err != nil {
		return err
	}
	s.store.Append(id, events)
	return nil
}

func (s *BankService) Withdraw(id string, amount int) error {
	events, err := s.load(id).Withdraw(amount)
	if err != nil {
		return err
	}
	s.store.Append(id, events)
	return nil
}

// GetAccount merekonstruksi state terkini dari event (sumber kebenaran).
func (s *BankService) GetAccount(id string) *Account {
	return s.load(id)
}

// ------------------------------------------------------------------
// SISI READ (Query/Projection) — CQRS
// ------------------------------------------------------------------

// BalanceProjection = READ MODEL: tampilan cepat saldo per rekening, dibangun
// dengan mendengarkan event. Terpisah dari sisi write -> bisa dioptimasi khusus
// baca (mis. tabel denormalisasi, cache). Ini inti CQRS.
type BalanceProjection struct {
	mu       sync.Mutex
	balances map[string]int
	owners   map[string]string
}

func NewBalanceProjection() *BalanceProjection {
	return &BalanceProjection{balances: map[string]int{}, owners: map[string]string{}}
}

// On dipanggil untuk tiap event (didaftarkan via store.Subscribe).
func (p *BalanceProjection) On(e Event) {
	p.mu.Lock()
	defer p.mu.Unlock()
	switch ev := e.(type) {
	case AccountOpened:
		p.owners[ev.AccountID] = ev.Owner
		p.balances[ev.AccountID] = 0
	case MoneyDeposited:
		p.balances[ev.AccountID] += ev.Amount
	case MoneyWithdrawn:
		p.balances[ev.AccountID] -= ev.Amount
	}
}

func (p *BalanceProjection) Balance(id string) int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.balances[id]
}
