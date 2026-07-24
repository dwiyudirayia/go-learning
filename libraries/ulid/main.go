// oklog/ulid — ID unik yang URUT WAKTU dan ramah dibaca manusia.
//
// Jalankan: go run ./libraries/ulid
// Test:     go test ./libraries/ulid
//
// 🔍 Analogi besar: ULID itu seperti NOMOR TIKET BIOSKOP model baru — bagian depannya
// stempel waktu ("dibeli jam 19:42"), bagian belakangnya kode acak biar tak bentrok.
// Karena depannya waktu, tiket yang dibeli belakangan selalu bernomor "lebih besar",
// jadi tumpukan tiket otomatis tersusun urut waktu tanpa perlu disortir.
//
// ULID vs UUID (lihat juga libraries/uuid):
//   - Keduanya ID unik tanpa koordinasi database.
//   - ULID **urut waktu** (seperti UUIDv7) DAN dikodekan Base32 tanpa tanda hubung —
//     lebih pendek (26 huruf vs 36), aman untuk URL, dan enak di-klik dua kali (tak
//     terputus di tanda hubung).
//   - UUID lebih universal & didukung native banyak database (tipe kolom UUID).
//
// Pilih ULID bila kamu suka ID urut-waktu yang ringkas & ramah URL; UUIDv7 bila ekosistemmu
// sudah UUID. Keduanya jauh lebih baik daripada auto-increment untuk sistem terdistribusi.
package main

import (
	"crypto/rand"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/oklog/ulid/v2"
)

func main() {
	fmt.Println("=== oklog/ulid ===")
	demoDasar()
	demoUrutWaktu()
	demoWaktuTerbaca()
	demoParse()
	demoMonotonic()
}

// ------------------------------------------------------------------
// 1. Membuat ULID
// ------------------------------------------------------------------

// Baru membuat ULID dengan waktu sekarang & keacakan kriptografis.
//
// 🔍 Analogi: ulid.Make() adalah jalan pintas paling umum — ia memakai waktu sekarang dan
// sumber acak default. Untuk kode yang bisa diuji (waktu bisa dikendalikan), kita pakai
// Baru(waktu) di bawah yang menyuntikkan waktunya.
func Baru() ulid.ULID {
	return ulid.Make()
}

// BaruPadaWaktu membuat ULID untuk waktu tertentu — dipakai test agar hasilnya pasti.
//
// 🔍 Analogi menyuntikkan waktu: sama seperti jwt & ratelimit di folder ini — memberi
// "jam" dari luar membuat kita bisa "melompat ke besok" di test tanpa menunggu.
func BaruPadaWaktu(t time.Time) ulid.ULID {
	return ulid.MustNew(ulid.Timestamp(t), rand.Reader)
}

func demoDasar() {
	fmt.Println("\n-- Membuat ULID --")
	for range 3 {
		fmt.Println("  ", Baru())
	}
	fmt.Println("   (26 karakter, Base32, tanpa tanda hubung, aman-URL)")
}

// ------------------------------------------------------------------
// 2. Sifat utama: urut waktu secara leksikografis
// ------------------------------------------------------------------

// 🔍 Analogi: karena stempel waktu ada di DEPAN, mengurutkan ULID sebagai TEKS BIASA
// (alfabetis) otomatis mengurutkannya berdasarkan WAKTU. Ini alasan ULID (seperti UUIDv7)
// ramah untuk primary key: baris baru selalu masuk di "ujung" index database, tak
// menyelip acak di tengah — sama seperti dijelaskan di libraries/uuid.

// UrutMenaik mengurutkan daftar ULID sebagai string.
func UrutMenaik(ids []ulid.ULID) []string {
	teks := make([]string, len(ids))
	for i, id := range ids {
		teks[i] = id.String()
	}
	sort.Strings(teks)
	return teks
}

// BuatBerurutan membuat n ULID pada waktu yang menaik (1 detik terpisah).
func BuatBerurutan(mulai time.Time, n int) []ulid.ULID {
	out := make([]ulid.ULID, n)
	for i := range out {
		out[i] = BaruPadaWaktu(mulai.Add(time.Duration(i) * time.Second))
	}
	return out
}

func demoUrutWaktu() {
	fmt.Println("\n-- Urut waktu --")
	mulai := time.Date(2026, 7, 24, 10, 0, 0, 0, time.UTC)
	ids := BuatBerurutan(mulai, 5)

	for i, id := range ids {
		fmt.Printf("   +%ddetik  %s\n", i, id)
	}

	// Bukti: urutkan sebagai teks -> tetap sesuai urutan pembuatan.
	terurut := UrutMenaik(ids)
	cocok := true
	for i, id := range ids {
		if id.String() != terurut[i] {
			cocok = false
			break
		}
	}
	fmt.Printf("   urutan teks == urutan waktu? %t\n", cocok)
}

// ------------------------------------------------------------------
// 3. Membaca waktu kembali dari ULID
// ------------------------------------------------------------------

// 🔍 Analogi: karena stempel waktu tertanam di dalam ID, kamu bisa "membaca tanggal beli"
// langsung dari nomor tiket — tanpa kolom created_at terpisah. Berguna untuk debugging
// ("ID ini dibuat kapan?") tanpa menyentuh database.

// WaktuDari mengekstrak waktu pembuatan dari sebuah ULID.
func WaktuDari(id ulid.ULID) time.Time {
	return time.UnixMilli(int64(id.Time())).UTC()
}

func demoWaktuTerbaca() {
	fmt.Println("\n-- Membaca waktu dari ULID --")
	asli := time.Date(2026, 7, 24, 15, 30, 0, 0, time.UTC)
	id := BaruPadaWaktu(asli)

	fmt.Printf("   ULID   : %s\n", id)
	fmt.Printf("   waktu  : %s\n", WaktuDari(id).Format("2006-01-02 15:04:05"))
	fmt.Printf("   cocok dengan waktu asli? %t\n", WaktuDari(id).Equal(asli.Truncate(time.Millisecond)))
}

// ------------------------------------------------------------------
// 4. Parse & validasi
// ------------------------------------------------------------------

// ErrULIDTidakValid sentinel milik aplikasi.
var ErrULIDTidakValid = errors.New("ulid tidak valid")

// ParseULID mengubah string jadi ULID dengan error yang bersih.
//
// 🔍 Analogi & JEBAKAN penting (Parse vs ParseStrict): ada DUA pintu masuk.
//
//	ulid.Parse       = satpam longgar. Demi kecepatan, ia TIDAK memeriksa apakah tiap
//	                   karakter sah — karakter aneh (I, U, "!", tanda hubung) diloloskan
//	                   begitu saja & menghasilkan ULID yang "salah diam-diam". Ia hanya
//	                   menolak panjang yang salah & nilai yang overflow.
//	ulid.ParseStrict = satpam teliti. Memeriksa SETIAP karakter harus Base32 Crockford
//	                   yang sah (tanpa I/L/O/U agar tak keliru dengan 1/0).
//
// Untuk input dari LUAR (pengguna, URL, jaringan), SELALU pakai ParseStrict — sama seperti
// gjson.Valid di libraries/gjson. Kita pakai ParseStrict di sini.
func ParseULID(s string) (ulid.ULID, error) {
	id, err := ulid.ParseStrict(s)
	if err != nil {
		return ulid.ULID{}, fmt.Errorf("%w: %q", ErrULIDTidakValid, s)
	}
	return id, nil
}

// Valid cek cepat (ketat) tanpa mengembalikan hasil parse.
func Valid(s string) bool {
	_, err := ulid.ParseStrict(s)
	return err == nil
}

func demoParse() {
	fmt.Println("\n-- Parse & validasi --")

	asli := Baru()
	kembali, err := ParseULID(asli.String())
	if err != nil {
		fmt.Println("   error:", err)
		return
	}
	fmt.Printf("   pulang-pergi utuh? %t\n", asli == kembali)

	for _, s := range []string{
		"01ARZ3NDEKTSV4RRFFQ69G5FAV", // valid
		"terlalu-pendek",             // tidak valid
		"",                           // kosong
	} {
		fmt.Printf("   %-30q valid? %t\n", s, Valid(s))
	}
}

// ------------------------------------------------------------------
// 5. Jebakan: ID dalam milidetik yang sama
// ------------------------------------------------------------------

// 🔍 Analogi & JEBAKAN: stempel waktu ULID beresolusi MILIDETIK. Kalau kamu membuat 1000
// ID dalam satu milidetik yang sama (mudah terjadi di loop cepat), bagian waktunya SAMA —
// urutannya lalu ditentukan bagian acak, yang TIDAK dijamin naik. Artinya dua ULID di
// milidetik sama bisa "keluar dari urutan".
//
// Solusinya: MonotonicEntropy. Ia menjamin bahwa dalam milidetik yang sama, tiap ID baru
// selalu > ID sebelumnya (bagian acak dinaikkan, bukan diacak ulang). Untuk primary key
// yang harus benar-benar urut, pakai ini.

// PembuatMonotonic membungkus sumber monotonic agar aman dipakai berulang.
type PembuatMonotonic struct {
	entropy *ulid.MonotonicEntropy
}

// NewPembuatMonotonic membuat pembuat ULID monotonic.
func NewPembuatMonotonic() *PembuatMonotonic {
	return &PembuatMonotonic{entropy: ulid.Monotonic(rand.Reader, 0)}
}

// Baru membuat ULID monotonic pada waktu tertentu.
func (p *PembuatMonotonic) Baru(t time.Time) (ulid.ULID, error) {
	id, err := ulid.New(ulid.Timestamp(t), p.entropy)
	if err != nil {
		return ulid.ULID{}, fmt.Errorf("gagal membuat ulid monotonic: %w", err)
	}
	return id, nil
}

func demoMonotonic() {
	fmt.Println("\n-- Monotonic (urut dalam milidetik sama) --")

	p := NewPembuatMonotonic()
	sama := time.Date(2026, 7, 24, 10, 0, 0, 0, time.UTC) // waktu identik untuk semua

	var ids []ulid.ULID
	for range 5 {
		id, err := p.Baru(sama)
		if err != nil {
			fmt.Println("   error:", err)
			return
		}
		ids = append(ids, id)
	}

	naik := true
	for i := 1; i < len(ids); i++ {
		if ids[i].Compare(ids[i-1]) <= 0 {
			naik = false
			break
		}
	}
	for _, id := range ids {
		fmt.Println("  ", id)
	}
	fmt.Printf("   5 ID di milidetik sama, tetap urut naik? %t\n", naik)
}
