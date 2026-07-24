// prometheus/client_golang — mengekspos metrik aplikasi agar bisa dipantau.
//
// Jalankan: go run ./libraries/prometheus
// Test:     go test ./libraries/prometheus
//
// 🔍 Analogi besar: log itu CATATAN HARIAN — bagus untuk menyelidiki satu kejadian
// ("kenapa pesanan #482 gagal?"), tapi payah untuk menjawab pertanyaan menyeluruh
// ("apakah layanan kita sedang melambat?"). Metrik itu PANEL DASBOR MOBIL: spidometer,
// pengukur bensin, lampu suhu. Ia tak menceritakan perjalananmu, tapi sekali lirik kamu
// tahu apakah keadaan sedang baik-baik saja.
//
// Cara kerja Prometheus juga khas: server Prometheus MENJEMPUT (scrape) angka-angka ini
// dari endpoint /metrics aplikasimu setiap beberapa detik. Aplikasimu tak perlu mengirim
// apa pun ke mana pun — ia cuma perlu menyediakan papan angka yang selalu terbaru.
package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	fmt.Println("=== prometheus/client_golang ===")

	m := NewMetrik()
	demoJenisMetrik(m)
	demoMiddleware(m)
	demoEndpoint(m)
	demoKardinalitas()
}

// ------------------------------------------------------------------
// 1. Kumpulan metrik
// ------------------------------------------------------------------

// Metrik menampung seluruh pengukur aplikasi beserta registry-nya sendiri.
//
// 🔍 Analogi registry sendiri vs registry global: prometheus punya registry bawaan yang
// dipakai bersama seluruh program — seperti PAPAN PENGUMUMAN UMUM di kelurahan. Praktis,
// tapi bermasalah: mendaftarkan metrik bernama sama dua kali akan PANIC, dan di test,
// angka dari test sebelumnya masih menempel di sana.
//
// Registry sendiri = papan pengumuman milikmu. Tiap test punya papan bersih, tak ada
// bentrok nama, dan tak ada angka nyasar. Ini juga alasan tak ada satu pun promauto
// di berkas ini — promauto diam-diam mendaftar ke registry global.
type Metrik struct {
	Registry *prometheus.Registry

	// Counter: angka yang HANYA BISA NAIK.
	permintaan *prometheus.CounterVec

	// Gauge: angka yang bisa naik DAN turun.
	sedangDiproses prometheus.Gauge

	// Histogram: sebaran nilai, untuk menjawab "berapa persentil ke-95?".
	durasi *prometheus.HistogramVec

	// Counter tanpa label, untuk peristiwa bisnis.
	pesananDibuat prometheus.Counter
}

// NewMetrik membuat & mendaftarkan seluruh metrik ke registry baru.
func NewMetrik() *Metrik {
	reg := prometheus.NewRegistry()

	m := &Metrik{
		Registry: reg,

		// 🔍 Analogi Counter = ODOMETER MOBIL. Ia tak pernah mundur; yang bermakna
		// bukan angkanya, melainkan SELISIHNYA antar waktu ("berapa km ditempuh
		// sejam terakhir"). Prometheus menghitung laju itu dengan fungsi rate().
		//
		// Konvensi penamaan: counter selalu berakhiran "_total".
		permintaan: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_permintaan_total",
				Help: "Jumlah total permintaan HTTP yang diterima.",
			},
			// Label = cara mengiris angka: per method, per rute, per status.
			[]string{"method", "rute", "status"},
		),

		// 🔍 Analogi Gauge = SPIDOMETER (atau pengukur bensin). Naik-turun bebas, dan
		// yang bermakna adalah nilainya SAAT INI. Cocok untuk: permintaan yang sedang
		// diproses, panjang antrean, jumlah koneksi database yang terpakai.
		sedangDiproses: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "http_permintaan_sedang_diproses",
			Help: "Banyaknya permintaan HTTP yang sedang ditangani saat ini.",
		}),

		// 🔍 Analogi Histogram = LACI BERSEKAT. Tiap pengukuran dijatuhkan ke sekat
		// sesuai besarnya (0-5ms, 5-10ms, 10-25ms, ...). Kenapa tidak menyimpan
		// rata-rata saja? Karena rata-rata MENIPU: 99 permintaan 10ms + 1 permintaan
		// 10 detik menghasilkan rata-rata ~110ms yang tampak wajar, padahal ada
		// pengguna yang menunggu 10 detik. Histogram membuatmu bisa bertanya
		// "95% pengguna dilayani di bawah berapa detik?" — itulah yang benar-benar
		// dirasakan pengguna.
		//
		// Konvensi penamaan: sertakan satuannya, dan pakai SATUAN DASAR (detik, byte).
		durasi: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name: "http_durasi_detik",
				Help: "Lama penanganan permintaan HTTP dalam detik.",
				// Sekat bawaan cocok untuk permintaan web pada umumnya.
				Buckets: prometheus.DefBuckets,
			},
			[]string{"rute"},
		),

		pesananDibuat: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "pesanan_dibuat_total",
			Help: "Jumlah pesanan yang berhasil dibuat.",
		}),
	}

	// MustRegister panic bila ada nama yang bentrok — itu disengaja: bentrok nama
	// metrik adalah kesalahan pemrograman yang harus ketahuan saat aplikasi mulai,
	// bukan diam-diam menghasilkan angka yang salah.
	reg.MustRegister(m.permintaan, m.sedangDiproses, m.durasi, m.pesananDibuat)
	return m
}

// ------------------------------------------------------------------
// 2. Memakai tiap jenis metrik
// ------------------------------------------------------------------

// CatatPermintaan menaikkan penghitung untuk satu permintaan.
func (m *Metrik) CatatPermintaan(method, rute string, status int) {
	m.permintaan.WithLabelValues(method, rute, strconv.Itoa(status)).Inc()
}

// CatatDurasi mencatat lama penanganan ke histogram.
func (m *Metrik) CatatDurasi(rute string, d time.Duration) {
	m.durasi.WithLabelValues(rute).Observe(d.Seconds())
}

// PesananBaru menaikkan penghitung peristiwa bisnis.
//
// 🔍 Analogi: metrik bukan cuma soal teknis (CPU, latensi). Metrik BISNIS seperti ini
// justru sering paling berharga — grafik "pesanan per menit" yang tiba-tiba jatuh ke nol
// memberi tahu ada masalah jauh lebih cepat daripada grafik penggunaan memori.
func (m *Metrik) PesananBaru() {
	m.pesananDibuat.Inc()
}

// MulaiProses & SelesaiProses menggerakkan gauge naik-turun.
func (m *Metrik) MulaiProses()   { m.sedangDiproses.Inc() }
func (m *Metrik) SelesaiProses() { m.sedangDiproses.Dec() }

func demoJenisMetrik(m *Metrik) {
	fmt.Println("\n-- Tiga jenis metrik --")

	for i := range 5 {
		m.CatatPermintaan(http.MethodGet, "/produk", 200)
		m.CatatDurasi("/produk", time.Duration(10+i*5)*time.Millisecond)
		m.PesananBaru()
	}
	m.CatatPermintaan(http.MethodGet, "/produk", 500)

	m.MulaiProses()
	m.MulaiProses()
	m.SelesaiProses()

	fmt.Println("   counter   : 5x status 200, 1x status 500 pada GET /produk")
	fmt.Println("   gauge     : 2 naik, 1 turun -> nilai sekarang 1")
	fmt.Println("   histogram : 5 pengamatan durasi antara 10ms dan 30ms")
}

// ------------------------------------------------------------------
// 3. Middleware HTTP
// ------------------------------------------------------------------

// perekamStatus mengintip status code yang ditulis handler.
//
// 🔍 Analogi: http.ResponseWriter itu seperti loket satu arah — begitu handler menulis
// jawabannya, kamu tak bisa bertanya "tadi kamu jawab apa?". Pembungkus ini adalah
// PENGINTIP yang duduk di loket: ia meneruskan semuanya apa adanya, sambil mencatat
// status yang lewat. Pola yang sama dipakai modul 18.
type perekamStatus struct {
	http.ResponseWriter
	status   int
	tertulis bool
}

func (p *perekamStatus) WriteHeader(kode int) {
	if !p.tertulis {
		p.status = kode
		p.tertulis = true
	}
	p.ResponseWriter.WriteHeader(kode)
}

func (p *perekamStatus) Write(b []byte) (int, error) {
	// Handler yang langsung menulis body tanpa WriteHeader berarti status 200.
	if !p.tertulis {
		p.status = http.StatusOK
		p.tertulis = true
	}
	return p.ResponseWriter.Write(b)
}

// Middleware membungkus handler agar tiap permintaan otomatis terukur.
func (m *Metrik) Middleware(rute string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mulai := time.Now()

		m.MulaiProses()
		// defer memastikan gauge tetap turun walau handler panic — tanpa ini,
		// satu panic akan membuat angka "sedang diproses" naik selamanya
		// dan dasbormu berbohong sampai aplikasi di-restart.
		defer m.SelesaiProses()

		p := &perekamStatus{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(p, r)

		m.CatatPermintaan(r.Method, rute, p.status)
		m.CatatDurasi(rute, time.Since(mulai))
	})
}

func demoMiddleware(m *Metrik) {
	fmt.Println("\n-- Middleware --")

	h := m.Middleware("/api", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(time.Duration(rand.Intn(5)) * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ok")
	}))

	for range 3 {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api", nil))
	}
	fmt.Println("   3 permintaan lewat middleware -> counter, gauge, histogram terisi")
}

// ------------------------------------------------------------------
// 4. Endpoint /metrics
// ------------------------------------------------------------------

// Handler menyajikan seluruh metrik dalam format teks yang dipahami Prometheus.
//
// 🔍 Analogi: ini PAPAN PENGUMUMAN di depan kantor. Prometheus datang tiap 15 detik,
// memotret papannya, lalu pulang. Aplikasimu tak pernah menghubungi Prometheus.
//
// Catatan keamanan: endpoint ini membocorkan bentuk internal sistemmu (nama rute,
// jumlah pengguna, versi). Di produksi, jangan buka ke internet — taruh di port
// terpisah atau lindungi di tingkat jaringan.
func (m *Metrik) Handler() http.Handler {
	return promhttp.HandlerFor(m.Registry, promhttp.HandlerOpts{})
}

func demoEndpoint(m *Metrik) {
	fmt.Println("\n-- Endpoint /metrics --")

	rec := httptest.NewRecorder()
	m.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/metrics", nil))

	// Tampilkan beberapa baris pertama saja sebagai contoh format.
	n := 0
	for _, baris := range splitBaris(rec.Body.String()) {
		if baris == "" || baris[0] == '#' {
			continue
		}
		fmt.Println("  ", baris)
		if n++; n >= 5 {
			break
		}
	}
	fmt.Println("   ... (dipotong)")
}

// splitBaris memecah teks jadi baris tanpa menarik dependensi tambahan.
func splitBaris(s string) []string {
	var out []string
	mulai := 0
	for i := range len(s) {
		if s[i] == '\n' {
			out = append(out, s[mulai:i])
			mulai = i + 1
		}
	}
	if mulai < len(s) {
		out = append(out, s[mulai:])
	}
	return out
}

// ------------------------------------------------------------------
// 5. Jebakan kardinalitas
// ------------------------------------------------------------------

// 🔍 Analogi kardinalitas — INI kesalahan yang paling sering merobohkan Prometheus:
// setiap KOMBINASI nilai label membuat satu deret waktu BARU yang harus disimpan
// selamanya. Bayangkan lemari arsip yang laciny bertambah tiap ada nilai label baru.
//
//	Label "status" (200, 404, 500, ...)  -> puluhan laci. Aman.
//	Label "rute" (/produk, /pesanan)     -> puluhan laci. Aman.
//	Label "user_id"                      -> SATU LACI PER PENGGUNA. Sejuta pengguna
//	                                        = sejuta deret waktu. Prometheus tumbang.
//
// Aturannya: label hanya untuk nilai yang jumlahnya SEDIKIT & TERBATAS. Kalau kamu ingin
// menelusuri satu pengguna atau satu request, itu pekerjaan log dan tracing (modul 33),
// bukan metrik.
//
// Jebakan turunan yang sering luput: memakai PATH MENTAH sebagai label. "/produk/1",
// "/produk/2", "/produk/3" tampak seperti tiga rute berbeda — padahal seharusnya satu
// pola "/produk/{id}". Selalu pakai POLA rute, bukan URL yang diminta.

// PolaRute mengubah path mentah jadi pola berkardinalitas rendah.
func PolaRute(path string) string {
	// Contoh sederhana: segmen yang berupa angka diganti "{id}".
	out := ""
	segmen := ""
	tambah := func() {
		if segmen == "" {
			return
		}
		if semuaAngka(segmen) {
			out += "/{id}"
		} else {
			out += "/" + segmen
		}
		segmen = ""
	}

	for i := range len(path) {
		if path[i] == '/' {
			tambah()
			continue
		}
		segmen += string(path[i])
	}
	tambah()

	if out == "" {
		return "/"
	}
	return out
}

func semuaAngka(s string) bool {
	if s == "" {
		return false
	}
	for i := range len(s) {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

func demoKardinalitas() {
	fmt.Println("\n-- Jebakan kardinalitas --")
	for _, p := range []string{"/produk/1", "/produk/42", "/pesanan/7/item/3", "/sehat", "/"} {
		fmt.Printf("   %-20s -> %s\n", p, PolaRute(p))
	}
	fmt.Println("   tanpa ini, tiap ID produk jadi deret waktu baru di Prometheus")
}
