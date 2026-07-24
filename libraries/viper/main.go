// spf13/viper — konfigurasi berlapis: nilai bawaan, berkas, dan variabel lingkungan.
//
// Jalankan: go run ./libraries/viper
//
//	APP_SERVER_PORT=9999 go run ./libraries/viper
//
// Test:     go test ./libraries/viper
//
// 🔍 Analogi besar: konfigurasi itu seperti ATURAN BERPAKAIAN yang berlapis-lapis.
// Ada aturan umum perusahaan (nilai bawaan di kode), ada aturan kantor cabang (berkas
// config.yaml), dan ada instruksi langsung dari bos hari ini (variabel lingkungan).
// Kalau ketiganya bicara soal hal yang sama, yang MENANG adalah yang paling dekat
// dengan situasi saat ini: instruksi bos > aturan cabang > aturan umum.
//
// Kenapa variabel lingkungan menang? Karena di produksi, kata sandi database dan kunci
// API TIDAK BOLEH ikut tersimpan di repositori. Ia disuntikkan saat aplikasi dijalankan.
// Ini prinsip 12-factor app, dan viper mewujudkannya tanpa kamu menulis logika berlapis.
package main

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

func main() {
	fmt.Println("=== spf13/viper ===")
	demoLapisan()
	demoBerkas()
	demoValidasi()
	demoJebakan()
}

// ------------------------------------------------------------------
// 1. Bentuk konfigurasi
// ------------------------------------------------------------------

// Konfig adalah seluruh pengaturan aplikasi dalam satu struct.
//
// 🔍 Analogi tag mapstructure: viper menyimpan konfigurasi sebagai map bertingkat
// ("server.port"), lalu perlu menuangkannya ke struct Go. Tag ini adalah LABEL yang
// memberi tahu "isi kunci 'port' ke field ini".
//
// Jebakan yang menghabiskan waktu banyak orang: viper memakai tag `mapstructure`,
// BUKAN `json`. Kalau kamu menulis tag json, viper akan diam-diam mengisi zero value
// tanpa error apa pun — kelihatan seperti konfigurasi tak terbaca.
type Konfig struct {
	Aplikasi string        `mapstructure:"aplikasi"`
	Debug    bool          `mapstructure:"debug"`
	Server   ServerKonfig  `mapstructure:"server"`
	DB       DBKonfig      `mapstructure:"db"`
	Timeout  time.Duration `mapstructure:"timeout"`
	Fitur    []string      `mapstructure:"fitur"`
}

type ServerKonfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

type DBKonfig struct {
	DSN         string `mapstructure:"dsn"`
	MaksKoneksi int    `mapstructure:"maks_koneksi"`
}

// ------------------------------------------------------------------
// 2. Menyiapkan viper
// ------------------------------------------------------------------

// NewViper membuat instance viper yang sudah disetel lapisan-lapisannya.
//
// Perhatikan: memakai viper.New(), BUKAN viper.SetDefault() global. Instance global
// itu variabel bersama — dua test yang berjalan berurutan akan saling mengotori.
func NewViper() *viper.Viper {
	v := viper.New()
	pasangNilaiBawaan(v)

	// 🔍 Analogi SetEnvPrefix: seperti memberi AWALAN NOMOR TELEPON kantor. Tanpa awalan,
	// variabel lingkungan bernama "PORT" milik sistem lain bisa tak sengaja terbaca
	// aplikasimu. Dengan awalan APP_, hanya variabel yang memang ditujukan untukmu
	// yang terpakai.
	v.SetEnvPrefix("APP")

	// 🔍 Analogi SetEnvKeyReplacer: kunci viper bertingkat memakai titik ("server.port"),
	// tapi nama variabel lingkungan TIDAK BOLEH mengandung titik. Penerjemah ini
	// mengubah titik jadi garis bawah, sehingga "server.port" dicari sebagai
	// APP_SERVER_PORT. Tanpa baris ini, seluruh konfigurasi bertingkat tak akan pernah
	// bisa ditimpa lewat environment — jebakan yang sangat sering menggigit.
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	v.AutomaticEnv()
	return v
}

// pasangNilaiBawaan mendaftarkan seluruh kunci beserta nilai amannya.
//
// 🔍 Analogi: ini DAFTAR ISI konfigurasi. Fungsinya dua:
//  1. Aplikasi tetap jalan walau tak ada berkas config sama sekali.
//  2. Lebih halus tapi penting: AutomaticEnv hanya menimpa kunci yang SUDAH DIKENAL
//     viper. Kunci yang tak pernah didaftarkan (lewat default atau berkas) tak akan
//     terbaca dari environment saat Unmarshal. Jadi mendaftarkan default di sini
//     sekaligus "mendaftarkan keberadaan" kunci tersebut.
func pasangNilaiBawaan(v *viper.Viper) {
	v.SetDefault("aplikasi", "go-learning")
	v.SetDefault("debug", false)
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 8080)
	v.SetDefault("db.dsn", "file::memory:")
	v.SetDefault("db.maks_koneksi", 10)
	v.SetDefault("timeout", "30s")
	v.SetDefault("fitur", []string{})
}

// MuatKonfig menuangkan isi viper ke struct lalu memvalidasinya.
//
// 🔍 Analogi validasi di sini: seperti PEMERIKSAAN SEBELUM LEPAS LANDAS. Lebih baik
// aplikasi menolak start dengan pesan jelas ("port 99999 tidak valid") daripada
// terbang lalu jatuh tiga jam kemudian saat ada pengguna yang mengakses.
func MuatKonfig(v *viper.Viper) (Konfig, error) {
	var k Konfig
	if err := v.Unmarshal(&k); err != nil {
		return Konfig{}, fmt.Errorf("gagal membaca konfigurasi: %w", err)
	}
	if err := k.Validasi(); err != nil {
		return Konfig{}, err
	}
	return k, nil
}

// ErrKonfigTidakValid sentinel agar pemanggil bisa membedakan salah konfigurasi
// dari kegagalan lain (mis. berkas rusak).
var ErrKonfigTidakValid = errors.New("konfigurasi tidak valid")

// Validasi memeriksa aturan yang tak bisa dijamin oleh tipe data saja.
func (k Konfig) Validasi() error {
	var masalah []string

	if k.Aplikasi == "" {
		masalah = append(masalah, "aplikasi tidak boleh kosong")
	}
	if k.Server.Port < 1 || k.Server.Port > 65535 {
		masalah = append(masalah, fmt.Sprintf("server.port %d di luar jangkauan 1-65535", k.Server.Port))
	}
	if k.DB.DSN == "" {
		masalah = append(masalah, "db.dsn wajib diisi")
	}
	if k.DB.MaksKoneksi < 1 {
		masalah = append(masalah, "db.maks_koneksi minimal 1")
	}
	if k.Timeout <= 0 {
		masalah = append(masalah, "timeout harus lebih besar dari nol")
	}

	if len(masalah) > 0 {
		// Laporkan SEMUA masalah sekaligus, bukan satu per satu — supaya pengguna
		// tak perlu memperbaiki, menjalankan ulang, gagal lagi, berulang kali.
		return fmt.Errorf("%w: %s", ErrKonfigTidakValid, strings.Join(masalah, "; "))
	}
	return nil
}

// Alamat merakit host:port untuk dipakai server.
func (k Konfig) Alamat() string {
	return fmt.Sprintf("%s:%d", k.Server.Host, k.Server.Port)
}

func demoLapisan() {
	fmt.Println("\n-- Lapisan konfigurasi --")

	k, err := MuatKonfig(NewViper())
	if err != nil {
		fmt.Println("   error:", err)
		return
	}
	fmt.Printf("   aplikasi : %s\n", k.Aplikasi)
	fmt.Printf("   alamat   : %s\n", k.Alamat())
	fmt.Printf("   timeout  : %s\n", k.Timeout)
	fmt.Println("   (coba: APP_SERVER_PORT=9999 go run ./libraries/viper)")
}

// ------------------------------------------------------------------
// 3. Membaca dari berkas
// ------------------------------------------------------------------

// MuatDariYAML membaca konfigurasi dari isi YAML apa pun (string, berkas, atau jaringan).
//
// 🔍 Analogi: memakai ReadConfig(io.Reader) alih-alih ReadInConfig() yang mencari berkas
// di disk membuat contoh ini bisa DIUJI tanpa membuat berkas sungguhan. Di aplikasi
// nyata kamu tetap memakai v.SetConfigName("config") + v.AddConfigPath("./configs")
// + v.ReadInConfig().
func MuatDariYAML(isi string) (Konfig, error) {
	v := NewViper()
	v.SetConfigType("yaml")

	if err := v.ReadConfig(strings.NewReader(isi)); err != nil {
		return Konfig{}, fmt.Errorf("berkas konfigurasi tidak bisa dibaca: %w", err)
	}
	return MuatKonfig(v)
}

const contohYAML = `
aplikasi: toko-online
debug: true
server:
  host: 127.0.0.1
  port: 3000
db:
  dsn: "postgres://localhost/toko"
  maks_koneksi: 25
timeout: 15s
fitur:
  - checkout-baru
  - rekomendasi
`

func demoBerkas() {
	fmt.Println("\n-- Dari berkas YAML --")

	k, err := MuatDariYAML(contohYAML)
	if err != nil {
		fmt.Println("   error:", err)
		return
	}
	fmt.Printf("   aplikasi : %s (debug=%t)\n", k.Aplikasi, k.Debug)
	fmt.Printf("   alamat   : %s\n", k.Alamat())
	fmt.Printf("   db       : %s (maks %d koneksi)\n", k.DB.DSN, k.DB.MaksKoneksi)
	fmt.Printf("   fitur    : %v\n", k.Fitur)
}

func demoValidasi() {
	fmt.Println("\n-- Gagal cepat saat konfigurasi salah --")

	rusak := `
aplikasi: ""
server:
  port: 99999
db:
  dsn: ""
  maks_koneksi: 0
timeout: 0s
`
	if _, err := MuatDariYAML(rusak); err != nil {
		fmt.Println("  ", err)
	}
}

// ------------------------------------------------------------------
// 4. Jebakan
// ------------------------------------------------------------------

// 🔍 Jebakan 1 — tag `json` DIABAIKAN, dan itu tidak selalu langsung terlihat.
//
// Saat tag mapstructure tak ada, viper jatuh ke cara cadangan: mencocokkan NAMA FIELD
// tanpa peduli huruf besar/kecil. Akibatnya jadi menyesatkan:
//
//   - Kunci satu kata seperti "aplikasi" KEBETULAN cocok dengan field Aplikasi,
//     sehingga sekilas tampak "tag json didukung". Padahal bukan.
//   - Kunci ber-garis-bawah seperti "maks_koneksi" TIDAK cocok dengan field MaksKoneksi
//     (dibandingkan sebagai "makskoneksi" vs "maks_koneksi"), jadi diam-diam bernilai nol.
//
// 🔍 Analogi: seperti kunci rumah yang kebetulan bisa membuka pintu depan tetangga karena
// modelnya sama, lalu kamu menyimpulkan "kunciku universal" — sampai suatu hari pintu
// gudang tak mau terbuka dan kamu bingung kenapa. Selalu tulis tag mapstructure secara
// eksplisit; jangan bergantung pada kebetulan.
type KonfigTagSalah struct {
	Aplikasi    string `json:"aplikasi"`     // kebetulan terisi: nama field == nama kunci
	MaksKoneksi int    `json:"maks_koneksi"` // DIAM-DIAM nol: nama field tak cocok kuncinya
}

// MuatDenganTagSalah memperagakan kegagalan diam-diam tersebut.
func MuatDenganTagSalah(isi string) (KonfigTagSalah, error) {
	v := viper.New()
	v.SetConfigType("yaml")
	if err := v.ReadConfig(strings.NewReader(isi)); err != nil {
		return KonfigTagSalah{}, err
	}
	var k KonfigTagSalah
	err := v.Unmarshal(&k)
	return k, err
}

// 🔍 Jebakan 2 — viper TIDAK peka huruf besar/kecil pada kunci, tapi variabel lingkungan
// di Unix PEKA. "APP_SERVER_PORT" bekerja; "app_server_port" tidak.

// 🔍 Jebakan 3 — viper bukan barang aman-goroutine untuk penulisan. Bacalah konfigurasi
// SEKALI saat aplikasi mulai, tuangkan ke struct, lalu edarkan struct itu. Jangan
// memanggil v.GetString() dari dalam handler HTTP yang berjalan bersamaan.

func demoJebakan() {
	fmt.Println("\n-- Jebakan: tag mapstructure vs json --")

	const isi = "aplikasi: toko-online\ndb:\n  maks_koneksi: 25\n"

	k, err := MuatDenganTagSalah("aplikasi: toko-online\nmaks_koneksi: 25\n")
	if err != nil {
		fmt.Println("   error:", err)
		return
	}
	fmt.Printf("   tag json -> Aplikasi=%q (terisi, KEBETULAN nama field cocok)\n", k.Aplikasi)
	fmt.Printf("   tag json -> MaksKoneksi=%d (nol diam-diam, tanpa error!)\n", k.MaksKoneksi)

	benar, err := MuatDariYAML(isi)
	if err != nil {
		fmt.Println("   error:", err)
		return
	}
	fmt.Printf("   mapstructure -> Aplikasi=%q MaksKoneksi=%d\n",
		benar.Aplikasi, benar.DB.MaksKoneksi)
}
