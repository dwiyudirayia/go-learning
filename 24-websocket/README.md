# 24 — Real-time: WebSocket & Server-Sent Events

HTTP biasa bersifat request-response (client bertanya, server menjawab). Untuk fitur **realtime** (chat, notifikasi, live dashboard, harga saham), server perlu **mendorong** data ke client. Dua cara utama: **WebSocket** (dua arah) & **SSE** (satu arah).

Jalankan:
```bash
go run ./24-websocket
# buka http://localhost:8080 di DUA tab -> chat realtime antar tab
```
Verifikasi otomatis: `go test ./24-websocket`

## WebSocket vs SSE vs Polling

| | Polling | SSE | WebSocket |
|-|---------|-----|-----------|
| Arah | client→server | server→client | **dua arah** |
| Protokol | HTTP berulang | HTTP (stream) | ws:// (upgrade) |
| Kompleksitas | rendah | rendah | sedang |
| Cocok untuk | data jarang berubah | notifikasi, feed | chat, game, kolaborasi |

## Arsitektur: Hub (pola pub/sub)

Inti realtime = **broadcast**. `hub.go` mengelola daftar subscriber & menyiarkan pesan — **murni Go, tanpa jaringan**, jadi mudah di-test.

```go
hub.Subscribe()   // client baru -> channel penerima
hub.Broadcast(m)  // kirim ke SEMUA subscriber (non-blocking)
hub.Unsubscribe() // client putus -> hapus
```

**Poin penting:** `Broadcast` **non-blocking** (`select { case ch<-m: default: }`) — satu client lambat tidak boleh menahan yang lain. Buffer channel + skip bila penuh.

## WebSocket (`ws.go`)
```go
c, _ := websocket.Accept(w, r, opts)   // upgrade HTTP -> WebSocket
// goroutine penulis: hub -> koneksi
// loop pembaca:      koneksi -> hub.Broadcast
```
Tiap koneksi = satu goroutine pembaca + satu penulis. Pola ini menskala ke ribuan koneksi karena goroutine murah (Modul 7). Pakai [`coder/websocket`](https://github.com/coder/websocket) (context-based, modern).

## Server-Sent Events (SSE)
Lebih sederhana — HTTP biasa dengan `Content-Type: text/event-stream`:
```go
w.Header().Set("Content-Type", "text/event-stream")
fmt.Fprintf(w, "data: %s\n\n", msg)  // format SSE
flusher.Flush()                       // dorong segera
```
Client pakai `new EventSource("/events")`. Auto-reconnect bawaan. Cocok untuk **notifikasi satu arah**.

## Hal yang perlu diperhatikan di produksi
- **Autentikasi** saat upgrade (cek token di handshake).
- **Origin check** (jangan `InsecureSkipVerify` di produksi — itu hanya untuk contoh).
- **Backpressure**: client lambat → drop pesan atau putus.
- **Skala horizontal**: banyak instance server → butuh broadcast lintas instance (Redis Pub/Sub atau NATS, Modul 22–23).
- **Heartbeat/ping** agar koneksi mati terdeteksi.

## Kapan & Di Mana Dipakai
- Chat, notifikasi push, live dashboard/metrics, kolaborasi (Google Docs), game, live tracking (ojek online).

## Latihan
1. Tambah "nama user" & tampilkan `[nama]: pesan`.
2. Tambah room/channel (subject seperti Modul 23) agar broadcast hanya ke room tertentu.
3. Tambah ping/pong heartbeat untuk mendeteksi koneksi mati.
4. Skala ke banyak instance: broadcast lewat Redis Pub/Sub (Modul 22).
5. Tambah endpoint SSE untuk "notifikasi order baru" yang dipicu event NATS (Modul 23).

## ✅ Solusi Latihan (Pembahasan)

1. **Nama user** — saat handshake, kirim pesan pertama berisi nama (atau baca query `?name=`). Simpan di struct `client{name string; conn *websocket.Conn}` dan format broadcast `fmt.Sprintf("[%s]: %s", c.name, msg)`.
2. **Room/channel** — ganti `map[*client]bool` jadi `map[string]map[*client]bool` (key = room). Broadcast hanya loop anggota room tersebut. Mirip konsep subject di Modul 23.
3. **Ping/pong heartbeat** — kirim ping periodik; jika tak ada pong dalam batas waktu, tutup koneksi:
   ```go
   ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
   if err := c.Ping(ctx); err != nil { c.Close(websocket.StatusPolicyViolation, "no pong") }
   ```
4. **Skala multi-instance** — tiap instance subscribe channel Redis; pesan lokal juga di-`PUBLISH` ke Redis, lalu tiap instance mem-broadcast ke klien-nya sendiri. Redis Pub/Sub jadi "bus" antar-instance (Modul 22).
5. **SSE dari NATS** — handler SSE (`Content-Type: text/event-stream`) subscribe subject `order.created` (Modul 23); tiap event tulis `fmt.Fprintf(w, "data: %s\n\n", json)` lalu `flush`. Cocok untuk notifikasi satu-arah server→browser.

---

## 🚀 Teknik Advanced (Level Up)
> 💻 **Contoh runnable + komentar detail** untuk teknik di bawah ada di folder [`advanced/`](advanced). Jalankan: `go run ./24-websocket/advanced`


- **Heartbeat** — ping/pong berkala + read deadline untuk deteksi koneksi mati; tanpa ini koneksi zombie menumpuk.
- **Hub pattern** — satu goroutine hub kelola register/unregister/broadcast; tiap koneksi punya goroutine baca & tulis terpisah.
- **Satu writer per koneksi** — WebSocket tak aman untuk tulis konkuren; serialkan via channel per-conn.
- **Backpressure** — jika client lambat, buffer penuh → putuskan atau drop, jangan biarkan memori meledak.
- **SSE vs WebSocket** — SSE (satu arah, server→client, di atas HTTP biasa, auto-reconnect) lebih sederhana untuk notifikasi; WS untuk dua arah.
- **Keamanan** — cek `Origin` (cegah CSWSH), autentikasi saat handshake, batasi ukuran pesan.
- **Graceful close** — kirim close frame, tunggu, lalu tutup TCP.
