// REST API CRUD memakai net/http MURNI (tanpa framework) — modul 12.
//
// Jalankan server:
//
//	go run ./12-http-stdlib      # dengar di :8080
//
// Coba dengan curl:
//
//	curl localhost:8080/books
//	curl -X POST localhost:8080/books -d '{"title":"Go","author":"RP"}'
//	curl localhost:8080/books/1
//	curl -X PUT localhost:8080/books/1 -d '{"title":"Go v2","author":"RP"}'
//	curl -X DELETE localhost:8080/books/1
//
// Verifikasi otomatis (tanpa server): go test ./12-http-stdlib
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

// 🔍 Analogi besar: server HTTP itu seperti RESTORAN. Pelanggan (browser/curl) mengirim
// PESANAN (request) via jalur tertentu; dapur (handler) memasak & mengirim balik PIRING
// (response). "REST/CRUD" = 4 aksi dasar: Create (POST=tambah), Read (GET=lihat),
// Update (PUT=ubah), Delete (DELETE=hapus) — persis buku menu operasi terhadap data.

// 🔍 Analogi: menaruh 'store' di dalam struct server itu DEPENDENCY INJECTION — seperti
// memberi koki bahan & alat lewat pintu, bukan menyuruhnya mengambil sendiri dari gudang global.
// Efeknya: saat menguji, kita bisa memberi "store palsu" (lihat modul 08). Menghindari variabel global.
// server memegang dependency (store). Handler jadi method-nya.
type server struct {
	store *BookStore
}

// 🔍 Analogi: mux (router) itu RESEPSIONIS yang membaca "mau ke mana" (GET /books) lalu
// mengarahkan ke handler yang tepat. Sejak Go 1.22, resepsionis paham METHOD+PATH langsung
// (mis. "POST /books" beda dari "GET /books") — dulu harus if-else manual, kini bawaan.
// routes mendaftarkan endpoint memakai routing method+path bawaan Go 1.22.
func (s *server) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /books", s.handleList)
	mux.HandleFunc("GET /books/search", s.handleSearch) // latihan 1 (lebih spesifik dari {id})
	mux.HandleFunc("POST /books", s.handleCreate)
	mux.HandleFunc("GET /books/{id}", s.handleGet)
	mux.HandleFunc("PUT /books/{id}", s.handleUpdate)
	mux.HandleFunc("DELETE /books/{id}", s.handleDelete)
	// Rantai middleware: recover (latihan 2) di luar, lalu logging.
	return recoverMW(logging(mux))
}

func main() {
	srv := &server{store: seed(NewBookStore())}
	log.Println("server berjalan di http://localhost:8080")
	if err := http.ListenAndServe(":8080", srv.routes()); err != nil {
		log.Fatal(err)
	}
}

// ---------- Handlers ----------

func (s *server) handleList(w http.ResponseWriter, r *http.Request) {
	// Latihan 4: pagination via ?limit=&offset=.
	books := s.store.List()
	offset := atoiDefault(r.URL.Query().Get("offset"), 0)
	limit := atoiDefault(r.URL.Query().Get("limit"), len(books))

	if offset > len(books) {
		offset = len(books)
	}
	end := offset + limit
	if end > len(books) {
		end = len(books)
	}
	writeJSON(w, http.StatusOK, books[offset:end])
}

func (s *server) handleCreate(w http.ResponseWriter, r *http.Request) {
	var b Book
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		writeError(w, http.StatusBadRequest, "JSON tidak valid")
		return
	}
	if b.Title == "" {
		writeError(w, http.StatusBadRequest, "title wajib diisi")
		return
	}
	writeJSON(w, http.StatusCreated, s.store.Create(b))
}

func (s *server) handleGet(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	b, err := s.store.Get(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, b)
}

func (s *server) handleUpdate(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	var b Book
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		writeError(w, http.StatusBadRequest, "JSON tidak valid")
		return
	}
	updated, err := s.store.Update(id, b)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (s *server) handleDelete(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	if err := s.store.Delete(id); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ---------- Helper ----------

// parseID mengambil {id} dari path dan memvalidasinya sebagai angka.
func parseID(w http.ResponseWriter, r *http.Request) (int, bool) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "id harus angka")
		return 0, false
	}
	return id, true
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// 🔍 Analogi: middleware itu SATPAM BERLAPIS di pintu masuk. Tiap request harus melewati
// mereka dulu sebelum sampai ke handler. logging = satpam yang mencatat tamu di buku tamu;
// recover = satpam yang menangkap kalau handler "pingsan" (panic) agar server tak ikut roboh.
// Polanya "bungkus": recoverMW(logging(mux)) — request masuk dari luar ke dalam, respons balik keluar.
// logging: middleware sederhana (mencatat method + path tiap request).
func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

// seed mengisi beberapa data awal (dipakai contoh sekaligus test).
func seed(s *BookStore) *BookStore {
	s.Create(Book{Title: "The Go Programming Language", Author: "Donovan & Kernighan"})
	s.Create(Book{Title: "Clean Architecture", Author: "Robert C. Martin"})
	return s
}
