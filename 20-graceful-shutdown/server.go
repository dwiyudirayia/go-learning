// Modul 20 — Graceful Shutdown: berhenti dengan rapi tanpa memutus request
// yang sedang berjalan.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"time"
)

// newServer membuat HTTP server contoh dengan sebuah handler "lambat".
func newServer(addr string, logger *slog.Logger) *http.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok\n"))
	})

	// /slow mensimulasikan pekerjaan yang butuh waktu (mis. query DB).
	// Saat shutdown, request yang SUDAH masuk ke sini tetap diselesaikan.
	mux.HandleFunc("GET /slow", func(w http.ResponseWriter, r *http.Request) {
		logger.Info("mulai proses lambat")
		time.Sleep(300 * time.Millisecond)
		w.Write([]byte("selesai diproses\n"))
		logger.Info("selesai proses lambat")
	})

	return &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second, // hindari Slowloris
	}
}

// runWithGracefulShutdown menjalankan server sampai ctx dibatalkan (biasanya oleh
// sinyal SIGINT/SIGTERM), lalu:
//  1. berhenti menerima koneksi BARU,
//  2. MENUNGGU request yang sedang berjalan selesai (sampai batas waktu),
//  3. baru benar-benar berhenti.
func runWithGracefulShutdown(ctx context.Context, srv *http.Server, ln net.Listener, logger *slog.Logger) error {
	serveErr := make(chan error, 1)
	go func() {
		// Serve mengembalikan ErrServerClosed saat Shutdown dipanggil — itu normal.
		err := srv.Serve(ln)
		if errors.Is(err, http.ErrServerClosed) {
			serveErr <- nil
		} else {
			serveErr <- err
		}
	}()

	select {
	case err := <-serveErr:
		return err // server gagal start
	case <-ctx.Done():
		logger.Info("sinyal shutdown diterima, mulai graceful shutdown")
	}

	// Beri waktu maksimal 10 detik untuk menyelesaikan request in-flight.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		// Bila melewati batas waktu, paksa tutup.
		_ = srv.Close()
		return err
	}
	logger.Info("server berhenti dengan rapi")
	return <-serveErr
}
