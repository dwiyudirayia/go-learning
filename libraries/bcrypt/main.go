// golang.org/x/crypto/bcrypt — menyimpan kata sandi dengan aman (hash searah).
//
// Jalankan: go run ./libraries/bcrypt
// Test:     go test ./libraries/bcrypt
//
// 🔍 Analogi besar: menyimpan kata sandi itu seperti mengurus daging di restoran.
// Menyimpan kata sandi APA ADANYA sama dengan menaruh daging mentah berlabel nama tamu
// di etalase — siapa pun yang membobol lemari langsung dapat semuanya. bcrypt MENGGILING
// daging itu jadi bakso: dari daging kamu bisa membuat bakso, tapi dari bakso kamu TAK
// BISA mengembalikannya jadi daging. Saat tamu datang lagi, kamu giling ulang daging yang
// ia bawa dan bandingkan baksonya — cocok berarti orangnya benar.
//
// Tiga sifat yang membuat bcrypt tepat untuk kata sandi, dan kenapa SHA-256/MD5 SALAH:
//  1. SEARAH  — hash tak bisa dibalik jadi kata sandi asli.
//  2. BERGARAM — tiap hash memakai "garam" acak, jadi dua orang berkata-sandi sama
//     menghasilkan hash BERBEDA. Ini mematikan serangan "rainbow table".
//  3. LAMBAT DENGAN SENGAJA — SHA-256 dirancang cepat (miliaran/detik), justru itu
//     buruk untuk kata sandi: penyerang bisa menebak miliaran/detik. bcrypt dibuat
//     lambat & bisa disetel makin lambat seiring komputer makin cepat (parameter "cost").
package main

import (
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	fmt.Println("=== golang.org/x/crypto/bcrypt ===")
	demoDasar()
	demoGaram()
	demoCost()
	demoRehash()
	demoBatas72()
}

// ------------------------------------------------------------------
// 1. Hash & verifikasi
// ------------------------------------------------------------------

// ErrKataSandiSalah dan ErrKataSandiTerlaluPanjang sentinel milik aplikasi,
// supaya pemanggil tak perlu tahu error internal bcrypt.
var (
	ErrKataSandiSalah          = errors.New("kata sandi salah")
	ErrKataSandiTerlaluPanjang = errors.New("kata sandi melebihi 72 byte")
	ErrKataSandiKosong         = errors.New("kata sandi tidak boleh kosong")
)

// HashKataSandi mengubah kata sandi jadi hash yang aman disimpan di database.
//
// 🔍 Analogi cost: angka cost adalah "berapa kali daging digiling ulang". Tiap kenaikan
// satu angka MELIPATGANDAKAN waktu (cost 10 = 2x lebih lama dari cost 9). DefaultCost=10
// wajar untuk kebanyakan aplikasi; naikkan bila servermu kuat & kamu ingin lebih tahan
// serangan. Terlalu tinggi = login jadi lambat bagi pengguna asli.
func HashKataSandi(kataSandi string) (string, error) {
	if kataSandi == "" {
		return "", ErrKataSandiKosong
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(kataSandi), bcrypt.DefaultCost)
	if err != nil {
		if errors.Is(err, bcrypt.ErrPasswordTooLong) {
			return "", ErrKataSandiTerlaluPanjang
		}
		return "", fmt.Errorf("gagal membuat hash: %w", err)
	}
	return string(hash), nil
}

// PeriksaKataSandi membandingkan kata sandi dengan hash tersimpan.
//
// 🔍 Analogi: perhatikan kamu TAK PERNAH "membuka" hash untuk membacanya. Kamu menggiling
// ulang kata sandi yang baru dimasukkan lalu membandingkan hasilnya — inilah kenapa bahkan
// admin database pun tak bisa melihat kata sandi asli pengguna. Kalau ada layanan yang bisa
// MENGIRIM ULANG kata sandimu lewat email, itu tanda bahaya: berarti mereka menyimpannya
// tanpa hash.
func PeriksaKataSandi(hash, kataSandi string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(kataSandi))
	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return ErrKataSandiSalah
	}
	if err != nil {
		return fmt.Errorf("gagal memeriksa kata sandi: %w", err)
	}
	return nil
}

func demoDasar() {
	fmt.Println("\n-- Hash & verifikasi --")

	hash, err := HashKataSandi("rahasia123")
	if err != nil {
		fmt.Println("   error:", err)
		return
	}
	fmt.Printf("   hash: %s\n", hash)

	if err := PeriksaKataSandi(hash, "rahasia123"); err == nil {
		fmt.Println("   kata sandi benar -> diterima ✅")
	}
	if err := PeriksaKataSandi(hash, "salah"); errors.Is(err, ErrKataSandiSalah) {
		fmt.Println("   kata sandi salah -> ditolak")
	}
}

// ------------------------------------------------------------------
// 2. Garam otomatis: kata sandi sama, hash berbeda
// ------------------------------------------------------------------

// 🔍 Analogi garam (salt): garam itu bumbu acak yang dicampurkan SEBELUM menggiling, dan
// resepnya (garamnya) ikut tertulis di dalam hasil hash. Karena tiap orang dapat garam
// berbeda, dua orang dengan kata sandi "123456" tetap menghasilkan hash yang tak mirip
// sama sekali. Penyerang tak bisa lagi membuat satu tabel tebakan raksasa (rainbow table)
// dan memakainya ke semua orang — ia harus menyerang tiap hash satu per satu.
//
// bcrypt mengurus garam OTOMATIS. Kamu tak perlu (dan tak boleh) mengelolanya sendiri.

func demoGaram() {
	fmt.Println("\n-- Garam otomatis --")

	h1, _ := HashKataSandi("samasekali")
	h2, _ := HashKataSandi("samasekali")

	fmt.Printf("   hash 1: %s\n", h1[:40])
	fmt.Printf("   hash 2: %s\n", h2[:40])
	fmt.Printf("   kata sandi sama, hash berbeda? %t\n", h1 != h2)
	fmt.Println("   -> keduanya tetap lolos verifikasi karena garam ada di dalam hash")
}

// ------------------------------------------------------------------
// 3. Cost: mengatur seberapa lambat
// ------------------------------------------------------------------

// HashDenganCost membuat hash dengan cost tertentu (untuk peragaan).
func HashDenganCost(kataSandi string, cost int) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(kataSandi), cost)
	if err != nil {
		return "", fmt.Errorf("gagal membuat hash (cost %d): %w", cost, err)
	}
	return string(hash), nil
}

// CostDariHash membaca cost yang tersimpan di dalam sebuah hash.
//
// 🔍 Analogi: cost ikut "tertulis di label bakso". Jadi kamu selalu bisa memeriksa hash
// lama dibuat dengan cost berapa — berguna untuk memutuskan apakah perlu di-hash ulang.
func CostDariHash(hash string) (int, error) {
	c, err := bcrypt.Cost([]byte(hash))
	if err != nil {
		return 0, fmt.Errorf("gagal membaca cost: %w", err)
	}
	return c, nil
}

func demoCost() {
	fmt.Println("\n-- Cost --")

	for _, cost := range []int{bcrypt.MinCost, bcrypt.DefaultCost, 12} {
		hash, err := HashDenganCost("uji", cost)
		if err != nil {
			fmt.Println("   error:", err)
			continue
		}
		terbaca, _ := CostDariHash(hash)
		fmt.Printf("   cost diminta %2d -> tersimpan di hash: %d\n", cost, terbaca)
	}
	fmt.Printf("   (rentang sah: %d..%d, default %d)\n",
		bcrypt.MinCost, bcrypt.MaxCost, bcrypt.DefaultCost)
}

// ------------------------------------------------------------------
// 4. Rehash saat cost dinaikkan
// ------------------------------------------------------------------

// 🔍 Analogi: seiring waktu, komputer makin cepat, jadi cost yang dulu aman jadi kurang
// aman. Kamu tak bisa mengubah hash lama tanpa kata sandi aslinya — tapi kamu PUNYA kata
// sandi asli tepat pada saat pengguna login. Itulah momen sempurna untuk diam-diam
// meng-hash ulang dengan cost lebih tinggi: pengguna tak merasakan apa-apa, keamanannya
// naik. Polanya: verifikasi -> kalau cost lama < target, hash ulang & simpan yang baru.

// PerluRehash memutuskan apakah hash perlu diperbarui ke cost target.
func PerluRehash(hash string, costTarget int) (bool, error) {
	c, err := CostDariHash(hash)
	if err != nil {
		return false, err
	}
	return c < costTarget, nil
}

// LoginDanMungkinRehash memverifikasi lalu meng-hash ulang bila cost-nya sudah usang.
// Mengembalikan hash baru (kosong bila tak perlu diganti).
func LoginDanMungkinRehash(hashLama, kataSandi string, costTarget int) (hashBaru string, err error) {
	if err := PeriksaKataSandi(hashLama, kataSandi); err != nil {
		return "", err
	}

	perlu, err := PerluRehash(hashLama, costTarget)
	if err != nil || !perlu {
		return "", err
	}

	baru, err := HashDenganCost(kataSandi, costTarget)
	if err != nil {
		return "", err
	}
	return baru, nil
}

func demoRehash() {
	fmt.Println("\n-- Rehash saat login --")

	// Hash lama dibuat dengan cost rendah (server jadul).
	lama, _ := HashDenganCost("rahasia123", bcrypt.MinCost)
	fmt.Printf("   hash lama cost %d\n", bcrypt.MinCost)

	baru, err := LoginDanMungkinRehash(lama, "rahasia123", 12)
	if err != nil {
		fmt.Println("   error:", err)
		return
	}
	if baru != "" {
		c, _ := CostDariHash(baru)
		fmt.Printf("   login sukses -> hash diperbarui ke cost %d (pengguna tak sadar)\n", c)
	}
}

// ------------------------------------------------------------------
// 5. Jebakan: batas 72 byte
// ------------------------------------------------------------------

// 🔍 Analogi: penggiling bcrypt punya corong yang hanya muat 72 byte. Versi lama diam-diam
// MEMBUANG sisanya — sehingga "kata sandi 100 huruf" dan "72 huruf pertamanya" dianggap
// sama, jebakan keamanan yang halus. Versi x/crypto sekarang lebih tegas: ia MENOLAK dengan
// error, bukan memotong diam-diam.
//
// Praktik yang lazim bila ingin mendukung kata sandi/frasa sangat panjang: lewatkan dulu
// melalui SHA-256 (hasilnya selalu muat), baru masukkan ke bcrypt. Untuk aplikasi biasa,
// cukup tolak input yang kelewat panjang seperti di sini.

func demoBatas72() {
	fmt.Println("\n-- Jebakan batas 72 byte --")

	if _, err := HashKataSandi(strings.Repeat("a", 72)); err == nil {
		fmt.Println("   72 byte -> masih diterima")
	}
	if _, err := HashKataSandi(strings.Repeat("a", 73)); errors.Is(err, ErrKataSandiTerlaluPanjang) {
		fmt.Println("   73 byte -> DITOLAK (bukan dipotong diam-diam)")
	}
}
