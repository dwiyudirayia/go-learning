// Teknik Advanced Modul 10 (project layout) — contoh runnable + penjelasan detail.
//
// Jalankan: go run ./10-project-layout/advanced
package main

import (
	_ "embed"
	"fmt"
	"runtime/debug"
	"strings"
)

// ===========================================================================
// 1. //go:embed — menyematkan file ke dalam binary
// ===========================================================================
// Directive di bawah menaruh isi version.txt ke variabel string saat COMPILE.
// Hasilnya: binary tunggal tanpa perlu menyertakan file terpisah saat deploy
// (dipakai untuk template, migrasi SQL, aset web, dsb).
//
//go:embed version.txt
var versi string

// ===========================================================================
// 2. Functional Options — konfigurasi opsional yang rapi & ekstensibel
// ===========================================================================
// Alih-alih konstruktor dengan banyak parameter (atau banyak struct config),
// kita terima daftar "opsi" berupa fungsi. Default masuk akal, pemanggil hanya
// menimpa yang perlu. Menambah opsi baru tak memecah kode pemanggil lama.

type Server struct {
	host    string
	port    int
	tls     bool
	timeout int
}

type Option func(*Server)

func WithHost(h string) Option { return func(s *Server) { s.host = h } }
func WithPort(p int) Option    { return func(s *Server) { s.port = p } }
func WithTLS() Option          { return func(s *Server) { s.tls = true } }
func WithTimeout(t int) Option { return func(s *Server) { s.timeout = t } }

func NewServer(opts ...Option) *Server {
	// Default yang aman:
	s := &Server{host: "localhost", port: 8080, tls: false, timeout: 30}
	for _, opt := range opts {
		opt(s) // tiap opsi menimpa field tertentu
	}
	return s
}

func (s *Server) String() string {
	skema := "http"
	if s.tls {
		skema = "https"
	}
	return fmt.Sprintf("%s://%s:%d (timeout %ds)", skema, s.host, s.port, s.timeout)
}

func main() {
	fmt.Println("== 1. //go:embed ==")
	fmt.Println("  versi (dari version.txt):", strings.TrimSpace(versi))

	fmt.Println("== 2. functional options ==")
	fmt.Println("  default          :", NewServer())
	fmt.Println("  hanya ubah port  :", NewServer(WithPort(9090)))
	fmt.Println("  gabungan opsi    :", NewServer(WithHost("api.internal"), WithPort(443), WithTLS(), WithTimeout(10)))

	// =====================================================================
	// 3. Build info via runtime/debug — metadata build tanpa variabel global
	// =====================================================================
	// ReadBuildInfo() membaca info yang ditanam toolchain: versi Go, path modul,
	// dan (bila dibangun dari git) revisi/vcs. Berguna untuk endpoint /version.
	fmt.Println("== 3. build info ==")
	if bi, ok := debug.ReadBuildInfo(); ok {
		fmt.Println("  go version :", bi.GoVersion)
		fmt.Println("  module path:", bi.Main.Path)
		for _, s := range bi.Settings {
			if s.Key == "GOARCH" || s.Key == "GOOS" {
				fmt.Printf("  %-11s: %s\n", s.Key, s.Value)
			}
		}
	}
	// Catatan build tags (tak bisa didemokan saat runtime): file dengan header
	//   //go:build linux
	// hanya ikut dikompilasi di OS/arch/fitur tertentu — dasar kode lintas
	// platform & stub WASM (lihat modul 47).
}
