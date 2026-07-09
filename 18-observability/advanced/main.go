// Teknik Advanced Modul 18 (observability) — contoh runnable + penjelasan detail.
//
// Fokus: structured logging dengan log/slog (stdlib). Jalankan:
//
//	go run ./18-observability/advanced
package main

import (
	"log/slog"
	"os"
)

func main() {
	// =====================================================================
	// 1. Handler JSON + ReplaceAttr (redaksi & normalisasi field)
	// =====================================================================
	// slog memisahkan "apa yang dicatat" (logger) dari "bagaimana formatnya"
	// (handler). JSON cocok untuk agregator (Loki/ELK). ReplaceAttr dipanggil
	// untuk SETIAP atribut -> tempat menyensor rahasia & merapikan field.
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo, // Debug akan disaring (tak tampil)
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey { // buang timestamp -> output deterministik utk demo
				return slog.Attr{}
			}
			if a.Key == "password" || a.Key == "token" { // JANGAN pernah log rahasia
				return slog.String(a.Key, "***redacted***")
			}
			return a
		},
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, opts))

	// =====================================================================
	// 2. Base attributes dengan With() — konteks yang menempel di tiap log
	// =====================================================================
	// Semua log dari logger turunan ini otomatis membawa service & version.
	log := logger.With("service", "api", "version", "1.4.2")

	// =====================================================================
	// 3. Level & structured key-value (bukan string interpolation)
	// =====================================================================
	log.Debug("query detail", "sql", "SELECT 1")                 // TIDAK tampil (di bawah Info)
	log.Info("user login", "user_id", 42, "password", "rahasia") // password tersensor
	log.Warn("latensi tinggi", "ms", 850)

	// =====================================================================
	// 4. Grouping — mengelompokkan atribut terkait
	// =====================================================================
	// slog.Group menyarangkan field -> {"http":{"method":"GET","status":200}}.
	log.Info("request selesai",
		slog.String("path", "/users/42"),
		slog.Group("http",
			slog.String("method", "GET"),
			slog.Int("status", 200),
		),
	)

	// Catatan metrik: di produksi lengkapi dgn Prometheus (histogram latensi
	// dgn bucket sesuai SLO, hindari label berkardinalitas tinggi spt user_id),
	// dan korelasikan log<->trace via trace_id (lihat modul 33).
}
