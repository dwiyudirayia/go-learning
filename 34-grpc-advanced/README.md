# 34 — gRPC Advanced (Streaming, Interceptor, Deadline)

Lanjutan Modul 16. Menutup **client streaming**, **bidirectional streaming**, **interceptor chain** (middleware gRPC), dan **deadline/cancellation**.

Jalankan:
```bash
go run ./34-grpc-advanced/server   # terminal 1 (:50053)
go run ./34-grpc-advanced/client   # terminal 2
```
Verifikasi otomatis (bufconn, tanpa port): `go test ./34-grpc-advanced/service`

## 4 tipe RPC (lengkap)

| Tipe | Proto | Contoh di sini |
|------|-------|----------------|
| Unary | `rpc F(Req) returns (Res)` | `SlowUnary` (Modul 16: `Add`) |
| Server streaming | `returns (stream Res)` | Modul 16: `Fibonacci` |
| **Client streaming** | `rpc F(stream Req) returns (Res)` | `Sum` |
| **Bidirectional** | `rpc F(stream Req) returns (stream Res)` | `Echo` |

### Client streaming (`Sum`)
Client kirim **banyak** angka, server balas **satu** total:
```go
// Server: baca sampai EOF, lalu SendAndClose.
for {
    req, err := stream.Recv()
    if err == io.EOF { return stream.SendAndClose(&SumResponse{Sum: total}) }
    total += req.Value
}
// Client: Send berkali-kali, lalu CloseAndRecv.
```
Berguna untuk **upload** / agregasi (kirim metrik, chunk file).

### Bidirectional (`Echo`)
Kedua arah mengalir bersamaan:
```go
for {
    msg, err := stream.Recv()
    if err == io.EOF { return nil }
    stream.Send(&EchoMessage{Text: "echo: " + msg.Text})
}
```
Berguna untuk **chat**, kolaborasi realtime, streaming dua arah.

## Interceptor Chain (middleware gRPC)

Rantai interceptor dijalankan berurutan sebelum handler — analog middleware HTTP (Modul 12):
```go
grpc.NewServer(
    grpc.ChainUnaryInterceptor(AuthUnaryInterceptor, LoggingUnaryInterceptor(logger)),
    grpc.ChainStreamInterceptor(AuthStreamInterceptor, LoggingStreamInterceptor(logger)),
)
```
- **Auth interceptor** memeriksa token di **metadata** (analog header HTTP). Test membuktikan: tanpa token → `Unauthenticated`, sebelum handler dipanggil.
- **Logging interceptor** mencatat tiap RPC.
- Metadata dikirim client: `metadata.AppendToOutgoingContext(ctx, "authorization", token)`.

## Deadline & Cancellation

Client menetapkan **deadline**; server wajib menghormatinya via `ctx`:
```go
// Client:
ctx, cancel := context.WithTimeout(ctx, 50*time.Millisecond); defer cancel()
client.SlowUnary(ctx, req) // -> DeadlineExceeded bila server terlalu lama

// Server:
select {
case <-time.After(delay): return result, nil
case <-ctx.Done(): return nil, status.Error(codes.DeadlineExceeded, "...")
}
```
**Deadline menular** (propagates) lintas service — bila A memanggil B dengan sisa deadline 30ms, B tahu batasnya. Mencegah request menggantung selamanya di rantai microservice.

## Status codes gRPC
`codes.OK/InvalidArgument/NotFound/Unauthenticated/DeadlineExceeded/Unavailable/...` — setara status HTTP (Modul 17 memetakannya).

## Kapan & Di Mana Dipakai
- Streaming: upload/download besar, feed realtime, chat, sinkronisasi.
- Interceptor: auth, logging, tracing (Modul 33), metrics (Modul 18), recovery — terpusat untuk semua RPC.

## Latihan
1. Tambah interceptor **recovery** (tangkap panic → `codes.Internal`).
2. Tambah interceptor **tracing** (Modul 33) memakai `otelgrpc`.
3. Tambah RPC server-streaming yang menghormati `ctx.Done()`.
4. Buat **gRPC-Gateway** (REST proxy di atas gRPC) dengan `protoc-gen-grpc-gateway`.
5. Tambah retry client-side dengan `grpc.WithDefaultServiceConfig` (retry policy).

## ✅ Solusi Latihan (Pembahasan)

1. **Interceptor recovery** — `defer func(){ if r:=recover(); r!=nil { err = status.Errorf(codes.Internal, "panic: %v", r) } }()` di dalam unary interceptor. Panic satu RPC tak menjatuhkan server.
2. **Interceptor tracing** — pakai `otelgrpc` (Modul 33) di chain interceptor. Rantai: recovery → tracing → logging → handler.
3. **Server-streaming hormati `ctx.Done()`** — dalam loop `stream.Send`, cek `select { case <-stream.Context().Done(): return ctx.Err(); default: }` agar berhenti saat klien pergi.
4. **gRPC-Gateway** — lihat Modul 48: hasilkan REST proxy dari `.proto` dengan `protoc-gen-grpc-gateway` (produksi) atau tulis manual (contoh 48).
5. **Retry client-side** — `grpc.WithDefaultServiceConfig` berisi JSON `retryPolicy` (`maxAttempts`, `retryableStatusCodes: ["UNAVAILABLE"]`). Retry transparan di layer transport.
