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

// 🔍 Analogi besar: graceful shutdown itu seperti RESTORAN yang mau tutup dengan sopan.
// Saat tanda "tutup" (sinyal SIGTERM dari Docker/Kubernetes) datang, restoran TIDAK langsung
// mengusir semua orang. Ia: (1) mengunci pintu depan (stop terima tamu baru), (2) tetap
// melayani tamu yang MASIH makan sampai selesai, (3) baru mematikan lampu. Tanpa ini, request
// yang sedang jalan terputus mendadak = data setengah jadi, pelanggan error. SIGTERM = "tolong
// tutup baik-baik"; kalau kelamaan (lewat 10 detik), baru dipaksa tutup (srv.Close).

// runWithGracefulShutdown menjalankan server sampai ctx dibatalkan (biasanya oleh
// sinyal SIGINT/SIGTERM), lalu:
//  1. berhenti menerima koneksi BARU,
//  2. MENUNGGU request yang sedang berjalan selesai (sampai batas waktu),
//  3. baru benar-benar berhenti.
func runWithGracefulShutdown(ctx context.Context, srv *http.Server, ln net.Listener, logger *slog.Logger) error {
	serveErr := make(chan error, 1)
	go func() {
		// 🔍 Analogi: ErrServerClosed itu BUKAN error sungguhan — ia cuma "pemberitahuan resmi"
		// bahwa server berhenti karena KITA yang menyuruh (Shutdown), bukan karena rusak. Jadi
		// kita perlakukan sebagai sukses (nil). Membedakan "berhenti sengaja" vs "berhenti gagal".
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
