package main

import "net/http"

// 🔍 Analogi: security headers itu seperti STIKER INSTRUKSI KEAMANAN yang ditempel di tiap paket
// yang kamu kirim ke browser: "jangan tebak isi paket" (nosniff), "jangan pajang aku di etalase
// orang lain" (X-Frame-Options = anti-clickjacking), "hanya percaya sumber dari rumah sendiri"
// (CSP). Murah (cuma menambah header) tapi menutup banyak celah serangan umum. "Defense in depth".
// securityHeaders menambahkan header keamanan standar ke setiap response.
// Ini pertahanan murah terhadap serangan umum (clickjacking, MIME sniffing, XSS).
func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		// Cegah browser menebak-nebak tipe konten (MIME sniffing).
		h.Set("X-Content-Type-Options", "nosniff")
		// Cegah halaman ditanam di <iframe> situs lain (clickjacking).
		h.Set("X-Frame-Options", "DENY")
		// Batasi sumber daya yang boleh dimuat (dasar; sesuaikan per app).
		h.Set("Content-Security-Policy", "default-src 'self'")
		// Paksa HTTPS di kunjungan berikutnya (hanya efektif via HTTPS).
		h.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		// Batasi informasi referrer yang dibocorkan.
		h.Set("Referrer-Policy", "no-referrer")
		next.ServeHTTP(w, r)
	})
}
