package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/redis/go-redis/v9"
)

// DISTRIBUTED LOCK memastikan hanya SATU instance (dari banyak) yang menjalankan
// suatu tugas pada satu waktu (mis. cron job yang tak boleh dobel). Memakai Redis:
//   - Acquire: SET key token NX PX ttl  (NX = hanya jika belum ada).
//   - Release: hapus HANYA jika token cocok (via Lua, atomik) -> tak menghapus
//     lock milik instance lain.
//   - TTL: bila pemegang lock crash, lock otomatis kadaluarsa (tak deadlock).
type DistributedLock struct {
	rdb *redis.Client
}

func NewDistributedLock(rdb *redis.Client) *DistributedLock {
	return &DistributedLock{rdb: rdb}
}

// randToken menghasilkan token unik agar hanya pemilik yang boleh melepas lock.
func randToken() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// Acquire mencoba mengambil lock. Mengembalikan token (untuk Release) & ok=true
// bila berhasil, ok=false bila lock sedang dipegang instance lain.
func (l *DistributedLock) Acquire(ctx context.Context, key string, ttl time.Duration) (token string, ok bool, err error) {
	token = randToken()
	ok, err = l.rdb.SetNX(ctx, key, token, ttl).Result()
	if err != nil {
		return "", false, err
	}
	return token, ok, nil
}

// releaseScript: hapus key HANYA bila nilainya == token (atomik, cegah race).
var releaseScript = redis.NewScript(`
if redis.call("get", KEYS[1]) == ARGV[1] then
    return redis.call("del", KEYS[1])
else
    return 0
end`)

// Release melepas lock (hanya bila token cocok).
func (l *DistributedLock) Release(ctx context.Context, key, token string) error {
	return releaseScript.Run(ctx, l.rdb, []string{key}, token).Err()
}
