// go-chi/chi — router HTTP idiomatik: tetap dekat net/http, tapi jauh lebih nyaman.
//
// Jalankan: go run ./libraries/chi     (server di :8085; atau lihat test)
// Test:     go test ./libraries/chi
//
// 🔍 Analogi besar: kalau net/http mentah itu MOBIL MANUAL dan Fiber (modul 13) itu MOBIL
// MATIC penuh fitur, maka chi itu MOBIL MATIC RINGAN — ia menambahkan hal-hal yang paling
// sering kamu butuhkan (parameter URL, grup rute, rantai middleware) TANPA menyeretmu jauh
// dari standar. Kuncinya: handler chi tetap `http.HandlerFunc` biasa, dan middleware-nya
// tetap `func(http.Handler) http.Handler` biasa. Artinya seluruh ekosistem net/http tetap
// bisa dipakai, dan kamu tak "terkunci" pada satu framework.
//
// Kapan pilih apa:
//   - net/http + ServeMux (Go 1.22+)  : kebutuhan sederhana, nol dependensi.
//   - chi                              : butuh grup rute, middleware bertingkat, sub-router,
//     tapi ingin tetap "terasa seperti Go standar".
//   - Fiber/Gin/Echo                   : ingin framework penuh dengan banyak baterai bawaan.
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	fmt.Println("=== go-chi/chi ===")
	fmt.Println("Server di http://localhost:8085 — coba: curl localhost:8085/produk")

	srv := &http.Server{
		Addr:              ":8085",
		Handler:           NewRouter(NewTokoStore()),
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())
}

// ------------------------------------------------------------------
// Data
// ------------------------------------------------------------------

// Produk adalah entitas contoh.
type Produk struct {
	ID    int    `json:"id"`
	Nama  string `json:"nama"`
	Harga int    `json:"harga"`
}

// TokoStore penyimpanan in-memory yang aman diakses banyak goroutine.
//
// 🔍 Analogi: sama seperti modul 12, store DISUNTIKKAN ke router lewat closure, bukan
// variabel global. Efeknya: tiap test punya store bersih sendiri.
type TokoStore struct {
	mu      sync.Mutex
	berikut int
	data    map[int]Produk
}

func NewTokoStore() *TokoStore {
	s := &TokoStore{berikut: 1, data: make(map[int]Produk)}
	s.Tambah(Produk{Nama: "Kopi", Harga: 25_000})
	s.Tambah(Produk{Nama: "Teh", Harga: 15_000})
	return s
}

func (s *TokoStore) Tambah(p Produk) Produk {
	s.mu.Lock()
	defer s.mu.Unlock()
	p.ID = s.berikut
	s.data[p.ID] = p
	s.berikut++
	return p
}

func (s *TokoStore) Ambil(id int) (Produk, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ada := s.data[id]
	return p, ada
}

func (s *TokoStore) Semua() []Produk {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Produk, 0, len(s.data))
	for _, p := range s.data {
		out = append(out, p)
	}
	return out
}

func (s *TokoStore) Hapus(id int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ada := s.data[id]; !ada {
		return false
	}
	delete(s.data, id)
	return true
}

// ------------------------------------------------------------------
// Router
// ------------------------------------------------------------------

// NewRouter merakit seluruh rute & middleware.
//
// 🔍 Analogi struktur rute chi: seperti DENAH GEDUNG berlapis. Middleware yang dipasang
// di lantai atas berlaku untuk semua ruangan di bawahnya; Route() membuat "sayap gedung"
// dengan awalan alamat & penjaga sendiri. Semuanya tetap memakai http.Handler standar.
func NewRouter(store *TokoStore) http.Handler {
	r := chi.NewRouter()

	// 🔍 Analogi middleware bawaan chi: ini "petugas gedung" siap pakai.
	//   RequestID  = memberi tiap tamu nomor antrean unik (untuk melacak di log).
	//   Recoverer  = jaring pengaman: handler yang panic tak merobohkan server,
	//                dijawab 500 dengan rapi.
	// Semuanya cuma func(http.Handler) http.Handler — kamu bisa menulis sendiri (lihat
	// contentTypeJSON di bawah) atau meminjam dari mana pun.
	//
	// Catatan: chi juga punya middleware.RealIP, TAPI kini deprecated karena rentan
	// pemalsuan IP (ia mempercayai header X-Forwarded-For mentah). Ambil IP asli hanya
	// dari proxy yang kamu percayai — jangan dari header yang bisa diisi siapa saja.
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(contentTypeJSON) // middleware buatan sendiri

	h := &tokoHandler{store: store}

	// Rute datar.
	r.Get("/sehat", func(w http.ResponseWriter, _ *http.Request) {
		tulisJSON(w, http.StatusOK, map[string]string{"status": "sehat"})
	})

	// 🔍 Analogi Route(): membuat "sayap /produk" dengan rutenya sendiri. Alih-alih
	// menulis "/produk" berulang-ulang, kamu menuliskannya sekali sebagai awalan.
	r.Route("/produk", func(r chi.Router) {
		r.Get("/", h.daftar)       // GET  /produk
		r.Post("/", h.buat)        // POST /produk
		r.Get("/{id}", h.ambil)    // GET  /produk/42
		r.Delete("/{id}", h.hapus) // DELETE /produk/42
	})

	// 🔍 Analogi Group(): seperti Route() tapi TANPA awalan alamat — dipakai untuk
	// menerapkan middleware ke sekelompok rute saja. Di sini "sayap admin" dijaga
	// pemeriksa token; rute di luar grup tak terpengaruh.
	r.Group(func(r chi.Router) {
		r.Use(butuhTokenAdmin)
		r.Get("/admin/statistik", func(w http.ResponseWriter, _ *http.Request) {
			tulisJSON(w, http.StatusOK, map[string]int{"jumlah_produk": len(store.Semua())})
		})
	})

	return r
}

// ------------------------------------------------------------------
// Handler
// ------------------------------------------------------------------

type tokoHandler struct {
	store *TokoStore
}

func (h *tokoHandler) daftar(w http.ResponseWriter, _ *http.Request) {
	tulisJSON(w, http.StatusOK, h.store.Semua())
}

func (h *tokoHandler) buat(w http.ResponseWriter, r *http.Request) {
	var p Produk
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		tulisError(w, http.StatusBadRequest, "JSON tidak valid")
		return
	}
	if p.Nama == "" {
		tulisError(w, http.StatusBadRequest, "nama wajib diisi")
		return
	}
	tulisJSON(w, http.StatusCreated, h.store.Tambah(p))
}

// ambil memperagakan pengambilan parameter URL — nilai jual utama chi dibanding
// ServeMux versi lama.
//
// 🔍 Analogi chi.URLParam: pola "/{id}" itu seperti KOLOM ISIAN di formulir alamat.
// chi.URLParam(r, "id") membaca apa yang diisi tamu di kolom itu. (Sejak Go 1.22,
// net/http punya r.PathValue yang setara — chi menang saat kamu butuh grup & sub-router.)
func (h *tokoHandler) ambil(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		tulisError(w, http.StatusBadRequest, "id harus angka")
		return
	}
	p, ada := h.store.Ambil(id)
	if !ada {
		tulisError(w, http.StatusNotFound, "produk tidak ditemukan")
		return
	}
	tulisJSON(w, http.StatusOK, p)
}

func (h *tokoHandler) hapus(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		tulisError(w, http.StatusBadRequest, "id harus angka")
		return
	}
	if !h.store.Hapus(id) {
		tulisError(w, http.StatusNotFound, "produk tidak ditemukan")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ------------------------------------------------------------------
// Middleware buatan sendiri
// ------------------------------------------------------------------

// contentTypeJSON memasang header Content-Type untuk semua respons.
//
// 🔍 Analogi: membuktikan middleware chi itu TIDAK ISTIMEWA — cuma fungsi biasa yang
// menerima handler & mengembalikan handler. Pola "bungkus" yang sama persis dengan modul 12.
func contentTypeJSON(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

// ErrTokenAdmin dipakai internal untuk kejelasan (token admin sederhana untuk contoh).
var ErrTokenAdmin = errors.New("token admin tidak sah")

// butuhTokenAdmin menolak permintaan tanpa header "X-Admin-Token: rahasia".
//
// (Contoh saja — di produksi pakai JWT/sesi sungguhan; lihat libraries/jwt.)
func butuhTokenAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Admin-Token") != "rahasia" {
			tulisError(w, http.StatusUnauthorized, "butuh token admin")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// ------------------------------------------------------------------
// Helper
// ------------------------------------------------------------------

func tulisJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func tulisError(w http.ResponseWriter, status int, pesan string) {
	tulisJSON(w, status, map[string]string{"error": pesan})
}
