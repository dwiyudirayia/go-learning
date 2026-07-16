// Teknik Advanced Modul 29 (clean architecture) — contoh runnable + penjelasan detail.
//
// Ports & Adapters (hexagonal) + Dependency Inversion + adapter DECORATOR.
// Jalankan:
//
//	go run ./29-clean-architecture/advanced
package main

import (
	"errors"
	"fmt"
)

// ===========================================================================
// DOMAIN (inti) — tak tahu apa pun soal DB, HTTP, atau framework.
// ===========================================================================
type Order struct {
	ID    string
	Total int
}

var ErrInvalid = errors.New("order tidak valid")

// PORT: interface DIDEFINISIKAN oleh domain (kebutuhannya), BUKAN oleh infra.
// Inilah Dependency Inversion — inti bergantung pada abstraksi; adapter di luar
// yang menyesuaikan diri. Panah ketergantungan mengarah KE DALAM.
//
// 🔍 Analogi: port itu seperti STOP KONTAK DI DINDING RUMAH — bentuk lubangnya
// ditentukan PEMILIK RUMAH (domain), dan produsen alat elektronik (infra:
// Postgres, Redis) yang wajib membuat colokan sesuai. Bukan sebaliknya rumah
// dibongkar tiap ganti merek alat.
type OrderRepo interface {
	Save(Order) error
	Get(id string) (Order, bool)
}

// USE CASE (interactor) — aturan aplikasi, memakai port (bukan tipe konkret).
type PlaceOrder struct {
	repo OrderRepo
}

func (uc PlaceOrder) Exec(id string, total int) error {
	if id == "" || total <= 0 {
		return fmt.Errorf("%w: id/total", ErrInvalid)
	}
	return uc.repo.Save(Order{ID: id, Total: total})
}

// ===========================================================================
// ADAPTERS (luar) — implementasi konkret dari port. Bisa ditukar bebas.
// ===========================================================================

// Adapter 1: penyimpanan in-memory (di produksi: Postgres/SQLite).
type memRepo struct{ data map[string]Order }

func newMemRepo() *memRepo                     { return &memRepo{data: map[string]Order{}} }
func (r *memRepo) Save(o Order) error          { r.data[o.ID] = o; return nil }
func (r *memRepo) Get(id string) (Order, bool) { o, ok := r.data[id]; return o, ok }

// Adapter 2: DECORATOR — membungkus repo lain untuk menambah logging TANPA
// mengubah domain maupun repo asli (Open/Closed Principle).
//
// 🔍 Analogi: decorator itu seperti SARUNG HP — menambah fungsi (pelindung,
// dudukan kartu) tanpa membongkar mesin HP-nya. Dari luar tetap "HP" (masih
// memenuhi interface OrderRepo), dan sarung bisa dilapis: anti-air di atas
// anti-benturan (logging di atas caching di atas repo asli).
type loggingRepo struct{ next OrderRepo }

func (r loggingRepo) Save(o Order) error {
	fmt.Printf("  [log] menyimpan order %s (total %d)\n", o.ID, o.Total)
	return r.next.Save(o)
}
func (r loggingRepo) Get(id string) (Order, bool) { return r.next.Get(id) }

func main() {
	// =====================================================================
	// COMPOSITION ROOT (di main) — merakit dependency dari luar ke dalam.
	// Domain & use case tak pernah tahu implementasi mana yang dipakai.
	//
	// 🔍 Analogi: composition root itu seperti MEJA PERAKITAN PABRIK — satu
	// tempat di mana semua komponen (repo, decorator, use case) dipasangkan.
	// Komponennya sendiri saling buta; hanya perakit (main) yang tahu
	// rangkaian lengkapnya, jadi mengganti satu suku cadang cukup di sini.
	// =====================================================================
	var repo OrderRepo = newMemRepo()
	repo = loggingRepo{next: repo} // bungkus dgn decorator; use case tak berubah
	uc := PlaceOrder{repo: repo}

	fmt.Println("== use case via port (implementasi tersembunyi) ==")
	if err := uc.Exec("ord-1", 150); err != nil {
		fmt.Println("  error:", err)
	}
	if err := uc.Exec("ord-2", 0); err != nil { // langgar aturan domain
		fmt.Println("  ditolak:", err)
	}

	if o, ok := repo.Get("ord-1"); ok {
		fmt.Printf("  tersimpan: %+v\n", o)
	}

	// Manfaat: ganti memRepo -> PostgresRepo cukup di main. Domain & use case
	// TIDAK disentuh, dan bisa diuji tanpa infra (isi port dgn mock).
}
