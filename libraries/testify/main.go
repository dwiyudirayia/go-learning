// stretchr/testify — assertion, mock, dan suite untuk test yang lebih ringkas.
//
// Jalankan: go run ./libraries/testify        (menjalankan kode yang diuji)
// Test:     go test -v ./libraries/testify    (INI bagian pentingnya — lihat testify_test.go)
//
// 🔍 Analogi besar: test bawaan Go itu seperti memasak tanpa alat ukur — kamu harus menulis
// sendiri "kalau hasilnya bukan X, laporkan begini" berulang-ulang. testify itu SET ALAT UKUR
// dapur: sendok takar, timbangan, termometer. Hasil masakannya sama, tapi kode test jadi jauh
// lebih pendek dan pesan kegagalannya otomatis rapi ("expected 100, actual 90").
//
// Catatan penting: testify itu pilihan, BUKAN keharusan. Banyak proyek Go besar (termasuk
// standard library) sengaja hanya memakai testing bawaan. File ini menunjukkan kapan
// testify benar-benar menolong, dan kapan bawaan Go sudah cukup.
package main

import (
	"errors"
	"fmt"
)

func main() {
	fmt.Println("=== stretchr/testify ===")
	fmt.Println("File ini berisi KODE YANG DIUJI. Contoh testify-nya ada di testify_test.go.")
	fmt.Println("Jalankan: go test -v ./libraries/testify")

	demoDompet()
	demoPembayaran()
}

// ------------------------------------------------------------------
// 1. Kode yang diuji: Dompet
// ------------------------------------------------------------------

// ErrSaldoKurang dan ErrJumlahTidakValid sentinel error.
// 🔍 Analogi: sentinel error itu KODE RESMI di struk penolakan ("saldo kurang"),
// bukan kalimat bebas — supaya kode lain bisa mencocokkannya dengan errors.Is.
var (
	ErrSaldoKurang        = errors.New("saldo tidak mencukupi")
	ErrJumlahTidakValid   = errors.New("jumlah harus lebih besar dari nol")
	ErrPenerimaTakDikenal = errors.New("penerima tidak dikenal")
)

// Dompet menyimpan saldo dalam satuan terkecil (rupiah).
//
// 🔍 Analogi: menyimpan uang sebagai bilangan bulat "rupiah" (bukan pecahan) itu seperti
// menghitung dengan satuan sen di kasir — menghindari sisa-sisa aneh akibat pecahan desimal.
// (Untuk perhitungan uang yang butuh pecahan, lihat libraries/decimal.)
type Dompet struct {
	Pemilik string
	Saldo   int
}

// NewDompet konstruktor — memastikan dompet selalu lahir dalam keadaan sah.
func NewDompet(pemilik string, saldoAwal int) *Dompet {
	return &Dompet{Pemilik: pemilik, Saldo: saldoAwal}
}

// Setor menambah saldo. Pointer receiver karena mengubah isi.
func (d *Dompet) Setor(jumlah int) error {
	if jumlah <= 0 {
		return fmt.Errorf("setor %d: %w", jumlah, ErrJumlahTidakValid)
	}
	d.Saldo += jumlah
	return nil
}

// Tarik mengurangi saldo bila mencukupi.
func (d *Dompet) Tarik(jumlah int) error {
	if jumlah <= 0 {
		return fmt.Errorf("tarik %d: %w", jumlah, ErrJumlahTidakValid)
	}
	if jumlah > d.Saldo {
		return fmt.Errorf("tarik %d dari saldo %d: %w", jumlah, d.Saldo, ErrSaldoKurang)
	}
	d.Saldo -= jumlah
	return nil
}

func demoDompet() {
	fmt.Println("\n-- Dompet --")
	d := NewDompet("Ana", 100_000)
	_ = d.Setor(50_000)
	fmt.Printf("   setelah setor 50.000 -> saldo %d\n", d.Saldo)

	if err := d.Tarik(500_000); err != nil {
		fmt.Println("   tarik 500.000 ditolak ->", err)
	}
}

// ------------------------------------------------------------------
// 2. Kode yang diuji: layanan dengan dependency (bahan untuk mock)
// ------------------------------------------------------------------

// Notifier adalah "colokan" untuk mengirim pemberitahuan.
//
// 🔍 Analogi: interface ini seperti LUBANG STOPKONTAK. Di produksi kita colok mesin
// sungguhan (SMS/email gateway yang mahal & lambat). Di test kita colok mesin boongan
// (mock) yang cuma mencatat "aku dipanggil dengan pesan apa" — cepat dan tanpa biaya.
type Notifier interface {
	Kirim(ke, pesan string) error
}

// LayananBayar memakai Notifier lewat interface, bukan tipe konkret.
// Inilah yang membuatnya bisa diuji tanpa jaringan.
type LayananBayar struct {
	notifier Notifier
}

// NewLayananBayar menyuntikkan dependency lewat konstruktor (dependency injection).
func NewLayananBayar(n Notifier) *LayananBayar {
	return &LayananBayar{notifier: n}
}

// Bayar menarik saldo lalu memberitahu penerima.
//
// Perhatikan urutannya: saldo ditarik DULU, baru notifikasi dikirim. Kalau notifikasi
// gagal, saldo dikembalikan — inilah perilaku yang akan kita buktikan lewat mock.
func (s *LayananBayar) Bayar(d *Dompet, jumlah int, ke string) error {
	if ke == "" {
		return ErrPenerimaTakDikenal
	}
	if err := d.Tarik(jumlah); err != nil {
		return err
	}
	pesan := fmt.Sprintf("%s mengirim %d ke %s", d.Pemilik, jumlah, ke)
	if err := s.notifier.Kirim(ke, pesan); err != nil {
		d.Saldo += jumlah // kompensasi: kembalikan saldo bila notifikasi gagal
		return fmt.Errorf("notifikasi gagal, pembayaran dibatalkan: %w", err)
	}
	return nil
}

// NotifierKonsol implementasi nyata sederhana (dipakai saat go run).
type NotifierKonsol struct{}

func (NotifierKonsol) Kirim(ke, pesan string) error {
	fmt.Printf("   [notif -> %s] %s\n", ke, pesan)
	return nil
}

func demoPembayaran() {
	fmt.Println("\n-- Pembayaran (dependency lewat interface) --")
	d := NewDompet("Ana", 200_000)
	s := NewLayananBayar(NotifierKonsol{})

	if err := s.Bayar(d, 75_000, "Budi"); err != nil {
		fmt.Println("   gagal:", err)
	}
	fmt.Printf("   sisa saldo Ana = %d\n", d.Saldo)

	fmt.Println("\n   Di test, NotifierKonsol diganti MockNotifier — lihat testify_test.go.")
}
