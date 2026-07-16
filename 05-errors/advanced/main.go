// Teknik Advanced Modul 05 (errors) — contoh runnable + penjelasan detail.
//
// Jalankan: go run ./05-errors/advanced
package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
)

// Sentinel error: nilai tetap yang bisa dibandingkan dengan errors.Is.
var ErrTidakDitemukan = errors.New("data tidak ditemukan")

// Typed error: membawa DATA tambahan; diambil kembali dengan errors.As.
type ValidasiError struct {
	Field string
	Pesan string
}

func (e *ValidasiError) Error() string {
	return fmt.Sprintf("validasi gagal pada %q: %s", e.Field, e.Pesan)
}

func main() {
	// =====================================================================
	// 1. errors.Join (Go 1.20) — gabung banyak error jadi satu
	// =====================================================================
	// Cocok untuk validasi banyak field: kumpulkan semua, laporkan sekaligus.
	// errors.Is/As tetap bisa menelusuri KE DALAM gabungan.
	//
	// 🔍 Analogi: errors.Join itu seperti MAP RAPOR — semua nilai merah
	// dikumpulkan dalam satu map, dan orang tua (errors.Is/As) tetap bisa
	// membuka lembar per lembar untuk mencari mata pelajaran tertentu.
	gabung := errors.Join(
		ErrTidakDitemukan,
		&ValidasiError{Field: "email", Pesan: "format salah"},
	)
	fmt.Println("== 1. errors.Join ==")
	fmt.Println("  isi:", gabung)
	fmt.Println("  errors.Is(.., ErrTidakDitemukan):", errors.Is(gabung, ErrTidakDitemukan))

	var ve *ValidasiError
	if errors.As(gabung, &ve) { // ambil kembali typed error + datanya
		fmt.Printf("  errors.As -> field bermasalah: %s\n", ve.Field)
	}

	// =====================================================================
	// 2. %w vs %v — mempertahankan vs menyegel rantai error
	// =====================================================================
	// %w  : membungkus, rantai TETAP bisa ditelusuri errors.Is/As.
	// %v  : menjadikan teks biasa, rantai TERPUTUS (sengaja sembunyikan detail).
	//
	// 🔍 Analogi: verb-w itu seperti MEMBUNGKUS KADO — isi aslinya masih ada
	// di dalam dan bisa dibuka lapis demi lapis. verb-v seperti MEMFOTO kado
	// lalu membuang isinya: yang tersisa cuma gambar (teks), isinya hilang.
	dibungkus := fmt.Errorf("muat profil: %w", ErrTidakDitemukan)
	disegel := fmt.Errorf("muat profil: %v", ErrTidakDitemukan)
	fmt.Println("== 2. bungkus (verb-w) vs segel (verb-v) ==")
	fmt.Println("  verb-w -> Is(ErrTidakDitemukan):", errors.Is(dibungkus, ErrTidakDitemukan)) // true
	fmt.Println("  verb-v -> Is(ErrTidakDitemukan):", errors.Is(disegel, ErrTidakDitemukan))   // false

	// =====================================================================
	// 3. Cek error stdlib modern: fs.ErrNotExist lewat errors.Is
	// =====================================================================
	// Jangan pakai os.IsNotExist (lawas, tak menembus pembungkus). Gunakan
	// errors.Is(err, fs.ErrNotExist) yang bekerja walau error sudah dibungkus.
	//
	// 🔍 Analogi: os.IsNotExist itu seperti satpam yang hanya mengecek KULIT
	// paket terluar; errors.Is seperti pemindai X-RAY — tetap menemukan barang
	// yang dicari walau sudah terbungkus berlapis-lapis.
	_, err := os.Open("/berkas/yang/pasti/tidak/ada")
	fmt.Println("== 3. errors.Is(err, fs.ErrNotExist) ==")
	fmt.Println("  berkas tak ada?", errors.Is(err, fs.ErrNotExist))

	// =====================================================================
	// 4. panic/recover di BATAS (boundary) — jaring pengaman
	// =====================================================================
	// panic tak boleh bocor lintas API. Tangkap di boundary (mis. middleware
	// HTTP) agar satu request bermasalah tak menjatuhkan seluruh server.
	//
	// 🔍 Analogi: recover di boundary itu seperti SEKAT KEDAP AIR di kapal —
	// satu kompartemen (request) boleh kebanjiran, tapi sekatnya menahan air
	// agar seluruh kapal (server) tidak ikut tenggelam.
	fmt.Println("== 4. recover di boundary ==")
	jalankanAman(func() { panic("kode dalam meledak") })
	fmt.Println("  server tetap hidup setelah panic ditangani")
}

// jalankanAman meniru middleware: apa pun panic di dalam fn diubah jadi log,
// bukan crash. Pola inti untuk ketahanan server.
func jalankanAman(fn func()) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("  [recover] menangkap panic: %v\n", r)
		}
	}()
	fn()
}
