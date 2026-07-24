package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/testutil"
)

// Tiap test membuat Metrik sendiri — inilah untungnya memakai registry sendiri,
// bukan registry global: tak ada angka dari test lain yang menempel.
func TestRegistryTerisolasiAntarTest(t *testing.T) {
	a := NewMetrik()
	b := NewMetrik()

	a.PesananBaru()
	a.PesananBaru()

	if got := testutil.ToFloat64(a.pesananDibuat); got != 2 {
		t.Errorf("metrik A = %v, ingin 2", got)
	}
	if got := testutil.ToFloat64(b.pesananDibuat); got != 0 {
		t.Errorf("metrik B = %v, ingin 0 — registry seharusnya terpisah", got)
	}
}

// Membuat dua Metrik tak boleh panic karena nama bentrok (bukti registry terpisah).
func TestMembuatBanyakMetrikTidakPanic(t *testing.T) {
	for range 5 {
		if m := NewMetrik(); m == nil {
			t.Fatal("NewMetrik mengembalikan nil")
		}
	}
}

func TestCounterHanyaNaik(t *testing.T) {
	m := NewMetrik()

	for range 3 {
		m.CatatPermintaan(http.MethodGet, "/produk", 200)
	}
	m.CatatPermintaan(http.MethodGet, "/produk", 500)
	m.CatatPermintaan(http.MethodPost, "/produk", 201)

	tests := []struct {
		nama   string
		method string
		rute   string
		status string
		want   float64
	}{
		{"GET 200 tiga kali", http.MethodGet, "/produk", "200", 3},
		{"GET 500 sekali", http.MethodGet, "/produk", "500", 1},
		{"POST 201 sekali", http.MethodPost, "/produk", "201", 1},
		{"kombinasi yang belum pernah muncul", http.MethodPut, "/produk", "200", 0},
	}

	for _, tt := range tests {
		t.Run(tt.nama, func(t *testing.T) {
			c := m.permintaan.WithLabelValues(tt.method, tt.rute, tt.status)
			if got := testutil.ToFloat64(c); got != tt.want {
				t.Errorf("counter = %v, ingin %v", got, tt.want)
			}
		})
	}
}

func TestGaugeBisaNaikDanTurun(t *testing.T) {
	m := NewMetrik()

	if got := testutil.ToFloat64(m.sedangDiproses); got != 0 {
		t.Fatalf("nilai awal = %v, ingin 0", got)
	}

	m.MulaiProses()
	m.MulaiProses()
	m.MulaiProses()
	if got := testutil.ToFloat64(m.sedangDiproses); got != 3 {
		t.Errorf("setelah 3 kali naik = %v, ingin 3", got)
	}

	m.SelesaiProses()
	m.SelesaiProses()
	if got := testutil.ToFloat64(m.sedangDiproses); got != 1 {
		t.Errorf("setelah 2 kali turun = %v, ingin 1 — inilah beda gauge dari counter", got)
	}
}

func TestHistogramMencatatPengamatan(t *testing.T) {
	m := NewMetrik()

	for _, d := range []time.Duration{5, 10, 25, 100, 500} {
		m.CatatDurasi("/produk", d*time.Millisecond)
	}

	// CollectAndCount menghitung banyaknya deret waktu, bukan banyaknya pengamatan.
	// Satu histogram berlabel "/produk" = satu deret.
	if got := testutil.CollectAndCount(m.durasi); got != 1 {
		t.Errorf("jumlah deret histogram = %d, ingin 1", got)
	}

	// Rute berbeda membuat deret baru.
	m.CatatDurasi("/pesanan", 20*time.Millisecond)
	if got := testutil.CollectAndCount(m.durasi); got != 2 {
		t.Errorf("jumlah deret histogram = %d, ingin 2", got)
	}
}

func TestMiddlewareMengukurPermintaan(t *testing.T) {
	m := NewMetrik()

	h := m.Middleware("/api", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))

	for range 4 {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api", nil))
	}

	c := m.permintaan.WithLabelValues(http.MethodGet, "/api", "200")
	if got := testutil.ToFloat64(c); got != 4 {
		t.Errorf("counter permintaan = %v, ingin 4", got)
	}
	// Gauge harus kembali ke nol setelah semua permintaan selesai.
	if got := testutil.ToFloat64(m.sedangDiproses); got != 0 {
		t.Errorf("sedang diproses = %v, ingin 0 setelah semua selesai", got)
	}
}

// Status code yang ditulis handler harus terekam apa adanya.
func TestMiddlewareMerekamStatusYangBenar(t *testing.T) {
	tests := []struct {
		nama         string
		tulisKode    int
		langsungBody bool
		wantLabel    string
	}{
		{"200 eksplisit", http.StatusOK, false, "200"},
		{"404", http.StatusNotFound, false, "404"},
		{"500", http.StatusInternalServerError, false, "500"},
		{"201", http.StatusCreated, false, "201"},
		{"tanpa WriteHeader dianggap 200", 0, true, "200"},
	}

	for _, tt := range tests {
		t.Run(tt.nama, func(t *testing.T) {
			m := NewMetrik()
			h := m.Middleware("/uji", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				if tt.langsungBody {
					_, _ = w.Write([]byte("langsung body"))
					return
				}
				w.WriteHeader(tt.tulisKode)
			}))

			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/uji", nil))

			c := m.permintaan.WithLabelValues(http.MethodGet, "/uji", tt.wantLabel)
			if got := testutil.ToFloat64(c); got != 1 {
				t.Errorf("counter untuk status %s = %v, ingin 1", tt.wantLabel, got)
			}
		})
	}
}

// Gauge harus tetap turun walau handler panic — kalau tidak, dasbor berbohong selamanya.
func TestGaugeTurunWalauHandlerPanic(t *testing.T) {
	m := NewMetrik()

	h := m.Middleware("/rusak", http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("handler sengaja rusak")
	}))

	func() {
		// Panic ditangkap di sini supaya test bisa lanjut memeriksa gauge-nya.
		defer func() { _ = recover() }()
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/rusak", nil))
	}()

	if got := testutil.ToFloat64(m.sedangDiproses); got != 0 {
		t.Errorf("sedang diproses = %v setelah panic, ingin 0 — defer sepertinya hilang", got)
	}
}

func TestHandlerMengeluarkanFormatPrometheus(t *testing.T) {
	m := NewMetrik()
	m.CatatPermintaan(http.MethodGet, "/produk", 200)
	m.CatatDurasi("/produk", 15*time.Millisecond)
	m.PesananBaru()

	rec := httptest.NewRecorder()
	m.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/metrics", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, ingin 200", rec.Code)
	}

	body := rec.Body.String()
	for _, harus := range []string{
		"# HELP http_permintaan_total",
		"# TYPE http_permintaan_total counter",
		`http_permintaan_total{method="GET",rute="/produk",status="200"} 1`,
		"pesanan_dibuat_total 1",
		"# TYPE http_durasi_detik histogram",
	} {
		if !strings.Contains(body, harus) {
			t.Errorf("keluaran /metrics tidak memuat %q", harus)
		}
	}
}

// Metrik yang belum pernah disentuh tetap muncul dengan nilai 0 —
// penting supaya grafik tidak "hilang" saat lalu lintas sepi.
func TestMetrikTanpaLabelMunculSejakAwal(t *testing.T) {
	m := NewMetrik()

	rec := httptest.NewRecorder()
	m.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/metrics", nil))

	body := rec.Body.String()
	if !strings.Contains(body, "pesanan_dibuat_total 0") {
		t.Error("counter tanpa label seharusnya sudah muncul bernilai 0")
	}
	if !strings.Contains(body, "http_permintaan_sedang_diproses 0") {
		t.Error("gauge seharusnya sudah muncul bernilai 0")
	}
}

// Kebalikannya, dan ini sering mengagetkan: metrik BER-LABEL sama sekali TIDAK muncul
// sampai kombinasi label pertamanya dipakai.
//
// 🔍 Analogi: metrik tanpa label itu seperti papan skor yang sudah terpasang di lapangan
// sejak awal — tertulis 0. Metrik ber-label itu papan skor yang baru DIBUATKAN saat
// pemain pertama masuk; sebelum itu tak ada papannya sama sekali.
//
// Akibat praktisnya: grafik "laju error 5xx" akan KOSONG (bukan nol) selama belum pernah
// ada error, dan alert seperti "rate(...) > 0.05" jadi tak pernah menyala pada layanan
// yang baru dinyalakan. Kalau itu penting, daftarkan kombinasi labelnya lebih dulu
// dengan WithLabelValues(...) saat aplikasi mulai.
func TestMetrikBerlabelBelumMunculSebelumDipakai(t *testing.T) {
	m := NewMetrik()

	ambilBody := func() string {
		rec := httptest.NewRecorder()
		m.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/metrics", nil))
		return rec.Body.String()
	}

	sebelum := ambilBody()
	if strings.Contains(sebelum, "http_permintaan_total{") {
		t.Error("counter ber-label seharusnya belum muncul sebelum ada label yang dipakai")
	}
	if strings.Contains(sebelum, "http_durasi_detik_bucket") {
		t.Error("histogram ber-label seharusnya belum muncul sebelum ada pengamatan")
	}

	m.CatatPermintaan(http.MethodGet, "/produk", 200)
	m.CatatDurasi("/produk", 15*time.Millisecond)

	sesudah := ambilBody()
	if !strings.Contains(sesudah, "http_permintaan_total{") {
		t.Error("setelah dipakai, counter ber-label harus muncul")
	}
	if !strings.Contains(sesudah, "http_durasi_detik_bucket") {
		t.Error("setelah ada pengamatan, histogram harus muncul")
	}
}

func TestPolaRuteMenekanKardinalitas(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"/produk/1", "/produk/{id}"},
		{"/produk/42", "/produk/{id}"},
		{"/produk/999999", "/produk/{id}"},
		{"/pesanan/7/item/3", "/pesanan/{id}/item/{id}"},
		{"/sehat", "/sehat"},
		{"/", "/"},
		{"", "/"},
		{"/produk", "/produk"},
		{"/produk/abc", "/produk/abc"},
		{"/produk/1a", "/produk/1a"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if got := PolaRute(tt.path); got != tt.want {
				t.Errorf("PolaRute(%q) = %q, ingin %q", tt.path, got, tt.want)
			}
		})
	}
}

// Inti pelajarannya: 1000 ID produk harus menjadi SATU deret waktu, bukan 1000.
func TestPolaRuteMencegahLedakanDeretWaktu(t *testing.T) {
	m := NewMetrik()

	for i := 1; i <= 1000; i++ {
		path := "/produk/" + strings.Repeat("", 0) + itoa(i)
		m.CatatPermintaan(http.MethodGet, PolaRute(path), 200)
	}

	if got := testutil.CollectAndCount(m.permintaan); got != 1 {
		t.Errorf("jumlah deret waktu = %d, ingin 1 — pola rute gagal menekan kardinalitas", got)
	}

	// Pembanding: tanpa PolaRute, tiap ID jadi deret sendiri.
	m2 := NewMetrik()
	for i := 1; i <= 50; i++ {
		m2.CatatPermintaan(http.MethodGet, "/produk/"+itoa(i), 200)
	}
	if got := testutil.CollectAndCount(m2.permintaan); got != 50 {
		t.Errorf("tanpa pola rute dapat %d deret, ingin 50 (inilah masalahnya)", got)
	}
}

// itoa kecil agar test tak perlu mengimpor strconv hanya untuk ini.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	return string(b)
}

func TestSplitBaris(t *testing.T) {
	tests := []struct {
		nama  string
		input string
		want  int
	}{
		{"tiga baris", "a\nb\nc", 3},
		{"baris terakhir berakhiran newline", "a\nb\n", 2},
		{"satu baris", "a", 1},
		{"kosong", "", 0},
	}
	for _, tt := range tests {
		t.Run(tt.nama, func(t *testing.T) {
			if got := len(splitBaris(tt.input)); got != tt.want {
				t.Errorf("splitBaris(%q) menghasilkan %d baris, ingin %d", tt.input, got, tt.want)
			}
		})
	}
}
