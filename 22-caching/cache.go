// Modul 22 — Caching dengan Redis: pola cache-aside, TTL, invalidation.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/redis/go-redis/v9"
)

type Product struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Price int    `json:"price"`
}

// ProductService memadukan "database" lambat + cache Redis.
type ProductService struct {
	rdb      *redis.Client
	ttl      time.Duration
	dbHits   int64 // penghitung berapa kali menyentuh DB (untuk membuktikan cache bekerja)
	products map[int]Product
}

func NewProductService(rdb *redis.Client) *ProductService {
	return &ProductService{
		rdb: rdb,
		ttl: 30 * time.Second,
		products: map[int]Product{
			1: {ID: 1, Name: "Keyboard", Price: 250000},
			2: {ID: 2, Name: "Mouse", Price: 120000},
		},
	}
}

func (s *ProductService) key(id int) string { return fmt.Sprintf("product:%d", id) }

// loadFromDB mensimulasikan query database yang lambat.
func (s *ProductService) loadFromDB(id int) (Product, bool) {
	atomic.AddInt64(&s.dbHits, 1)
	time.Sleep(20 * time.Millisecond) // simulasi latensi DB
	p, ok := s.products[id]
	return p, ok
}

// Get menerapkan pola CACHE-ASIDE:
//  1. cek cache; kalau ada (hit) -> kembalikan.
//  2. kalau tidak ada (miss) -> baca DB, lalu isi cache dengan TTL.
//
// Mengembalikan juga 'fromCache' agar mudah didemokan.
func (s *ProductService) Get(ctx context.Context, id int) (Product, bool, error) {
	// 1. Coba dari cache.
	val, err := s.rdb.Get(ctx, s.key(id)).Result()
	if err == nil {
		var p Product
		if jsonErr := json.Unmarshal([]byte(val), &p); jsonErr == nil {
			return p, true, nil // cache HIT
		}
	} else if !errors.Is(err, redis.Nil) {
		return Product{}, false, err // error Redis sungguhan (bukan sekadar "tak ada")
	}

	// 2. Cache MISS -> baca DB.
	p, ok := s.loadFromDB(id)
	if !ok {
		return Product{}, false, fmt.Errorf("produk %d tidak ada", id)
	}

	// 3. Isi cache dengan TTL (biar tidak basi selamanya).
	if data, mErr := json.Marshal(p); mErr == nil {
		s.rdb.Set(ctx, s.key(id), data, s.ttl)
	}
	return p, false, nil
}

// Invalidate menghapus cache saat data berubah (mis. setelah update produk).
// Tanpa ini, cache bisa menyajikan data basi (stale).
func (s *ProductService) Invalidate(ctx context.Context, id int) error {
	return s.rdb.Del(ctx, s.key(id)).Err()
}

// UpdatePrice mengubah harga di "DB" lalu meng-INVALIDATE cache-nya.
func (s *ProductService) UpdatePrice(ctx context.Context, id, price int) error {
	p, ok := s.products[id]
	if !ok {
		return fmt.Errorf("produk %d tidak ada", id)
	}
	p.Price = price
	s.products[id] = p
	return s.Invalidate(ctx, id) // penting: buang cache lama
}

func (s *ProductService) DBHits() int64 { return atomic.LoadInt64(&s.dbHits) }
