// REAL-CASE Modul 41 (capstone) — URL shortener: POSTGRESQL (store) + REDIS (cache-aside).
//
// Gabungan stack produksi: sumber kebenaran di Postgres, dibaca cepat lewat
// cache Redis. Ini pola yang sama dgn capstone advanced/ (in-memory) — hanya
// store & cache-nya diganti infra nyata.
//
// Auto-skip bila POSTGRES_DSN atau REDIS_ADDR kosong. Jalankan nyata:
//
//	docker compose -f 41-capstone/real-case/docker-compose.yml up -d
//	POSTGRES_DSN='postgres://postgres:secret@127.0.0.1:5432/appdb' \
//	REDIS_ADDR=127.0.0.1:6379 go run ./41-capstone/real-case
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type App struct {
	pool  *pgxpool.Pool
	cache *redis.Client
}

// Shorten menyimpan URL ke Postgres (sumber kebenaran) dan mengembalikan kode.
func (a *App) Shorten(ctx context.Context, url string) (string, error) {
	var id int64
	if err := a.pool.QueryRow(ctx,
		`INSERT INTO links(url) VALUES($1) RETURNING id`, url).Scan(&id); err != nil {
		return "", err
	}
	return fmt.Sprintf("c%d", id), nil
}

// Resolve memakai CACHE-ASIDE: cek Redis -> miss -> Postgres -> isi Redis.
func (a *App) Resolve(ctx context.Context, code string) (string, bool, error) {
	if url, err := a.cache.Get(ctx, "link:"+code).Result(); err == nil {
		return url, true, nil // cache HIT
	} else if !errors.Is(err, redis.Nil) {
		return "", false, err
	}
	var url string
	err := a.pool.QueryRow(ctx, `SELECT url FROM links WHERE id=$1`, kodeKeID(code)).Scan(&url)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	_ = a.cache.Set(ctx, "link:"+code, url, 10*time.Minute).Err() // isi cache
	return url, false, nil
}

func kodeKeID(code string) int64 { // "c123" -> 123
	var id int64
	fmt.Sscanf(code, "c%d", &id)
	return id
}

func main() {
	dsn := os.Getenv("POSTGRES_DSN")
	addr := os.Getenv("REDIS_ADDR")
	if dsn == "" || addr == "" {
		fmt.Println("⏭️  DILEWATI: set POSTGRES_DSN & REDIS_ADDR untuk versi nyata.")
		fmt.Println("   docker compose -f 41-capstone/real-case/docker-compose.yml up -d")
		fmt.Println("   POSTGRES_DSN='postgres://postgres:secret@127.0.0.1:5432/appdb' \\")
		fmt.Println("   REDIS_ADDR=127.0.0.1:6379 go run ./41-capstone/real-case")
		return
	}
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		panic(err)
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		panic("gagal konek Postgres: " + err.Error())
	}
	if _, err := pool.Exec(ctx, `CREATE TABLE IF NOT EXISTS links(id BIGSERIAL PRIMARY KEY, url TEXT NOT NULL)`); err != nil {
		panic(err)
	}

	cache := redis.NewClient(&redis.Options{Addr: addr})
	defer cache.Close()
	if err := cache.Ping(ctx).Err(); err != nil {
		panic("gagal konek Redis: " + err.Error())
	}

	app := &App{pool: pool, cache: cache}

	code, err := app.Shorten(ctx, "https://go.dev")
	if err != nil {
		panic(err)
	}
	fmt.Println("== URL shortener (Postgres + Redis) ==")
	fmt.Println("  shorten https://go.dev ->", code)

	url1, hit1, _ := app.Resolve(ctx, code) // miss -> Postgres -> isi cache
	url2, hit2, _ := app.Resolve(ctx, code) // hit  -> Redis
	fmt.Printf("  resolve #1: %s (cache hit=%v)\n", url1, hit1)
	fmt.Printf("  resolve #2: %s (cache hit=%v)\n", url2, hit2)
	_, ada, _ := app.Resolve(ctx, "c999999")
	fmt.Println("  resolve kode tak ada -> ketemu?", ada)
}
