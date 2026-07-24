// rs/zerolog — logging terstruktur (JSON) tanpa alokasi memori berlebih.
//
// Jalankan: go run ./libraries/zerolog
// Test:     go test ./libraries/zerolog
//
// 🔍 Analogi besar: log biasa (fmt.Println) itu seperti CATATAN TANGAN di buku tulis —
// enak dibaca manusia satu-satu, tapi coba cari "semua transaksi milik user 42 yang gagal
// kemarin" di antara 10 juta baris tulisan tangan. Log terstruktur itu FORMULIR ISIAN:
// tiap informasi punya kolomnya sendiri (user_id, status, durasi). Mesin (Elasticsearch,
// Loki, CloudWatch) bisa langsung memfilter kolom — pencarian yang tadinya mustahil jadi
// sekali klik. Harganya: log jadi kurang enak dibaca mata telanjang, makanya ada
// ConsoleWriter yang mempercantiknya khusus saat kamu ngoding di laptop.
package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/rs/zerolog"
)

func main() {
	fmt.Println("=== rs/zerolog ===")
	demoDasar()
	demoSubLogger()
	demoLevel()
	demoConsoleWriter()
	demoJebakan()
	demoBandingSlog()
}

// ------------------------------------------------------------------
// 1. Logger dasar
// ------------------------------------------------------------------

// NewLogger membuat logger JSON ke tujuan mana pun (io.Writer).
//
// 🔍 Analogi: menerima io.Writer, bukan langsung menulis ke layar, itu seperti mesin cetak
// yang bisa disambungkan ke kertas, ke layar, atau ke amplop. Di produksi tujuannya
// os.Stdout (lalu diambil agen log); di test tujuannya buffer di memori — sehingga kita
// bisa MEMERIKSA log yang dihasilkan, bukan sekadar berharap.
func NewLogger(w io.Writer, level zerolog.Level) zerolog.Logger {
	return zerolog.New(w).Level(level)
}

// NewLoggerProduksi versi lengkap dengan waktu — dipakai di aplikasi sungguhan.
func NewLoggerProduksi(w io.Writer) zerolog.Logger {
	return zerolog.New(w).
		Level(zerolog.InfoLevel).
		With().
		Timestamp().
		Logger()
}

// LogPesanan mencatat satu peristiwa bisnis dengan kolom-kolom terstruktur.
//
// Perhatikan pola berantai: Info() membuka peristiwa, .Str/.Int/.Dur menambah kolom,
// dan .Msg() MENUTUP sekaligus mengirimkannya.
func LogPesanan(l zerolog.Logger, orderID string, total int, durasi time.Duration) {
	l.Info().
		Str("order_id", orderID).
		Int("total", total).
		Dur("durasi", durasi).
		Msg("pesanan dibuat")
}

// LogGagal mencatat error. Err() menaruhnya di kolom baku "error".
func LogGagal(l zerolog.Logger, orderID string, err error) {
	l.Error().
		Str("order_id", orderID).
		Err(err).
		Msg("pesanan gagal diproses")
}

func demoDasar() {
	fmt.Println("\n-- Logger dasar (JSON) --")
	l := NewLogger(os.Stdout, zerolog.DebugLevel)
	LogPesanan(l, "ORD-001", 250_000, 42*time.Millisecond)
	LogGagal(l, "ORD-002", errors.New("stok habis"))
}

// ------------------------------------------------------------------
// 2. Sub-logger — konteks yang ikut ke mana-mana
// ------------------------------------------------------------------

// SubLogger membuat turunan logger yang SELALU membawa kolom tertentu.
//
// 🔍 Analogi: seperti KOP SURAT. Sekali kamu pasang kop "Divisi Pembayaran — No. Berkas 77",
// setiap surat yang dicetak dari mesin itu otomatis membawa kop tersebut. Kamu tak perlu
// menulis ulang request_id di 30 tempat berbeda — inilah cara melacak satu request
// dari awal sampai akhir di tumpukan log yang bercampur.
func SubLogger(l zerolog.Logger, layanan, requestID string) zerolog.Logger {
	return l.With().
		Str("layanan", layanan).
		Str("request_id", requestID).
		Logger()
}

func demoSubLogger() {
	fmt.Println("\n-- Sub-logger (kolom yang selalu ikut) --")
	induk := NewLogger(os.Stdout, zerolog.InfoLevel)
	req := SubLogger(induk, "checkout", "req-abc-123")

	req.Info().Msg("mulai proses")
	req.Info().Int("item", 3).Msg("keranjang divalidasi")
	req.Info().Msg("selesai")
}

// ------------------------------------------------------------------
// 3. Level — keran yang mengatur seberapa cerewet log-nya
// ------------------------------------------------------------------

// 🔍 Analogi level: seperti PENGATUR VOLUME radio. Debug = dengar semua bisikan (berguna
// saat mencari bug, tapi bikin tagihan penyimpanan log membengkak). Info = kejadian normal
// yang layak dicatat. Warn = "ada yang aneh tapi masih jalan". Error = "ini gagal, seseorang
// perlu tahu". Di produksi biasanya Info; saat menyelidiki masalah, dinaikkan sementara ke Debug.

// LevelDariString mengubah teks konfigurasi jadi level zerolog.
// Kalau tak dikenali, jatuh ke InfoLevel — pilihan aman ketimbang mematikan log.
//
// Catatan: zerolog.ParseLevel TIDAK peka huruf besar/kecil ("TRACE" = "trace"),
// tapi tetap peka spasi (" info" ditolak). Karena itu nilai dari file konfigurasi
// sebaiknya di-trim dulu sebelum masuk ke sini.
func LevelDariString(s string) zerolog.Level {
	lvl, err := zerolog.ParseLevel(s)
	if err != nil || lvl == zerolog.NoLevel {
		return zerolog.InfoLevel
	}
	return lvl
}

// TulisSemuaLevel menulis satu baris untuk tiap level — dipakai membuktikan penyaringan.
func TulisSemuaLevel(l zerolog.Logger) {
	l.Debug().Msg("pesan debug")
	l.Info().Msg("pesan info")
	l.Warn().Msg("pesan warn")
	l.Error().Msg("pesan error")
}

func demoLevel() {
	fmt.Println("\n-- Penyaringan level (logger di level Warn) --")
	l := NewLogger(os.Stdout, zerolog.WarnLevel)
	TulisSemuaLevel(l)
	fmt.Println("   (debug & info tidak muncul — tersaring sebelum dibentuk)")
}

// ------------------------------------------------------------------
// 4. ConsoleWriter — JSON yang dipercantik untuk mata manusia
// ------------------------------------------------------------------

// 🔍 Analogi: ConsoleWriter itu "mode baca" — mengubah formulir JSON yang padat jadi baris
// berwarna yang enak dibaca. HANYA untuk mode pengembangan: ia jauh lebih lambat dan
// hasilnya tak bisa diurai mesin. Di produksi, tetap JSON polos.
func NewLoggerDev(w io.Writer) zerolog.Logger {
	cw := zerolog.ConsoleWriter{
		Out:        w,
		TimeFormat: time.Kitchen,
		NoColor:    true, // warna dimatikan agar keluaran tetap rapi bila dialihkan ke file
	}
	return zerolog.New(cw).With().Timestamp().Logger()
}

func demoConsoleWriter() {
	fmt.Println("\n-- ConsoleWriter (khusus mode ngoding) --")
	// Logger disimpan ke variabel dulu: method zerolog memakai pointer receiver,
	// jadi tak bisa dipanggil langsung pada nilai balik fungsi.
	dev := NewLoggerDev(os.Stdout)
	dev.Info().
		Str("order_id", "ORD-003").
		Int("total", 99_000).
		Msg("pesanan dibuat")
}

// ------------------------------------------------------------------
// 5. Jebakan paling sering: lupa .Msg()
// ------------------------------------------------------------------

// 🔍 Analogi: rantai Info().Str(...) itu seperti MENGISI FORMULIR. Formulir yang sudah diisi
// tapi tak pernah dimasukkan ke kotak pos ya tak sampai ke mana-mana. .Msg() (atau .Send())
// adalah tindakan memasukkannya ke kotak pos. Ini jebakan nomor satu pemakai zerolog:
// kompiler TIDAK akan menegurmu — log-nya cuma diam-diam hilang.

// LogTanpaMsg sengaja lupa memanggil .Msg() — tidak ada apa pun yang tertulis.
func LogTanpaMsg(l zerolog.Logger) {
	l.Info().Str("penting", "data ini hilang") // <- tidak ada .Msg(), tidak terkirim
}

// LogDenganSend memakai .Send() bila kamu memang tak butuh teks pesan.
func LogDenganSend(l zerolog.Logger) {
	l.Info().Str("penting", "data ini terkirim").Send()
}

func demoJebakan() {
	fmt.Println("\n-- Jebakan: lupa .Msg() --")
	var buf bytes.Buffer
	l := NewLogger(&buf, zerolog.InfoLevel)

	LogTanpaMsg(l)
	fmt.Printf("   setelah LogTanpaMsg, panjang keluaran = %d byte (kosong!)\n", buf.Len())

	LogDenganSend(l)
	fmt.Printf("   setelah LogDenganSend  -> %s", buf.String())
}

// ------------------------------------------------------------------
// 6. Perbandingan dengan log/slog (bawaan Go 1.21+)
// ------------------------------------------------------------------

// 🔍 Analogi: slog itu "alat bawaan yang sudah ada di kotak perkakas rumah" — cukup untuk
// hampir semua pekerjaan, tanpa menambah dependensi. zerolog itu alat khusus yang lebih
// cepat & nyaris nol alokasi memori — berguna saat kamu menulis jutaan baris log per menit.
//
// Pilih slog kalau: proyek baru, ingin sedikit dependensi, performa log bukan leher botol.
// Pilih zerolog kalau: throughput log sangat tinggi, atau timmu sudah terbiasa dengannya.

// LogPesananSlog menulis peristiwa yang sama memakai stdlib, untuk dibandingkan.
func LogPesananSlog(w io.Writer, orderID string, total int) {
	l := slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{Level: slog.LevelInfo}))
	l.Info("pesanan dibuat", "order_id", orderID, "total", total)
}

func demoBandingSlog() {
	fmt.Println("\n-- zerolog vs log/slog (keluaran serupa) --")
	fmt.Print("   zerolog: ")
	zl := NewLogger(os.Stdout, zerolog.InfoLevel)
	zl.Info().
		Str("order_id", "ORD-004").Int("total", 10_000).Msg("pesanan dibuat")
	fmt.Print("   slog   : ")
	LogPesananSlog(os.Stdout, "ORD-004", 10_000)
}
