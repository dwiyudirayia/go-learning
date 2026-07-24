// robfig/cron — menjadwalkan pekerjaan berulang di dalam aplikasi Go.
//
// Jalankan: go run ./libraries/cron
// Test:     go test ./libraries/cron
//
// 🔍 Analogi besar: cron itu ALARM BERULANG di ponselmu. Kamu tak perlu mengingat-ingat
// "jam 7 pagi bangunkan aku" setiap hari — cukup pasang sekali, alarmnya yang mengingat.
// Bedanya, alarm cron hidup DI DALAM proses aplikasimu: begitu aplikasi mati, alarmnya
// ikut mati. Itu pembeda penting dari crontab sistem operasi (yang hidup di luar aplikasi)
// dan dari penjadwal terdistribusi seperti Kubernetes CronJob.
//
// Konsekuensinya (penting!): kalau aplikasimu jalan di 3 replika, alarm yang sama akan
// berbunyi 3 KALI — tiap replika punya penjadwalnya sendiri. Bagian terakhir file ini
// membahas cara mengatasinya.
package main

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/robfig/cron/v3"
)

func main() {
	fmt.Println("=== robfig/cron ===")
	demoFormatSpec()
	demoJadwalBerikutnya()
	demoMenjalankan()
	demoPelindungJob()
	demoBanyakReplika()
}

// ------------------------------------------------------------------
// 1. Format spec — membaca "* * * * *"
// ------------------------------------------------------------------

// 🔍 Analogi: spec cron itu seperti FORMULIR JADWAL dengan lima kotak, dibaca kiri ke kanan
// dari satuan terkecil ke terbesar:
//
//	┌──────── menit        (0-59)
//	│ ┌────── jam          (0-23)
//	│ │ ┌──── tanggal      (1-31)
//	│ │ │ ┌── bulan        (1-12)
//	│ │ │ │ ┌ hari pekan   (0-6, Minggu=0)
//	* * * * *
//
// Tanda bintang berarti "setiap". Jadi "0 3 * * *" = menit 0, jam 3, setiap tanggal,
// setiap bulan, setiap hari = "tiap hari jam 3 pagi".
//
// Pintasan yang jauh lebih mudah dibaca: "@hourly", "@daily", "@weekly", "@monthly",
// dan "@every 30m". Pakai pintasan ini kalau memang cukup — kode yang bisa dibaca
// rekan setim lebih berharga daripada spec yang pintar.

// ValidasiSpec memeriksa apakah sebuah spec cron sah, tanpa menjalankan apa pun.
//
// Berguna untuk memvalidasi jadwal yang diinput pengguna di halaman admin — kamu ingin
// menolaknya saat disimpan, bukan saat alarm gagal berbunyi tengah malam.
func ValidasiSpec(spec string) error {
	if _, err := cron.ParseStandard(spec); err != nil {
		return fmt.Errorf("spec cron %q tidak valid: %w", spec, err)
	}
	return nil
}

func demoFormatSpec() {
	fmt.Println("\n-- Validasi spec --")
	for _, s := range []string{
		"0 3 * * *",      // tiap hari jam 3 pagi
		"*/15 * * * *",   // tiap 15 menit
		"0 9 * * 1-5",    // jam 9 pagi, Senin s.d. Jumat
		"@daily",         // pintasan
		"@every 1h30m",   // interval bebas
		"tiap hari pagi", // tidak sah
		"99 * * * *",     // menit 99 tidak ada
	} {
		if err := ValidasiSpec(s); err != nil {
			fmt.Printf("   %-16q DITOLAK\n", s)
			continue
		}
		fmt.Printf("   %-16q sah\n", s)
	}
}

// ------------------------------------------------------------------
// 2. Menghitung jadwal berikutnya (tanpa menunggu!)
// ------------------------------------------------------------------

// 🔍 Analogi: ini seperti bertanya ke alarm "kalau sekarang jam 10 pagi, kapan kamu akan
// berbunyi 3 kali berikutnya?" — tanpa harus benar-benar menunggu sampai besok.
//
// Inilah kunci MENGUJI penjadwal: jangan pernah menulis test yang tidur 60 detik menunggu
// job berjalan (lambat dan rapuh). Uji saja perhitungan jadwalnya — itu murni fungsi
// matematika waktu, cepat dan pasti.

// JadwalBerikutnya menghitung n waktu eksekusi berikutnya setelah waktu 'dari'.
func JadwalBerikutnya(spec string, dari time.Time, n int) ([]time.Time, error) {
	jadwal, err := cron.ParseStandard(spec)
	if err != nil {
		return nil, fmt.Errorf("spec cron %q tidak valid: %w", spec, err)
	}

	hasil := make([]time.Time, 0, n)
	t := dari
	for i := 0; i < n; i++ {
		t = jadwal.Next(t)
		hasil = append(hasil, t)
	}
	return hasil, nil
}

func demoJadwalBerikutnya() {
	fmt.Println("\n-- Kapan berikutnya berjalan? --")
	dari := time.Date(2026, 7, 23, 10, 30, 0, 0, time.UTC) // Kamis
	for _, spec := range []string{"0 3 * * *", "*/15 * * * *", "0 9 * * 1-5"} {
		waktu, err := JadwalBerikutnya(spec, dari, 3)
		if err != nil {
			fmt.Println("   error:", err)
			continue
		}
		fmt.Printf("   %-14s ->", spec)
		for _, w := range waktu {
			fmt.Printf("  %s", w.Format("Mon 02 Jan 15:04"))
		}
		fmt.Println()
	}
}

// ------------------------------------------------------------------
// 3. Menjalankan penjadwal sungguhan
// ------------------------------------------------------------------

// 🔍 Analogi WithSeconds: formulir cron standar kotak terkecilnya adalah MENIT — tak ada
// kotak untuk detik, karena cron lahir di dunia tugas pemeliharaan harian. cron.WithSeconds()
// menambahkan satu kotak di paling depan, sehingga specnya jadi enam kotak: "*/2 * * * * *"
// artinya "tiap 2 detik". Pakai hanya kalau memang perlu — kalau intervalnya sedetik-dua
// detik, biasanya time.Ticker lebih tepat daripada cron.

// PenghitungJob adalah job sederhana yang mencatat berapa kali ia dipanggil.
// Memakai atomic karena job cron berjalan di goroutine terpisah.
type PenghitungJob struct {
	jumlah atomic.Int64
}

func (p *PenghitungJob) Jalankan() {
	p.jumlah.Add(1)
}

func (p *PenghitungJob) Jumlah() int64 {
	return p.jumlah.Load()
}

func demoMenjalankan() {
	fmt.Println("\n-- Menjalankan penjadwal (3 detik) --")

	job := &PenghitungJob{}
	// WithSeconds mengaktifkan spec 6 kotak; DiscardLogger membungkam log bawaan cron.
	c := cron.New(cron.WithSeconds(), cron.WithLogger(cron.DiscardLogger))

	id, err := c.AddFunc("*/1 * * * * *", job.Jalankan) // tiap detik
	if err != nil {
		fmt.Println("   gagal menjadwalkan:", err)
		return
	}

	c.Start()
	time.Sleep(3200 * time.Millisecond)

	// Stop mengembalikan context yang selesai setelah job yang SEDANG jalan rampung.
	// 🔍 Analogi: seperti menutup toko — pintu dikunci untuk pelanggan baru, tapi
	// pelanggan yang sudah di dalam tetap dilayani sampai selesai.
	ctx := c.Stop()
	<-ctx.Done()

	fmt.Printf("   job berjalan %d kali dalam ~3 detik\n", job.Jumlah())
	fmt.Printf("   entry id = %d, sisa entri setelah Stop = %d\n", id, len(c.Entries()))
}

// ------------------------------------------------------------------
// 4. Pelindung job: Recover, SkipIfStillRunning, DelayIfStillRunning
// ------------------------------------------------------------------

// 🔍 Analogi tiga pelindung — bayangkan petugas kebersihan yang datang tiap 5 menit:
//
//	Recover              = JARING PENGAMAN. Kalau petugas terpeleset (job panic), tanpa ini
//	                       SELURUH aplikasi ikut roboh. Hampir selalu pantas dipasang.
//	SkipIfStillRunning   = "kalau petugas sebelumnya BELUM selesai, jadwal kali ini
//	                       DILEWATI saja." Cocok untuk tugas yang tak apa-apa terlewat,
//	                       mis. menyegarkan cache.
//	DelayIfStillRunning  = "tunggu petugas sebelumnya selesai, baru mulai." Antreannya
//	                       menumpuk, tapi tak ada jadwal yang hilang. Cocok untuk tugas
//	                       yang wajib dijalankan, mis. mengirim tagihan.
//
// Jebakan yang sering menggigit: TANPA salah satu dari dua yang terakhir, job yang lambat
// akan TUMPANG TINDIH dengan dirinya sendiri. Job "sinkronisasi 10 menit" yang ternyata
// butuh 15 menit akan punya dua salinan berjalan bersamaan, saling rebut data.

// NewPenjadwalAman membuat penjadwal dengan pelindung yang masuk akal sebagai default.
func NewPenjadwalAman() *cron.Cron {
	return cron.New(
		cron.WithLogger(cron.DiscardLogger),
		cron.WithChain(
			cron.Recover(cron.DiscardLogger), // job panic tak merobohkan aplikasi
			cron.SkipIfStillRunning(cron.DiscardLogger),
		),
	)
}

func demoPelindungJob() {
	fmt.Println("\n-- Pelindung job --")

	c := NewPenjadwalAman()
	if _, err := c.AddFunc("@every 1h", func() {
		panic("job ini sengaja rusak")
	}); err != nil {
		fmt.Println("   gagal menjadwalkan:", err)
		return
	}
	fmt.Println("   penjadwal dibuat dengan Recover + SkipIfStillRunning")
	fmt.Println("   -> job yang panic tidak akan merobohkan aplikasi")
	fmt.Println("   -> job yang lambat tidak akan tumpang tindih dengan dirinya sendiri")
}

// ------------------------------------------------------------------
// 5. Jebakan produksi: zona waktu & banyak replika
// ------------------------------------------------------------------

// 🔍 Analogi zona waktu: "tiap hari jam 3 pagi" itu jam 3 pagi DI MANA? Server cloud
// hampir selalu berjalan dalam UTC, sedangkan laporan harian yang diminta bos maksudnya
// jam 3 pagi WIB. Selisihnya 7 jam — laporan "harian" jadi memuat rentang tanggal yang
// salah. Selalu SEBUTKAN zona waktunya secara eksplisit, jangan mengandalkan zona server.

// NewPenjadwalWIB membuat penjadwal yang jam-jamnya dihitung dalam waktu Jakarta.
func NewPenjadwalWIB() (*cron.Cron, error) {
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		// Basis data zona waktu bisa tak tersedia di image container yang sangat ramping
		// (mis. scratch/distroless). Solusinya: impor paket time/tzdata di aplikasi.
		return nil, fmt.Errorf("zona waktu Asia/Jakarta tak tersedia: %w", err)
	}
	return cron.New(cron.WithLocation(loc), cron.WithLogger(cron.DiscardLogger)), nil
}

// 🔍 Analogi banyak replika: kamu memasang alarm yang sama di 3 ponsel berbeda, lalu heran
// kenapa tagihan terkirim 3 kali ke pelanggan. Penjadwal in-process TIDAK tahu ada replika
// lain. Tiga cara umum mengatasinya:
//
//  1. Kunci terdistribusi (mis. SETNX di Redis dengan TTL): replika yang berhasil merebut
//     kunci saja yang menjalankan job. Paling praktis bila kamu sudah punya Redis.
//  2. Serahkan ke platform: Kubernetes CronJob menjalankan pod terpisah — penjadwalannya
//     bukan lagi urusan aplikasimu.
//  3. Jadikan job-nya idempoten + tandai di database ("laporan 2026-07-23 sudah dibuat"),
//     sehingga dijalankan berulang pun aman.

// BolehJalan meniru pemeriksaan kunci terdistribusi.
// Di produksi, fungsi ini akan memanggil Redis SETNX atau tabel kunci di database.
func BolehJalan(rebutKunci func(kunci string) bool, namaJob string) bool {
	return rebutKunci("cron:" + namaJob)
}

func demoBanyakReplika() {
	fmt.Println("\n-- Jebakan produksi --")

	if _, err := NewPenjadwalWIB(); err != nil {
		fmt.Println("   zona waktu:", err)
	} else {
		fmt.Println("   zona waktu: penjadwal WIB berhasil dibuat (Asia/Jakarta)")
	}

	// Peragaan kunci: hanya replika pertama yang menang.
	sudahDiambil := map[string]bool{}
	kunci := func(k string) bool {
		if sudahDiambil[k] {
			return false
		}
		sudahDiambil[k] = true
		return true
	}
	for i := 1; i <= 3; i++ {
		fmt.Printf("   replika %d boleh menjalankan 'kirim-tagihan'? %t\n",
			i, BolehJalan(kunci, "kirim-tagihan"))
	}
}
