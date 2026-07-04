# 31 — Saga & Outbox Pattern

Di microservices, satu operasi bisnis sering menyentuh **banyak service/database**. Transaksi ACID biasa tak cukup (tak ada transaksi lintas DB). Dua pola kunci: **Saga** (konsistensi via kompensasi) & **Outbox** (event tak hilang).

Jalankan:
```bash
go run ./31-saga-outbox
```
Verifikasi otomatis: `go test ./31-saga-outbox`

## 1. Saga — transaksi terdistribusi via kompensasi

Alih-alih mengunci semua service (mahal & rapuh), saga menjalankan **langkah lokal** berurutan. Bila satu gagal, jalankan **aksi kompensasi** (undo) untuk langkah yang sudah sukses, dalam **urutan terbalik**.

```
reserve-stok  ->  bayar  ->  kirim ❌
     ✔            ✔          gagal
                             ↓ kompensasi mundur:
                undo:bayar  <-  undo:reserve-stok
```
Output membuktikan: gagal di "kirim" → `undo:bayar` lalu `undo:reserve-stok` → semua state kembali bersih.

### Orkestrasi vs Koreografi
- **Orkestrasi** (modul ini): satu koordinator memanggil langkah + kompensasi. Mudah dipahami & di-debug.
- **Koreografi**: tiap service bereaksi ke event (Modul 23), tanpa koordinator pusat. Lebih decoupled tapi sulit dilacak.

### Catatan penting
- Kompensasi harus **idempotent** & bisa gagal → siapkan alert + retry (semantik "eventually consistent").
- Saga bukan rollback ACID — ada jendela waktu di mana state "setengah jalan".

## 2. Outbox — mengatasi masalah "dual write"

Masalah: kamu perlu **update DB** DAN **publish event** (Modul 23). Kalau app crash setelah commit DB tapi sebelum publish → event **hilang**. Kalau publish dulu lalu DB gagal → event **hantu**.

**Solusi Outbox:** tulis event ke tabel `outbox` dalam **transaksi yang sama** dengan perubahan bisnis. Sebuah **relay** membaca outbox lalu publish.
```go
tx.Exec("INSERT INTO orders ...")                 // perubahan bisnis
tx.Exec("INSERT INTO outbox(topic,payload) ...")  // event — TRANSAKSI SAMA
tx.Commit()                                        // keduanya atomik

// Relay (background):
relay.ProcessOnce(ctx) // SELECT pending -> publish -> UPDATE sent=1
```
- **Atomik**: order & event berhasil/gagal bersama.
- **Tak hilang**: crash sebelum publish? Relay tetap menemukannya nanti.
- **At-least-once**: relay bisa publish 2× (crash setelah publish, sebelum `sent=1`) → konsumen wajib **idempotent** (Modul 25).
- Publish gagal → `sent` tetap 0 → dicoba lagi (test membuktikan event tak hilang).

> Produksi: relay bisa polling (modul ini) atau membaca **change data capture** (Debezium membaca WAL Postgres).

## Idempotency Key (pelengkap)
Untuk operasi yang dipicu ulang (retry, pesan ganda), sematkan **idempotency key** unik per operasi; simpan key yang sudah diproses → operasi kedua dengan key sama diabaikan.

## Kapan & Di Mana Dipakai
- Checkout e-commerce (stok + bayar + kirim), transfer antar-akun, booking multi-langkah.
- Setiap kali kamu "update DB lalu kirim event" (Outbox) — sangat umum di event-driven.

## Latihan
1. Tambah langkah `notify` pada saga + kompensasinya.
2. Tambah kolom `idempotency_key` unik di outbox agar event tak ganda.
3. Jalankan relay sebagai loop (goroutine + ticker, Modul 25) alih-alih sekali jalan.
4. Integrasikan relay ke broker Modul 23 (publish sungguhan ke NATS/Kafka).
5. Implementasikan saga **koreografi** memakai event (tanpa koordinator pusat).

## ✅ Solusi Latihan (Pembahasan)

1. **Langkah `notify` + kompensasi** — tambah step ke daftar saga dengan pasangan `do`/`undo`. Bila step setelahnya gagal, koordinator memanggil `undo` tiap step yang sudah sukses (urutan mundur).
2. **`idempotency_key` unik** — kolom `UNIQUE` di tabel outbox; sebelum insert cek/`INSERT OR IGNORE`. Mencegah event ganda saat relay retry.
3. **Relay sebagai loop** — `ticker := time.NewTicker(1*time.Second)`; tiap tick baca outbox `WHERE published = 0`, publish, tandai `published = 1`. Pola worker Modul 25.
4. **Publish ke broker nyata** — ganti "publish" simulasi dengan `broker.Publish(topic, event)` (Modul 23). Outbox menjamin **atomic**: tulis state + event dalam satu transaksi DB, relay kirim belakangan (at-least-once).
5. **Saga koreografi** — tanpa koordinator pusat: tiap service bereaksi terhadap event dan menerbitkan event berikutnya. Lebih longgar (loose coupling) tapi alur lebih sulit dilacak → andalkan tracing (Modul 33).
