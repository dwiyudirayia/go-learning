# 46 — Integrasi Layanan Pihak Ketiga

Aplikasi nyata memanggil layanan eksternal: **payment** (Stripe), **storage** (S3), **email** (SendGrid), dan menerima **webhook**. Modul ini mengajarkan **pola integrasi yang testable & tahan vendor**.

Jalankan:
```bash
go run ./46-service-integrations
go test ./46-service-integrations
```

## Pola inti: Interface + Mock (Adapter)

**Jangan panggil SDK vendor langsung dari logika bisnismu.** Bungkus di **interface**; produksi pakai implementasi nyata, test pakai mock.
```go
type PaymentGateway interface { Charge(ctx, ChargeRequest) (ChargeResult, error) }
type Storage        interface { Put(ctx, key, data) error; Get(ctx, key) ([]byte, error) }
type Emailer        interface { Send(ctx, Email) error }
```
Manfaat:
- **Testable** — `OrderService.Checkout` diuji **tanpa** memanggil Stripe/S3 nyata (test membuktikan bayar+simpan+email terjadi lewat mock).
- **Tahan vendor** — ganti Stripe→Midtrans, S3→GCS = tulis adapter baru, logika bisnis tak berubah (Modul 29).
- **Aman** — test tak butuh kredensial/kuota.

```go
// Produksi: NewOrderService(stripeGW, s3Storage, sendgridEmailer)
// Test:     NewOrderService(&MockGateway{}, NewInMemoryStorage(), &MockEmailer{})
```

### Implementasi nyata (di produksi)
| Layanan | Library Go |
|---------|-----------|
| Payment | [stripe-go](https://github.com/stripe/stripe-go) |
| Storage | [aws-sdk-go-v2/s3](https://github.com/aws/aws-sdk-go-v2) / GCS / MinIO |
| Email | [sendgrid-go](https://github.com/sendgrid/sendgrid-go) / SMTP |

## Webhook — verifikasi WAJIB 🔒

Layanan mengirim event ke endpoint-mu ("pembayaran berhasil"). Endpoint itu **publik** → siapa pun bisa mengirim event palsu. **Wajib verifikasi tanda tangan** (`webhook.go`):
```
tanda tangan = HMAC-SHA256(signing_secret, "timestamp.payload")
```
```go
err := VerifyWebhook(secret, signatureHeader, payload, 5*time.Minute, time.Now())
```
Tiga proteksi (test membuktikan semuanya):
1. **Integritas** — payload dipalsukan → HMAC tak cocok → **ditolak**.
2. **Autentisitas** — hanya yang tahu `signing_secret` bisa membuat tanda tangan valid.
3. **Anti-replay** — timestamp lama (> toleransi) → **ditolak** (cegah event lama dikirim ulang).

Perbandingan pakai `hmac.Equal` (**constant-time**) → cegah timing attack.

## Praktik penting
- **Idempotensi** — webhook bisa terkirim ganda (at-least-once). Simpan `event_id` yang sudah diproses & lewati duplikat (Modul 25, 31).
- **Retry & timeout** (Modul 32) untuk panggilan keluar yang gagal.
- **Circuit breaker** (Modul 32) agar vendor yang down tak menjatuhkan aplikasimu.
- **Balas 2xx cepat** ke webhook, proses berat di background (Modul 25) — pengirim akan retry bila lambat.
- **Secret dari env** (Modul 19), jangan di kode.

## Kapan & Di Mana Dipakai
- Hampir semua produk: pembayaran, upload file, notifikasi email/SMS, integrasi SaaS.

## Latihan
1. Tambah `SMSSender` interface + mock (mis. Twilio).
2. Tambah idempotensi webhook: simpan `event_id` diproses (Modul 22/25), tolak duplikat.
3. Bungkus `PaymentGateway` dengan retry + circuit breaker (Modul 32).
4. Tambah handler HTTP (Modul 12/13) yang menerima & memverifikasi webhook.
5. Ganti `InMemoryStorage` dengan MinIO (S3-compatible) lewat aws-sdk-go-v2.

## ✅ Solusi Latihan (Pembahasan)

1. **`SMSSender` + mock** — interface `SMSSender interface{ Send(to, body string) error }`; mock merekam panggilan untuk assert di test (pola Modul 08). Implementasi nyata mis. Twilio.
2. **Idempotensi webhook** — simpan `event_id` yang sudah diproses (Redis SET/tabel, Modul 22/25); bila `event_id` sudah ada → balas 200 tanpa proses ulang. Cegah double-charge saat provider retry.
3. **`PaymentGateway` + resiliency** — bungkus dengan retry + circuit breaker (Modul 32); panggilan pembayaran harus tahan jaringan flaky tapi tak double-charge (kombinasikan dengan idempotensi #2).
4. **Handler webhook** — endpoint HTTP (Modul 13) yang **verifikasi HMAC-SHA256** signature + cek timestamp (anti-replay) sebelum memproses. Tolak bila signature tak cocok.
5. **MinIO** — ganti `InMemoryStorage` dengan client S3-compatible (`aws-sdk-go-v2`, endpoint MinIO). Interface `Storage` membuat swap ini satu baris di `main`.

---

## 🚀 Teknik Advanced (Level Up)
> 💻 **Contoh runnable + komentar detail** untuk teknik di bawah ada di folder [`advanced/`](advanced). Jalankan: `go run ./46-service-integrations/advanced`


- **Interface + mock** — bungkus tiap 3rd-party (Stripe/S3/email) di interface; uji dengan mock (tanpa panggil API nyata / tanpa biaya).
- **Webhook aman** — verifikasi **HMAC signature** (`hmac.Equal`, constant-time) + **anti-replay** (tolak timestamp lama/nonce dipakai ulang). Pola repo ini.
- **Idempotency key** — kirim key agar retry ke pihak ketiga tak menggandakan (mis. double charge).
- **Circuit breaker & timeout** — lindungi dari dependency eksternal lambat/mati [[32-resiliency-patterns]].
- **Retry idempoten + backoff** — hanya retry operasi aman.
- **Secrets & rate limits** — kelola API key aman [[27-security]]; hormati rate limit vendor.
