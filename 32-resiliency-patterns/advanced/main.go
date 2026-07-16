// Teknik Advanced Modul 32 (resiliency) — contoh runnable + penjelasan detail.
//
// Circuit breaker DENGAN half-open recovery + bulkhead isolation. Jalankan:
//
//	go run ./32-resiliency-patterns/advanced
package main

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// ===========================================================================
// 1. CIRCUIT BREAKER dengan 3 state: Closed -> Open -> HalfOpen -> Closed
// ===========================================================================
// Closed  : lewatkan panggilan; hitung kegagalan beruntun.
// Open     : tolak cepat (fail-fast) selama cooldown, beri dependency waktu pulih.
// HalfOpen : setelah cooldown, izinkan SATU panggilan percobaan. Sukses -> Closed;
//
//	gagal -> Open lagi.
//
// 🔍 Analogi: half-open itu seperti MENJENGUK TEMAN YANG SAKIT — setelah masa
// istirahat (cooldown), kirim SATU orang dulu untuk mengetuk pintu. Kalau ia
// disambut (sukses), umumkan "sudah sehat, silakan berkunjung" (Closed);
// kalau pintu tak dibuka, biarkan istirahat lagi (Open). Jangan langsung
// mengajak serombongan — bisa kambuh.
type State int

const (
	Closed State = iota
	Open
	HalfOpen
)

func (s State) String() string { return [...]string{"CLOSED", "OPEN", "HALF-OPEN"}[s] }

type Breaker struct {
	ambang    int
	cooldown  time.Duration
	state     State
	kegagalan int
	bukaPada  time.Time
}

var ErrOpen = errors.New("circuit terbuka (fail-fast)")

func (b *Breaker) Call(fn func() error) error {
	// Transisi Open -> HalfOpen bila cooldown lewat.
	if b.state == Open && time.Since(b.bukaPada) >= b.cooldown {
		b.state = HalfOpen
	}
	if b.state == Open {
		return ErrOpen // tolak tanpa menyentuh dependency
	}

	err := fn()
	if err != nil {
		b.kegagalan++
		if b.state == HalfOpen || b.kegagalan >= b.ambang {
			b.state = Open // percobaan gagal / ambang tercapai -> buka
			b.bukaPada = time.Now()
		}
		return err
	}
	// sukses -> reset
	b.kegagalan = 0
	b.state = Closed
	return nil
}

// ===========================================================================
// 2. BULKHEAD — isolasi resource agar satu dependency lambat tak menyedot semua
// ===========================================================================
// Tiap dependency dapat "kolam" konkurensi terpisah (semaphore). Bila kolam
// dependency A penuh, request ke A ditolak/antre TANPA menghabiskan slot untuk B.
//
// 🔍 Analogi: bulkhead memang diambil dari SEKAT LAMBUNG KAPAL — lambung
// dibagi beberapa kompartemen kedap; satu bocor (dependency lambat), air
// terkurung di kompartemen itu saja dan kapal tetap mengapung. Tanpa sekat,
// satu kebocoran menenggelamkan semuanya (semua goroutine tersedot menunggu
// satu dependency).
type Bulkhead struct{ slot chan struct{} }

func NewBulkhead(n int) *Bulkhead { return &Bulkhead{slot: make(chan struct{}, n)} }

func (b *Bulkhead) Coba(fn func()) bool {
	select {
	case b.slot <- struct{}{}: // dapat slot
		defer func() { <-b.slot }()
		fn()
		return true
	default:
		return false // kolam penuh -> tolak cepat (lindungi resource)
	}
}

func main() {
	// --- Circuit breaker: dependency "mati" lalu "pulih" ---
	fmt.Println("== 1. circuit breaker ==")
	cb := &Breaker{ambang: 3, cooldown: 50 * time.Millisecond}
	sehat := false // fase awal: dependency mati
	panggil := func() error {
		if !sehat {
			return errors.New("dependency down")
		}
		return nil
	}

	for i := 1; i <= 4; i++ {
		err := cb.Call(panggil)
		fmt.Printf("  call %d: state=%-9s err=%v\n", i, cb.state, err)
	}
	fmt.Println("  ...tunggu cooldown, dependency pulih...")
	time.Sleep(60 * time.Millisecond)
	sehat = true
	for i := 5; i <= 6; i++ {
		err := cb.Call(panggil)
		fmt.Printf("  call %d: state=%-9s err=%v\n", i, cb.state, err)
	}

	// --- Bulkhead: kapasitas 2, tiga pekerjaan bersamaan -> satu ditolak ---
	fmt.Println("== 2. bulkhead (kapasitas 2) ==")
	bh := NewBulkhead(2)
	var wg sync.WaitGroup
	var mu sync.Mutex
	diterima, ditolak := 0, 0
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ok := bh.Coba(func() { time.Sleep(30 * time.Millisecond) })
			mu.Lock()
			if ok {
				diterima++
			} else {
				ditolak++
			}
			mu.Unlock()
		}()
	}
	wg.Wait()
	fmt.Printf("  diterima=%d ditolak=%d (kolam penuh melindungi dependency)\n", diterima, ditolak)
}
