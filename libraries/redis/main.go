// redis/go-redis — klien Redis, dengan miniredis sebagai Redis palsu untuk contoh & test.
//
// Jalankan: go run ./libraries/redis     (memakai miniredis — tanpa perlu Redis sungguhan)
// Test:     go test ./libraries/redis
//
// 🔍 Analogi besar: database itu GUDANG BESAR di lantai bawah — muat segalanya, tapi
// mengambil sesuatu dari sana butuh waktu. Redis itu MEJA KERJA di sebelahmu: kecil,
// isinya cuma yang sedang sering dipakai, tapi mengambilnya nyaris seketika.
//
// Karena meja kerja bisa dibersihkan sewaktu-waktu (Redis menyimpan di RAM), aturannya:
// apa pun yang ada di Redis harus BISA DIBANGUN ULANG dari gudang. Kalau kehilangan isi
// Redis membuat datamu hilang selamanya, berarti kamu memakainya sebagai gudang —
// dan itu keputusan yang perlu dipikirkan ulang.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func main() {
	fmt.Println("=== redis/go-redis ===")

	// 🔍 Analogi miniredis: ini "Redis mainan" yang hidup di dalam proses Go ini sendiri.
	// Perilakunya meniru Redis asli, jadi contoh & test bisa jalan di laptop mana pun
	// tanpa memasang apa pun. Pola yang sama dipakai modul 22.
	mr, err := miniredis.Run()
	if err != nil {
		fmt.Println("gagal menyalakan miniredis:", err)
		return
	}
	defer mr.Close()

	c := NewCache(NewKlien(mr.Addr()))
	ctx := context.Background()

	demoDasar(ctx, c)
	demoCacheAside(ctx, c)
	demoTTL(ctx, c, mr)
	demoPembatalan(ctx, c)
	demoPenghitung(ctx, c)
	demoPipeline(ctx, c)
}

// ------------------------------------------------------------------
// Klien & cache
// ------------------------------------------------------------------

// NewKlien membuat klien Redis dengan pengaturan yang wajar untuk produksi.
func NewKlien(alamat string) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: alamat,
		// 🔍 Analogi timeout: Redis sangat cepat (biasanya di bawah 1 milidetik).
		// Kalau ia tak menjawab dalam 2 detik, hampir pasti ada yang salah — lebih baik
		// menyerah dan mengambil dari gudang daripada menahan permintaan pengguna.
		DialTimeout:  2 * time.Second,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
		// Kolam koneksi: dibuka sekali, dipakai berulang.
		PoolSize: 10,
	})
}

// Produk contoh data yang di-cache.
type Produk struct {
	ID    int    `json:"id"`
	Nama  string `json:"nama"`
	Harga int    `json:"harga"`
}

// Cache membungkus klien Redis dengan operasi yang sudah bermakna bagi aplikasi.
type Cache struct {
	rdb        *redis.Client
	dariGudang int // penghitung: berapa kali sumber asli benar-benar dibaca
}

func NewCache(rdb *redis.Client) *Cache {
	return &Cache{rdb: rdb}
}

func (c *Cache) Tutup() error { return c.rdb.Close() }

// kunciProduk menyusun nama kunci.
//
// 🔍 Analogi penamaan kunci: Redis itu satu ruangan besar tanpa sekat — semua kunci
// bercampur. Konvensi "namespace:jenis:id" (mis. "toko:produk:42") itu seperti MEMBERI
// LABEL RAK. Tanpanya, kunci "42" milik produk bisa tertimpa kunci "42" milik pesanan,
// dan kamu tak akan pernah tahu kapan itu terjadi.
func kunciProduk(id int) string {
	return fmt.Sprintf("toko:produk:%d", id)
}

// ------------------------------------------------------------------
// 1. Operasi dasar & jebakan redis.Nil
// ------------------------------------------------------------------

// ErrTidakAdaDiCache dikembalikan bila kunci tidak ada.
var ErrTidakAdaDiCache = errors.New("tidak ada di cache")

// Simpan menaruh produk ke cache dengan masa berlaku.
func (c *Cache) Simpan(ctx context.Context, p Produk, ttl time.Duration) error {
	data, err := json.Marshal(p)
	if err != nil {
		return fmt.Errorf("gagal mengubah produk jadi JSON: %w", err)
	}
	if err := c.rdb.Set(ctx, kunciProduk(p.ID), data, ttl).Err(); err != nil {
		return fmt.Errorf("gagal menyimpan ke cache: %w", err)
	}
	return nil
}

// Ambil membaca produk dari cache.
//
// 🔍 Analogi redis.Nil — INI jebakan nomor satu pemakai go-redis:
// "meja kosong" (kunci tak ada) itu HAL NORMAL, bukan kerusakan. Tapi go-redis
// mengembalikannya sebagai error bernama redis.Nil. Kalau kamu memperlakukan setiap
// error sebagai kegagalan sistem, aplikasimu akan mengira Redis rusak setiap kali
// ada cache miss — padahal cache miss itu justru kejadian sehari-hari.
//
// Bedakan dengan tegas: redis.Nil = "tidak ada" (lanjut ambil dari gudang);
// error lain = "Redis bermasalah" (catat di log, mungkin nyalakan circuit breaker).
func (c *Cache) Ambil(ctx context.Context, id int) (Produk, error) {
	data, err := c.rdb.Get(ctx, kunciProduk(id)).Bytes()
	if errors.Is(err, redis.Nil) {
		return Produk{}, fmt.Errorf("produk %d: %w", id, ErrTidakAdaDiCache)
	}
	if err != nil {
		return Produk{}, fmt.Errorf("redis bermasalah saat membaca produk %d: %w", id, err)
	}

	var p Produk
	if err := json.Unmarshal(data, &p); err != nil {
		return Produk{}, fmt.Errorf("isi cache rusak untuk produk %d: %w", id, err)
	}
	return p, nil
}

func demoDasar(ctx context.Context, c *Cache) {
	fmt.Println("\n-- Simpan & ambil --")

	p := Produk{ID: 1, Nama: "Kopi Arabika", Harga: 85_000}
	if err := c.Simpan(ctx, p, time.Minute); err != nil {
		fmt.Println("   error:", err)
		return
	}

	got, err := c.Ambil(ctx, 1)
	if err != nil {
		fmt.Println("   error:", err)
		return
	}
	fmt.Printf("   dari cache -> %+v\n", got)

	if _, err := c.Ambil(ctx, 999); errors.Is(err, ErrTidakAdaDiCache) {
		fmt.Println("   produk 999 -> cache miss (normal, bukan kerusakan)")
	}
}

// ------------------------------------------------------------------
// 2. Pola cache-aside
// ------------------------------------------------------------------

// AmbilAtauMuat adalah pola cache-aside: lihat cache dulu, kalau kosong ambil dari
// gudang lalu simpan untuk permintaan berikutnya.
//
// 🔍 Analogi: kamu butuh berkas. Lihat dulu di meja kerja (cache). Kalau tak ada,
// turun ke gudang (database), FOTOKOPI-nya ditaruh di meja, lalu kembali bekerja.
// Permintaan berikutnya cukup melihat meja.
//
// Perhatikan sikap terhadap kegagalan Redis: kalau meja kerjanya terbakar, kamu tetap
// bisa bekerja — cukup turun ke gudang. Cache yang mati seharusnya membuat aplikasi
// LEBIH LAMBAT, bukan MATI. Karena itu error dari cache di sini tidak dikembalikan.
func (c *Cache) AmbilAtauMuat(
	ctx context.Context,
	id int,
	ttl time.Duration,
	dariGudang func(context.Context, int) (Produk, error),
) (Produk, bool, error) {
	if p, err := c.Ambil(ctx, id); err == nil {
		return p, true, nil // true = berasal dari cache
	}

	p, err := dariGudang(ctx, id)
	if err != nil {
		return Produk{}, false, fmt.Errorf("gagal memuat produk %d: %w", id, err)
	}
	c.dariGudang++

	// Gagal menyimpan ke cache TIDAK boleh menggagalkan permintaan pengguna —
	// datanya sudah ada, cache hanya optimasi.
	_ = c.Simpan(ctx, p, ttl)
	return p, false, nil
}

// JumlahAksesGudang untuk membuktikan cache benar-benar mengurangi beban database.
func (c *Cache) JumlahAksesGudang() int { return c.dariGudang }

func demoCacheAside(ctx context.Context, c *Cache) {
	fmt.Println("\n-- Cache-aside --")

	gudang := func(_ context.Context, id int) (Produk, error) {
		time.Sleep(10 * time.Millisecond) // meniru query database
		return Produk{ID: id, Nama: "Teh Melati", Harga: 35_000}, nil
	}

	for i := 1; i <= 3; i++ {
		p, dariCache, err := c.AmbilAtauMuat(ctx, 2, time.Minute, gudang)
		if err != nil {
			fmt.Println("   error:", err)
			return
		}
		fmt.Printf("   percobaan %d -> %s (dari cache? %t)\n", i, p.Nama, dariCache)
	}
	fmt.Printf("   database disentuh %d kali untuk 3 permintaan\n", c.JumlahAksesGudang())
}

// ------------------------------------------------------------------
// 3. TTL — tanggal kedaluwarsa
// ------------------------------------------------------------------

// 🔍 Analogi TTL: seperti tanggal kedaluwarsa pada makanan di kulkas. Ia adalah jawaban
// paling sederhana untuk pertanyaan tersulit di dunia cache: "sampai kapan salinan ini
// boleh dipercaya?" TTL pendek = data lebih segar tapi database lebih sering disentuh;
// TTL panjang = database santai tapi pengguna bisa melihat data basi.
//
// Jangan pernah menyimpan tanpa TTL kecuali kamu benar-benar berniat: kunci tanpa masa
// berlaku akan menumpuk sampai memori Redis penuh.

// SisaMasaBerlaku mengembalikan sisa waktu hidup sebuah kunci.
func (c *Cache) SisaMasaBerlaku(ctx context.Context, id int) (time.Duration, error) {
	d, err := c.rdb.TTL(ctx, kunciProduk(id)).Result()
	if err != nil {
		return 0, fmt.Errorf("gagal membaca TTL: %w", err)
	}
	// -2 berarti kunci tak ada; -1 berarti ada tapi tanpa masa berlaku.
	return d, nil
}

func demoTTL(ctx context.Context, c *Cache, mr *miniredis.Miniredis) {
	fmt.Println("\n-- TTL --")

	p := Produk{ID: 3, Nama: "Roti Gandum", Harga: 25_000}
	if err := c.Simpan(ctx, p, 30*time.Second); err != nil {
		fmt.Println("   error:", err)
		return
	}

	sisa, _ := c.SisaMasaBerlaku(ctx, 3)
	fmt.Printf("   sisa masa berlaku: %v\n", sisa)

	// 🔍 Analogi FastForward: miniredis punya MESIN WAKTU. Kita bisa melompat 31 detik
	// ke depan seketika, tanpa test yang benar-benar tidur setengah menit.
	mr.FastForward(31 * time.Second)

	if _, err := c.Ambil(ctx, 3); errors.Is(err, ErrTidakAdaDiCache) {
		fmt.Println("   setelah lompat 31 detik -> kunci sudah kedaluwarsa")
	}
}

// ------------------------------------------------------------------
// 4. Pembatalan cache
// ------------------------------------------------------------------

// Hapus membuang satu produk dari cache.
//
// 🔍 Analogi: begitu harga produk berubah di gudang, FOTOKOPI di meja kerja jadi
// menyesatkan. Ada dua sikap: buang fotokopinya (hapus) atau langsung ganti dengan
// yang baru (tulis ulang). Membuang lebih aman — kalau proses penulisan gagal di
// tengah jalan, yang terjadi cuma cache miss, bukan data salah yang bertahan lama.
func (c *Cache) Hapus(ctx context.Context, id int) error {
	if err := c.rdb.Del(ctx, kunciProduk(id)).Err(); err != nil {
		return fmt.Errorf("gagal menghapus cache produk %d: %w", id, err)
	}
	return nil
}

func demoPembatalan(ctx context.Context, c *Cache) {
	fmt.Println("\n-- Pembatalan cache --")

	p := Produk{ID: 4, Nama: "Keju", Harga: 120_000}
	_ = c.Simpan(ctx, p, time.Minute)
	fmt.Println("   disimpan, lalu harga berubah di database...")

	if err := c.Hapus(ctx, 4); err != nil {
		fmt.Println("   error:", err)
		return
	}
	if _, err := c.Ambil(ctx, 4); errors.Is(err, ErrTidakAdaDiCache) {
		fmt.Println("   cache dibuang -> permintaan berikutnya membaca data terbaru")
	}
}

// ------------------------------------------------------------------
// 5. Penghitung atomik
// ------------------------------------------------------------------

// Naikkan menambah penghitung dan mengembalikan nilai barunya.
//
// 🔍 Analogi: INCR itu seperti MESIN PENGHITUNG KLIK di pintu masuk — kalau sepuluh
// petugas menekan bersamaan, angkanya tetap benar. Redis mengerjakan perintah satu per
// satu, jadi tak ada dua proses yang bisa "membaca 5, sama-sama menambah, sama-sama
// menulis 6". Ini alasan penghitung kunjungan & pembatas laju sering ditaruh di Redis.
func (c *Cache) Naikkan(ctx context.Context, kunci string, ttl time.Duration) (int64, error) {
	n, err := c.rdb.Incr(ctx, kunci).Result()
	if err != nil {
		return 0, fmt.Errorf("gagal menaikkan penghitung %q: %w", kunci, err)
	}
	// Pasang masa berlaku HANYA saat kunci baru dibuat (nilai pertama = 1),
	// supaya jendela hitungnya tidak diperpanjang terus-menerus.
	if n == 1 && ttl > 0 {
		_ = c.rdb.Expire(ctx, kunci, ttl).Err()
	}
	return n, nil
}

func demoPenghitung(ctx context.Context, c *Cache) {
	fmt.Println("\n-- Penghitung atomik --")

	for range 5 {
		n, err := c.Naikkan(ctx, "statistik:kunjungan:2026-07-23", time.Hour)
		if err != nil {
			fmt.Println("   error:", err)
			return
		}
		fmt.Printf("   kunjungan ke-%d\n", n)
	}
}

// ------------------------------------------------------------------
// 6. Pipeline
// ------------------------------------------------------------------

// SimpanBanyak mengirim banyak perintah dalam SATU perjalanan jaringan.
//
// 🔍 Analogi: menyimpan 100 produk satu per satu itu seperti bolak-balik ke toko
// 100 kali untuk membeli 100 barang. Pipeline adalah DAFTAR BELANJA: satu perjalanan,
// semua barang terbawa. Biaya terbesar berbicara dengan Redis bukanlah pekerjaannya
// (Redis sangat cepat), melainkan perjalanan bolak-baliknya.
func (c *Cache) SimpanBanyak(ctx context.Context, produk []Produk, ttl time.Duration) error {
	pipe := c.rdb.Pipeline()

	for _, p := range produk {
		data, err := json.Marshal(p)
		if err != nil {
			return fmt.Errorf("gagal mengubah produk %d jadi JSON: %w", p.ID, err)
		}
		pipe.Set(ctx, kunciProduk(p.ID), data, ttl)
	}

	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("pipeline gagal: %w", err)
	}
	return nil
}

func demoPipeline(ctx context.Context, c *Cache) {
	fmt.Println("\n-- Pipeline --")

	batch := []Produk{
		{ID: 10, Nama: "Gelas", Harga: 45_000},
		{ID: 11, Nama: "Piring", Harga: 55_000},
		{ID: 12, Nama: "Sendok", Harga: 15_000},
	}
	if err := c.SimpanBanyak(ctx, batch, time.Minute); err != nil {
		fmt.Println("   error:", err)
		return
	}
	fmt.Printf("   %d produk disimpan dalam satu perjalanan jaringan\n", len(batch))

	got, err := c.Ambil(ctx, 11)
	if err != nil {
		fmt.Println("   error:", err)
		return
	}
	fmt.Printf("   cek satu -> %+v\n", got)
}
