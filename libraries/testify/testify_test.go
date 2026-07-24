package main

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// ------------------------------------------------------------------
// 1. assert vs require — perbedaan yang paling sering salah dipakai
// ------------------------------------------------------------------

// 🔍 Analogi: assert = MENCATAT pelanggaran lalu inspeksi jalan terus (test lanjut ke baris
// berikutnya). require = MENGHENTIKAN inspeksi di tempat (test berhenti seketika).
//
// Aturan praktis: pakai require untuk hal yang kalau gagal membuat baris berikutnya
// tak masuk akal (mis. err harus nil sebelum kamu memakai hasilnya — kalau tetap lanjut,
// kamu akan kena nil pointer dan pesan errornya jadi membingungkan). Pakai assert untuk
// pemeriksaan mandiri yang tak saling bergantung.

func TestAssertVsRequire(t *testing.T) {
	d := NewDompet("Ana", 100)

	// require: kalau ini gagal, tak ada gunanya memeriksa saldo di bawahnya.
	err := d.Setor(50)
	require.NoError(t, err, "setor jumlah sah tidak boleh error")

	// assert: pemeriksaan mandiri, boleh lanjut walau salah satu gagal.
	assert.Equal(t, 150, d.Saldo, "saldo harus bertambah 50")
	assert.Equal(t, "Ana", d.Pemilik)
	assert.Positive(t, d.Saldo)
}

// ------------------------------------------------------------------
// 2. Table-driven test — pola idiomatik Go, dipercantik testify
// ------------------------------------------------------------------

func TestTarik(t *testing.T) {
	tests := []struct {
		nama      string
		saldoAwal int
		jumlah    int
		wantErr   error
		wantSaldo int
	}{
		{"penarikan normal", 100_000, 30_000, nil, 70_000},
		{"tarik tepat sejumlah saldo", 50_000, 50_000, nil, 0},
		{"saldo kurang", 10_000, 30_000, ErrSaldoKurang, 10_000},
		{"jumlah nol ditolak", 10_000, 0, ErrJumlahTidakValid, 10_000},
		{"jumlah negatif ditolak", 10_000, -5, ErrJumlahTidakValid, 10_000},
	}

	for _, tt := range tests {
		t.Run(tt.nama, func(t *testing.T) {
			d := NewDompet("Uji", tt.saldoAwal)
			err := d.Tarik(tt.jumlah)

			if tt.wantErr != nil {
				// ErrorIs menembus rantai pembungkus, sama seperti errors.Is.
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
			// Saldo harus tetap utuh saat penarikan ditolak.
			assert.Equal(t, tt.wantSaldo, d.Saldo)
		})
	}
}

// ------------------------------------------------------------------
// 3. Mock — mengganti dependency dengan boneka yang bisa diinterogasi
// ------------------------------------------------------------------

// MockNotifier menyematkan mock.Mock (embedding), sehingga otomatis mewarisi
// kemampuan mencatat panggilan, mengatur nilai balik, dan memverifikasi harapan.
//
// 🔍 Analogi: ini seperti AKTOR PENGGANTI di syuting. Ia tak benar-benar melompat dari
// gedung (tak benar-benar mengirim SMS), tapi kamu bisa bertanya sesudahnya: "tadi kamu
// dipanggil berapa kali? adegannya persis seperti skenario?"
type MockNotifier struct {
	mock.Mock
}

func (m *MockNotifier) Kirim(ke, pesan string) error {
	// Called mencatat argumen yang masuk & mengambil nilai balik yang sudah diatur.
	args := m.Called(ke, pesan)
	return args.Error(0)
}

func TestBayarSuksesMemanggilNotifier(t *testing.T) {
	m := new(MockNotifier)
	// On menetapkan harapan: dipanggil dengan "Budi" + string apa pun, balikkan nil.
	m.On("Kirim", "Budi", mock.AnythingOfType("string")).Return(nil).Once()

	d := NewDompet("Ana", 100_000)
	err := NewLayananBayar(m).Bayar(d, 40_000, "Budi")

	require.NoError(t, err)
	assert.Equal(t, 60_000, d.Saldo)
	// AssertExpectations gagal bila ada harapan yang tak terpenuhi (mis. tak pernah dipanggil).
	m.AssertExpectations(t)
}

func TestBayarMengembalikanSaldoSaatNotifikasiGagal(t *testing.T) {
	gagalKirim := errors.New("gateway sedang mati")

	m := new(MockNotifier)
	m.On("Kirim", mock.Anything, mock.Anything).Return(gagalKirim)

	d := NewDompet("Ana", 100_000)
	err := NewLayananBayar(m).Bayar(d, 40_000, "Budi")

	require.Error(t, err)
	assert.ErrorIs(t, err, gagalKirim, "error asli harus tetap bisa dilacak")
	// Inti pengujian: saldo dikembalikan utuh karena pembayaran dibatalkan.
	assert.Equal(t, 100_000, d.Saldo)
}

func TestBayarTidakMemanggilNotifierSaatSaldoKurang(t *testing.T) {
	m := new(MockNotifier)
	// Sengaja TIDAK memasang harapan apa pun.

	d := NewDompet("Ana", 10_000)
	err := NewLayananBayar(m).Bayar(d, 40_000, "Budi")

	assert.ErrorIs(t, err, ErrSaldoKurang)
	// Membuktikan notifier sama sekali tak tersentuh saat validasi gagal lebih dulu.
	m.AssertNotCalled(t, "Kirim", mock.Anything, mock.Anything)
}

// ------------------------------------------------------------------
// 4. Suite — sekelompok test yang berbagi persiapan
// ------------------------------------------------------------------

// 🔍 Analogi: suite itu seperti DAPUR yang dibersihkan ulang sebelum tiap resep dicoba.
// SetupTest berjalan sebelum SETIAP test method, jadi tiap test mulai dari keadaan bersih
// tanpa kamu menulis ulang kode persiapan. Berguna saat persiapannya panjang (koneksi DB,
// server palsu). Untuk test sederhana, fungsi biasa + helper sudah cukup — jangan berlebihan.

type DompetSuite struct {
	suite.Suite
	dompet  *Dompet
	mock    *MockNotifier
	layanan *LayananBayar
}

// SetupTest dijalankan ulang sebelum tiap test method di suite ini.
func (s *DompetSuite) SetupTest() {
	s.dompet = NewDompet("Ana", 100_000)
	s.mock = new(MockNotifier)
	s.layanan = NewLayananBayar(s.mock)
}

func (s *DompetSuite) TestSetorMenambahSaldo() {
	s.Require().NoError(s.dompet.Setor(25_000))
	s.Equal(125_000, s.dompet.Saldo)
}

func (s *DompetSuite) TestBayarMenguranginSaldo() {
	s.mock.On("Kirim", mock.Anything, mock.Anything).Return(nil)

	s.Require().NoError(s.layanan.Bayar(s.dompet, 25_000, "Budi"))
	s.Equal(75_000, s.dompet.Saldo)
}

// Bukti isolasi: test ini melihat saldo 100.000 lagi, bukan sisa dari test sebelumnya.
func (s *DompetSuite) TestTiapTestMulaiBersih() {
	s.Equal(100_000, s.dompet.Saldo, "SetupTest harus mengulang keadaan awal")
}

func (s *DompetSuite) TestPenerimaKosongDitolak() {
	err := s.layanan.Bayar(s.dompet, 10_000, "")
	s.ErrorIs(err, ErrPenerimaTakDikenal)
	s.Equal(100_000, s.dompet.Saldo)
}

// Satu fungsi Test* biasa sebagai pintu masuk suite ke `go test`.
func TestDompetSuite(t *testing.T) {
	suite.Run(t, new(DompetSuite))
}
