package main

import "time"

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
