package main

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
)

// klienUji menyalakan API palsu lalu mengembalikan klien yang menunjuk ke sana.
// Server otomatis ditutup saat test selesai lewat t.Cleanup.
func klienUji(t *testing.T) *resty.Client {
	t.Helper()
	srv := httptest.NewServer(APIPalsu())
	t.Cleanup(srv.Close)
	return NewKlien(srv.URL)
}

func TestAmbilUser(t *testing.T) {
	c := klienUji(t)

	u, err := AmbilUser(c, 2)
	if err != nil {
		t.Fatalf("AmbilUser(2) error: %v", err)
	}
	if u.ID != 2 || u.Nama != "Budi" {
		t.Errorf("AmbilUser(2) = %+v, ingin ID 2 / Budi", u)
	}
}

// 404 BUKAN error transport — pastikan ia diterjemahkan jadi sentinel yang bisa dikenali.
func TestAmbilUserTidakDitemukan(t *testing.T) {
	c := klienUji(t)

	_, err := AmbilUser(c, 999)
	if err == nil {
		t.Fatal("ingin error untuk id yang tak ada")
	}
	if !errors.Is(err, ErrTidakDitemukan) {
		t.Errorf("error = %v, ingin membungkus ErrTidakDitemukan", err)
	}
}

func TestAmbilUserAlamatMati(t *testing.T) {
	// Port 1 hampir pasti tak ada yang mendengarkan -> kegagalan transport.
	c := NewKlien("http://127.0.0.1:1")

	_, err := AmbilUser(c, 1)
	if err == nil {
		t.Fatal("ingin error transport")
	}
	// Ini kegagalan jaringan, BUKAN 404 — jangan sampai tertukar.
	if errors.Is(err, ErrTidakDitemukan) {
		t.Error("kegagalan jaringan seharusnya tidak dikira 'tidak ditemukan'")
	}
}

func TestBuatUser(t *testing.T) {
	c := klienUji(t)

	u, err := BuatUser(c, User{Nama: "Citra", Email: "citra@contoh.id"})
	if err != nil {
		t.Fatalf("BuatUser error: %v", err)
	}
	if u.ID == 0 {
		t.Error("server seharusnya mengisi ID pada user baru")
	}
	if u.Nama != "Citra" {
		t.Errorf("nama = %q, ingin Citra", u.Nama)
	}
}

func TestBuatUserValidasiGagal(t *testing.T) {
	c := klienUji(t)

	tests := []struct {
		nama  string
		input User
	}{
		{"nama kosong", User{Nama: "", Email: "a@b.id"}},
		{"email tanpa @", User{Nama: "Dedi", Email: "bukan-email"}},
		{"dua-duanya kosong", User{}},
	}

	for _, tt := range tests {
		t.Run(tt.nama, func(t *testing.T) {
			_, err := BuatUser(c, tt.input)
			if err == nil {
				t.Fatal("ingin error validasi dari server")
			}
			// Badan error dari server harus terurai ke APIError, bukan sekadar teks mentah.
			var apiErr *APIError
			if !errors.As(err, &apiErr) {
				t.Fatalf("error = %v, ingin bertipe *APIError", err)
			}
			if apiErr.Kode != "validasi_gagal" {
				t.Errorf("kode = %q, ingin validasi_gagal", apiErr.Kode)
			}
		})
	}
}

func TestCariUser(t *testing.T) {
	c := klienUji(t)

	tests := []struct {
		nama      string
		q         string
		limit     int
		wantMin   int
		wantTepat int // -1 berarti tidak diperiksa
	}{
		{"tanpa kata kunci ambil semua", "", 10, 3, 3},
		{"limit memotong hasil", "", 2, 2, 2},
		{"cari huruf a", "a", 10, 1, -1},
		{"kata kunci tak cocok", "zzz", 10, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.nama, func(t *testing.T) {
			us, err := CariUser(c, tt.q, tt.limit)
			if err != nil {
				t.Fatalf("CariUser error: %v", err)
			}
			if len(us) < tt.wantMin {
				t.Errorf("dapat %d hasil, ingin minimal %d", len(us), tt.wantMin)
			}
			if tt.wantTepat >= 0 && len(us) != tt.wantTepat {
				t.Errorf("dapat %d hasil, ingin tepat %d", len(us), tt.wantTepat)
			}
		})
	}
}

// Middleware OnBeforeRequest harus memasang header di SETIAP request.
func TestMiddlewareMemasangRequestID(t *testing.T) {
	var terlihat atomic.Value
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		terlihat.Store(r.Header.Get("X-Request-ID"))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":1,"nama":"Ana","email":"ana@contoh.id"}`))
	}))
	defer srv.Close()

	if _, err := AmbilUser(NewKlien(srv.URL), 1); err != nil {
		t.Fatalf("AmbilUser error: %v", err)
	}
	id, _ := terlihat.Load().(string)
	if id == "" {
		t.Error("X-Request-ID tidak terpasang oleh middleware")
	}
}

// Retry harus mencoba ulang 503 sampai berhasil.
func TestRetryMengulangSampaiSukses(t *testing.T) {
	srv, percobaan := ServerRapuh(2) // gagal 2x, sukses di percobaan ke-3
	defer srv.Close()

	c := NewKlienRetry(srv.URL, 3)
	u, err := AmbilUser(c, 1)
	if err != nil {
		t.Fatalf("ingin sukses setelah retry, dapat error: %v", err)
	}
	if u.Nama != "Ana" {
		t.Errorf("nama = %q, ingin Ana", u.Nama)
	}
	if got := percobaan.Load(); got != 3 {
		t.Errorf("jumlah percobaan = %d, ingin 3 (1 asli + 2 ulangan)", got)
	}
}

// Kalau jatah retry habis dan server tetap gagal, error harus dikembalikan.
func TestRetryMenyerahSetelahJatahHabis(t *testing.T) {
	srv, percobaan := ServerRapuh(100) // selalu gagal
	defer srv.Close()

	c := NewKlienRetry(srv.URL, 2)
	if _, err := AmbilUser(c, 1); err == nil {
		t.Fatal("ingin error setelah jatah retry habis")
	}
	if got := percobaan.Load(); got != 3 {
		t.Errorf("jumlah percobaan = %d, ingin 3 (1 asli + 2 ulangan)", got)
	}
}

// Inti disiplin retry: 4xx TIDAK boleh diulang — mengulangnya sia-sia & membebani server.
func TestRetryTidakMengulangKesalahanKlien(t *testing.T) {
	var percobaan atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		percobaan.Add(1)
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	c := NewKlienRetry(srv.URL, 3)
	if _, err := AmbilUser(c, 1); !errors.Is(err, ErrTidakDitemukan) {
		t.Errorf("error = %v, ingin ErrTidakDitemukan", err)
	}
	if got := percobaan.Load(); got != 1 {
		t.Errorf("jumlah percobaan = %d, ingin tepat 1 (404 tak boleh diulang)", got)
	}
}

// Server lambat + timeout pendek = error transport, bukan menggantung selamanya.
func TestTimeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(300 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := NewKlien(srv.URL).SetTimeout(50 * time.Millisecond)

	mulai := time.Now()
	_, err := AmbilUser(c, 1)
	lama := time.Since(mulai)

	if err == nil {
		t.Fatal("ingin error timeout")
	}
	if lama > 250*time.Millisecond {
		t.Errorf("permintaan berjalan %v — timeout sepertinya tidak berlaku", lama)
	}
}

func TestContainsFold(t *testing.T) {
	tests := []struct {
		s, sub string
		want   bool
	}{
		{"Ana", "a", true},
		{"Ana", "A", true},
		{"Budi", "ud", true},
		{"Budi", "UD", true},
		{"Budi", "x", false},
		{"", "a", false},
		{"apa pun", "", true},
	}
	for _, tt := range tests {
		if got := containsFold(tt.s, tt.sub); got != tt.want {
			t.Errorf("containsFold(%q, %q) = %t, ingin %t", tt.s, tt.sub, got, tt.want)
		}
	}
}
