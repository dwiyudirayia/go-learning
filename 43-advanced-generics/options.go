package main

import "time"

// 🔍 Analogi besar functional options: seperti PESAN KOPI. Default-nya sudah enak (localhost:8080).
// Kamu cukup menyebut yang mau diubah: "kopi, less sugar, extra shot" -> NewServer(WithPort(9090),
// WithTLS()). Tak perlu menyebut SEMUA parameter berurutan (WithHost tak ditulis? pakai default).
// Ini mengalahkan dua pola buruk: (a) konstruktor dgn 10 argumen posisional yang membingungkan
// (NewServer("localhost",8080,30,false,...) — argumen mana yang mana?), dan (b) struct config raksasa.
// Tiap "WithXxx" adalah fungsi kecil yang mengubah satu setelan; NewServer menerapkannya berurutan.

// FUNCTIONAL OPTIONS: pola idiomatik Go untuk konstruktor dengan banyak
// parameter opsional (menggantikan konstruktor bertumpuk atau struct config
// besar). Pemanggil hanya menyebut yang ingin diubah; sisanya default.

type Server struct {
	Host    string
	Port    int
	Timeout time.Duration
	TLS     bool
}

// Option = fungsi yang memodifikasi Server.
type Option func(*Server)

func WithHost(h string) Option           { return func(s *Server) { s.Host = h } }
func WithPort(p int) Option              { return func(s *Server) { s.Port = p } }
func WithTimeout(d time.Duration) Option { return func(s *Server) { s.Timeout = d } }
func WithTLS() Option                    { return func(s *Server) { s.TLS = true } }

// NewServer memulai dari DEFAULT lalu menerapkan tiap opsi.
func NewServer(opts ...Option) *Server {
	s := &Server{ // default masuk akal
		Host:    "localhost",
		Port:    8080,
		Timeout: 30 * time.Second,
		TLS:     false,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}
