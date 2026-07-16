// Teknik Advanced Modul 46 (service integrations) — contoh runnable + penjelasan detail.
//
// Interface + MOCK untuk API pihak ketiga, dan verifikasi WEBHOOK (HMAC +
// anti-replay). Jalankan:
//
//	go run ./46-service-integrations/advanced
package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
)

// ===========================================================================
//  1. Interface + MOCK — bungkus dependency eksternal agar bisa diuji tanpa
//     memanggil API nyata (tanpa biaya, tanpa jaringan, deterministik).
//
// 🔍 Analogi: MockGateway itu seperti UANG MAINAN UNTUK LATIHAN KASIR — kasir
// baru berlatih transaksi lengkap tanpa menyentuh uang sungguhan (API Stripe
// yang berbiaya). Skenario "kartu ditolak" pun bisa dilatih kapan saja tanpa
// menunggu kejadian nyata.
// ===========================================================================
type PaymentGateway interface {
	Charge(amountCents int) (string, error)
}

// MockGateway menggantikan Stripe/dsb saat test.
type MockGateway struct{ tolak bool }

func (m MockGateway) Charge(amount int) (string, error) {
	if m.tolak {
		return "", errors.New("kartu ditolak")
	}
	return fmt.Sprintf("ch_%d", amount), nil
}

// ===========================================================================
// 2. Verifikasi WEBHOOK: HMAC signature + ANTI-REPLAY
// ===========================================================================
// Webhook dari pihak ketiga harus dibuktikan asli: tanda tangan HMAC memakai
// shared secret. Tambahan anti-replay: tolak timestamp lama & signature yang
// sudah pernah dipakai (nonce) -> penyerang tak bisa memutar ulang request sah.
//
// 🔍 Analogi: webhook itu SURAT PENTING dari mitra. HMAC = segel lilin yang
// hanya bisa dicetak pemegang stempel rahasia (shared secret). Anti-replay =
// dua pemeriksaan tambahan: tanggal surat tak boleh terlalu tua (timestamp),
// dan nomor surat yang sama tak boleh diproses dua kali (dedup) — sebab
// pencuri bisa saja MEMFOTOKOPI surat asli bersegel dan mengirimnya ulang.
var secret = []byte("whsec_rahasia")

func tandaTangan(ts int64, body string) string {
	mac := hmac.New(sha256.New, secret)
	fmt.Fprintf(mac, "%d.%s", ts, body) // timestamp ikut ditandatangani
	return hex.EncodeToString(mac.Sum(nil))
}

var (
	ErrSignature = errors.New("signature tidak valid")
	ErrExpired   = errors.New("timestamp kedaluwarsa (replay?)")
	ErrReplay    = errors.New("signature sudah dipakai (replay)")
)

func verifikasiWebhook(ts int64, body, sig string, now int64, terlihat map[string]bool) error {
	// (a) tolak yang terlalu tua -> mempersempit jendela serangan replay.
	if now-ts > 300 { // 5 menit
		return ErrExpired
	}
	// (b) bandingkan signature dalam waktu KONSTAN (hindari timing attack).
	if !hmac.Equal([]byte(tandaTangan(ts, body)), []byte(sig)) {
		return ErrSignature
	}
	// (c) tolak signature yang sudah pernah diproses (dedup nonce).
	if terlihat[sig] {
		return ErrReplay
	}
	terlihat[sig] = true
	return nil
}

func main() {
	// 1. interface + mock
	fmt.Println("== 1. interface + mock (uji tanpa API nyata) ==")
	var gw PaymentGateway = MockGateway{}
	id, _ := gw.Charge(2500)
	fmt.Println("  charge sukses -> id:", id)
	_, err := MockGateway{tolak: true}.Charge(2500)
	fmt.Println("  charge ditolak ->", err)

	// 2. webhook
	fmt.Println("== 2. verifikasi webhook (HMAC + anti-replay) ==")
	const now = 1_000_000
	body := `{"event":"payment.succeeded"}`
	ts := int64(now - 10) // 10 detik lalu (masih segar)
	sig := tandaTangan(ts, body)
	terlihat := map[string]bool{}

	fmt.Println("  valid & segar       :", verifikasiWebhook(ts, body, sig, now, terlihat)) // nil
	fmt.Println("  REPLAY (sig sama)   :", verifikasiWebhook(ts, body, sig, now, terlihat)) // ErrReplay
	fmt.Println("  signature dipalsu   :", verifikasiWebhook(ts, body, "deadbeef", now, terlihat))
	old := int64(now - 999)
	fmt.Println("  timestamp kedaluwarsa:", verifikasiWebhook(old, body, tandaTangan(old, body), now, terlihat))

	// Catatan: simpan secret di env/secret manager; gunakan idempotency key saat
	// MENGIRIM ke pihak ketiga agar retry tak menggandakan (mis. double charge).
}
