package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestAmbilBanyakMenjagaUrutan(t *testing.T) {
	ambil := func(_ context.Context, id int) (string, error) {
		// Sengaja: id kecil justru lebih lambat, supaya kalau urutan hasil bergantung
		// pada siapa yang selesai duluan, test ini akan gagal.
		time.Sleep(time.Duration(10-id) * time.Millisecond)
		return fmt.Sprintf("user-%d", id), nil
	}

	got, err := AmbilBanyak(context.Background(), []int{1, 2, 3, 4, 5}, ambil)
	if err != nil {
		t.Fatalf("AmbilBanyak error: %v", err)
	}

	want := []string{"user-1", "user-2", "user-3", "user-4", "user-5"}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("hasil[%d] = %q, ingin %q — urutan tidak terjaga", i, got[i], want[i])
		}
	}
}

func TestAmbilBanyakBenarBenarBersamaan(t *testing.T) {
	const n = 10
	const jeda = 30 * time.Millisecond

	ambil := func(_ context.Context, id int) (string, error) {
		time.Sleep(jeda)
		return "ok", nil
	}

	mulai := time.Now()
	if _, err := AmbilBanyak(context.Background(), make([]int, n), ambil); err != nil {
		t.Fatalf("error: %v", err)
	}
	lama := time.Since(mulai)

	// Kalau berurutan, ini butuh 300ms. Bersamaan seharusnya jauh lebih cepat.
	if lama > 5*jeda {
		t.Errorf("selesai dalam %v — sepertinya berjalan berurutan, bukan bersamaan", lama)
	}
}

func TestAmbilBanyakMengembalikanErrorPertama(t *testing.T) {
	sengaja := errors.New("sumber sedang mati")
	ambil := func(_ context.Context, id int) (string, error) {
		if id == 3 {
			return "", sengaja
		}
		return "ok", nil
	}

	hasil, err := AmbilBanyak(context.Background(), []int{1, 2, 3, 4}, ambil)
	if err == nil {
		t.Fatal("ingin error")
	}
	if !errors.Is(err, sengaja) {
		t.Errorf("error = %v, ingin membungkus error asli", err)
	}
	if !strings.Contains(err.Error(), "id 3") {
		t.Errorf("error = %q, ingin menyebut id yang gagal", err)
	}
	if hasil != nil {
		t.Error("saat gagal, hasil parsial tidak boleh dikembalikan")
	}
}

func TestAmbilBanyakDaftarKosong(t *testing.T) {
	got, err := AmbilBanyak(context.Background(), nil,
		func(context.Context, int) (string, error) { return "", nil })
	if err != nil {
		t.Fatalf("error tak terduga: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("ingin hasil kosong, dapat %v", got)
	}
}

// Inti WithContext: kegagalan satu pekerjaan membatalkan yang lain.
func TestGagalCepatMembatalkanSisanya(t *testing.T) {
	mulai := time.Now()
	selesai, err := AmbilDenganGagalCepat(context.Background(), 10, 0)
	lama := time.Since(mulai)

	if err == nil {
		t.Fatal("ingin error dari pekerjaan yang gagal")
	}
	// Tiap pekerjaan butuh 100ms; kalau pembatalan bekerja, totalnya jauh di bawah itu.
	if lama > 80*time.Millisecond {
		t.Errorf("selesai dalam %v — pembatalan sepertinya tidak bekerja", lama)
	}
	if selesai != 0 {
		t.Errorf("%d pekerjaan sempat selesai, ingin 0 (semua dibatalkan)", selesai)
	}
}

func TestGagalCepatTanpaKegagalan(t *testing.T) {
	// gagalDiIndeks di luar jangkauan -> tak ada yang gagal.
	_, err := AmbilDenganGagalCepat(context.Background(), 3, -1)
	if err != nil {
		t.Errorf("ingin sukses, dapat: %v", err)
	}
}

// Context yang sudah dibatalkan sejak awal harus menghentikan semuanya seketika.
func TestContextSudahDibatalkan(t *testing.T) {
	ctx, batal := context.WithCancel(context.Background())
	batal()

	_, err := AmbilDenganGagalCepat(ctx, 5, -1)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("error = %v, ingin context.Canceled", err)
	}
}

func TestSetLimitMenahanJumlahPekerja(t *testing.T) {
	tests := []struct {
		nama  string
		batas int
	}{
		{"satu pekerja (berurutan)", 1},
		{"tiga pekerja", 3},
		{"sepuluh pekerja", 10},
	}

	for _, tt := range tests {
		t.Run(tt.nama, func(t *testing.T) {
			puncak := PuncakBersamaan(30, tt.batas)
			if int(puncak) > tt.batas {
				t.Errorf("puncak %d pekerja bersamaan MELEBIHI batas %d", puncak, tt.batas)
			}
			if puncak < 1 {
				t.Errorf("puncak = %d, seharusnya minimal 1", puncak)
			}
		})
	}
}

// 100 permintaan serentak untuk kunci sama harus menghasilkan SATU panggilan sumber.
func TestSingleflightMenggabungkanPermintaan(t *testing.T) {
	p := NewPengambilData(func(kunci string) (string, error) {
		time.Sleep(50 * time.Millisecond)
		return "nilai-" + kunci, nil
	})

	const n = 100
	hasil := make([]string, n)

	var wg sync.WaitGroup
	for i := range n {
		wg.Add(1)
		go func() {
			defer wg.Done()
			v, err := p.Ambil("kunci-panas")
			if err != nil {
				t.Errorf("Ambil error: %v", err)
				return
			}
			hasil[i] = v
		}()
	}
	wg.Wait()

	if got := p.JumlahPanggilanSumber(); got != 1 {
		t.Errorf("sumber dihubungi %d kali, ingin 1 — inilah inti singleflight", got)
	}
	// Semua pemanggil harus menerima jawaban yang sama.
	for i, h := range hasil {
		if h != "nilai-kunci-panas" {
			t.Fatalf("hasil[%d] = %q, ingin nilai-kunci-panas", i, h)
		}
	}
}

// Kunci berbeda TIDAK digabungkan — masing-masing punya penerbangannya sendiri.
func TestSingleflightMemisahkanKunciBerbeda(t *testing.T) {
	p := NewPengambilData(func(kunci string) (string, error) {
		time.Sleep(20 * time.Millisecond)
		return "nilai-" + kunci, nil
	})

	var wg sync.WaitGroup
	for _, k := range []string{"a", "b", "c"} {
		for range 10 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, _ = p.Ambil(k)
			}()
		}
	}
	wg.Wait()

	if got := p.JumlahPanggilanSumber(); got != 3 {
		t.Errorf("sumber dihubungi %d kali, ingin 3 (satu per kunci)", got)
	}
}

// Panggilan berurutan (tidak tumpang tindih) TIDAK digabungkan.
// singleflight bukan cache — ia hanya menggabungkan yang sedang berlangsung.
func TestSingleflightBukanCache(t *testing.T) {
	p := NewPengambilData(func(kunci string) (string, error) {
		return "nilai", nil
	})

	for range 3 {
		if _, err := p.Ambil("sama"); err != nil {
			t.Fatalf("Ambil error: %v", err)
		}
	}

	if got := p.JumlahPanggilanSumber(); got != 3 {
		t.Errorf("sumber dihubungi %d kali, ingin 3 — singleflight bukan pengganti cache", got)
	}
}

func TestSingleflightMeneruskanError(t *testing.T) {
	sengaja := errors.New("sumber rusak")
	p := NewPengambilData(func(string) (string, error) { return "", sengaja })

	_, err := p.Ambil("apa saja")
	if !errors.Is(err, sengaja) {
		t.Errorf("error = %v, ingin %v", err, sengaja)
	}
}

func TestSemaphoreMembatasiTotalBobot(t *testing.T) {
	p := NewPengunggahBerkas(100)
	ctx := context.Background()

	var sedang, puncak atomic.Int64
	catat := func(mb int64) func() {
		return func() {
			kini := sedang.Add(mb)
			for {
				lama := puncak.Load()
				if kini <= lama || puncak.CompareAndSwap(lama, kini) {
					break
				}
			}
			time.Sleep(20 * time.Millisecond)
			sedang.Add(-mb)
		}
	}

	var wg sync.WaitGroup
	for _, mb := range []int64{60, 50, 30, 10, 40} {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := p.Unggah(ctx, mb, catat(mb)); err != nil {
				t.Errorf("Unggah(%d MB) error: %v", mb, err)
			}
		}()
	}
	wg.Wait()

	if got := puncak.Load(); got > 100 {
		t.Errorf("puncak pemakaian %d MB MELEBIHI jatah 100 MB", got)
	}
	if got := p.JumlahTerunggah(); got != 5 {
		t.Errorf("terunggah = %d, ingin 5 (semua akhirnya kebagian)", got)
	}
}

// Permintaan yang lebih besar dari kapasitas total harus DITOLAK, bukan menggantung.
func TestSemaphoreMenolakPermintaanTerlaluBesar(t *testing.T) {
	p := NewPengunggahBerkas(100)

	err := p.Unggah(context.Background(), 500, func() {
		t.Error("kerja tidak boleh dijalankan untuk permintaan yang ditolak")
	})
	if err == nil {
		t.Fatal("ingin error untuk berkas melebihi kapasitas")
	}
	if !strings.Contains(err.Error(), "melebihi kapasitas") {
		t.Errorf("error = %q, ingin menyebut kapasitas", err)
	}
}

func TestSemaphoreMenghormatiPembatalan(t *testing.T) {
	p := NewPengunggahBerkas(10)

	// Habiskan seluruh jatah dengan satu unggahan yang lambat.
	mulai := make(chan struct{})
	selesai := make(chan struct{})
	go func() {
		_ = p.Unggah(context.Background(), 10, func() {
			close(mulai)
			<-selesai
		})
	}()
	<-mulai

	// Pemanggil kedua tak akan kebagian jatah; context-nya keburu habis.
	ctx, batal := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer batal()

	err := p.Unggah(ctx, 5, func() { t.Error("seharusnya tak pernah berjalan") })
	close(selesai)

	if err == nil {
		t.Fatal("ingin error karena context habis saat menunggu jatah")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("error = %v, ingin context.DeadlineExceeded", err)
	}
}
