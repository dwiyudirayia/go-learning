// caarlos0/env — memuat konfigurasi dari environment ke struct lewat tag. Minimalis.
//
// Jalankan: go run ./libraries/env
//
//	APP_PORT=9999 DEBUG=true go run ./libraries/env
//
// Test:     go test ./libraries/env
//
// 🔍 Analogi besar: kalau Viper (libraries/viper) itu DAPUR LENGKAP yang bisa membaca dari
// file YAML, JSON, env, flag, bahkan remote — maka caarlos0/env itu PISAU LIPAT: satu
// pekerjaan, dikerjakan dengan sangat rapi. Ia HANYA membaca environment, langsung ke
// struct, lewat tag. Tak ada file, tak ada sumber lain.
//
// Kapan pilih yang mana:
//   - env    : aplikasi 12-factor / cloud-native yang memang HANYA dikonfigurasi lewat
//     environment (kontainer, Kubernetes). Ringan, nol kerumitan.
//   - Viper  : butuh berlapis (file default + override env + flag), hot-reload, banyak format.
//
// Filosofi 12-factor (dan alasan pendekatan ini populer di era kontainer): konfigurasi
// hidup di ENVIRONMENT, bukan di file yang ikut ter-commit. Kata sandi & kunci API
// disuntikkan saat runtime, tak pernah tersimpan di repositori.
package main

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/caarlos0/env/v11"
)

func main() {
	fmt.Println("=== caarlos0/env ===")
	demoDasar()
	demoTipe()
	demoWajib()
	demoValidasi()
}

// ------------------------------------------------------------------
// 1. Struct konfigurasi
// ------------------------------------------------------------------

// Konfig — seluruh pengaturan aplikasi. Perhatikan tag `env`.
//
// 🔍 Analogi tag env: tiap tag adalah ALAMAT di papan environment.
//
//	`env:"PORT"`                    -> baca variabel PORT.
//	`envDefault:"8080"`             -> kalau tak diset, pakai ini (aplikasi tetap jalan).
//	`env:"...,required"`            -> WAJIB diisi; kalau tak ada, Parse gagal (fail fast).
//	`envPrefix:"DB_"`               -> untuk struct bersarang: semua di dalamnya berawalan DB_.
//	`envSeparator:","`             -> untuk slice: pemisah antar-nilai.
type Konfig struct {
	Nama    string        `env:"APP_NAME" envDefault:"go-learning"`
	Port    int           `env:"APP_PORT" envDefault:"8080"`
	Debug   bool          `env:"DEBUG" envDefault:"false"`
	Timeout time.Duration `env:"TIMEOUT" envDefault:"30s"`
	Fitur   []string      `env:"FITUR" envSeparator:"," envDefault:""`

	// Struct bersarang dengan awalan sendiri: DB_DSN, DB_MAX_CONN.
	DB DBKonfig `envPrefix:"DB_"`
}

type DBKonfig struct {
	DSN     string `env:"DSN" envDefault:"file::memory:"`
	MaxConn int    `env:"MAX_CONN" envDefault:"10"`
}

// ErrKonfigTidakValid sentinel agar pemanggil bisa membedakan salah-konfigurasi.
var ErrKonfigTidakValid = errors.New("konfigurasi tidak valid")

// Muat mengurai environment ke Konfig lalu memvalidasinya.
//
// 🔍 Analogi validasi setelah parse: env memastikan TIPE-nya benar (PORT harus angka),
// tapi tak tahu ATURAN BISNIS-mu (port harus 1..65535). Jadi: parse dulu (env), lalu
// periksa akal sehat (Validasi). Sama seperti "pemeriksaan sebelum lepas landas" di
// libraries/viper — lebih baik menolak start dengan pesan jelas daripada crash nanti.
func Muat() (Konfig, error) {
	var k Konfig
	if err := env.Parse(&k); err != nil {
		return Konfig{}, fmt.Errorf("gagal membaca environment: %w", err)
	}
	if err := k.Validasi(); err != nil {
		return Konfig{}, err
	}
	return k, nil
}

// MuatDari mengurai memakai sekumpulan variabel yang diberikan (untuk test yang bersih).
//
// 🔍 Analogi: alih-alih mengutak-atik environment proses sungguhan (yang bocor antar-test),
// env.ParseWithOptions menerima "environment palsu" berupa map. Test jadi terisolasi &
// tak saling mengganggu.
func MuatDari(vars map[string]string) (Konfig, error) {
	var k Konfig
	if err := env.ParseWithOptions(&k, env.Options{Environment: vars}); err != nil {
		return Konfig{}, fmt.Errorf("gagal membaca environment: %w", err)
	}
	if err := k.Validasi(); err != nil {
		return Konfig{}, err
	}
	return k, nil
}

// Validasi memeriksa aturan yang tak bisa dijamin tipe data saja.
func (k Konfig) Validasi() error {
	var masalah []string
	if k.Nama == "" {
		masalah = append(masalah, "APP_NAME tidak boleh kosong")
	}
	if k.Port < 1 || k.Port > 65535 {
		masalah = append(masalah, fmt.Sprintf("APP_PORT %d di luar 1..65535", k.Port))
	}
	if k.DB.MaxConn < 1 {
		masalah = append(masalah, "DB_MAX_CONN minimal 1")
	}
	if k.Timeout <= 0 {
		masalah = append(masalah, "TIMEOUT harus > 0")
	}
	if len(masalah) > 0 {
		return fmt.Errorf("%w: %s", ErrKonfigTidakValid, strings.Join(masalah, "; "))
	}
	return nil
}

func demoDasar() {
	fmt.Println("\n-- Memuat konfigurasi --")

	k, err := Muat()
	if err != nil {
		fmt.Println("   error:", err)
		return
	}
	fmt.Printf("   nama    : %s\n", k.Nama)
	fmt.Printf("   port    : %d\n", k.Port)
	fmt.Printf("   debug   : %t\n", k.Debug)
	fmt.Printf("   db.dsn  : %s (max %d koneksi)\n", k.DB.DSN, k.DB.MaxConn)
	fmt.Println("   (coba: APP_PORT=9999 DEBUG=true go run ./libraries/env)")
}

// ------------------------------------------------------------------
// 2. Konversi tipe otomatis
// ------------------------------------------------------------------

// 🔍 Analogi: environment SELALU berupa teks ("8080", "true", "30s"). env menerjemahkannya
// ke tipe Go yang tepat — int, bool, time.Duration, bahkan slice — persis seperti PETUGAS
// PABEAN yang membaca formulir teks lalu mengisinya ke sistem sesuai jenis datanya.
// Kalau teksnya tak bisa jadi tipe yang diminta (PORT="abc"), Parse GAGAL dengan jelas.

func demoTipe() {
	fmt.Println("\n-- Konversi tipe --")

	k, err := MuatDari(map[string]string{
		"APP_PORT": "3000",
		"DEBUG":    "true",
		"TIMEOUT":  "1m30s",
		"FITUR":    "checkout,rekomendasi,chat",
	})
	if err != nil {
		fmt.Println("   error:", err)
		return
	}
	fmt.Printf("   port (int)      : %d\n", k.Port)
	fmt.Printf("   debug (bool)    : %t\n", k.Debug)
	fmt.Printf("   timeout (durasi): %s\n", k.Timeout)
	fmt.Printf("   fitur (slice)   : %v (%d item)\n", k.Fitur, len(k.Fitur))
}

// ------------------------------------------------------------------
// 3. Field wajib (required)
// ------------------------------------------------------------------

// KonfigRahasia memperagakan field yang WAJIB ada & tak boleh kosong.
//
// 🔍 Analogi & BEDA HALUS yang penting (required vs notEmpty):
//
//	`required` = variabel harus HADIR di environment, TAPI boleh bernilai kosong.
//	             (SECRET_KEY="" lolos required — sering mengejutkan!)
//	`notEmpty` = harus hadir DAN tidak kosong. Ini yang benar untuk rahasia:
//	             kunci kosong sama bahayanya dengan tak ada kunci.
//
// Aturan main: kata sandi/kunci API TIDAK boleh punya default yang "aman". Lebih baik
// aplikasi MENOLAK START daripada diam-diam jalan dengan kunci kosong. notEmpty
// mewujudkan sikap fail-fast itu sepenuhnya.
type KonfigRahasia struct {
	SecretKey string `env:"SECRET_KEY,notEmpty"`
}

// MuatRahasia mengurai KonfigRahasia dari environment yang diberikan.
func MuatRahasia(vars map[string]string) (KonfigRahasia, error) {
	var k KonfigRahasia
	if err := env.ParseWithOptions(&k, env.Options{Environment: vars}); err != nil {
		return KonfigRahasia{}, fmt.Errorf("konfigurasi rahasia gagal: %w", err)
	}
	return k, nil
}

func demoWajib() {
	fmt.Println("\n-- Field wajib (notEmpty) --")

	if k, err := MuatRahasia(map[string]string{"SECRET_KEY": "s3raha51a"}); err == nil {
		fmt.Printf("   dengan SECRET_KEY -> dimuat (panjang %d)\n", len(k.SecretKey))
	}
	if _, err := MuatRahasia(map[string]string{}); err != nil {
		fmt.Println("   tanpa SECRET_KEY  -> DITOLAK (fail fast, aplikasi tak start)")
	}
	if _, err := MuatRahasia(map[string]string{"SECRET_KEY": ""}); err != nil {
		fmt.Println("   SECRET_KEY kosong -> DITOLAK juga (beda dari 'required' biasa)")
	}
}

// ------------------------------------------------------------------
// 4. Validasi aturan bisnis
// ------------------------------------------------------------------

func demoValidasi() {
	fmt.Println("\n-- Validasi aturan bisnis --")

	// Tipe benar (angka), tapi nilainya di luar akal (port 99999).
	if _, err := MuatDari(map[string]string{"APP_PORT": "99999"}); err != nil {
		fmt.Println("  ", err)
	}
}
