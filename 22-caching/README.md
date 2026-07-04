# 22 — Caching dengan Redis

Cache menyimpan data yang sering diakses di memori cepat (Redis) agar tak selalu menyentuh database yang lambat. Modul ini pakai [go-redis](https://github.com/redis/go-redis) + [miniredis](https://github.com/alicebob/miniredis) (Redis in-memory untuk demo/test **tanpa server**).

Jalankan:
```bash
go run ./22-caching     # pakai miniredis, tak perlu Redis sungguhan
```
Verifikasi otomatis: `go test ./22-caching`

## Pola Cache-Aside (paling umum)

```
       ┌─────────┐   1. cek cache
       │  Redis  │ ◄───────────────┐
       └─────────┘                 │
            ▲ 3. isi cache         │
            │ (dengan TTL)         │
       ┌─────────┐   2. kalau miss ┌─────────┐
       │ Service │ ───────────────►│ Database│
       └─────────┘   baca DB       └─────────┘
```
```go
func Get(ctx, id) {
    if val, err := rdb.Get(ctx, key); err == nil { return val }  // 1. HIT
    p := db.Load(id)                                             // 2. MISS -> DB
    rdb.Set(ctx, key, p, ttl)                                    // 3. isi cache
    return p
}
```
Bukti di output: call 1 = **22ms** (DB), call 2 = **0ms** (cache), DB hanya disentuh **1×**.

## Konsep penting

### TTL (Time To Live)
`rdb.Set(ctx, key, val, 30*time.Second)` — cache otomatis kadaluarsa. Mencegah data basi menetap selamanya. Test memakai `mr.FastForward()` untuk menguji expiry tanpa menunggu.

### Invalidation (masalah tersulit di caching)
> *"There are only two hard things in Computer Science: cache invalidation and naming things."*

Saat data berubah, cache **harus dibuang** agar tidak menyajikan data lama:
```go
func UpdatePrice(id, price) {
    db.Update(id, price)
    rdb.Del(ctx, key)   // buang cache -> Get berikutnya ambil data segar
}
```

### `redis.Nil`
`Get` untuk key yang tak ada mengembalikan error `redis.Nil` — itu **cache miss normal**, bukan error sistem. Bedakan dengan error koneksi.

## Strategi cache lain (sekilas)
- **Write-through**: tulis ke cache & DB bersamaan.
- **Write-behind**: tulis cache dulu, DB belakangan (async).
- **Read-through**: cache yang otomatis load dari DB.
Cache-aside (modul ini) paling sederhana & fleksibel.

## Bahaya yang perlu diwaspadai
- **Stale data** — lupa invalidate → user lihat data lama.
- **Cache stampede** — banyak request miss bersamaan menyerbu DB. Solusi: singleflight, lock, atau TTL acak.
- **Cardinality** — jangan cache semuanya; pilih data yang sering dibaca & jarang berubah.

## Kapan & Di Mana Dipakai
- Data sering dibaca, jarang berubah: profil user, katalog produk, config, hasil query berat.
- Session store, rate limiter, leaderboard (Redis punya struktur data khusus).

## Latihan
1. Tambah `GetMany` yang mem-batch beberapa produk (pakai `MGet`).
2. Tambah TTL acak (jitter) untuk mencegah cache stampede.
3. Tambah metrik hit/miss ratio (integrasikan Prometheus, Modul 18).
4. Ganti miniredis ke Redis sungguhan via env `REDIS_ADDR` (Modul 19).
5. Implementasikan `singleflight` (`golang.org/x/sync/singleflight`) agar miss bersamaan hanya 1× ke DB.

## ✅ Solusi Latihan (Pembahasan)

1. **`GetMany` dengan `MGet`** — satu round-trip untuk banyak key:
   ```go
   keys := []string{"product:1","product:2"}
   vals, _ := rdb.MGet(ctx, keys...).Result() // []any, nil untuk yang miss
   ```
   Kumpulkan yang miss → ambil dari DB → `MSet` balik.
2. **TTL jitter (anti stampede)** — acak sedikit TTL agar tak semua expired serempak:
   ```go
   ttl := base + time.Duration(rand.Intn(60))*time.Second
   rdb.Set(ctx, key, val, ttl)
   ```
3. **Metrik hit/miss** — dua Counter (Modul 18): `cacheHits.Inc()` saat ketemu, `cacheMiss.Inc()` saat tidak. Ratio = hits/(hits+miss).
4. **Redis sungguhan via env** — `addr := getenv("REDIS_ADDR","localhost:6379")`; kalau kosong pakai miniredis (test), kalau ada pakai `redis.NewClient(&redis.Options{Addr: addr})`. Kode aplikasi tak berubah (Modul 19).
5. **`singleflight`** — saat 100 request miss bersamaan, hanya 1 yang menembus DB:
   ```go
   v, err, _ := g.Do(key, func()(any,error){ return loadFromDB(key) })
   ```
   `g` adalah `singleflight.Group`. Sisanya menunggu hasil yang sama → DB tak terbebani (Modul 38).
