package main

import (
	"context"
	"log/slog"
	"math"
	"math/rand"
	"time"
)

// RetryConfig mengatur perilaku retry & backoff.
type RetryConfig struct {
	MaxAttempts int           // 0 = tak terbatas (sampai ctx dibatalkan)
	BaseDelay   time.Duration // jeda awal
	MaxDelay    time.Duration // batas atas jeda
}

func DefaultRetry() RetryConfig {
	return RetryConfig{MaxAttempts: 0, BaseDelay: 100 * time.Millisecond, MaxDelay: 10 * time.Second}
}

// backoff menghitung jeda EKSPONENSIAL + JITTER, dibatasi MaxDelay.
// Jitter (acak) mencegah "thundering herd" — semua client reconnect serempak.
func backoff(cfg RetryConfig, attempt int) time.Duration {
	d := float64(cfg.BaseDelay) * math.Pow(2, float64(attempt))
	if d > float64(cfg.MaxDelay) {
		d = float64(cfg.MaxDelay)
	}
	jitter := d * 0.2 * rand.Float64() // + 0..20%
	return time.Duration(d + jitter)
}

// Retry menjalankan fn sampai sukses, MaxAttempts habis, atau ctx dibatalkan.
// Dipakai untuk PUBLISH yang harus tahan gangguan sesaat.
func Retry(ctx context.Context, cfg RetryConfig, fn func() error) error {
	var lastErr error
	for attempt := 0; cfg.MaxAttempts == 0 || attempt < cfg.MaxAttempts; attempt++ {
		if err := fn(); err == nil {
			return nil
		} else {
			lastErr = err
		}
		select {
		case <-time.After(backoff(cfg, attempt)):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return lastErr
}

// superviseConsumer = SKEMA RECONNECT untuk KONSUMEN.
//
// Ia terus memanggil connectAndServe (yang menyambung + memproses pesan sampai
// koneksi putus). Bila putus (connectAndServe mengembalikan error), ia menunggu
// backoff lalu menyambung ulang — sampai ctx dibatalkan. Inilah cara konsumen
// "bangkit sendiri" setelah broker sempat mati/restart.
func superviseConsumer(ctx context.Context, name string, cfg RetryConfig, logger *slog.Logger, connectAndServe func(context.Context) error) {
	for attempt := 0; ctx.Err() == nil; attempt++ {
		err := connectAndServe(ctx) // blok sampai putus atau ctx selesai
		if ctx.Err() != nil {
			return // shutdown normal
		}
		wait := backoff(cfg, attempt)
		if logger != nil {
			logger.Warn("konsumen terputus, menyambung ulang",
				slog.String("name", name), slog.Any("err", err), slog.Duration("retry_dalam", wait))
		}
		select {
		case <-time.After(wait):
		case <-ctx.Done():
			return
		}
	}
}

// PublishWithRetry membungkus Broker.Publish dengan retry+backoff -> gangguan
// sesaat (broker restart) tak langsung menggagalkan pengiriman.
func PublishWithRetry(ctx context.Context, b Broker, topic string, data []byte, logger *slog.Logger) error {
	cfg := RetryConfig{MaxAttempts: 12, BaseDelay: 50 * time.Millisecond, MaxDelay: 2 * time.Second}
	return Retry(ctx, cfg, func() error {
		err := b.Publish(ctx, topic, data)
		if err != nil && logger != nil {
			logger.Warn("publish gagal, akan retry", slog.Any("err", err))
		}
		return err
	})
}
