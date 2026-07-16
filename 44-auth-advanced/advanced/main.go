// Teknik Advanced Modul 44 (auth advanced) — contoh runnable + penjelasan detail.
//
// JWT (golang-jwt v5) + otorisasi RBAC + ABAC. Jalankan:
//
//	go run ./44-auth-advanced/advanced
package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var secret = []byte("kunci-rahasia-server")

// ===========================================================================
// 1. JWT — terbitkan & verifikasi token (stateless)
// ===========================================================================
// 🔍 Analogi: JWT itu seperti GELANG KONSER anti-palsu — identitas & kelas
// tiket (claims) tercetak di gelangnya, dan hologram pabrik (signature)
// membuat pemalsuan ketahuan. Penjaga pintu mana pun bisa memeriksanya TANPA
// menelepon loket pusat (stateless) — tapi karena itu pula gelang yang sudah
// beredar sulit ditarik kembali (revoke), maka masa berlakunya dibuat pendek.
func terbitkanToken(user, role string) (string, error) {
	claims := jwt.MapClaims{
		"sub":  user,
		"role": role,
		"exp":  time.Now().Add(time.Hour).Unix(), // kedaluwarsa -> divalidasi otomatis
		"iat":  time.Now().Unix(),
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tok.SignedString(secret)
}

func verifikasiToken(s string) (jwt.MapClaims, error) {
	tok, err := jwt.Parse(s, func(t *jwt.Token) (any, error) {
		// WAJIB cek metode signing -> cegah serangan "alg=none"/pergantian algoritma.
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("metode signing tak terduga")
		}
		return secret, nil
	})
	if err != nil {
		return nil, err
	}
	return tok.Claims.(jwt.MapClaims), nil
}

// ===========================================================================
// 2. RBAC — akses berdasarkan PERAN (role -> daftar izin)
// ===========================================================================
// 🔍 Analogi: RBAC itu seperti SERAGAM DI RUMAH SAKIT — dokter, perawat, dan
// tamu punya akses pintu yang berbeda berdasar seragamnya saja. Sederhana dan
// mudah diaudit, tapi kasar: semua dokter dianggap sama.
var izinPeran = map[string][]string{
	"admin":  {"baca", "tulis", "hapus"},
	"editor": {"baca", "tulis"},
	"viewer": {"baca"},
}

func rbacBoleh(role, aksi string) bool {
	for _, izin := range izinPeran[role] {
		if izin == aksi {
			return true
		}
	}
	return false
}

// ===========================================================================
// 3. ABAC — akses berdasarkan ATRIBUT (kebijakan lebih kaya dari sekadar peran)
// ===========================================================================
// Contoh kebijakan: user hanya boleh mengedit dokumen MILIKNYA, atau bila ia admin.
//
// 🔍 Analogi: ABAC itu RBAC yang naik level — bukan cuma melihat seragam,
// tapi juga KONTEKSNYA: "dokter boleh membuka rekam medis HANYA milik
// pasiennya sendiri, kecuali kepala rumah sakit". Aturannya menimbang atribut
// subjek + objek, bukan sekadar jabatan.
type Subjek struct {
	User string
	Role string
}
type Dokumen struct{ Pemilik string }

func abacBolehEdit(s Subjek, d Dokumen) bool {
	return s.Role == "admin" || s.User == d.Pemilik
}

func main() {
	// 1. JWT
	fmt.Println("== 1. JWT ==")
	token, _ := terbitkanToken("budi", "editor")
	fmt.Printf("  token (potong): %s...\n", token[:32])
	claims, err := verifikasiToken(token)
	fmt.Printf("  verifikasi OK: sub=%v role=%v err=%v\n", claims["sub"], claims["role"], err)
	_, errPalsu := verifikasiToken(token + "tamper")
	fmt.Println("  token dirusak -> ditolak:", errPalsu != nil)

	// 2. RBAC
	fmt.Println("== 2. RBAC ==")
	role := claims["role"].(string)
	for _, aksi := range []string{"baca", "tulis", "hapus"} {
		fmt.Printf("  role %q boleh %q? %v\n", role, aksi, rbacBoleh(role, aksi))
	}

	// 3. ABAC
	fmt.Println("== 3. ABAC ==")
	dok := Dokumen{Pemilik: "ani"}
	fmt.Printf("  editor budi edit dok milik ani? %v\n", abacBolehEdit(Subjek{"budi", "editor"}, dok))
	fmt.Printf("  admin edit dok milik ani?        %v\n", abacBolehEdit(Subjek{"cici", "admin"}, dok))
	fmt.Printf("  ani edit dok miliknya sendiri?   %v\n", abacBolehEdit(Subjek{"ani", "editor"}, dok))

	// Catatan: JWT stateless sulit di-revoke -> pakai access pendek + refresh
	// dengan rotasi & deteksi reuse. Untuk aturan kompleks, gunakan policy engine
	// (OPA/Casbin) alih-alih if-else tersebar.
}
