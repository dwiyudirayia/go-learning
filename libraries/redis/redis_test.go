package main

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
)

// cacheUji menyalakan Redis palsu dan mengembalikan cache yang siap dipakai.
// Keduanya otomatis ditutup saat test selesai.
func cacheUji(t *testing.T) (*Cache, *miniredis.Miniredis) {
	t.Helper()

	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("gagal menyalakan miniredis: %v", err)
	}
	t.Cleanup(mr.Close)

	c := NewCache(NewKlien(mr.Addr()))
	t.Cleanup(func() { _ = c.Tutup() })
	return c, mr
}

func TestSimpanDanAmbil(t *testing.T) {
	c, _ := cacheUji(t)
	ctx := context.Background()

	p := Produk{ID: 1, Nama: "Kopi Arabika", Harga: 85_000}
	if err := c.Simpan(ctx, p, time.Minute); err != nil {
		t.Fatalf("Simpan gagal: %v", err)
	}

	got, err := c.Ambil(ctx, 1)
	if err != nil {
		t.Fatalf("Ambil gagal: %v", err)
	}
	if got != p {
		t.Errorf("Ambil = %+v, ingin %+v", got, p)
	}
}

// Cache miss adalah kejadian NORMAL — harus bisa dibedakan dari kerusakan Redis.
func TestCacheMissBukanKerusakan(t *testing.T) {
	c, _ := cacheUji(t)

	_, err := c.Ambil(context.Background(), 999)
	if err == nil {
		t.Fatal("ingin error untuk kunci yang tak ada")
	}
	if !errors.Is(err, ErrTidakAdaDiCache) {
		t.Errorf("error = %v, ingin ErrTidakAdaDiCache (bukan kegagalan sistem)", err)
	}
}

func TestIsiCacheRusak(t *testing.T) {
	c, mr := cacheUji(t)

	// Tanam data yang bukan JSON langsung ke Redis palsu.
	if err := mr.Set(kunciProduk(7), "ini bukan json"); err != nil {
		t.Fatalf("penyiapan gagal: %v", err)
	}

	_, err := c.Ambil(context.Background(), 7)
	if err == nil {
		t.Fatal("ingin error untuk isi cache yang rusak")
	}
	// Rusak BUKAN sama dengan tidak ada — penanganannya berbeda.
	if errors.Is(err, ErrTidakAdaDiCache) {
		t.Error("isi rusak seharusnya tidak dilaporkan sebagai cache miss")
	}
}

// Inti cache-aside: permintaan pertama menyentuh database, sisanya tidak.
func TestCacheAsideMengurangiAksesDatabase(t *testing.T) {
	c, _ := cacheUji(t)
	ctx := context.Background()

	akses := 0
	gudang := func(_ context.Context, id int) (Produk, error) {
		akses++
		return Produk{ID: id, Nama: "Teh Melati", Harga: 35_000}, nil
	}

	for i := 1; i <= 5; i++ {
		p, dariCache, err := c.AmbilAtauMuat(ctx, 2, time.Minute, gudang)
		if err != nil {
			t.Fatalf("percobaan %d gagal: %v", i, err)
		}
		if p.Nama != "Teh Melati" {
			t.Errorf("percobaan %d: nama = %q", i, p.Nama)
		}
		// Yang pertama pasti dari gudang; sisanya harus dari cache.
		wantDariCache := i > 1
		if dariCache != wantDariCache {
			t.Errorf("percobaan %d: dariCache = %t, ingin %t", i, dariCache, wantDariCache)
		}
	}

	if akses != 1 {
		t.Errorf("database disentuh %d kali untuk 5 permintaan, ingin 1", akses)
	}
}

func TestCacheAsideMeneruskanErrorGudang(t *testing.T) {
	c, _ := cacheUji(t)
	sengaja := errors.New("database sedang mati")

	_, _, err := c.AmbilAtauMuat(context.Background(), 5, time.Minute,
		func(context.Context, int) (Produk, error) { return Produk{}, sengaja })

	if !errors.Is(err, sengaja) {
		t.Errorf("error = %v, ingin membungkus error gudang", err)
	}
}

// TTL diuji dengan mesin waktu miniredis — tanpa menunggu waktu nyata.
func TestTTLKedaluwarsa(t *testing.T) {
	c, mr := cacheUji(t)
	ctx := context.Background()

	p := Produk{ID: 3, Nama: "Roti Gandum", Harga: 25_000}
	if err := c.Simpan(ctx, p, 30*time.Second); err != nil {
		t.Fatalf("Simpan gagal: %v", err)
	}

	tests := []struct {
		nama      string
		maju      time.Duration
		wantMasih bool
	}{
		{"belum lewat", 10 * time.Second, true},
		{"mendekati batas", 19 * time.Second, true},
		{"lewat batas", 2 * time.Second, false},
	}

	for _, tt := range tests {
		t.Run(tt.nama, func(t *testing.T) {
			mr.FastForward(tt.maju)

			_, err := c.Ambil(ctx, 3)
			masih := err == nil
			if masih != tt.wantMasih {
				t.Errorf("masih ada = %t, ingin %t (err=%v)", masih, tt.wantMasih, err)
			}
		})
	}
}

func TestSisaMasaBerlaku(t *testing.T) {
	c, _ := cacheUji(t)
	ctx := context.Background()

	if err := c.Simpan(ctx, Produk{ID: 8, Nama: "X"}, time.Minute); err != nil {
		t.Fatalf("Simpan gagal: %v", err)
	}

	sisa, err := c.SisaMasaBerlaku(ctx, 8)
	if err != nil {
		t.Fatalf("SisaMasaBerlaku gagal: %v", err)
	}
	if sisa <= 0 || sisa > time.Minute {
		t.Errorf("sisa = %v, ingin di antara 0 dan 1 menit", sisa)
	}

	// Kunci yang tak ada dilaporkan sebagai -2 (konvensi Redis), bukan error.
	if got, _ := c.SisaMasaBerlaku(ctx, 12345); got >= 0 {
		t.Errorf("TTL kunci tak ada = %v, ingin negatif", got)
	}
}

func TestHapusMembatalkanCache(t *testing.T) {
	c, _ := cacheUji(t)
	ctx := context.Background()

	if err := c.Simpan(ctx, Produk{ID: 4, Nama: "Keju", Harga: 120_000}, time.Minute); err != nil {
		t.Fatalf("Simpan gagal: %v", err)
	}
	if err := c.Hapus(ctx, 4); err != nil {
		t.Fatalf("Hapus gagal: %v", err)
	}

	if _, err := c.Ambil(ctx, 4); !errors.Is(err, ErrTidakAdaDiCache) {
		t.Errorf("setelah dihapus, Ambil = %v, ingin cache miss", err)
	}

	// Menghapus kunci yang memang tak ada bukan error (idempoten).
	if err := c.Hapus(ctx, 4); err != nil {
		t.Errorf("Hapus kedua kali seharusnya aman, dapat: %v", err)
	}
}

// Setelah dibatalkan, permintaan berikutnya harus membaca data TERBARU dari gudang.
func TestPembatalanMembuatDataSegarTerbaca(t *testing.T) {
	c, _ := cacheUji(t)
	ctx := context.Background()

	harga := 100
	gudang := func(_ context.Context, id int) (Produk, error) {
		return Produk{ID: id, Nama: "Berubah", Harga: harga}, nil
	}

	p, _, _ := c.AmbilAtauMuat(ctx, 9, time.Minute, gudang)
	if p.Harga != 100 {
		t.Fatalf("harga awal = %d, ingin 100", p.Harga)
	}

	// Harga berubah di database, TAPI cache masih menyimpan yang lama.
	harga = 250
	p, dariCache, _ := c.AmbilAtauMuat(ctx, 9, time.Minute, gudang)
	if !dariCache || p.Harga != 100 {
		t.Errorf("tanpa pembatalan seharusnya masih data lama, dapat harga=%d dariCache=%t",
			p.Harga, dariCache)
	}

	// Setelah dibatalkan, barulah harga baru terbaca.
	if err := c.Hapus(ctx, 9); err != nil {
		t.Fatalf("Hapus gagal: %v", err)
	}
	p, dariCache, _ = c.AmbilAtauMuat(ctx, 9, time.Minute, gudang)
	if dariCache || p.Harga != 250 {
		t.Errorf("setelah pembatalan: harga=%d dariCache=%t, ingin 250 & false", p.Harga, dariCache)
	}
}

func TestPenghitungNaikBerurutan(t *testing.T) {
	c, _ := cacheUji(t)
	ctx := context.Background()

	for i := int64(1); i <= 5; i++ {
		got, err := c.Naikkan(ctx, "kunjungan", time.Hour)
		if err != nil {
			t.Fatalf("Naikkan gagal: %v", err)
		}
		if got != i {
			t.Errorf("nilai penghitung = %d, ingin %d", got, i)
		}
	}
}

// TTL hanya dipasang saat kunci pertama kali dibuat, supaya jendela hitungnya
// tidak diperpanjang terus setiap ada kenaikan.
func TestPenghitungTTLTidakDiperpanjang(t *testing.T) {
	c, mr := cacheUji(t)
	ctx := context.Background()

	if _, err := c.Naikkan(ctx, "harian", 10*time.Second); err != nil {
		t.Fatalf("Naikkan gagal: %v", err)
	}
	mr.FastForward(6 * time.Second)
	if _, err := c.Naikkan(ctx, "harian", 10*time.Second); err != nil {
		t.Fatalf("Naikkan gagal: %v", err)
	}

	// Kalau TTL diperpanjang tiap kenaikan, kunci masih hidup di detik ke-11.
	mr.FastForward(5 * time.Second)
	if mr.Exists("harian") {
		t.Error("kunci masih ada — TTL sepertinya ikut diperpanjang setiap kenaikan")
	}
}

func TestSimpanBanyakLewatPipeline(t *testing.T) {
	c, _ := cacheUji(t)
	ctx := context.Background()

	batch := []Produk{
		{ID: 10, Nama: "Gelas", Harga: 45_000},
		{ID: 11, Nama: "Piring", Harga: 55_000},
		{ID: 12, Nama: "Sendok", Harga: 15_000},
	}
	if err := c.SimpanBanyak(ctx, batch, time.Minute); err != nil {
		t.Fatalf("SimpanBanyak gagal: %v", err)
	}

	for _, want := range batch {
		got, err := c.Ambil(ctx, want.ID)
		if err != nil {
			t.Errorf("produk %d tidak tersimpan: %v", want.ID, err)
			continue
		}
		if got != want {
			t.Errorf("produk %d = %+v, ingin %+v", want.ID, got, want)
		}
	}
}

func TestSimpanBanyakKosong(t *testing.T) {
	c, _ := cacheUji(t)
	if err := c.SimpanBanyak(context.Background(), nil, time.Minute); err != nil {
		t.Errorf("pipeline kosong seharusnya aman, dapat: %v", err)
	}
}

// Kalau Redis mati, aplikasi harus tetap melayani lewat gudang — hanya lebih lambat.
func TestRedisMatiTidakMerobohkanAplikasi(t *testing.T) {
	c, mr := cacheUji(t)
	ctx := context.Background()

	mr.Close() // Redis "mati" mendadak

	p, dariCache, err := c.AmbilAtauMuat(ctx, 1, time.Minute,
		func(_ context.Context, id int) (Produk, error) {
			return Produk{ID: id, Nama: "Dari gudang", Harga: 1}, nil
		})

	if err != nil {
		t.Fatalf("aplikasi seharusnya tetap jalan saat Redis mati, dapat: %v", err)
	}
	if dariCache {
		t.Error("dariCache = true padahal Redis mati")
	}
	if p.Nama != "Dari gudang" {
		t.Errorf("nama = %q, ingin diambil dari gudang", p.Nama)
	}
}

func TestKunciProdukMemakaiNamespace(t *testing.T) {
	if got := kunciProduk(42); got != "toko:produk:42" {
		t.Errorf("kunciProduk(42) = %q, ingin toko:produk:42", got)
	}
}
