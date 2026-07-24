// go-playground/validator — memvalidasi struct lewat tag, dengan pesan Bahasa Indonesia.
//
// Jalankan: go run ./libraries/validator
// Test:     go test ./libraries/validator
//
// 🔍 Analogi besar: validator itu SATPAM DI PINTU MASUK gedung. Tanpa satpam, kamu harus
// menaruh pemeriksaan identitas di setiap ruangan (if di handler, if lagi di service,
// if lagi sebelum simpan ke database) — melelahkan dan pasti ada yang terlewat.
// Dengan satpam di satu pintu, aturannya ditulis SEKALI di tag struct, dan semua yang
// masuk sudah dijamin bersih.
//
// Aturan emas: validasi di BATAS LUAR aplikasi (handler HTTP, konsumer antrean), lalu
// bagian dalam boleh percaya bahwa datanya sudah sah. Ini juga alasan kenapa modul 13
// memakai validator di lapisan handler, bukan di dalam service.
package main

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
)

func main() {
	fmt.Println("=== go-playground/validator ===")

	v, err := NewValidator()
	if err != nil {
		fmt.Println("gagal menyiapkan validator:", err)
		return
	}
	demoValid(v)
	demoTidakValid(v)
	demoAturanKhusus(v)
	demoLintasField(v)
	demoJebakan(v)
}

// ------------------------------------------------------------------
// 1. Struct dengan aturan di tag
// ------------------------------------------------------------------

// Alamat adalah struct bersarang — validator ikut turun memeriksanya.
type Alamat struct {
	Jalan    string `json:"jalan" validate:"required,min=5"`
	Kota     string `json:"kota" validate:"required"`
	KodePos  string `json:"kode_pos" validate:"required,len=5,number"`
	Provinsi string `json:"provinsi" validate:"required,oneof=Jakarta Jabar Jateng Jatim Bali"`
}

// Pendaftaran memuat hampir semua jenis aturan yang sering dipakai.
//
// 🔍 Analogi tiap tag = satu baris syarat di formulir pendaftaran:
//
//	required      = "wajib diisi" (bintang merah di formulir)
//	email         = "harus berbentuk alamat surel"
//	min=8 / max=50= panjang teks minimal/maksimal
//	gte=17,lte=99 = untuk ANGKA: lebih besar/kecil sama dengan
//	oneof=a b c   = hanya boleh salah satu dari daftar (seperti pilihan ganda)
//	eqfield=X     = harus sama dengan isi kolom X (konfirmasi kata sandi)
//	omitempty     = "kalau kosong ya sudah, tapi kalau diisi harus benar" (kolom opsional)
//	dive          = "periksa juga SETIAP ISI di dalam slice/map ini"
//	notelp_id     = aturan buatan sendiri, didaftarkan di NewValidator()
type Pendaftaran struct {
	Nama       string   `json:"nama" validate:"required,min=3,max=50"`
	Email      string   `json:"email" validate:"required,email"`
	Umur       int      `json:"umur" validate:"required,gte=17,lte=99"`
	Telepon    string   `json:"telepon" validate:"required,notelp_id"`
	Peran      string   `json:"peran" validate:"required,oneof=admin editor pembaca"`
	Kata       string   `json:"kata_sandi" validate:"required,min=8"`
	KonfirKata string   `json:"konfirmasi_kata_sandi" validate:"required,eqfield=Kata"`
	Situs      string   `json:"situs" validate:"omitempty,url"`
	Minat      []string `json:"minat" validate:"required,min=1,dive,min=3"`
	Alamat     Alamat   `json:"alamat" validate:"required"`
}

// ------------------------------------------------------------------
// 2. Menyiapkan validator
// ------------------------------------------------------------------

// NewValidator membuat validator yang sudah disetel untuk kebutuhan aplikasi Indonesia.
//
// Buat SEKALI lalu pakai ulang: validator menyimpan cache hasil pembacaan tag,
// jadi membuatnya berulang kali membuang kerja. Ia juga aman dipakai banyak goroutine.
func NewValidator() (*validator.Validate, error) {
	v := validator.New(validator.WithRequiredStructEnabled())

	// 🔍 Analogi: tanpa ini, pesan error menyebut nama field Go ("KonfirKata") yang tak
	// pernah dilihat pengguna. Dengan ini, pesan memakai nama dari tag json
	// ("konfirmasi_kata_sandi") — persis yang dikirim klien, sehingga frontend bisa
	// menyorot kolom yang tepat.
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		nama := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if nama == "" || nama == "-" {
			return fld.Name
		}
		return nama
	})

	// Aturan buatan sendiri.
	if err := v.RegisterValidation("notelp_id", validasiTeleponID); err != nil {
		return nil, fmt.Errorf("gagal mendaftarkan aturan notelp_id: %w", err)
	}

	// Aturan tingkat struct (butuh melihat beberapa field sekaligus).
	v.RegisterStructValidation(validasiPemesanan, Pemesanan{})

	return v, nil
}

// ------------------------------------------------------------------
// 3. Menjalankan validasi & menerjemahkan errornya
// ------------------------------------------------------------------

// KesalahanField adalah bentuk error yang siap dikirim ke klien sebagai JSON.
type KesalahanField struct {
	Field string `json:"field"`
	Pesan string `json:"pesan"`
}

// Validasi menjalankan pemeriksaan dan menerjemahkan hasilnya ke Bahasa Indonesia.
//
// 🔍 Analogi: keluaran mentah validator itu seperti KODE PELANGGARAN polisi ("pasal 287").
// Pengguna tak paham. Fungsi ini adalah JURU BAHASA yang mengubahnya jadi kalimat manusia
// ("umur minimal 17"). Selalu terjemahkan sebelum ditampilkan ke pengguna.
func Validasi(v *validator.Validate, s any) []KesalahanField {
	err := v.Struct(s)
	if err == nil {
		return nil
	}

	// InvalidValidationError terjadi bila yang dikirim bukan struct (mis. string atau nil).
	// Ini kesalahan PROGRAMMER, bukan kesalahan pengguna — bedakan penanganannya.
	var salahPakai *validator.InvalidValidationError
	if errors.As(err, &salahPakai) {
		return []KesalahanField{{Field: "-", Pesan: "objek yang divalidasi bukan struct"}}
	}

	var ve validator.ValidationErrors
	if !errors.As(err, &ve) {
		return []KesalahanField{{Field: "-", Pesan: err.Error()}}
	}

	hasil := make([]KesalahanField, 0, len(ve))
	for _, fe := range ve {
		hasil = append(hasil, KesalahanField{
			Field: fe.Field(),
			Pesan: pesanIndonesia(fe),
		})
	}
	return hasil
}

// pesanIndonesia mengubah satu pelanggaran jadi kalimat yang bisa dibaca pengguna.
func pesanIndonesia(fe validator.FieldError) string {
	field := fe.Field()
	param := fe.Param()

	switch fe.Tag() {
	case "required":
		return field + " wajib diisi"
	case "email":
		return field + " harus berupa alamat email yang sah"
	case "url":
		return field + " harus berupa URL yang sah"
	case "min":
		if fe.Kind() == reflect.String {
			return fmt.Sprintf("%s minimal %s karakter", field, param)
		}
		return fmt.Sprintf("%s minimal %s", field, param)
	case "max":
		if fe.Kind() == reflect.String {
			return fmt.Sprintf("%s maksimal %s karakter", field, param)
		}
		return fmt.Sprintf("%s maksimal %s", field, param)
	case "len":
		return fmt.Sprintf("%s harus tepat %s karakter", field, param)
	case "gte":
		return fmt.Sprintf("%s minimal %s", field, param)
	case "lte":
		return fmt.Sprintf("%s maksimal %s", field, param)
	case "oneof":
		return fmt.Sprintf("%s harus salah satu dari: %s", field, strings.ReplaceAll(param, " ", ", "))
	case "eqfield":
		return field + " harus sama dengan " + param
	case "number", "numeric":
		return field + " hanya boleh berisi angka"
	case "notelp_id":
		return field + " harus nomor Indonesia yang sah (08... atau +62...)"
	case "gtfield":
		return field + " harus setelah " + param
	default:
		return fmt.Sprintf("%s tidak memenuhi aturan %q", field, fe.Tag())
	}
}

func contohValid() Pendaftaran {
	return Pendaftaran{
		Nama:       "Ana Pratiwi",
		Email:      "ana@contoh.id",
		Umur:       28,
		Telepon:    "081234567890",
		Peran:      "editor",
		Kata:       "rahasia123",
		KonfirKata: "rahasia123",
		Situs:      "https://ana.contoh.id",
		Minat:      []string{"golang", "musik"},
		Alamat: Alamat{
			Jalan:    "Jl. Merdeka No. 10",
			Kota:     "Bandung",
			KodePos:  "40111",
			Provinsi: "Jabar",
		},
	}
}

func demoValid(v *validator.Validate) {
	fmt.Println("\n-- Data yang benar --")
	if errs := Validasi(v, contohValid()); len(errs) == 0 {
		fmt.Println("   lolos semua pemeriksaan ✅")
	} else {
		fmt.Println("   tak terduga:", errs)
	}
}

func demoTidakValid(v *validator.Validate) {
	fmt.Println("\n-- Data yang bermasalah --")

	buruk := Pendaftaran{
		Nama:       "Ab",               // terlalu pendek
		Email:      "bukan-email",      // bukan email
		Umur:       12,                 // di bawah 17
		Telepon:    "12345",            // bukan nomor Indonesia
		Peran:      "raja",             // di luar daftar
		Kata:       "pendek",           // kurang dari 8
		KonfirKata: "beda",             // tak sama
		Situs:      "bukan-url",        // diisi tapi salah bentuk
		Minat:      []string{"go", ""}, // isi slice terlalu pendek
		// Alamat sengaja dibiarkan kosong
	}
	for _, e := range Validasi(v, buruk) {
		fmt.Printf("   %-22s %s\n", e.Field, e.Pesan)
	}
}

// ------------------------------------------------------------------
// 4. Aturan buatan sendiri
// ------------------------------------------------------------------

// validasiTeleponID memeriksa format nomor telepon Indonesia.
//
// 🔍 Analogi: tag bawaan itu perkakas serba guna. Begitu aturannya khas negaramu/bisnismu
// (NIK, NPWP, kode cabang), kamu membuat "cap" sendiri lalu memakainya di banyak struct
// dengan satu kata — bukan menyalin-tempel logika if ke mana-mana.
func validasiTeleponID(fl validator.FieldLevel) bool {
	s := fl.Field().String()

	// Normalkan awalan: +62 dan 62 diperlakukan sama dengan 0.
	switch {
	case strings.HasPrefix(s, "+62"):
		s = "0" + s[3:]
	case strings.HasPrefix(s, "62"):
		s = "0" + s[2:]
	}

	if !strings.HasPrefix(s, "08") {
		return false
	}
	if len(s) < 10 || len(s) > 14 {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func demoAturanKhusus(v *validator.Validate) {
	fmt.Println("\n-- Aturan buatan sendiri: notelp_id --")
	for _, n := range []string{
		"081234567890", "+6281234567890", "6281234567890",
		"0812", "12345678901", "0812-3456-7890",
	} {
		// Var memvalidasi satu nilai tanpa perlu membungkusnya ke struct.
		err := v.Var(n, "notelp_id")
		status := "sah"
		if err != nil {
			status = "DITOLAK"
		}
		fmt.Printf("   %-18s %s\n", n, status)
	}
}

// ------------------------------------------------------------------
// 5. Validasi lintas field (tingkat struct)
// ------------------------------------------------------------------

// Pemesanan butuh aturan yang tak bisa ditulis di satu tag: tanggal selesai harus
// setelah tanggal mulai, DAN durasinya tak boleh lebih dari 30 hari.
type Pemesanan struct {
	Kode    string    `json:"kode" validate:"required"`
	Mulai   time.Time `json:"mulai" validate:"required"`
	Selesai time.Time `json:"selesai" validate:"required"`
	Tamu    int       `json:"tamu" validate:"required,gte=1,lte=10"`
}

// validasiPemesanan dijalankan setelah tag biasa selesai diperiksa.
//
// 🔍 Analogi: tag itu memeriksa tiap kolom SENDIRI-SENDIRI ("tanggal ini terisi?").
// Validasi tingkat struct memeriksa HUBUNGAN antar kolom ("tanggal pulang setelah
// tanggal berangkat?") — pertanyaan yang mustahil dijawab bila hanya melihat satu kolom.
func validasiPemesanan(sl validator.StructLevel) {
	p, ok := sl.Current().Interface().(Pemesanan)
	if !ok {
		return
	}

	if !p.Selesai.After(p.Mulai) {
		sl.ReportError(p.Selesai, "selesai", "Selesai", "gtfield", "mulai")
		return
	}
	if p.Selesai.Sub(p.Mulai) > 30*24*time.Hour {
		sl.ReportError(p.Selesai, "selesai", "Selesai", "maks30hari", "")
	}
}

func demoLintasField(v *validator.Validate) {
	fmt.Println("\n-- Validasi lintas field --")

	mulai := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	kasus := []struct {
		nama string
		p    Pemesanan
	}{
		{"normal", Pemesanan{Kode: "BK-1", Mulai: mulai, Selesai: mulai.Add(72 * time.Hour), Tamu: 2}},
		{"selesai sebelum mulai", Pemesanan{Kode: "BK-2", Mulai: mulai, Selesai: mulai.Add(-24 * time.Hour), Tamu: 2}},
		{"terlalu lama", Pemesanan{Kode: "BK-3", Mulai: mulai, Selesai: mulai.Add(60 * 24 * time.Hour), Tamu: 2}},
		{"tamu terlalu banyak", Pemesanan{Kode: "BK-4", Mulai: mulai, Selesai: mulai.Add(24 * time.Hour), Tamu: 50}},
	}
	for _, k := range kasus {
		errs := Validasi(v, k.p)
		if len(errs) == 0 {
			fmt.Printf("   %-24s lolos\n", k.nama)
			continue
		}
		fmt.Printf("   %-24s %s\n", k.nama, errs[0].Pesan)
	}
}

// ------------------------------------------------------------------
// 6. Jebakan yang perlu diketahui
// ------------------------------------------------------------------

// 🔍 Jebakan 1 — "required" TIDAK BISA membedakan "nol" dari "tidak diisi".
// Bagi validator, Umur=0 sama saja dengan kolom yang tak dikirim, karena keduanya
// adalah zero value. Kalau nol adalah nilai yang SAH di bisnismu (mis. diskon 0%,
// stok 0), pakai pointer (*int) supaya nil berarti "tak dikirim" dan 0 berarti "nol".

// Diskon memperagakan perbedaan int vs *int untuk kolom yang boleh bernilai nol.
type Diskon struct {
	// Salah untuk kasus ini: 0 akan dianggap "tidak diisi" dan ditolak.
	PersenSalah int `json:"persen_salah" validate:"required,gte=0,lte=100"`
	// Benar: nil = tak dikirim, &0 = nol persen dan itu sah.
	PersenBenar *int `json:"persen_benar" validate:"required,gte=0,lte=100"`
}

// 🔍 Jebakan 2 — validator hanya melihat field yang DIEKSPOR (huruf besar di awal).
// Field kecil seperti "harga" akan dilewati diam-diam, tanpa peringatan apa pun.

// 🔍 Jebakan 3 — validasi BUKAN pengganti pemeriksaan aturan bisnis. "email berbentuk sah"
// tidak sama dengan "email ini belum terdaftar" — yang kedua butuh query ke database
// dan tempatnya di lapisan service, bukan di tag.

func demoJebakan(v *validator.Validate) {
	fmt.Println("\n-- Jebakan: required vs nilai nol --")

	nol := 0
	d := Diskon{PersenSalah: 0, PersenBenar: &nol}
	for _, e := range Validasi(v, d) {
		fmt.Printf("   %-15s %s  <- padahal 0%% itu sah!\n", e.Field, e.Pesan)
	}
	fmt.Println("   persen_benar (pointer ke 0) lolos, karena nil != 0")
}
