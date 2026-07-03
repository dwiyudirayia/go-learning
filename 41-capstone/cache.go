package main

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

// Cache membungkus Redis untuk pola cache-aside (Modul 22): lookup code->url
// yang sering diakses disimpan di memori cepat agar tak selalu menyentuh DB.
type Cache struct {
	rdb *redis.Client
	ttl time.Duration
}

func NewCache(rdb *redis.Client) *Cache {
	return &Cache{rdb: rdb, ttl: time.Hour}
}

func (c *Cache) key(code string) string { return "url:" + code }

// GetURL mengembalikan url dari cache; ok=false bila miss.
func (c *Cache) GetURL(ctx context.Context, code string) (string, bool, error) {
	url, err := c.rdb.Get(ctx, c.key(code)).Result()
	if errors.Is(err, redis.Nil) {
		return "", false, nil // miss (bukan error)
	}
	if err != nil {
		return "", false, err
	}
	return url, true, nil
}

func (c *Cache) SetURL(ctx context.Context, code, url string) {
	c.rdb.Set(ctx, c.key(code), url, c.ttl)
}
