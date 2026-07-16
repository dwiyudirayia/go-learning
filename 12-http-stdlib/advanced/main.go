// Teknik Advanced Modul 12 (net/http) — contoh runnable + penjelasan detail.
//
// Semua berjalan IN-PROCESS via httptest (tanpa port nyata). Jalankan:
//
//	go run ./12-http-stdlib/advanced
package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"
)

// ===========================================================================
// 1. Middleware = func(http.Handler) http.Handler
// ===========================================================================
// Middleware membungkus handler: kerjakan sesuatu sebelum/sesudah, lalu
// panggil next. Dirantai agar tiap lapis punya satu tanggung jawab.
//
// 🔍 Analogi: middleware itu seperti POS PEMERIKSAAN BERLAPIS di bandara —
// tiap pos (logging, auth, rate limit) satu tugas, lalu mempersilakan
// penumpang lanjut ke pos berikutnya (next). Penumpang yang sama pasti
// melewati semua pos dalam urutan yang sama.

func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mulai := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("[akses] %s %s (%s)", r.Method, r.URL.Path, time.Since(mulai))
	})
}

// Chain menerapkan middleware berurutan: Chain(h, a, b) => a(b(h)).
func Chain(h http.Handler, mws ...func(http.Handler) http.Handler) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}

func main() {
	// =====================================================================
	// 2. Routing bawaan Go 1.22: METHOD + wildcard {id} + PathValue
	// =====================================================================
	// ServeMux modern bisa mencocokkan method dan menangkap segmen path —
	// tak perlu router pihak ketiga untuk kasus umum.
	//
	// 🔍 Analogi: pattern "GET /users/{id}" itu seperti FORMULIR DENGAN KOLOM
	// KOSONG — resepsionis (mux) mencocokkan formatnya, lalu isi kolom {id}
	// tinggal dibaca lewat r.PathValue("id").
	mux := http.NewServeMux()

	mux.HandleFunc("GET /users/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id") // ambil segmen wildcard
		fmt.Fprintf(w, "detail user #%s", id)
	})

	// =====================================================================
	// 3. Batasi ukuran body dengan http.MaxBytesReader (cegah abuse memori)
	// =====================================================================
	// 🔍 Analogi: MaxBytesReader itu seperti TIMBANGAN BAGASI di check-in —
	// koper (body) di atas batas langsung ditolak di konter, bukan dibiarkan
	// masuk sampai pesawat (memori server) kelebihan muatan.
	mux.HandleFunc("POST /upload", func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 16) // maksimal 16 byte
		b, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "body terlalu besar", http.StatusRequestEntityTooLarge)
			return
		}
		fmt.Fprintf(w, "diterima %d byte", len(b))
	})

	// Bungkus seluruh mux dengan middleware logging.
	handler := Chain(mux, logging)

	// =====================================================================
	// httptest.Server menjalankan handler di alamat lokal sementara.
	// =====================================================================
	// 🔍 Analogi: httptest.Server itu seperti PANGGUNG GLADI RESIK — set
	// lengkap dan lampu menyala (server sungguhan di port acak), tapi hanya
	// untuk latihan; selesai gladi, panggung dibongkar (srv.Close).
	srv := httptest.NewServer(handler)
	defer srv.Close()

	fmt.Println("== 2. routing method + wildcard ==")
	fmt.Println(" ", get(srv.URL+"/users/42"))

	fmt.Println("== 3. MaxBytesReader ==")
	fmt.Println("  body kecil :", post(srv.URL+"/upload", "halo"))
	fmt.Println("  body besar :", post(srv.URL+"/upload", strings.Repeat("x", 100)))
}

// get & post: helper kecil untuk memanggil server uji dan mengembalikan ringkasan.
func get(url string) string {
	resp, err := http.Get(url)
	if err != nil {
		return "err: " + err.Error()
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	return fmt.Sprintf("%d %s", resp.StatusCode, b)
}

func post(url, body string) string {
	resp, err := http.Post(url, "text/plain", strings.NewReader(body))
	if err != nil {
		return "err: " + err.Error()
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	return fmt.Sprintf("%d %s", resp.StatusCode, strings.TrimSpace(string(b)))
}
