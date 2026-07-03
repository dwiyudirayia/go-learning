// CAPSTONE — URL Shortener siap-produksi (mini).
//
// Jalankan (tanpa infra apa pun — SQLite temp + Redis in-memory):
//
//	go run ./41-capstone
//
// Coba:
//
//	curl -X POST localhost:8080/auth/register -d '{"email":"a@b.c","password":"secret123"}'
//	TOKEN=$(curl -s -X POST localhost:8080/auth/login -d '{"email":"a@b.c","password":"secret123"}' | jq -r .token)
//	curl -X POST localhost:8080/api/shorten -H "Authorization: Bearer $TOKEN" -d '{"url":"https://go.dev"}'
//	curl -i localhost:8080/<code>     # -> 302 redirect
//
// Env: PORT, DB_PATH, REDIS_ADDR, JWT_SECRET.
// Verifikasi otomatis: go test ./41-capstone
package main

import (
	"context"
	"database/sql"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"

	_ "modernc.org/sqlite"
)

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Config (Modul 19).
	port := getenv("PORT", "8080")
	dbPath := getenv("DB_PATH", "capstone.db")

	// Database (Modul 14).
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	if err := setupSchema(db); err != nil {
		log.Fatal(err)
	}

	// Redis: pakai server sungguhan bila REDIS_ADDR diset, else in-memory (Modul 22).
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		mr, _ := miniredis.Run()
		defer mr.Close()
		redisAddr = mr.Addr()
		logger.Info("Redis in-memory (set REDIS_ADDR untuk Redis sungguhan)")
	}
	rdb := redis.NewClient(&redis.Options{Addr: redisAddr})
	defer rdb.Close()

	// Wiring berlapis: store -> cache -> service -> handler (Modul 29).
	svc := NewService(NewStore(db), NewCache(rdb))
	app := buildApp(svc, logger)

	// Graceful shutdown (Modul 20).
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		logger.Info("server jalan", slog.String("addr", ":"+port))
		if err := app.Listen(":" + port); err != nil {
			logger.Error("listen", slog.Any("err", err))
		}
	}()

	<-ctx.Done()
	logger.Info("shutdown: menyelesaikan request in-flight...")
	_ = app.ShutdownWithContext(context.Background())
	logger.Info("selesai.")
}
