# 27 — Security

Keamanan bukan fitur tambahan — harus dibangun sejak awal. Modul ini menutup celah umum: **rate limiting**, **security headers**, **refresh token**, plus **TLS/mTLS** (dijelaskan).

Jalankan:
```bash
go run ./27-security
```
Verifikasi otomatis: `go test ./27-security`

## 1. Rate Limiting (per-IP token bucket)

Cegah brute-force, scraping, & DoS ringan. Pakai `golang.org/x/time/rate`:
```go
limiter := rate.NewLimiter(rate.Limit(5), 5) // 5 req/detik, burst 5
if !limiter.Allow() { http.Error(w, "...", http.StatusTooManyRequests) } // 429
```
Kita simpan **satu limiter per IP** (map). Test membuktikan: request ke-6 dari IP sama → **429**, IP berbeda tetap lolos.

> Produksi: rate limit terdistribusi pakai Redis (Modul 22) agar konsisten lintas instance.

## 2. Security Headers

Pertahanan murah di setiap response:
| Header | Melindungi dari |
|--------|-----------------|
| `X-Content-Type-Options: nosniff` | MIME sniffing |
| `X-Frame-Options: DENY` | clickjacking (iframe) |
| `Content-Security-Policy` | XSS, injeksi resource |
| `Strict-Transport-Security` | downgrade ke HTTP |
| `Referrer-Policy` | kebocoran URL referrer |

## 3. Access + Refresh Token

Menyempurnakan JWT dari Modul 15:
```
login  -> { access (15 menit), refresh (7 hari) }
access token kadaluarsa -> POST /refresh dengan refresh token -> access baru
```
- **Access** pendek → kalau bocor, cepat mati.
- **Refresh** panjang, disimpan aman (HttpOnly cookie), bisa **dicabut** (revoke).
- Token menyimpan `typ` ("access"/"refresh") → access token **ditolak** untuk endpoint refresh (test membuktikan).

## 4. TLS & mTLS (untuk gRPC/HTTP)

**TLS** mengenkripsi lalu lintas:
```go
http.ListenAndServeTLS(":443", "cert.pem", "key.pem", handler)
// gRPC:
creds, _ := credentials.NewServerTLSFromFile("cert.pem", "key.pem")
grpc.NewServer(grpc.Creds(creds))
```
**mTLS** (mutual TLS) — client **dan** server saling memverifikasi sertifikat. Umum untuk komunikasi internal antar-microservice (Modul 17) sehingga hanya service tepercaya bisa saling bicara. Di produksi sering ditangani service mesh (Istio/Linkerd).

## Prinsip keamanan lain (checklist)
- ✅ Password **di-hash** (bcrypt, Modul 15) — jangan plaintext.
- ✅ Query pakai **placeholder** (Modul 14) — cegah SQL injection.
- ✅ **Validasi & sanitasi** input (Modul 13).
- ✅ Secret dari **env/secret manager** (Modul 19) — jangan di kode.
- ✅ **HTTPS** di mana-mana; **least privilege** untuk akses DB/cloud.
- ✅ Jangan bocorkan detail error ke user (log internal, balas pesan generik).
- ✅ Update dependency (`govulncheck ./...` untuk cek kerentanan).

## Kapan & Di Mana Dipakai
- **Setiap** aplikasi yang menghadap internet. Wajib untuk auth, API publik, dan komunikasi antar service.

## Latihan
1. Tambah **revoke** refresh token (simpan daftar token dicabut di Redis/DB).
2. Tambah rate limit **berbeda** untuk `/login` (lebih ketat) vs endpoint lain.
3. Buat sertifikat self-signed (`crypto/tls`, `crypto/x509`) & jalankan `ListenAndServeTLS`.
4. Tambah `govulncheck` ke CI (Modul CI) untuk memindai kerentanan.
5. Terapkan security headers & rate limit ke studi kasus Modul 15.

## ✅ Solusi Latihan (Pembahasan)

1. **Revoke refresh token** — simpan daftar token dicabut (atau whitelist token aktif) di Redis: `SADD revoked <jti>` dengan TTL = sisa umur token. Saat refresh, tolak bila `SISMEMBER revoked <jti>`.
2. **Rate limit berbeda untuk `/login`** — pasang limiter lebih ketat (mis. 5/menit per IP) khusus route login, terpisah dari limiter global. Cegah brute-force.
3. **TLS self-signed** — generate dengan `crypto/x509` + `crypto/ecdsa`, tulis `cert.pem`/`key.pem`, lalu `srv.ListenAndServeTLS("cert.pem","key.pem")`. Untuk dev; produksi pakai Let's Encrypt/`autocert`.
4. **`govulncheck` di CI** — tambah step: `go install golang.org/x/vuln/cmd/govulncheck@latest && govulncheck ./...`. Gagalkan build bila ada kerentanan pada dependency.
5. **Terapkan ke Modul 15** — pasang middleware security headers (`X-Content-Type-Options`, `Strict-Transport-Security`, dll) + rate limit di server REST studi kasus. Keamanan = lapisan default, bukan tambalan.
