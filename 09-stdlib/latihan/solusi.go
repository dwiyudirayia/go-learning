// Solusi latihan Modul 09 — Standard Library.
// Jalankan: go run ./09-stdlib/latihan
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"sort"
	"strings"
	"time"
)

func main() {
	fmt.Println("=== Solusi Latihan Modul 09 ===")
	latihan1()
	latihan2()
	latihan3()
	latihan4()
	latihan5()
}

// ------------------------------------------------------------------
// Latihan 1: Book + JSON marshal/unmarshal
// ------------------------------------------------------------------
type Book struct {
	Title  string   `json:"title"`
	Author string   `json:"author"`
	Year   int      `json:"year"`
	Tags   []string `json:"tags,omitempty"`
}

func latihan1() {
	fmt.Println("\n-- Latihan 1: JSON Book --")
	b := Book{Title: "The Go Programming Language", Author: "Donovan & Kernighan", Year: 2015, Tags: []string{"go", "programming"}}

	data, _ := json.Marshal(b)
	fmt.Printf("marshal   : %s\n", data)

	var back Book
	_ = json.Unmarshal(data, &back)
	fmt.Printf("unmarshal : %+v\n", back)
}

// ------------------------------------------------------------------
// Latihan 2: parse tanggal + tambah 90 hari + format
// ------------------------------------------------------------------
func latihan2() {
	fmt.Println("\n-- Latihan 2: time --")
	t, err := time.Parse("2006-01-02", "2026-07-01")
	if err != nil {
		fmt.Println("parse gagal:", err)
		return
	}
	plus90 := t.AddDate(0, 0, 90) // tambah 90 hari (AddDate lebih benar dari Add utk hari/bulan)
	fmt.Printf("%s + 90 hari = %s\n", t.Format("02 Jan 2006"), plus90.Format("02 Jan 2006"))
}

// ------------------------------------------------------------------
// Latihan 3: wordCount dari io.Reader apa pun
// ------------------------------------------------------------------
func wordCount(r io.Reader) (map[string]int, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	counts := map[string]int{}
	for _, w := range strings.Fields(strings.ToLower(string(data))) {
		counts[w]++
	}
	return counts, nil
}

func latihan3() {
	fmt.Println("\n-- Latihan 3: wordCount(io.Reader) --")
	// Sumbernya strings.Reader, tapi fungsi jalan untuk Reader apa pun (file, http body, dll).
	counts, _ := wordCount(strings.NewReader("go go go rust go rust"))
	// Urutkan untuk output stabil.
	keys := make([]string, 0, len(counts))
	for k := range counts {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Printf("  %-6s : %d\n", k, counts[k])
	}
}

// ------------------------------------------------------------------
// Latihan 4: tulis & baca file JSON sementara berisi []Book
// ------------------------------------------------------------------
func latihan4() {
	fmt.Println("\n-- Latihan 4: file JSON []Book --")
	books := []Book{
		{Title: "Buku A", Author: "X", Year: 2020},
		{Title: "Buku B", Author: "Y", Year: 2023, Tags: []string{"baru"}},
	}

	f, err := os.CreateTemp("", "books-*.json")
	if err != nil {
		fmt.Println("gagal buat file:", err)
		return
	}
	defer os.Remove(f.Name())
	_ = f.Close()

	// Tulis (pretty) ke file.
	data, _ := json.MarshalIndent(books, "", "  ")
	_ = os.WriteFile(f.Name(), data, 0o644)

	// Baca kembali & decode.
	raw, _ := os.ReadFile(f.Name())
	var loaded []Book
	_ = json.Unmarshal(raw, &loaded)
	fmt.Printf("disimpan & dibaca kembali %d buku:\n", len(loaded))
	for _, b := range loaded {
		fmt.Printf("  - %q (%d) tags=%v\n", b.Title, b.Year, b.Tags)
	}
}

// ------------------------------------------------------------------
// Latihan 5: handler HTTP -> JSON, diuji dengan httptest
// ------------------------------------------------------------------
func booksHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode([]Book{
		{Title: "REST in Go", Author: "Gopher", Year: 2026},
	})
}

func latihan5() {
	fmt.Println("\n-- Latihan 5: handler + httptest --")
	mux := http.NewServeMux()
	mux.HandleFunc("GET /books", booksHandler)

	srv := httptest.NewServer(mux)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/books")
	if err != nil {
		fmt.Println("request gagal:", err)
		return
	}
	defer resp.Body.Close()

	var books []Book
	_ = json.NewDecoder(resp.Body).Decode(&books)
	fmt.Printf("status %s, dapat %d buku: %+v\n", resp.Status, len(books), books)
}
