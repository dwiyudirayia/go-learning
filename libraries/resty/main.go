// go-resty/resty — HTTP client yang enak dipakai (JSON, retry, timeout, middleware).
//
// Jalankan: go run ./libraries/resty     (menyalakan API palsu sendiri — tanpa internet)
// Test:     go test ./libraries/resty
//
// 🔍 Analogi besar: net/http bawaan Go itu MOBIL MANUAL — kamu sendiri yang mengurus
// kopling: bikin request, set header, cek status code, baca body, tutup body, urai JSON.
// Semuanya bisa, tapi tiap panggilan API butuh 15 baris kode yang mirip-mirip.
// resty itu MOBIL MATIC: satu rantai pemanggilan mengurus semuanya, plus fitur yang
// biasanya kamu tulis sendiri — retry, timeout, dan header bersama.
//
// Kapan tetap pakai net/http: kalau kamu menulis library yang dipakai orang lain
// (jangan paksakan dependensi ke pengguna), atau kebutuhannya cuma satu-dua panggilan
// sederhana. resty menang saat aplikasimu bicara ke banyak API eksternal.
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/go-resty/resty/v2"
)

func main() {
	fmt.Println("=== go-resty/resty ===")

	// API palsu dinyalakan di dalam proses ini, jadi contohnya jalan tanpa internet.
	srv := httptest.NewServer(APIPalsu())
	defer srv.Close()
	fmt.Println("API palsu berjalan di", srv.URL)

	c := NewKlien(srv.URL)
	demoAmbil(c)
	demoBuat(c)
	demoCari(c)
	demoError(c)
	demoRetry()
}

// ------------------------------------------------------------------
// Model & error API
// ------------------------------------------------------------------

// User adalah bentuk data yang dipertukarkan dengan API.
type User struct {
	ID    int    `json:"id"`
	Nama  string `json:"nama"`
	Email string `json:"email"`
}

// APIError adalah bentuk BADAN error yang dikirim server (bukan error Go).
//
// 🔍 Analogi: bedakan dua jenis kegagalan. "Surat tak sampai ke kantor pos" (jaringan mati,
// timeout) itu error Go — resty mengembalikannya di variabel err. "Surat sampai, tapi
// isinya balasan penolakan" (HTTP 404/422) BUKAN error Go — resty tetap mengembalikan
// err=nil, dan status penolakannya ada di respons. Ini jebakan nomor satu pemakai baru:
// mengecek err saja tidak cukup, kamu WAJIB memeriksa status kodenya juga.
type APIError struct {
	Kode  string `json:"kode"`
	Pesan string `json:"pesan"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API menolak (%s): %s", e.Kode, e.Pesan)
}

// ErrTidakDitemukan sentinel agar pemanggil bisa membedakan 404 dari kegagalan lain.
var ErrTidakDitemukan = errors.New("user tidak ditemukan")

// ------------------------------------------------------------------
// 1. Membuat klien
// ------------------------------------------------------------------

// NewKlien membuat klien yang sudah dikonfigurasi sekali untuk dipakai berulang.
//
// 🔍 Analogi: klien resty itu seperti KANTOR PERWAKILAN yang kamu buka sekali di satu kota.
// Alamat kantor pusat (BaseURL), kop surat (header), dan aturan main (timeout, retry)
// diatur sekali di awal; setiap surat berikutnya tinggal ditulis isinya saja.
//
// Penting: buat SATU klien lalu pakai ulang. Membuat klien baru tiap panggilan membuang
// keuntungan connection pooling — seperti membangun kantor baru untuk tiap surat.
func NewKlien(baseURL string) *resty.Client {
	return resty.New().
		SetBaseURL(baseURL).
		// 🔍 Analogi timeout: tanpa ini, panggilan yang menggantung bisa menahan
		// goroutine-mu SELAMANYA. Timeout itu "kalau 5 detik belum dijawab, pulang saja."
		// Klien HTTP tanpa timeout adalah salah satu penyebab kebocoran goroutine
		// paling umum di layanan Go.
		SetTimeout(5*time.Second).
		SetHeader("Accept", "application/json").
		SetHeader("User-Agent", "go-learning/1.0").
		// Middleware: dijalankan sebelum SETIAP request keluar.
		OnBeforeRequest(func(_ *resty.Client, r *resty.Request) error {
			r.SetHeader("X-Request-ID", "req-"+strconv.FormatInt(time.Now().UnixNano(), 36))
			return nil
		})
}

// ------------------------------------------------------------------
// 2. GET dengan path parameter
// ------------------------------------------------------------------

// AmbilUser mengambil satu user berdasarkan ID.
//
// SetResult memberi tahu resty "kalau sukses, uraikan JSON-nya ke sini".
// SetError memberi tahu "kalau ditolak, uraikan JSON errornya ke sini".
// Keduanya menghapus kebutuhan menulis json.Unmarshal manual.
func AmbilUser(c *resty.Client, id int) (User, error) {
	var hasil User
	var apiErr APIError

	resp, err := c.R().
		SetPathParam("id", strconv.Itoa(id)).
		SetResult(&hasil).
		SetError(&apiErr).
		Get("/users/{id}")
	if err != nil {
		// Kegagalan transport: jaringan mati, DNS gagal, timeout.
		return User{}, fmt.Errorf("gagal menghubungi API: %w", err)
	}

	// Kegagalan di tingkat aplikasi: surat sampai, tapi jawabannya penolakan.
	if resp.StatusCode() == http.StatusNotFound {
		return User{}, fmt.Errorf("id %d: %w", id, ErrTidakDitemukan)
	}
	if resp.IsError() {
		return User{}, fmt.Errorf("ambil user %d: %w", id, &apiErr)
	}
	return hasil, nil
}

func demoAmbil(c *resty.Client) {
	fmt.Println("\n-- GET /users/{id} --")
	u, err := AmbilUser(c, 1)
	if err != nil {
		fmt.Println("   error:", err)
		return
	}
	fmt.Printf("   %+v\n", u)
}

// ------------------------------------------------------------------
// 3. POST dengan badan JSON
// ------------------------------------------------------------------

// BuatUser mengirim user baru. SetBody otomatis mengubah struct jadi JSON
// dan memasang header Content-Type — dua hal yang harus manual di net/http.
func BuatUser(c *resty.Client, u User) (User, error) {
	var hasil User
	var apiErr APIError

	resp, err := c.R().
		SetBody(u).
		SetResult(&hasil).
		SetError(&apiErr).
		Post("/users")
	if err != nil {
		return User{}, fmt.Errorf("gagal menghubungi API: %w", err)
	}
	if resp.IsError() {
		return User{}, fmt.Errorf("buat user: %w", &apiErr)
	}
	return hasil, nil
}

func demoBuat(c *resty.Client) {
	fmt.Println("\n-- POST /users --")
	u, err := BuatUser(c, User{Nama: "Citra", Email: "citra@contoh.id"})
	if err != nil {
		fmt.Println("   error:", err)
		return
	}
	fmt.Printf("   dibuat -> %+v\n", u)

	// Badan yang tak lolos validasi server dijawab 422, bukan error jaringan.
	if _, err := BuatUser(c, User{Nama: "", Email: "bukan-email"}); err != nil {
		fmt.Println("   validasi ditolak ->", err)
	}
}

// ------------------------------------------------------------------
// 4. Query parameter
// ------------------------------------------------------------------

// CariUser memakai query string: /users?q=...&limit=...
func CariUser(c *resty.Client, q string, limit int) ([]User, error) {
	var hasil []User

	resp, err := c.R().
		SetQueryParam("q", q).
		SetQueryParam("limit", strconv.Itoa(limit)).
		SetResult(&hasil).
		Get("/users")
	if err != nil {
		return nil, fmt.Errorf("gagal menghubungi API: %w", err)
	}
	if resp.IsError() {
		return nil, fmt.Errorf("cari user: status %d", resp.StatusCode())
	}
	return hasil, nil
}

func demoCari(c *resty.Client) {
	fmt.Println("\n-- GET /users?q=&limit= --")
	us, err := CariUser(c, "a", 2)
	if err != nil {
		fmt.Println("   error:", err)
		return
	}
	for _, u := range us {
		fmt.Printf("   %d %s\n", u.ID, u.Nama)
	}
}

func demoError(c *resty.Client) {
	fmt.Println("\n-- Membedakan jenis kegagalan --")

	if _, err := AmbilUser(c, 999); errors.Is(err, ErrTidakDitemukan) {
		fmt.Println("   404 dikenali sebagai 'tidak ditemukan':", err)
	}

	// Klien yang menunjuk ke alamat mati: INI baru error transport.
	mati := NewKlien("http://127.0.0.1:1")
	if _, err := AmbilUser(mati, 1); err != nil {
		fmt.Println("   alamat mati -> error transport (bukan status HTTP)")
	}
}

// ------------------------------------------------------------------
// 5. Retry — mencoba lagi, tapi hanya untuk kegagalan yang PANTAS diulang
// ------------------------------------------------------------------

// 🔍 Analogi: retry itu seperti menelepon ulang saat nada sibuk. Masuk akal untuk nada
// sibuk (server kelebihan muatan, jaringan tersendat) — TAPI konyol kalau nomornya
// memang salah (404) atau datamu memang tak valid (422). Menelepon ulang nomor salah
// seribu kali tak akan membuatnya jadi benar; ia cuma menambah beban server yang
// sudah kepayahan.
//
// Aturannya: ulangi HANYA untuk 5xx, 429, dan kegagalan jaringan. Jangan pernah
// mengulang 4xx (kecuali 429), dan jangan pernah mengulang POST yang tidak idempoten
// tanpa kunci idempotensi — kamu bisa membuat pesanan ganda.

// NewKlienRetry membuat klien yang mencoba ulang dengan jeda yang MENINGKAT.
//
// 🔍 Analogi jeda meningkat (backoff): seperti mengetuk pintu — kalau tak dijawab, jangan
// mengetuk makin cepat. Beri jeda makin lama. Kalau 1000 klien serentak mencoba ulang
// tiap 100ms, server yang sedang sekarat akan makin tertimbun (istilahnya: thundering herd).
func NewKlienRetry(baseURL string, maks int) *resty.Client {
	return NewKlien(baseURL).
		SetRetryCount(maks).
		SetRetryWaitTime(20 * time.Millisecond).
		SetRetryMaxWaitTime(200 * time.Millisecond).
		AddRetryCondition(func(r *resty.Response, err error) bool {
			if err != nil {
				return true // kegagalan jaringan: pantas diulang
			}
			// 429 = diminta pelan-pelan; 5xx = server sedang bermasalah.
			return r.StatusCode() == http.StatusTooManyRequests || r.StatusCode() >= 500
		})
}

// ServerRapuh membuat server yang gagal 'gagalDulu' kali pertama, lalu sukses.
// Dipakai untuk membuktikan retry bekerja — dan menghitung berapa kali dicoba.
func ServerRapuh(gagalDulu int32) (*httptest.Server, *atomic.Int32) {
	var percobaan atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if percobaan.Add(1) <= gagalDulu {
			w.WriteHeader(http.StatusServiceUnavailable) // 503: pantas diulang
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(User{ID: 1, Nama: "Ana", Email: "ana@contoh.id"})
	}))
	return srv, &percobaan
}

func demoRetry() {
	fmt.Println("\n-- Retry --")

	srv, percobaan := ServerRapuh(2) // gagal 2 kali, sukses di percobaan ke-3
	defer srv.Close()

	c := NewKlienRetry(srv.URL, 3)
	u, err := AmbilUser(c, 1)
	if err != nil {
		fmt.Println("   error:", err)
		return
	}
	fmt.Printf("   sukses setelah %d percobaan -> %s\n", percobaan.Load(), u.Nama)
}

// ------------------------------------------------------------------
// API palsu — supaya seluruh contoh jalan tanpa internet
// ------------------------------------------------------------------

// APIPalsu mengembalikan handler HTTP berisi beberapa endpoint contoh.
//
// 🔍 Analogi: ini RESTORAN MAINAN. Kita tak mau contoh belajar bergantung pada API
// internet yang bisa mati, berubah, atau memblokir. Server palsu di dalam proses membuat
// contoh (dan test) selalu bisa diulang dengan hasil sama — pola yang sama dipakai di
// modul 09 dengan httptest.
func APIPalsu() http.Handler {
	users := []User{
		{ID: 1, Nama: "Ana", Email: "ana@contoh.id"},
		{ID: 2, Nama: "Budi", Email: "budi@contoh.id"},
		{ID: 3, Nama: "Cakra", Email: "cakra@contoh.id"},
	}

	tulisJSON := func(w http.ResponseWriter, status int, v any) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(v)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /users/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			tulisJSON(w, http.StatusBadRequest, APIError{Kode: "id_tidak_valid", Pesan: "id harus angka"})
			return
		}
		for _, u := range users {
			if u.ID == id {
				tulisJSON(w, http.StatusOK, u)
				return
			}
		}
		tulisJSON(w, http.StatusNotFound, APIError{Kode: "tidak_ditemukan", Pesan: "user tidak ada"})
	})

	mux.HandleFunc("GET /users", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query().Get("q")
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		if limit <= 0 {
			limit = len(users)
		}

		hasil := make([]User, 0, len(users))
		for _, u := range users {
			if q == "" || containsFold(u.Nama, q) {
				hasil = append(hasil, u)
			}
			if len(hasil) == limit {
				break
			}
		}
		tulisJSON(w, http.StatusOK, hasil)
	})

	mux.HandleFunc("POST /users", func(w http.ResponseWriter, r *http.Request) {
		var u User
		if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
			tulisJSON(w, http.StatusBadRequest, APIError{Kode: "json_rusak", Pesan: "badan bukan JSON valid"})
			return
		}
		if u.Nama == "" || !containsFold(u.Email, "@") {
			tulisJSON(w, http.StatusUnprocessableEntity,
				APIError{Kode: "validasi_gagal", Pesan: "nama wajib diisi & email harus sah"})
			return
		}
		u.ID = len(users) + 1
		tulisJSON(w, http.StatusCreated, u)
	})

	return mux
}

// containsFold cek substring tanpa peduli huruf besar/kecil (versi ringkas untuk contoh).
func containsFold(s, sub string) bool {
	if sub == "" {
		return true
	}
	for i := 0; i+len(sub) <= len(s); i++ {
		cocok := true
		for j := range len(sub) {
			a, b := s[i+j], sub[j]
			if 'A' <= a && a <= 'Z' {
				a += 'a' - 'A'
			}
			if 'A' <= b && b <= 'Z' {
				b += 'a' - 'A'
			}
			if a != b {
				cocok = false
				break
			}
		}
		if cocok {
			return true
		}
	}
	return false
}
