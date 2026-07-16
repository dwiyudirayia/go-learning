// Teknik Advanced Modul 17 (microservices) — contoh runnable + penjelasan detail.
//
// Simulasi pure-Go pola panggilan antar-service: propagasi deadline, retry
// dengan backoff+jitter, correlation ID, dan circuit breaker sederhana.
// Jalankan:
//
//	go run ./17-studi-kasus-microservices/advanced
package main

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"
)

type ctxKey string

const correlationKey ctxKey = "correlation_id"

var ErrHulu = errors.New("layanan inventory error sementara")

// rng lokal & ber-seed tetap -> jitter deterministik untuk demo (tanpa
// rand.Seed yang sudah deprecated sejak Go 1.20).
var rng = rand.New(rand.NewSource(1))

// ===========================================================================
// 1. "Downstream service" yang FLAKY + menghormati DEADLINE context
// ===========================================================================
// Meniru layanan inventory: gagal beberapa kali lalu sukses. Ia mengecek
// ctx.Err() — bila deadline dari pemanggil (hulu) sudah lewat, langsung batal.
// Inilah propagasi deadline lintas hop: batas waktu mengalir ke bawah.
//
// 🔍 Analogi: propagasi deadline itu seperti ESTAFET DENGAN SATU STOPWATCH —
// pelari pertama (API gateway) menyalakan stopwatch 2 detik, dan SEMUA pelari
// berikutnya (service hilir) melirik stopwatch yang sama. Begitu waktu habis,
// pelari mana pun langsung berhenti — tak ada yang lari sia-sia.
type inventory struct{ gagalTersisa int }

func (in *inventory) cekStok(ctx context.Context, sku string) (int, error) {
	if err := ctx.Err(); err != nil {
		return 0, err // deadline/cancel dari hulu -> jangan kerjakan apa pun
	}
	cid, _ := ctx.Value(correlationKey).(string)
	if in.gagalTersisa > 0 {
		in.gagalTersisa--
		fmt.Printf("    [inventory cid=%s] gagal (sisa %d)\n", cid, in.gagalTersisa)
		return 0, ErrHulu
	}
	fmt.Printf("    [inventory cid=%s] OK -> stok %s\n", cid, sku)
	return 42, nil
}

// ===========================================================================
// 2. Retry dengan BACKOFF + JITTER, dibatasi deadline context
// ===========================================================================
// Retry hanya untuk error sementara & operasi idempoten. Backoff eksponensial
// mengurangi tekanan; jitter mencegah semua klien retry serempak (thundering
// herd). Loop berhenti bila context habis (menghormati deadline hulu).
//
// 🔍 Analogi: backoff itu seperti menelepon nomor yang sibuk — coba lagi
// setelah 1 menit, lalu 2, lalu 4 (memberi napas, bukan menyerbu). Jitter =
// tiap orang menunggu WAKTU ACAK sedikit berbeda; kalau seribu penelepon
// menunggu tepat 1 menit yang sama, sentral langsung tumbang lagi bersamaan
// (thundering herd).
func panggilDenganRetry(ctx context.Context, fn func(context.Context) (int, error)) (int, error) {
	const maxPercobaan = 5
	backoff := 10 * time.Millisecond
	var errTerakhir error
	for percobaan := 1; percobaan <= maxPercobaan; percobaan++ {
		hasil, err := fn(ctx)
		if err == nil {
			return hasil, nil
		}
		errTerakhir = err
		jitter := time.Duration(rng.Int63n(int64(backoff))) // 0..backoff
		tunggu := backoff + jitter
		fmt.Printf("  percobaan %d gagal, tunggu %v...\n", percobaan, tunggu.Truncate(time.Millisecond))
		select {
		case <-ctx.Done():
			return 0, fmt.Errorf("deadline habis saat retry: %w", ctx.Err())
		case <-time.After(tunggu):
			backoff *= 2 // eksponensial
		}
	}
	return 0, fmt.Errorf("menyerah setelah %d percobaan: %w", maxPercobaan, errTerakhir)
}

// ===========================================================================
// 3. CIRCUIT BREAKER sederhana — berhenti memanggil dependency yang rusak
// ===========================================================================
// Setelah sejumlah kegagalan beruntun, breaker "terbuka" dan menolak panggilan
// cepat (fail-fast) tanpa menyentuh dependency, memberi waktu ia pulih dan
// mencegah cascade failure.
//
// 🔍 Analogi: persis SEKRING LISTRIK di rumah — korsleting beruntun membuat
// sekring "putus" (breaker terbuka) sehingga arus berhenti mengalir ke
// perangkat yang rusak. Lebih baik satu ruangan gelap sebentar daripada
// seisi rumah terbakar (cascade failure).
type breaker struct {
	gagalBeruntun int
	ambang        int
	terbuka       bool
}

var ErrBreakerTerbuka = errors.New("circuit breaker terbuka (fail-fast)")

func (b *breaker) call(fn func() error) error {
	if b.terbuka {
		return ErrBreakerTerbuka
	}
	err := fn()
	if err != nil {
		b.gagalBeruntun++
		if b.gagalBeruntun >= b.ambang {
			b.terbuka = true
		}
		return err
	}
	b.gagalBeruntun = 0
	return nil
}

func main() {
	// --- Skenario A: retry berhasil dalam deadline ---
	fmt.Println("== 1&2. deadline + retry backoff/jitter (sukses) ==")
	inv := &inventory{gagalTersisa: 2} // gagal 2x lalu sukses
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	ctx = context.WithValue(ctx, correlationKey, "req-777") // correlation id lintas hop
	defer cancel()
	stok, err := panggilDenganRetry(ctx, func(c context.Context) (int, error) {
		return inv.cekStok(c, "SKU-1")
	})
	fmt.Printf("  hasil: stok=%d err=%v\n", stok, err)

	// --- Skenario B: deadline terlalu ketat -> menyerah ---
	fmt.Println("== 2. deadline ketat -> retry menyerah ==")
	inv2 := &inventory{gagalTersisa: 99}
	ctx2, cancel2 := context.WithTimeout(context.Background(), 25*time.Millisecond)
	ctx2 = context.WithValue(ctx2, correlationKey, "req-888")
	defer cancel2()
	_, err = panggilDenganRetry(ctx2, func(c context.Context) (int, error) {
		return inv2.cekStok(c, "SKU-2")
	})
	fmt.Println("  hasil:", err)

	// --- Skenario C: circuit breaker terbuka setelah 3 kegagalan ---
	fmt.Println("== 3. circuit breaker ==")
	cb := &breaker{ambang: 3}
	for i := 1; i <= 5; i++ {
		err := cb.call(func() error { return ErrHulu }) // selalu gagal
		fmt.Printf("  call %d -> %v (breaker terbuka=%v)\n", i, err, cb.terbuka)
	}
}
