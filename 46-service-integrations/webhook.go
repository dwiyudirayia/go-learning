package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// 🔍 Analogi besar WEBHOOK: kebalikan dari "kamu menelepon layanan". Webhook = layanan yang
// MENELEPONMU saat ada kabar ("pembayaranmu berhasil!"). Kamu memberi mereka nomor (URL endpoint),
// mereka menelepon saat ada event. Hemat: kamu tak perlu bertanya berulang "sudah bayar belum?".
//
// 🔍 Analogi tanda tangan HMAC: karena siapa saja bisa menelepon nomormu, bagaimana tahu ini benar
// dari Stripe, bukan penipu? Pengirim asli menyegel pesan dengan "CAP LILIN rahasia" (HMAC pakai
// signing secret yang cuma kalian berdua tahu). Kamu hitung ulang cap dari isi pesan; kalau cocok,
// asli. Penipu tak punya secret -> capnya salah -> ditolak.
//
// 🔍 Analogi anti-replay: pengecekan timestamp mencegah penyerang MEREKAM pesan asli lalu MENGIRIM
// ULANG nanti (seperti memutar rekaman izin lama). Pesan lebih tua dari toleransi -> ditolak.

// WEBHOOK: layanan (Stripe/GitHub) mengirim event ke endpoint-mu (mis. "pembayaran
// berhasil"). WAJIB verifikasi TANDA TANGAN agar tak ada yang memalsukan event.
// Pola Stripe: HMAC-SHA256 atas "timestamp.payload" dengan signing secret.

// SignPayload menghitung header tanda tangan (dipakai pengirim & untuk test).
func SignPayload(secret string, timestamp int64, payload []byte) string {
	signed := fmt.Sprintf("%d.%s", timestamp, payload)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signed))
	sig := hex.EncodeToString(mac.Sum(nil))
	return fmt.Sprintf("t=%d,v1=%s", timestamp, sig)
}

var (
	ErrBadSignature = errors.New("tanda tangan webhook tidak valid")
	ErrTooOld       = errors.New("webhook kadaluarsa (kemungkinan replay)")
)

// VerifyWebhook memverifikasi header terhadap payload memakai signing secret.
// tolerance mencegah serangan REPLAY (event lama dikirim ulang).
func VerifyWebhook(secret, header string, payload []byte, tolerance time.Duration, now time.Time) error {
	var ts int64
	var sig string
	for _, part := range strings.Split(header, ",") {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			continue
		}
		switch kv[0] {
		case "t":
			ts, _ = strconv.ParseInt(kv[1], 10, 64)
		case "v1":
			sig = kv[1]
		}
	}
	if ts == 0 || sig == "" {
		return ErrBadSignature
	}

	// Cek umur (anti-replay).
	if d := now.Sub(time.Unix(ts, 0)); d < 0 || d > tolerance {
		return ErrTooOld
	}

	// Hitung ulang & bandingkan secara CONSTANT-TIME (cegah timing attack).
	expected := fmt.Sprintf("%d.%s", ts, payload)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(expected))
	want := mac.Sum(nil)

	got, err := hex.DecodeString(sig)
	if err != nil || !hmac.Equal(got, want) {
		return ErrBadSignature
	}
	return nil
}
