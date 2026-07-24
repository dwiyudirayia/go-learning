package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// minta menjalankan satu permintaan lewat router dan mengembalikan recorder-nya.
//
// 🔍 Analogi: httptest.NewRecorder itu "layar palsu" yang menangkap jawaban handler.
// Karena router chi hanyalah http.Handler biasa, kita bisa mengujinya tanpa menyalakan
// server sungguhan — sama seperti modul 12.
func minta(t *testing.T, r http.Handler, method, path string, body any, header map[string]string) *httptest.ResponseRecorder {
	t.Helper()

	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("gagal encode body: %v", err)
		}
	}

	req := httptest.NewRequest(method, path, &buf)
	for k, v := range header {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func TestSehat(t *testing.T) {
	r := NewRouter(NewTokoStore())
	rec := minta(t, r, http.MethodGet, "/sehat", nil, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, ingin 200", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, ingin application/json (middleware buatan sendiri)", ct)
	}
}

func TestDaftarProduk(t *testing.T) {
	r := NewRouter(NewTokoStore())
	rec := minta(t, r, http.MethodGet, "/produk", nil, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, ingin 200", rec.Code)
	}
	var produk []Produk
	if err := json.Unmarshal(rec.Body.Bytes(), &produk); err != nil {
		t.Fatalf("gagal decode: %v", err)
	}
	if len(produk) != 2 {
		t.Errorf("dapat %d produk awal, ingin 2", len(produk))
	}
}

// Inti nilai jual chi: parameter URL "/{id}".
func TestAmbilProdukLewatParameterURL(t *testing.T) {
	r := NewRouter(NewTokoStore())

	tests := []struct {
		nama       string
		path       string
		wantStatus int
	}{
		{"produk ada", "/produk/1", http.StatusOK},
		{"produk lain ada", "/produk/2", http.StatusOK},
		{"id tak ada", "/produk/999", http.StatusNotFound},
		{"id bukan angka", "/produk/abc", http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.nama, func(t *testing.T) {
			rec := minta(t, r, http.MethodGet, tt.path, nil, nil)
			if rec.Code != tt.wantStatus {
				t.Errorf("GET %s = %d, ingin %d", tt.path, rec.Code, tt.wantStatus)
			}
		})
	}
}

func TestBuatProduk(t *testing.T) {
	r := NewRouter(NewTokoStore())

	rec := minta(t, r, http.MethodPost, "/produk", Produk{Nama: "Gula", Harga: 12_000}, nil)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, ingin 201", rec.Code)
	}

	var dibuat Produk
	if err := json.Unmarshal(rec.Body.Bytes(), &dibuat); err != nil {
		t.Fatalf("gagal decode: %v", err)
	}
	if dibuat.ID == 0 {
		t.Error("server seharusnya mengisi ID")
	}
	if dibuat.Nama != "Gula" {
		t.Errorf("nama = %q, ingin Gula", dibuat.Nama)
	}
}

func TestBuatProdukTidakValid(t *testing.T) {
	r := NewRouter(NewTokoStore())

	t.Run("nama kosong", func(t *testing.T) {
		rec := minta(t, r, http.MethodPost, "/produk", Produk{Harga: 100}, nil)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("status = %d, ingin 400", rec.Code)
		}
	})

	t.Run("JSON rusak", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/produk", bytes.NewBufferString("{rusak"))
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("status = %d, ingin 400", rec.Code)
		}
	})
}

func TestHapusProduk(t *testing.T) {
	store := NewTokoStore()
	r := NewRouter(store)

	rec := minta(t, r, http.MethodDelete, "/produk/1", nil, nil)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, ingin 204", rec.Code)
	}
	// Sungguh terhapus dari store.
	if _, ada := store.Ambil(1); ada {
		t.Error("produk 1 seharusnya sudah terhapus")
	}
	// Menghapus lagi -> 404.
	rec = minta(t, r, http.MethodDelete, "/produk/1", nil, nil)
	if rec.Code != http.StatusNotFound {
		t.Errorf("hapus kedua = %d, ingin 404", rec.Code)
	}
}

// chi otomatis membalas 405 (bukan 404) saat path cocok tapi method salah.
func TestMethodTidakDiizinkan(t *testing.T) {
	r := NewRouter(NewTokoStore())

	rec := minta(t, r, http.MethodPut, "/produk/1", nil, nil)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("PUT /produk/1 = %d, ingin 405 (path ada, method tak didukung)", rec.Code)
	}
}

func TestRuteTidakDikenal(t *testing.T) {
	r := NewRouter(NewTokoStore())

	rec := minta(t, r, http.MethodGet, "/tidak-ada", nil, nil)
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, ingin 404", rec.Code)
	}
}

// Grup middleware: rute admin butuh token, rute lain tidak.
func TestGrupMiddlewareAdmin(t *testing.T) {
	r := NewRouter(NewTokoStore())

	t.Run("tanpa token ditolak", func(t *testing.T) {
		rec := minta(t, r, http.MethodGet, "/admin/statistik", nil, nil)
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, ingin 401", rec.Code)
		}
	})

	t.Run("token salah ditolak", func(t *testing.T) {
		rec := minta(t, r, http.MethodGet, "/admin/statistik", nil,
			map[string]string{"X-Admin-Token": "salah"})
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, ingin 401", rec.Code)
		}
	})

	t.Run("token benar diterima", func(t *testing.T) {
		rec := minta(t, r, http.MethodGet, "/admin/statistik", nil,
			map[string]string{"X-Admin-Token": "rahasia"})
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, ingin 200", rec.Code)
		}
		var stat map[string]int
		if err := json.Unmarshal(rec.Body.Bytes(), &stat); err != nil {
			t.Fatalf("gagal decode: %v", err)
		}
		if stat["jumlah_produk"] != 2 {
			t.Errorf("jumlah_produk = %d, ingin 2", stat["jumlah_produk"])
		}
	})

	// Bukti isolasi grup: middleware admin TIDAK berlaku untuk rute biasa.
	t.Run("rute biasa tidak butuh token", func(t *testing.T) {
		rec := minta(t, r, http.MethodGet, "/produk", nil, nil)
		if rec.Code != http.StatusOK {
			t.Errorf("status = %d, ingin 200 — rute non-admin tak boleh butuh token", rec.Code)
		}
	})
}

// Recoverer harus mengubah panic jadi 500, bukan merobohkan proses.
//
// Router kecil dirakit di sini dengan rantai middleware yang sama seperti NewRouter,
// lalu diberi satu rute yang sengaja panic — cara paling langsung membuktikan Recoverer
// benar-benar menangkap.
func TestRecovererMenangkapPanic(t *testing.T) {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Get("/panic", func(http.ResponseWriter, *http.Request) {
		panic("handler sengaja rusak")
	})

	rec := minta(t, r, http.MethodGet, "/panic", nil, nil)
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, ingin 500 — Recoverer sepertinya tak menangkap panic", rec.Code)
	}
}
