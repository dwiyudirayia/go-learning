package main

import (
	"log"
	"net/http"
	"strconv"
	"strings"
)

// Latihan 1: GET /books/search?author=X — filter berdasarkan query string.
func (s *server) handleSearch(w http.ResponseWriter, r *http.Request) {
	author := strings.ToLower(r.URL.Query().Get("author"))
	var hasil []Book
	for _, b := range s.store.List() {
		if author == "" || strings.Contains(strings.ToLower(b.Author), author) {
			hasil = append(hasil, b)
		}
	}
	writeJSON(w, http.StatusOK, hasil)
}

// Latihan 2: middleware recover — panic di handler diubah jadi 500,
// server tetap hidup (lihat Modul 5).
func recoverMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("panic dipulihkan: %v", rec)
				writeError(w, http.StatusInternalServerError, "kesalahan server")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// atoiDefault mengubah string ke int; kembalikan def bila kosong/invalid.
func atoiDefault(s string, def int) int {
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < 0 {
		return def
	}
	return n
}
