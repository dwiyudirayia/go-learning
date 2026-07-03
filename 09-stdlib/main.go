// Package main untuk modul 09 — Standard Library.
// Jalankan: go run ./09-stdlib
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"time"
)

func main() {
	fmt.Println("=== 09 — Standard Library ===")
	contohJSON()
	contohTime()
	contohIO()
	contohOS()
	contohHTTP()
}

// ------------------------------------------------------------------
// 1. encoding/json
// ------------------------------------------------------------------
type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email,omitempty"` // hilang bila kosong
	Admin bool   `json:"is_admin"`
	pass  string // unexported -> tidak ikut JSON
}

func contohJSON() {
	fmt.Println("\n-- encoding/json --")

	u := User{ID: 1, Name: "Ana", Admin: true, pass: "rahasia"}
	b, _ := json.Marshal(u)
	fmt.Printf("Marshal          : %s\n", b) // Email hilang (omitempty), pass tak ikut

	pretty, _ := json.MarshalIndent(u, "", "  ")
	fmt.Printf("MarshalIndent    :\n%s\n", pretty)

	// Unmarshal JSON -> struct
	var u2 User
	_ = json.Unmarshal([]byte(`{"id":2,"name":"Budi","email":"budi@mail.id"}`), &u2)
	fmt.Printf("Unmarshal        : %+v\n", u2)

	// JSON dinamis / tak dikenal strukturnya -> map[string]any
	var dyn map[string]any
	_ = json.Unmarshal([]byte(`{"kota":"Bandung","populasi":2500000}`), &dyn)
	// Angka JSON masuk sebagai float64 saat target 'any'.
	fmt.Printf("dinamis (any)    : kota=%v populasi=%.0f\n", dyn["kota"], dyn["populasi"])
}

// ------------------------------------------------------------------
// 2. time (pakai waktu tetap agar output deterministik)
// ------------------------------------------------------------------
func contohTime() {
	fmt.Println("\n-- time --")

	t := time.Date(2026, time.July, 1, 15, 4, 5, 0, time.UTC)
	fmt.Printf("Format            : %s\n", t.Format("2006-01-02 15:04:05"))
	fmt.Printf("Format lain       : %s\n", t.Format("Mon, 02 Jan 2006"))

	// Parse string -> time
	parsed, _ := time.Parse("2006-01-02", "2026-12-25")
	fmt.Printf("Parse             : %s (bulan: %s)\n", parsed.Format("02/01/2006"), parsed.Month())

	// Duration & aritmetika waktu
	besok := t.Add(24 * time.Hour)
	fmt.Printf("Add 24 jam        : %s\n", besok.Format("2006-01-02"))
	selisih := besok.Sub(t)
	fmt.Printf("Sub (Duration)    : %s = %.0f jam\n", selisih, selisih.Hours())
	fmt.Printf("Perbandingan      : t sebelum besok? %t\n", t.Before(besok))
}

// ------------------------------------------------------------------
// 3. io: Reader/Writer, Copy, ReadAll
// ------------------------------------------------------------------
func contohIO() {
	fmt.Println("\n-- io --")

	// strings.Reader (sumber) -> bytes.Buffer (tujuan) via io.Copy.
	src := strings.NewReader("data mengalir lewat io.Reader")
	var dst bytes.Buffer
	n, _ := io.Copy(&dst, src)
	fmt.Printf("io.Copy menyalin %d byte: %q\n", n, dst.String())

	// io.ReadAll membaca habis Reader apa pun.
	data, _ := io.ReadAll(strings.NewReader("baca semua"))
	fmt.Printf("io.ReadAll       : %q\n", data)

	// bytes.Buffer sebagai Writer untuk fmt.Fprintf.
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "halo %s, %d tahun", "Ciko", 3)
	fmt.Printf("bytes.Buffer     : %q\n", buf.String())
}

// ------------------------------------------------------------------
// 4. os: env & file (temp)
// ------------------------------------------------------------------
func contohOS() {
	fmt.Println("\n-- os --")

	// Environment variable
	_ = os.Setenv("APP_ENV", "production")
	fmt.Printf("Getenv APP_ENV   : %q\n", os.Getenv("APP_ENV"))

	// Tulis lalu baca file sementara, kemudian hapus.
	f, err := os.CreateTemp("", "modul9-*.txt")
	if err != nil {
		fmt.Println("gagal buat temp:", err)
		return
	}
	defer os.Remove(f.Name()) // bersihkan
	_ = f.Close()

	_ = os.WriteFile(f.Name(), []byte("isi file sementara\n"), 0o644)
	content, _ := os.ReadFile(f.Name())
	fmt.Printf("tulis+baca file  : %q (%s)\n", strings.TrimSpace(string(content)), f.Name())
}

// ------------------------------------------------------------------
// 5. net/http: server (via httptest) + client
// ------------------------------------------------------------------
func contohHTTP() {
	fmt.Println("\n-- net/http --")

	// Server: routing method+path (Go 1.22+) mengembalikan JSON.
	mux := http.NewServeMux()
	mux.HandleFunc("GET /users/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id") // path parameter bawaan
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(User{ID: 99, Name: "User-" + id, Admin: false})
	})

	// httptest menjalankan server sungguhan di port acak, lalu kita panggil.
	srv := httptest.NewServer(mux)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/users/7")
	if err != nil {
		fmt.Println("request gagal:", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("status           : %s\n", resp.Status)

	// Decode JSON response langsung dari body (io.Reader).
	var got User
	_ = json.NewDecoder(resp.Body).Decode(&got)
	fmt.Printf("body (decoded)   : %+v\n", got)
}
