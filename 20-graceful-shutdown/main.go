// Jalankan: go run ./20-graceful-shutdown
//
// Lalu di terminal lain:
//
//	curl localhost:8080/slow &   # request lambat (300ms)
//	# tekan Ctrl+C di server -> request /slow tetap diselesaikan sebelum berhenti
//
// Verifikasi otomatis: go test ./20-graceful-shutdown
package main

import (
	"context"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// signal.NotifyContext membatalkan ctx saat menerima SIGINT (Ctrl+C) atau
	// SIGTERM (dikirim Docker/Kubernetes saat menghentikan container).
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	srv := newServer(":8080", logger)

	ln, err := net.Listen("tcp", srv.Addr)
	if err != nil {
		logger.Error("gagal listen", slog.Any("err", err))
		os.Exit(1)
	}

	logger.Info("server berjalan", slog.String("addr", srv.Addr))
	if err := runWithGracefulShutdown(ctx, srv, ln, logger); err != nil {
		logger.Error("server error", slog.Any("err", err))
		os.Exit(1)
	}
}
