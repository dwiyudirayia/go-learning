// google/uuid — membuat ID unik global tanpa perlu bertanya ke database.
//
// Jalankan: go run ./libraries/uuid
// Test:     go test ./libraries/uuid
//
// 🔍 Analogi besar: UUID itu seperti NOMOR RANGKA KENDARAAN. Pabrik di Jepang dan pabrik
// di Jerman bisa mencetak nomor rangka masing-masing TANPA saling menelepon, dan nomornya
// tetap dijamin tak pernah bentrok di seluruh dunia. Bandingkan dengan nomor antrian bank
// (AUTO_INCREMENT di database): satu mesin harus jadi wasit tunggal yang membagikan nomor.
// Di sistem terdistribusi, wasit tunggal itu jadi leher botol — UUID menghapus kebutuhan itu.
package main

import (
	"errors"
	"fmt"
	"sort"

	"github.com/google/uuid"
)

func main() {
	fmt.Println("=== google/uuid ===")
	demoV4()
	demoV7()
	demoParse()
	demoJebakan()
}

// ------------------------------------------------------------------
// 1. UUID v4 — acak murni
// ------------------------------------------------------------------

// 🔍 Analogi v4: seperti mengambil nomor undian dari drum berisi 2^122 bola.
// Peluang dua orang mengambil bola yang sama praktis nol (kamu lebih mungkin
// tertimpa meteor). Karena murni acak, urutannya TIDAK bermakna apa-apa.

// NewV4 membuat UUID acak. uuid.New() akan panic bila sumber acak sistem rusak —
// aman untuk kebanyakan aplikasi, tapi versi ber-error ada di NewV4Safe.
func NewV4() string {
	return uuid.New().String()
}

// NewV4Safe versi yang mengembalikan error alih-alih panic.
// Dipakai bila kamu tak mau proses mati gara-gara sumber entropi bermasalah.
func NewV4Safe() (string, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return "", fmt.Errorf("gagal membuat uuid v4: %w", err)
	}
	return id.String(), nil
}

func demoV4() {
	fmt.Println("\n-- v4 (acak) --")
	for i := 0; i < 3; i++ {
		fmt.Println("  ", NewV4())
	}
	id, err := NewV4Safe()
	if err != nil {
		fmt.Println("   error:", err)
		return
	}
	fmt.Println("   versi aman ->", id)
}

// ------------------------------------------------------------------
// 2. UUID v7 — acak TAPI urut waktu
// ------------------------------------------------------------------

// 🔍 Analogi v7: seperti nomor rangka yang 48 bit pertamanya adalah STEMPEL WAKTU produksi,
// sisanya baru acak. Hasilnya: ID yang dibuat belakangan selalu "lebih besar" daripada yang
// dibuat lebih dulu. Kenapa penting? Bayangkan lemari arsip yang disusun alfabetis. Kalau
// map baru selalu diselipkan ACAK di tengah-tengah (v4), petugas harus terus menggeser-geser
// isi laci — itulah yang dialami index B-tree database. Kalau map baru selalu ditaruh di
// UJUNG (v7), tak ada yang perlu digeser. Inilah alasan v7 jauh lebih ramah sebagai primary key.

// NewV7 membuat UUID v7 (berbasis waktu). Sejak Go module uuid v1.4+.
func NewV7() (string, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return "", fmt.Errorf("gagal membuat uuid v7: %w", err)
	}
	return id.String(), nil
}

// GenerateV7Batch membuat n buah UUID v7 berurutan.
// Dipakai di test untuk membuktikan sifat "urut waktu"-nya.
func GenerateV7Batch(n int) ([]string, error) {
	ids := make([]string, 0, n)
	for i := 0; i < n; i++ {
		id, err := NewV7()
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

// IsSorted memeriksa apakah daftar string sudah urut menaik.
// Untuk v7 hasilnya true; untuk v4 hampir pasti false.
func IsSorted(ids []string) bool {
	return sort.StringsAreSorted(ids)
}

func demoV7() {
	fmt.Println("\n-- v7 (urut waktu) --")
	v7s, err := GenerateV7Batch(5)
	if err != nil {
		fmt.Println("   error:", err)
		return
	}
	for _, id := range v7s {
		fmt.Println("  ", id)
	}
	fmt.Printf("   v7 urut menaik? %t  <- ramah index database\n", IsSorted(v7s))

	v4s := []string{NewV4(), NewV4(), NewV4(), NewV4(), NewV4()}
	fmt.Printf("   v4 urut menaik? %t  <- acak, bikin index 'loncat-loncat'\n", IsSorted(v4s))
}

// ------------------------------------------------------------------
// 3. Parse & validasi
// ------------------------------------------------------------------

// ErrUUIDTidakValid sentinel error agar pemanggil bisa memakai errors.Is.
var ErrUUIDTidakValid = errors.New("uuid tidak valid")

// ParseUUID mengubah string jadi uuid.UUID, membungkus kegagalan dengan sentinel.
//
// 🔍 Analogi: seperti petugas yang memeriksa nomor rangka di STNK — bentuknya harus
// tepat 36 karakter dengan tanda hubung di posisi yang benar, kalau tidak: ditolak.
func ParseUUID(s string) (uuid.UUID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil, fmt.Errorf("%w: %q", ErrUUIDTidakValid, s)
	}
	return id, nil
}

// IsValid cek cepat tanpa mengalokasikan hasil parse.
func IsValid(s string) bool {
	return uuid.Validate(s) == nil
}

// VersionOf mengembalikan nomor versi UUID (4, 7, dst).
func VersionOf(s string) (int, error) {
	id, err := ParseUUID(s)
	if err != nil {
		return 0, err
	}
	return int(id.Version()), nil
}

func demoParse() {
	fmt.Println("\n-- Parse & validasi --")

	contoh := []string{
		"550e8400-e29b-41d4-a716-446655440000", // valid, v4
		"550e8400e29b41d4a716446655440000",     // valid juga (tanpa tanda hubung)
		"bukan-uuid",                           // tidak valid
		"",                                     // kosong
	}
	for _, s := range contoh {
		if v, err := VersionOf(s); err != nil {
			fmt.Printf("   %-40q -> DITOLAK (%v)\n", s, err)
		} else {
			fmt.Printf("   %-40q -> valid, versi %d\n", s, v)
		}
	}
}

// ------------------------------------------------------------------
// 4. Kasus tepi & jebakan
// ------------------------------------------------------------------

// 🔍 Analogi uuid.Nil: seperti kolom nomor rangka yang masih KOSONG di formulir.
// Nilainya 00000000-0000-0000-0000-000000000000 — ini "zero value"-nya UUID.
// Jebakan: uuid.Nil BUKAN error. Kalau kamu lupa mengisi ID, kodenya tetap jalan
// tapi seluruh baris di database punya ID kembar yang sama. Selalu cek "== uuid.Nil".

// IsNil memeriksa apakah UUID masih zero value (belum diisi).
func IsNil(id uuid.UUID) bool {
	return id == uuid.Nil
}

// 🔍 Analogi UUID sebagai key map: uuid.UUID itu array [16]byte — tipe yang COMPARABLE,
// jadi bisa langsung jadi kunci map (seperti nomor rangka jadi kunci di lemari arsip).
// Ini alasan menyimpan uuid.UUID lebih hemat daripada string: 16 byte vs 36 byte.

// HitungUnik memakai UUID langsung sebagai kunci map (tanpa .String()).
func HitungUnik(ids []uuid.UUID) int {
	set := make(map[uuid.UUID]struct{}, len(ids))
	for _, id := range ids {
		set[id] = struct{}{}
	}
	return len(set)
}

// MustParseAman membungkus uuid.MustParse yang PANIC bila string salah.
//
// 🔍 Analogi: MustParse itu seperti tombol tanpa pengaman — hanya boleh dipakai untuk
// nilai yang kamu tulis sendiri di kode (konstanta), BUKAN untuk input dari pengguna.
// Input pengguna selalu pakai uuid.Parse yang mengembalikan error.
func MustParseAman(s string) (id uuid.UUID, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%w: MustParse panic pada %q", ErrUUIDTidakValid, s)
		}
	}()
	return uuid.MustParse(s), nil
}

func demoJebakan() {
	fmt.Println("\n-- Kasus tepi & jebakan --")

	var kosong uuid.UUID
	fmt.Printf("   zero value UUID = %v (IsNil=%t)\n", kosong, IsNil(kosong))

	a := uuid.New()
	b := uuid.New()
	fmt.Printf("   HitungUnik([a,b,a]) = %d (UUID jadi kunci map langsung)\n",
		HitungUnik([]uuid.UUID{a, b, a}))

	if _, err := MustParseAman("jelas-bukan-uuid"); err != nil {
		fmt.Println("   MustParse pada input kotor ->", err)
	}

	fmt.Println("\n   Ringkas: v7 untuk primary key (urut waktu), v4 untuk token/nonce acak,")
	fmt.Println("   simpan sebagai uuid.UUID (16 byte) bukan string (36 byte).")
}
