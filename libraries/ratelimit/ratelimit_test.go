package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// titikNol adalah "sekarang" tetap agar seluruh test berjalan tanpa menunggu waktu nyata.
var titikNol = time.Date(2026, 7, 23, 10, 0, 0, 0, time.UTC)

func TestKapasitasMenentukanLonjakanAwal(t *testing.T) {
	tests := []struct {
		nama      string
		perDetik  float64
		kapasitas int
		kirim     int
		wantLolos int
	}{
		{"ember 5 menerima 5 dari 10", 5, 5, 10, 5},
		{"ember 1 hanya menerima 1", 5, 1, 10, 1},
		{"ember 10 menerima semua", 5, 10, 10, 10},
		{"ember lebih besar dari kiriman", 5, 20, 3, 3},
	}

	for _, tt := range tests {
		t.Run(tt.nama, func(t *testing.T) {
			l := NewPembatas(tt.perDetik, tt.kapasitas)
			if got := HitungLolos(l, titikNol, tt.kirim); got != tt.wantLolos {
				t.Errorf("lolos = %d, ingin %d", got, tt.wantLolos)
			}
		})
	}
}

// Setelah ember kosong, token terisi kembali sesuai laju — bukan sekaligus.
func TestEmberTerisiKembaliSesuaiWaktu(t *testing.T) {
	l := NewPembatas(5, 5) // 5 token/detik

	// Kuras habis.
	if got := HitungLolos(l, titikNol, 5); got != 5 {
		t.Fatalf("penyiapan gagal: lolos %d, ingin 5", got)
	}
	if l.AllowN(titikNol, 1) {
		t.Fatal("ember seharusnya sudah kosong")
	}

	tests := []struct {
		nama      string
		maju      time.Duration
		wantLolos int
	}{
		{"200ms mengisi 1 token", 200 * time.Millisecond, 1},
		{"600ms mengisi 3 token", 600 * time.Millisecond, 3},
		{"1 detik mengisi 5 token", time.Second, 5},
		{"10 detik tetap maksimal sekapasitas ember", 10 * time.Second, 5},
	}

	for _, tt := range tests {
		t.Run(tt.nama, func(t *testing.T) {
			// Pembatas baru tiap sub-test supaya perhitungannya bersih.
			l := NewPembatas(5, 5)
			HitungLolos(l, titikNol, 5) // kuras

			got := HitungLolos(l, titikNol.Add(tt.maju), 10)
			if got != tt.wantLolos {
				t.Errorf("setelah %v lolos = %d, ingin %d", tt.maju, got, tt.wantLolos)
			}
		})
	}
}

func TestNewPembatasInterval(t *testing.T) {
	// Satu izin tiap 100ms = 10 per detik.
	l := NewPembatasInterval(100*time.Millisecond, 1)

	if !l.AllowN(titikNol, 1) {
		t.Fatal("permintaan pertama seharusnya lolos")
	}
	if l.AllowN(titikNol.Add(50*time.Millisecond), 1) {
		t.Error("50ms kemudian seharusnya masih ditolak")
	}
	if !l.AllowN(titikNol.Add(100*time.Millisecond), 1) {
		t.Error("100ms kemudian seharusnya sudah lolos")
	}
}

func TestWaitMengantreSampaiContextHabis(t *testing.T) {
	// 100 per detik; dalam 120ms wajar bila hanya belasan yang selesai.
	l := NewPembatasInterval(10*time.Millisecond, 1)

	ctx, batal := context.WithTimeout(context.Background(), 120*time.Millisecond)
	defer batal()

	selesai, err := KerjaDenganAntrean(ctx, l, 1000)
	if err == nil {
		t.Fatal("ingin error karena context habis sebelum 1000 pekerjaan selesai")
	}
	if selesai == 0 {
		t.Error("setidaknya beberapa pekerjaan seharusnya sempat berjalan")
	}
	if selesai >= 1000 {
		t.Error("seharusnya tidak mungkin menyelesaikan semuanya secepat itu")
	}
}

func TestWaitContextSudahDibatalkan(t *testing.T) {
	l := NewPembatas(1000, 1000)
	ctx, batal := context.WithCancel(context.Background())
	batal()

	if _, err := KerjaDenganAntrean(ctx, l, 5); err == nil {
		t.Error("context yang sudah dibatalkan seharusnya langsung menghentikan pekerjaan")
	}
}

// Tiap kunci punya embernya sendiri — penyalahguna tak boleh merugikan pengguna lain.
func TestPembatasPerKunciTerpisah(t *testing.T) {
	p := NewPembatasPerKunci(5, 2).DenganJam(func() time.Time { return titikNol })

	// Kuras jatah IP pertama (kapasitas ember = 2).
	for i := 1; i <= 2; i++ {
		if !p.Izinkan("1.1.1.1") {
			t.Fatalf("permintaan ke-%d seharusnya lolos", i)
		}
	}
	if p.Izinkan("1.1.1.1") {
		t.Error("permintaan ketiga dari IP yang sama seharusnya ditolak")
	}

	// IP lain sama sekali tak terpengaruh.
	if !p.Izinkan("2.2.2.2") {
		t.Error("IP berbeda seharusnya punya jatah sendiri")
	}
	if p.Jumlah() != 2 {
		t.Errorf("kunci dilacak = %d, ingin 2", p.Jumlah())
	}
}

// Map pembatas harus bisa dibersihkan, kalau tidak ia bocor tanpa batas.
func TestBersihkanMembuangKunciLama(t *testing.T) {
	kini := titikNol
	p := NewPembatasPerKunci(5, 2).DenganJam(func() time.Time { return kini })

	p.Izinkan("lama-1")
	p.Izinkan("lama-2")

	// Majukan waktu, lalu satu kunci dipakai lagi (jadi "segar").
	kini = titikNol.Add(10 * time.Minute)
	p.Izinkan("baru")

	if got := p.Jumlah(); got != 3 {
		t.Fatalf("sebelum dibersihkan ada %d kunci, ingin 3", got)
	}

	dibuang := p.Bersihkan(5 * time.Minute)
	if dibuang != 2 {
		t.Errorf("dibuang %d kunci, ingin 2", dibuang)
	}
	if got := p.Jumlah(); got != 1 {
		t.Errorf("sisa %d kunci, ingin 1 (yang masih segar)", got)
	}
}

func TestBersihkanTidakMembuangYangMasihAktif(t *testing.T) {
	p := NewPembatasPerKunci(5, 2).DenganJam(func() time.Time { return titikNol })
	p.Izinkan("aktif")

	if dibuang := p.Bersihkan(time.Hour); dibuang != 0 {
		t.Errorf("dibuang %d, ingin 0 — kunci aktif tidak boleh dihapus", dibuang)
	}
}

func TestMiddlewareMenolakDengan429(t *testing.T) {
	p := NewPembatasPerKunci(1, 2).DenganJam(func() time.Time { return titikNol })

	h := MiddlewarePembatas(p, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	kirim := func() *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "10.0.0.1:12345"
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		return rec
	}

	// Dua permintaan pertama masuk jatah ember.
	for i := 1; i <= 2; i++ {
		if got := kirim().Code; got != http.StatusOK {
			t.Errorf("permintaan %d = %d, ingin 200", i, got)
		}
	}

	// Yang ketiga ditolak, dan harus memberi tahu kapan boleh mencoba lagi.
	rec := kirim()
	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("permintaan ketiga = %d, ingin 429", rec.Code)
	}
	if rec.Header().Get("Retry-After") == "" {
		t.Error("respons 429 seharusnya menyertakan header Retry-After")
	}
}

// Klien berbeda tidak boleh saling menghabiskan jatah.
func TestMiddlewareMemisahkanAntarKlien(t *testing.T) {
	p := NewPembatasPerKunci(1, 1).DenganJam(func() time.Time { return titikNol })
	h := MiddlewarePembatas(p, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	kirim := func(ip string) int {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = ip + ":9999"
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		return rec.Code
	}

	if got := kirim("10.0.0.1"); got != http.StatusOK {
		t.Errorf("klien A permintaan 1 = %d, ingin 200", got)
	}
	if got := kirim("10.0.0.1"); got != http.StatusTooManyRequests {
		t.Errorf("klien A permintaan 2 = %d, ingin 429", got)
	}
	if got := kirim("10.0.0.2"); got != http.StatusOK {
		t.Errorf("klien B = %d, ingin 200 — jatahnya terpisah", got)
	}
}

func TestKunciDariRequest(t *testing.T) {
	tests := []struct {
		nama       string
		remoteAddr string
		want       string
	}{
		{"IPv4 dengan port", "192.168.1.10:54321", "192.168.1.10"},
		{"IPv6 dengan port", "[::1]:8080", "::1"},
		{"tanpa port dipakai apa adanya", "192.168.1.10", "192.168.1.10"},
	}

	for _, tt := range tests {
		t.Run(tt.nama, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.RemoteAddr = tt.remoteAddr
			if got := KunciDariRequest(req); got != tt.want {
				t.Errorf("KunciDariRequest(%q) = %q, ingin %q", tt.remoteAddr, got, tt.want)
			}
		})
	}
}
