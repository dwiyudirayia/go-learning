// golang.org/x/sync — errgroup, singleflight, dan semaphore.
//
// Jalankan: go run ./libraries/errgroup
// Test:     go test -race ./libraries/errgroup
//
// 🔍 Analogi besar: sync.WaitGroup bawaan Go itu ABSENSI polos — ia cuma tahu "sudah
// pulang semua atau belum". Ia tak tahu apakah ada pegawai yang GAGAL mengerjakan
// tugasnya, dan tak bisa menyuruh yang lain berhenti kalau satu orang menemukan masalah.
//
// errgroup adalah KETUA TIM: ia menunggu semua anggota, mengumpulkan kegagalan pertama,
// dan (kalau diminta) meniup peluit agar semua berhenti begitu ada satu yang gagal.
//
// Catatan: modul 38 membahas pola-pola ini lebih dalam dari sisi konsep konkurensi.
// Berkas ini fokus pada RESEP SIAP PAKAI yang paling sering dibutuhkan sehari-hari.
package main

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
	"golang.org/x/sync/singleflight"
)

func main() {
	fmt.Println("=== golang.org/x/sync ===")
	demoErrgroup()
	demoGagalCepat()
	demoBatasKonkurensi()
	demoSingleflight()
	demoSemaphore()
}

// ------------------------------------------------------------------
// 1. errgroup — jalankan bersamaan, kumpulkan hasil & error
// ------------------------------------------------------------------

// AmbilBanyak memanggil 'ambil' untuk setiap id secara bersamaan.
//
// 🔍 Analogi hasil berurutan: perhatikan tiap goroutine menulis ke hasil[i] —
// SLOT-nya sendiri, seperti loker bernomor. Karena tak ada dua goroutine yang menyentuh
// slot yang sama, TIDAK PERLU mutex sama sekali, dan urutan hasilnya tetap sesuai
// urutan masukan. Ini pola yang jauh lebih rapi daripada mengunci satu slice bersama
// lalu mengurutkannya lagi belakangan.
func AmbilBanyak(ctx context.Context, ids []int, ambil func(context.Context, int) (string, error)) ([]string, error) {
	hasil := make([]string, len(ids))

	// WithContext memberi ctx turunan yang otomatis DIBATALKAN saat ada yang gagal.
	g, ctx := errgroup.WithContext(ctx)

	for i, id := range ids {
		g.Go(func() error {
			s, err := ambil(ctx, id)
			if err != nil {
				return fmt.Errorf("id %d: %w", id, err)
			}
			hasil[i] = s
			return nil
		})
	}

	// Wait mengembalikan error PERTAMA yang terjadi (nil bila semua sukses).
	if err := g.Wait(); err != nil {
		return nil, err
	}
	return hasil, nil
}

func demoErrgroup() {
	fmt.Println("\n-- errgroup: ambil banyak sekaligus --")

	ambil := func(_ context.Context, id int) (string, error) {
		time.Sleep(20 * time.Millisecond) // meniru panggilan jaringan
		return fmt.Sprintf("user-%d", id), nil
	}

	mulai := time.Now()
	hasil, err := AmbilBanyak(context.Background(), []int{1, 2, 3, 4, 5}, ambil)
	if err != nil {
		fmt.Println("   error:", err)
		return
	}
	fmt.Printf("   %v\n", hasil)
	fmt.Printf("   5 panggilan @20ms selesai dalam %v (bukan 100ms)\n",
		time.Since(mulai).Round(10*time.Millisecond))
}

// ------------------------------------------------------------------
// 2. Gagal cepat — satu gagal, sisanya dihentikan
// ------------------------------------------------------------------

// 🔍 Analogi: kamu memesan 5 komponen untuk merakit satu barang. Begitu diketahui
// komponen ke-2 habis stok, barang itu MUSTAHIL jadi — jadi buat apa menunggu 3 komponen
// lain selesai dikirim? errgroup.WithContext meniup peluit: ctx dibatalkan, dan pekerjaan
// yang menghormati ctx berhenti seketika. Ini menghemat waktu DAN kuota panggilan API.
//
// Syaratnya: fungsi pekerjamu HARUS memperhatikan ctx.Done(). Goroutine yang mengabaikan
// context tak bisa dihentikan oleh siapa pun — peluit ditiup, tapi ia tetap lari.

// AmbilDenganGagalCepat sama seperti AmbilBanyak, tapi mencatat berapa pekerjaan
// yang sempat berjalan — untuk membuktikan sisanya benar-benar dibatalkan.
func AmbilDenganGagalCepat(ctx context.Context, jumlah int, gagalDiIndeks int) (int32, error) {
	var selesai atomic.Int32

	g, ctx := errgroup.WithContext(ctx)
	for i := range jumlah {
		g.Go(func() error {
			if i == gagalDiIndeks {
				return fmt.Errorf("pekerjaan %d gagal", i)
			}
			// Menghormati pembatalan: berhenti begitu peluit ditiup.
			select {
			case <-time.After(100 * time.Millisecond):
				selesai.Add(1)
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		})
	}
	return selesai.Load(), g.Wait()
}

func demoGagalCepat() {
	fmt.Println("\n-- errgroup: gagal cepat --")

	mulai := time.Now()
	_, err := AmbilDenganGagalCepat(context.Background(), 10, 0)
	fmt.Printf("   error: %v\n", err)
	fmt.Printf("   10 pekerjaan @100ms berhenti dalam %v\n",
		time.Since(mulai).Round(10*time.Millisecond))
}

// ------------------------------------------------------------------
// 3. SetLimit — membatasi jumlah pekerja
// ------------------------------------------------------------------

// 🔍 Analogi: punya 10.000 URL untuk diambil bukan berarti kamu boleh membuka 10.000
// koneksi sekaligus. Itu seperti mengirim 10.000 orang serentak ke satu pintu toko —
// pintunya jebol (server tujuan kewalahan, atau kuota file descriptor-mu habis).
// SetLimit memasang PENJAGA PINTU: maksimal N orang di dalam, sisanya antre dengan tertib.

// PuncakBersamaan menjalankan 'jumlah' pekerjaan dengan batas 'batas' pekerja,
// lalu mengembalikan berapa banyak yang PERNAH berjalan bersamaan.
func PuncakBersamaan(jumlah, batas int) int32 {
	var sedangJalan, puncak atomic.Int32

	g := new(errgroup.Group)
	g.SetLimit(batas)

	for range jumlah {
		g.Go(func() error {
			kini := sedangJalan.Add(1)
			// Catat rekor tertinggi (loop compare-and-swap, aman dari balapan).
			for {
				lama := puncak.Load()
				if kini <= lama || puncak.CompareAndSwap(lama, kini) {
					break
				}
			}
			time.Sleep(5 * time.Millisecond)
			sedangJalan.Add(-1)
			return nil
		})
	}
	_ = g.Wait()
	return puncak.Load()
}

func demoBatasKonkurensi() {
	fmt.Println("\n-- errgroup: SetLimit --")
	for _, batas := range []int{1, 3, 10} {
		fmt.Printf("   batas %2d -> puncak pekerja bersamaan = %d\n",
			batas, PuncakBersamaan(30, batas))
	}
}

// ------------------------------------------------------------------
// 4. singleflight — satu pertanyaan, satu jawaban, dibagi ke semua
// ------------------------------------------------------------------

// 🔍 Analogi: bayangkan 1.000 orang di ruang tunggu serentak bertanya "sekarang jam
// berapa?". Tanpa singleflight, kamu menelepon operator 1.000 kali. Dengan singleflight,
// SATU orang menelepon, 999 lainnya menunggu sebentar lalu ikut mendengar jawabannya.
//
// Di dunia nyata masalah ini bernama CACHE STAMPEDE: satu item cache kedaluwarsa,
// lalu ribuan request serentak menyerbu database untuk mengisinya kembali — dan
// database yang tadinya sehat langsung tumbang.

// PengambilData membungkus sumber data lambat dengan pelindung singleflight.
type PengambilData struct {
	sf      singleflight.Group
	panggil atomic.Int32 // berapa kali sumber asli benar-benar dihubungi
	sumber  func(kunci string) (string, error)
}

func NewPengambilData(sumber func(string) (string, error)) *PengambilData {
	return &PengambilData{sumber: sumber}
}

// Ambil menjamin: untuk kunci yang sama, hanya SATU pemanggilan sumber yang berjalan
// pada satu waktu. Pemanggil lain ikut menunggu dan menerima hasil yang sama.
func (p *PengambilData) Ambil(kunci string) (string, error) {
	v, err, _ := p.sf.Do(kunci, func() (any, error) {
		p.panggil.Add(1)
		return p.sumber(kunci)
	})
	if err != nil {
		return "", err
	}
	s, ok := v.(string)
	if !ok {
		return "", errors.New("tipe hasil tak terduga dari singleflight")
	}
	return s, nil
}

// JumlahPanggilanSumber untuk membuktikan penggabungan benar-benar terjadi.
func (p *PengambilData) JumlahPanggilanSumber() int32 {
	return p.panggil.Load()
}

func demoSingleflight() {
	fmt.Println("\n-- singleflight --")

	p := NewPengambilData(func(kunci string) (string, error) {
		time.Sleep(50 * time.Millisecond) // sumber lambat
		return "nilai-" + kunci, nil
	})

	var wg sync.WaitGroup
	for range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = p.Ambil("kunci-panas")
		}()
	}
	wg.Wait()

	fmt.Printf("   100 permintaan serentak -> sumber dihubungi %d kali\n",
		p.JumlahPanggilanSumber())
}

// ------------------------------------------------------------------
// 5. semaphore — jatah berbobot
// ------------------------------------------------------------------

// 🔍 Analogi: SetLimit itu seperti "maksimal 5 ORANG di lift". semaphore.Weighted itu
// "maksimal 500 KILOGRAM di lift" — satu orang gemuk bisa memakan jatah tiga orang kurus.
// Pakai semaphore bila pekerjaanmu tak sama beratnya: mengunggah berkas 1 GB jelas
// memakan memori jauh lebih besar daripada berkas 1 MB, jadi tak adil bila keduanya
// dihitung "satu pekerjaan".

// PengunggahBerkas membatasi total MEMORI yang dipakai unggahan bersamaan,
// bukan jumlah unggahannya.
type PengunggahBerkas struct {
	sem       *semaphore.Weighted
	maksMB    int64
	terunggah atomic.Int32
}

func NewPengunggahBerkas(maksMB int64) *PengunggahBerkas {
	return &PengunggahBerkas{sem: semaphore.NewWeighted(maksMB), maksMB: maksMB}
}

// Unggah menahan jatah sebesar ukuran berkas selama proses berjalan.
func (p *PengunggahBerkas) Unggah(ctx context.Context, ukuranMB int64, kerja func()) error {
	if ukuranMB > p.maksMB {
		// Penting: permintaan yang lebih besar dari kapasitas total akan MENGGANTUNG
		// selamanya bila tidak ditolak lebih dulu — jatahnya takkan pernah cukup.
		return fmt.Errorf("berkas %d MB melebihi kapasitas total %d MB", ukuranMB, p.maksMB)
	}
	if err := p.sem.Acquire(ctx, ukuranMB); err != nil {
		return fmt.Errorf("gagal mendapat jatah: %w", err)
	}
	defer p.sem.Release(ukuranMB)

	kerja()
	p.terunggah.Add(1)
	return nil
}

func (p *PengunggahBerkas) JumlahTerunggah() int32 {
	return p.terunggah.Load()
}

func demoSemaphore() {
	fmt.Println("\n-- semaphore berbobot --")

	p := NewPengunggahBerkas(100) // total jatah 100 MB
	ctx := context.Background()

	var wg sync.WaitGroup
	for _, mb := range []int64{60, 50, 30, 10} {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = p.Unggah(ctx, mb, func() { time.Sleep(20 * time.Millisecond) })
		}()
	}
	wg.Wait()
	fmt.Printf("   4 berkas terunggah = %d (berbagi jatah 100 MB)\n", p.JumlahTerunggah())

	if err := p.Unggah(ctx, 500, func() {}); err != nil {
		fmt.Println("   berkas raksasa ->", err)
	}
}
