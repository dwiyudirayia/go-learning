// golang-jwt/jwt — menerbitkan & memverifikasi JSON Web Token (access + refresh).
//
// Jalankan: go run ./libraries/jwt
// Test:     go test ./libraries/jwt
//
// 🔍 Analogi besar: JWT itu GELANG KONSER yang sudah dicap panitia. Begitu kamu masuk,
// petugas di setiap panggung cukup melihat capnya — mereka tak perlu menelepon loket
// untuk bertanya "orang ini sudah bayar belum?". Itulah arti "stateless": server tak
// menyimpan daftar sesi, semua informasi ada di gelangnya, dan capnya membuktikan
// gelang itu tak dipalsukan.
//
// Konsekuensi yang WAJIB dipahami: gelang yang sudah terlanjur dicap TIDAK BISA
// dibatalkan dari jauh. Kalau seseorang mencuri gelangmu, ia tetap bisa masuk sampai
// gelangnya kedaluwarsa. Inilah alasan masa berlaku access token dibuat pendek
// (5-15 menit) dan dipasangkan dengan refresh token — dibahas di bagian 4.
package main

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func main() {
	fmt.Println("=== golang-jwt/jwt ===")

	p := NewPenerbit("rahasia-yang-sangat-panjang-dan-acak", "go-learning")
	demoTerbitVerifikasi(p)
	demoKedaluwarsa()
	demoTandaTanganPalsu(p)
	demoSeranganAlgNone(p)
	demoRefresh(p)
	demoIsiTokenTerbaca(p)
}

// ------------------------------------------------------------------
// 1. Bentuk klaim
// ------------------------------------------------------------------

// Jenis token dibedakan supaya refresh token tak bisa dipakai sebagai access token.
const (
	JenisAkses   = "akses"
	JenisRefresh = "refresh"
)

// Klaim adalah isi token: data milikmu + klaim baku JWT.
//
// 🔍 Analogi: RegisteredClaims yang disematkan itu KOLOM BAKU di gelang konser yang
// dimengerti semua petugas di dunia: kapan kedaluwarsa (exp), siapa penerbitnya (iss),
// gelang ini milik siapa (sub). Field di atasnya adalah kolom tambahan khas acaramu
// sendiri ("akses ke area VIP").
//
// PERINGATAN yang sering disalahpahami: isi token TIDAK DIENKRIPSI, hanya di-encode
// base64. Siapa pun yang memegang tokennya bisa membaca seluruh klaim. Jangan pernah
// menaruh kata sandi, NIK, atau data pribadi sensitif di dalamnya.
type Klaim struct {
	Peran string `json:"peran"`
	Jenis string `json:"jenis"`
	jwt.RegisteredClaims
}

// ------------------------------------------------------------------
// 2. Penerbit token
// ------------------------------------------------------------------

// Penerbit membungkus semua pengaturan penandatanganan.
//
// 🔍 Analogi field 'sekarang': ini JAM DINDING yang bisa dicabut dan diganti. Di produksi
// ia menunjuk time.Now. Di test, kita menggantinya dengan jam palsu supaya bisa
// "melompat ke besok" secara instan — tanpa test yang tidur 15 menit menunggu token
// kedaluwarsa. Menyuntikkan waktu adalah pola yang sangat berharga untuk kode apa pun
// yang bergantung pada jam.
type Penerbit struct {
	rahasia     []byte
	penerbit    string
	masaAkses   time.Duration
	masaRefresh time.Duration
	sekarang    func() time.Time
}

// NewPenerbit membuat penerbit dengan masa berlaku yang lazim dipakai.
func NewPenerbit(rahasia, namaPenerbit string) *Penerbit {
	return &Penerbit{
		rahasia:  []byte(rahasia),
		penerbit: namaPenerbit,
		// 🔍 Analogi masa berlaku: access token itu TIKET HARIAN — sengaja pendek,
		// supaya kalau dicuri, kerugiannya cepat berakhir. Refresh token itu KARTU
		// MEMBER — berlaku lama, tapi disimpan lebih aman (cookie HttpOnly, bukan
		// localStorage) dan bisa dicabut lewat daftar di database.
		masaAkses:   15 * time.Minute,
		masaRefresh: 7 * 24 * time.Hour,
		sekarang:    time.Now,
	}
}

// DenganJam mengganti sumber waktu — dipakai test untuk melompati waktu.
func (p *Penerbit) DenganJam(jam func() time.Time) *Penerbit {
	salinan := *p
	salinan.sekarang = jam
	return &salinan
}

// buatToken adalah mesin bersama untuk kedua jenis token.
func (p *Penerbit) buatToken(subjek, peran, jenis string, masa time.Duration) (string, error) {
	if subjek == "" {
		return "", errors.New("subjek token tidak boleh kosong")
	}
	kini := p.sekarang()

	klaim := Klaim{
		Peran: peran,
		Jenis: jenis,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    p.penerbit,
			Subject:   subjek,
			IssuedAt:  jwt.NewNumericDate(kini),
			NotBefore: jwt.NewNumericDate(kini),
			ExpiresAt: jwt.NewNumericDate(kini.Add(masa)),
		},
	}

	// HS256 = tanda tangan simetris: kunci yang sama dipakai untuk mencap DAN memeriksa.
	// Cocok bila penerbit & pemeriksa adalah aplikasi yang sama.
	// Untuk banyak layanan yang perlu memverifikasi tapi tak boleh menerbitkan,
	// pakai RS256/ES256 (kunci privat mencap, kunci publik memeriksa).
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, klaim)
	s, err := token.SignedString(p.rahasia)
	if err != nil {
		return "", fmt.Errorf("gagal menandatangani token: %w", err)
	}
	return s, nil
}

// BuatAkses menerbitkan access token berumur pendek.
func (p *Penerbit) BuatAkses(subjek, peran string) (string, error) {
	return p.buatToken(subjek, peran, JenisAkses, p.masaAkses)
}

// BuatRefresh menerbitkan refresh token berumur panjang (tanpa peran).
func (p *Penerbit) BuatRefresh(subjek string) (string, error) {
	return p.buatToken(subjek, "", JenisRefresh, p.masaRefresh)
}

// ------------------------------------------------------------------
// 3. Verifikasi
// ------------------------------------------------------------------

// ErrJenisTokenSalah dikembalikan bila refresh token dipakai sebagai access token.
var ErrJenisTokenSalah = errors.New("jenis token tidak sesuai")

// Verifikasi memeriksa tanda tangan, masa berlaku, penerbit, DAN algoritmanya.
func (p *Penerbit) Verifikasi(tokenStr, jenisDiharapkan string) (*Klaim, error) {
	var klaim Klaim

	_, err := jwt.ParseWithClaims(tokenStr, &klaim,
		func(t *jwt.Token) (any, error) {
			return p.rahasia, nil
		},
		// 🔍 INI BARIS PALING PENTING DI SELURUH FILE.
		// WithValidMethods mengunci algoritma yang diterima. Tanpanya, penyerang bisa
		// mengirim token yang mengaku beralgoritma "none" (tanpa tanda tangan sama sekali)
		// atau menukar RS256 jadi HS256 lalu menandatanganinya dengan kunci PUBLIK yang
		// memang terbuka untuk umum. Keduanya adalah kerentanan JWT paling terkenal, dan
		// keduanya dimatikan oleh satu baris ini.
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer(p.penerbit),
		jwt.WithExpirationRequired(),
		// Toleransi kecil untuk jam server yang tak persis sama.
		jwt.WithLeeway(5*time.Second),
		jwt.WithTimeFunc(p.sekarang),
	)
	if err != nil {
		return nil, fmt.Errorf("token ditolak: %w", err)
	}

	// Pemeriksaan aturan bisnis, dilakukan SETELAH token terbukti sah.
	if klaim.Jenis != jenisDiharapkan {
		return nil, fmt.Errorf("%w: ingin %q, dapat %q", ErrJenisTokenSalah, jenisDiharapkan, klaim.Jenis)
	}
	return &klaim, nil
}

func demoTerbitVerifikasi(p *Penerbit) {
	fmt.Println("\n-- Terbitkan & verifikasi --")

	tok, err := p.BuatAkses("user-42", "admin")
	if err != nil {
		fmt.Println("   error:", err)
		return
	}
	fmt.Printf("   token: %s...\n", tok[:40])

	k, err := p.Verifikasi(tok, JenisAkses)
	if err != nil {
		fmt.Println("   error:", err)
		return
	}
	fmt.Printf("   sah -> subjek=%s peran=%s kedaluwarsa=%s\n",
		k.Subject, k.Peran, k.ExpiresAt.Format(time.Kitchen))
}

func demoKedaluwarsa() {
	fmt.Println("\n-- Token kedaluwarsa --")

	awal := time.Date(2026, 7, 23, 10, 0, 0, 0, time.UTC)
	p := NewPenerbit("rahasia-uji", "go-learning").DenganJam(func() time.Time { return awal })

	tok, err := p.BuatAkses("user-1", "pembaca")
	if err != nil {
		fmt.Println("   error:", err)
		return
	}

	// Lompat 16 menit ke depan — masa akses cuma 15 menit.
	nanti := p.DenganJam(func() time.Time { return awal.Add(16 * time.Minute) })
	if _, err := nanti.Verifikasi(tok, JenisAkses); err != nil {
		fmt.Println("   16 menit kemudian ->", err)
		fmt.Printf("   dikenali sebagai kedaluwarsa? %t\n", errors.Is(err, jwt.ErrTokenExpired))
	}
}

func demoTandaTanganPalsu(p *Penerbit) {
	fmt.Println("\n-- Tanda tangan palsu --")

	tok, err := p.BuatAkses("user-42", "pembaca")
	if err != nil {
		fmt.Println("   error:", err)
		return
	}

	// Penyerang mencoba mengubah isi token tanpa tahu kunci rahasianya.
	rusak := TukarKlaim(tok)
	if _, err := p.Verifikasi(rusak, JenisAkses); err != nil {
		fmt.Printf("   isi diubah -> ditolak (tanda tangan invalid? %t)\n",
			errors.Is(err, jwt.ErrTokenSignatureInvalid))
	}

	// Penerbit lain (kunci berbeda) juga harus ditolak.
	lain := NewPenerbit("kunci-yang-berbeda", "go-learning")
	if _, err := lain.Verifikasi(tok, JenisAkses); err != nil {
		fmt.Println("   diverifikasi dengan kunci lain -> ditolak")
	}
}

// TukarKlaim merusak bagian klaim sebuah token (meniru upaya pemalsuan).
//
// 🔍 Analogi: seperti menghapus tulisan "kelas ekonomi" di tiket lalu menulis "bisnis",
// tapi lupa bahwa hologram di sudut tiket dibuat berdasarkan tulisan aslinya.
// Tanda tangan JWT dihitung dari header+klaim — begitu klaim diubah, capnya tak cocok lagi.
func TukarKlaim(token string) string {
	bagian := strings.Split(token, ".")
	if len(bagian) != 3 {
		return token
	}
	// Ganti satu karakter di bagian klaim.
	klaim := []byte(bagian[1])
	if klaim[0] == 'e' {
		klaim[0] = 'f'
	} else {
		klaim[0] = 'e'
	}
	return bagian[0] + "." + string(klaim) + "." + bagian[2]
}

// ------------------------------------------------------------------
// 4. Serangan "alg: none"
// ------------------------------------------------------------------

// TokenTanpaTandaTangan membuat token beralgoritma "none" — tanpa cap sama sekali.
//
// 🔍 Analogi: ini seperti membuat tiket sendiri di rumah lalu menulis "TIDAK PERLU
// HOLOGRAM" di sudutnya, berharap petugas percaya begitu saja. Sebagian pustaka JWT
// zaman dulu memang percaya — dan itulah kerentanan yang membobol banyak sistem.
// Pertahanannya: WithValidMethods di Verifikasi().
func TokenTanpaTandaTangan(subjek, peran string) (string, error) {
	klaim := Klaim{
		Peran: peran,
		Jenis: JenisAkses,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "go-learning",
			Subject:   subjek,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodNone, klaim)
	s, err := t.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		return "", fmt.Errorf("gagal membuat token none: %w", err)
	}
	return s, nil
}

func demoSeranganAlgNone(p *Penerbit) {
	fmt.Println("\n-- Serangan alg:none --")

	jahat, err := TokenTanpaTandaTangan("user-1", "admin")
	if err != nil {
		fmt.Println("   error:", err)
		return
	}
	fmt.Println("   penyerang membuat token 'admin' tanpa tanda tangan")

	if _, err := p.Verifikasi(jahat, JenisAkses); err != nil {
		fmt.Println("   -> DITOLAK berkat WithValidMethods")
	} else {
		fmt.Println("   -> DITERIMA (ini akan jadi lubang keamanan serius!)")
	}
}

// ------------------------------------------------------------------
// 5. Refresh token
// ------------------------------------------------------------------

// Segarkan menukar refresh token yang sah dengan access token baru.
//
// 🔍 Analogi: menunjukkan KARTU MEMBER di loket untuk mendapat TIKET HARIAN baru.
// Di sistem produksi, langkah ini juga tempat kamu memeriksa daftar cabut (apakah
// kartu ini sudah dilaporkan hilang?) dan idealnya menerbitkan refresh token baru
// sekaligus mematikan yang lama — namanya refresh token rotation.
func (p *Penerbit) Segarkan(refreshToken string, peranSekarang string) (string, error) {
	klaim, err := p.Verifikasi(refreshToken, JenisRefresh)
	if err != nil {
		return "", fmt.Errorf("refresh gagal: %w", err)
	}
	return p.BuatAkses(klaim.Subject, peranSekarang)
}

func demoRefresh(p *Penerbit) {
	fmt.Println("\n-- Refresh token --")

	rt, err := p.BuatRefresh("user-42")
	if err != nil {
		fmt.Println("   error:", err)
		return
	}

	// Refresh token TIDAK boleh diterima sebagai access token.
	if _, err := p.Verifikasi(rt, JenisAkses); err != nil {
		fmt.Printf("   dipakai sebagai access token -> ditolak (jenis salah? %t)\n",
			errors.Is(err, ErrJenisTokenSalah))
	}

	baru, err := p.Segarkan(rt, "editor")
	if err != nil {
		fmt.Println("   error:", err)
		return
	}
	k, err := p.Verifikasi(baru, JenisAkses)
	if err != nil {
		fmt.Println("   error:", err)
		return
	}
	fmt.Printf("   access token baru -> subjek=%s peran=%s\n", k.Subject, k.Peran)
}

// ------------------------------------------------------------------
// 6. Isi token bisa dibaca siapa saja
// ------------------------------------------------------------------

// BacaKlaimTanpaVerifikasi menguraikan token TANPA memeriksa tanda tangannya.
//
// 🔍 Analogi: ini seperti membaca tulisan di gelang konser tanpa memeriksa hologramnya.
// Berguna untuk hal remeh (menampilkan nama pengguna di frontend, atau melihat isi token
// saat menyelidiki masalah). TIDAK BOLEH dipakai untuk keputusan keamanan apa pun —
// isinya bisa saja karangan penyerang.
func BacaKlaimTanpaVerifikasi(tokenStr string) (*Klaim, error) {
	var klaim Klaim
	parser := jwt.NewParser()
	if _, _, err := parser.ParseUnverified(tokenStr, &klaim); err != nil {
		return nil, fmt.Errorf("token tidak bisa diurai: %w", err)
	}
	return &klaim, nil
}

func demoIsiTokenTerbaca(p *Penerbit) {
	fmt.Println("\n-- Isi token tidak rahasia --")

	tok, err := p.BuatAkses("user-42", "admin")
	if err != nil {
		fmt.Println("   error:", err)
		return
	}
	k, err := BacaKlaimTanpaVerifikasi(tok)
	if err != nil {
		fmt.Println("   error:", err)
		return
	}
	fmt.Printf("   dibaca tanpa kunci apa pun -> subjek=%s peran=%s\n", k.Subject, k.Peran)
	fmt.Println("   -> jangan pernah menaruh data rahasia di dalam klaim!")
}
