// tidwall/gjson + sjson — membaca & menulis JSON TANPA mendefinisikan struct.
//
// Jalankan: go run ./libraries/gjson
// Test:     go test ./libraries/gjson
//
// 🔍 Analogi besar: `encoding/json` bawaan itu seperti MEMBONGKAR SELURUH ISI KOPER untuk
// mengambil satu kaus kaki — kamu harus mendefinisikan struct yang mencakup seluruh bentuk
// JSON, lalu Unmarshal semuanya, baru mengambil field yang kamu mau. Bagus kalau bentuknya
// tetap & kamu butuh semuanya.
//
// gjson itu MERAIH LANGSUNG ke saku koper: `gjson.Get(data, "user.alamat.kota")` mengambil
// satu nilai tanpa membongkar apa pun & tanpa struct. sjson pasangannya untuk MENULIS:
// `sjson.Set(data, "user.aktif", true)`.
//
// Kapan pakai ini vs encoding/json:
//   - Bentuk JSON TIDAK TETAP / tak dikenal (webhook pihak ketiga, config dinamis, respons
//     API yang cuma butuh 2-3 field dari 100).
//   - Sekadar mengintip/menambal satu-dua field tanpa repot bikin struct.
//
// Kapan tetap encoding/json:
//   - Bentuknya tetap & kamu memang mengelola seluruh objeknya -> struct lebih jelas,
//     type-safe, dan tervalidasi kompiler.
package main

import (
	"errors"
	"fmt"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// Contoh JSON bersarang & ber-array — sengaja "berantakan" seperti data dunia nyata.
const dataPesanan = `{
  "id": "ORD-2026-042",
  "pelanggan": {
    "nama": "Ana Pratiwi",
    "email": "ana@contoh.id",
    "alamat": {"kota": "Bandung", "provinsi": "Jabar", "kodepos": "40111"}
  },
  "item": [
    {"nama": "Kopi Arabika", "harga": 85000, "qty": 2},
    {"nama": "Teh Melati",   "harga": 35000, "qty": 1},
    {"nama": "Gula Aren",    "harga": 15000, "qty": 3}
  ],
  "lunas": false,
  "kupon": null
}`

func main() {
	fmt.Println("=== tidwall/gjson + sjson ===")
	demoBaca()
	demoArray()
	demoAda()
	demoTulis()
	demoValidasi()
}

// ------------------------------------------------------------------
// 1. Membaca nilai dengan path bertitik
// ------------------------------------------------------------------

// 🔍 Analogi path "pelanggan.alamat.kota": seperti ALAMAT LENGKAP yang dibaca dari luar ke
// dalam — provinsi, lalu kota, lalu jalan. gjson menuruni JSON mengikuti titik. Tiap segmen
// bisa nama field ATAU indeks array (item.0.nama = item pertama).

// KotaPelanggan mengambil kota tanpa struct apa pun.
func KotaPelanggan(json string) string {
	return gjson.Get(json, "pelanggan.alamat.kota").String()
}

// NamaPelanggan mengambil nama.
func NamaPelanggan(json string) string {
	return gjson.Get(json, "pelanggan.nama").String()
}

// StatusLunas membaca boolean.
//
// 🔍 Analogi tipe: gjson menyimpan hasil sebagai Result yang bisa "diminta" jadi tipe apa
// pun lewat .String()/.Int()/.Bool()/.Float(). Kalau path tak ada atau tipenya beda, kamu
// dapat ZERO VALUE (0, "", false) — BUKAN panic. Nyaman, tapi hati-hati: "field hilang" dan
// "field bernilai false" sama-sama menghasilkan false. Bila bedanya penting, cek Exists().
func StatusLunas(json string) bool {
	return gjson.Get(json, "lunas").Bool()
}

func demoBaca() {
	fmt.Println("\n-- Membaca nilai --")
	fmt.Println("   nama  :", NamaPelanggan(dataPesanan))
	fmt.Println("   kota  :", KotaPelanggan(dataPesanan))
	fmt.Println("   lunas :", StatusLunas(dataPesanan))
	fmt.Println("   item pertama:", gjson.Get(dataPesanan, "item.0.nama").String())
}

// ------------------------------------------------------------------
// 2. Query array — di sinilah gjson bersinar
// ------------------------------------------------------------------

// 🔍 Analogi query array gjson: ini seperti "mesin pencari mini" di dalam JSON.
//   "item.#"           = HITUNG isi array (# = berapa banyak).
//   "item.#.nama"      = ambil field 'nama' dari SETIAP item -> jadi array nama.
//   "item.#(harga>30000)#.nama" = ambil nama item yang harganya > 30000 (filter!).
// Melakukan ini dengan encoding/json butuh Unmarshal + for-loop + if. gjson: satu string.

// SemuaNamaItem mengambil nama seluruh item sebagai slice.
func SemuaNamaItem(json string) []string {
	hasil := gjson.Get(json, "item.#.nama")
	var out []string
	hasil.ForEach(func(_, value gjson.Result) bool {
		out = append(out, value.String())
		return true // true = lanjut; return false untuk berhenti lebih awal
	})
	return out
}

// JumlahItem menghitung banyaknya item.
func JumlahItem(json string) int64 {
	return gjson.Get(json, "item.#").Int()
}

// TotalHarga menjumlahkan harga*qty seluruh item.
//
// gjson tak punya "sum" bawaan, jadi kita iterasi — tapi tanpa struct & tanpa Unmarshal.
func TotalHarga(json string) int64 {
	var total int64
	gjson.Get(json, "item").ForEach(func(_, item gjson.Result) bool {
		total += item.Get("harga").Int() * item.Get("qty").Int()
		return true
	})
	return total
}

// ItemMahal mengambil nama item yang harganya di atas ambang, memakai sintaks filter gjson.
func ItemMahal(json string, ambang int) []string {
	path := fmt.Sprintf("item.#(harga>%d)#.nama", ambang)
	var out []string
	gjson.Get(json, path).ForEach(func(_, v gjson.Result) bool {
		out = append(out, v.String())
		return true
	})
	return out
}

func demoArray() {
	fmt.Println("\n-- Query array --")
	fmt.Println("   jumlah item :", JumlahItem(dataPesanan))
	fmt.Println("   semua nama  :", SemuaNamaItem(dataPesanan))
	fmt.Printf("   total harga : Rp%d\n", TotalHarga(dataPesanan))
	fmt.Println("   item >30rb  :", ItemMahal(dataPesanan, 30_000))
}

// ------------------------------------------------------------------
// 3. Exists — membedakan "tidak ada" dari "bernilai kosong/false/null"
// ------------------------------------------------------------------

// 🔍 Analogi: di JSON, tiga hal ini BERBEDA tapi mudah tertukar:
//   field tak ada        -> Exists()=false
//   field bernilai null  -> Exists()=true,  Type=Null
//   field bernilai false -> Exists()=true,  Bool()=false
// Untuk webhook & config, bedanya sering penting ("pengguna tak mengirim field" vs
// "pengguna sengaja mengosongkannya"). Selalu pakai Exists() saat ketiadaan itu bermakna.

// PunyaKupon memeriksa apakah field kupon ada DAN tidak null.
func PunyaKupon(json string) bool {
	r := gjson.Get(json, "kupon")
	return r.Exists() && r.Type != gjson.Null
}

// PunyaField cek keberadaan path apa pun.
func PunyaField(json, path string) bool {
	return gjson.Get(json, path).Exists()
}

func demoAda() {
	fmt.Println("\n-- Exists --")
	fmt.Printf("   punya 'kupon' (null)?        %t\n", PunyaKupon(dataPesanan))
	fmt.Printf("   ada 'pelanggan.email'?       %t\n", PunyaField(dataPesanan, "pelanggan.email"))
	fmt.Printf("   ada 'pelanggan.telepon'?     %t\n", PunyaField(dataPesanan, "pelanggan.telepon"))
}

// ------------------------------------------------------------------
// 4. Menulis dengan sjson (imutabel: mengembalikan JSON BARU)
// ------------------------------------------------------------------

// 🔍 Analogi sjson: seperti fotokopi-dengan-koreksi. sjson TIDAK mengubah string aslinya;
// ia mengembalikan SALINAN baru dengan perubahannya. Ini idiomatik & aman — tak ada efek
// samping tersembunyi. Path yang belum ada akan DIBUAT otomatis (termasuk objek bersarang).

// TandaiLunas mengubah status lunas jadi true.
func TandaiLunas(json string) (string, error) {
	baru, err := sjson.Set(json, "lunas", true)
	if err != nil {
		return "", fmt.Errorf("gagal mengubah status lunas: %w", err)
	}
	return baru, nil
}

// PasangKupon menetapkan kode kupon (membuat field yang tadinya null).
func PasangKupon(json, kode string) (string, error) {
	baru, err := sjson.Set(json, "kupon", kode)
	if err != nil {
		return "", fmt.Errorf("gagal memasang kupon: %w", err)
	}
	return baru, nil
}

// TambahCatatan membuat path bersarang yang belum ada sama sekali.
func TambahCatatan(json, catatan string) (string, error) {
	baru, err := sjson.Set(json, "meta.catatan", catatan)
	if err != nil {
		return "", fmt.Errorf("gagal menambah catatan: %w", err)
	}
	return baru, nil
}

// HapusField membuang sebuah field.
func HapusField(json, path string) (string, error) {
	baru, err := sjson.Delete(json, path)
	if err != nil {
		return "", fmt.Errorf("gagal menghapus %q: %w", path, err)
	}
	return baru, nil
}

func demoTulis() {
	fmt.Println("\n-- Menulis (sjson) --")

	lunas, _ := TandaiLunas(dataPesanan)
	fmt.Printf("   setelah TandaiLunas -> lunas=%t\n", gjson.Get(lunas, "lunas").Bool())

	berkupon, _ := PasangKupon(dataPesanan, "HEMAT10")
	fmt.Printf("   pasang kupon        -> kupon=%q\n", gjson.Get(berkupon, "kupon").String())

	bercatatan, _ := TambahCatatan(dataPesanan, "kirim pagi")
	fmt.Printf("   tambah meta.catatan -> %q (path bersarang dibuat otomatis)\n",
		gjson.Get(bercatatan, "meta.catatan").String())

	// Asli TIDAK berubah — inilah sifat imutabel sjson.
	fmt.Printf("   data asli masih lunas=%t (tak tersentuh)\n",
		gjson.Get(dataPesanan, "lunas").Bool())
}

// ------------------------------------------------------------------
// 5. Validasi input
// ------------------------------------------------------------------

// ErrJSONTidakValid dikembalikan untuk input yang bukan JSON.
var ErrJSONTidakValid = errors.New("bukan JSON yang valid")

// AmbilAman memvalidasi dulu sebelum membaca.
//
// 🔍 Analogi & JEBAKAN PENTING: demi kecepatan, gjson.Get TIDAK memvalidasi JSON — ia
// "berjalan" melewati byte dan bisa memberi hasil aneh pada input rusak, tanpa error.
// Untuk data dari LUAR (pengguna, jaringan), SELALU gjson.Valid() dulu. Untuk data yang
// kamu yakini sudah benar (hasil Marshal sendiri), boleh langsung Get demi performa.
func AmbilAman(json, path string) (string, error) {
	if !gjson.Valid(json) {
		return "", ErrJSONTidakValid
	}
	return gjson.Get(json, path).String(), nil
}

func demoValidasi() {
	fmt.Println("\n-- Validasi --")

	if v, err := AmbilAman(dataPesanan, "pelanggan.nama"); err == nil {
		fmt.Println("   JSON valid   -> nama:", v)
	}
	if _, err := AmbilAman("{ini rusak", "apa.saja"); errors.Is(err, ErrJSONTidakValid) {
		fmt.Println("   JSON rusak   -> ditolak sebelum dibaca")
	}
}
