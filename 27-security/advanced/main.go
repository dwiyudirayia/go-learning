// Teknik Advanced Modul 27 (security) — contoh runnable + penjelasan detail.
//
// Rate limiting, security headers, dan perbandingan tahan timing-attack.
// Jalankan:
//
//	go run ./27-security/advanced
package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"golang.org/x/time/rate"
)

func main() {
	// =====================================================================
	// 1. RATE LIMITING (token bucket) — lindungi endpoint mahal / login
	// =====================================================================
	// rate.NewLimiter(r, b): isi ulang r token/detik, kapasitas burst b.
	// Allow() true bila ada token. Di produksi: satu limiter PER IP/user.
	//
	// 🔍 Analogi: token bucket itu seperti EMBER KUPON antrean — kupon
	// menetes masuk dengan kecepatan tetap (r/detik), ember memuat maksimal
	// b kupon (burst). Mau dilayani? Ambil satu kupon; ember kosong berarti
	// tunggu tetesan berikutnya.
	lim := rate.NewLimiter(rate.Every(100*time.Millisecond), 2) // 10/dtk, burst 2
	fmt.Println("== 1. rate limiting (burst 2) ==")
	for i := 1; i <= 5; i++ {
		fmt.Printf("  request %d -> diizinkan=%v\n", i, lim.Allow())
	}

	// =====================================================================
	// 2. Security headers via middleware
	// =====================================================================
	// 🔍 Analogi: security headers itu seperti STIKER ATURAN di pintu masuk
	// ("dilarang merokok", "wajib helm") — server menitipkan aturan main, dan
	// BROWSER-lah yang menegakkannya: jangan tebak-tebak isi file (nosniff),
	// jangan mau dipajang di bingkai orang (DENY), wajib jalur HTTPS (HSTS).
	secure := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := w.Header()
			h.Set("X-Content-Type-Options", "nosniff")             // cegah MIME sniffing
			h.Set("X-Frame-Options", "DENY")                       // cegah clickjacking
			h.Set("Strict-Transport-Security", "max-age=31536000") // paksa HTTPS
			h.Set("Content-Security-Policy", "default-src 'self'")
			next.ServeHTTP(w, r)
		})
	}
	handler := secure(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "ok")
	}))
	srv := httptest.NewServer(handler)
	defer srv.Close()
	resp, err := http.Get(srv.URL)
	if err != nil {
		panic(err)
	}
	resp.Body.Close()
	fmt.Println("== 2. security headers ==")
	for _, k := range []string{"X-Content-Type-Options", "X-Frame-Options", "Content-Security-Policy"} {
		fmt.Printf("  %-25s: %s\n", k, resp.Header.Get(k))
	}

	// =====================================================================
	// 3. Perbandingan TAHAN TIMING-ATTACK (constant time)
	// =====================================================================
	// Membandingkan token/HMAC dengan == bocor lewat waktu eksekusi (berhenti
	// di byte pertama yang beda). hmac.Equal membandingkan dalam waktu KONSTAN.
	//
	// 🔍 Analogi: perbandingan dengan == itu seperti satpam yang bilang
	// "SALAH!" begitu digit pertama PIN meleset — pencuri tinggal mengukur
	// seberapa lama satpam berpikir untuk menebak digit demi digit.
	// hmac.Equal = satpam yang selalu memeriksa SEMUA digit dengan durasi
	// sama persis, benar atau salah, jadi tak ada petunjuk yang bocor.
	secret := []byte("kunci-rahasia")
	pesan := []byte("payload-webhook")
	tandaTanganAsli := tandaTangan(secret, pesan)

	fmt.Println("== 3. constant-time compare (verifikasi webhook) ==")
	fmt.Printf("  signature valid?   %v\n", hmac.Equal(tandaTanganAsli, tandaTangan(secret, pesan)))
	fmt.Printf("  signature dipalsu? %v\n", hmac.Equal(tandaTanganAsli, tandaTangan([]byte("kunci-salah"), pesan)))

	// Catatan: password -> bcrypt/argon2id (bukan sha); token -> access pendek +
	// refresh dgn rotasi & revocation; secret -> env/secret manager (modul 19).
}

// tandaTangan menghitung HMAC-SHA256 (dasar verifikasi webhook, lihat modul 46).
func tandaTangan(secret, msg []byte) []byte {
	mac := hmac.New(sha256.New, secret)
	mac.Write(msg)
	return mac.Sum(nil)
}
