# 48 — gRPC-Gateway (satu proto → REST + gRPC)

Kadang kamu ingin **satu** definisi service melayani **dua** klien: gRPC (untuk service internal, cepat) **dan** REST/JSON (untuk browser/mobile/partner). gRPC-Gateway mewujudkannya dari satu `.proto`.

Jalankan:
```bash
go run ./48-grpc-gateway
# gRPC di :50054, REST gateway di :8081
curl -X POST localhost:8081/v1/greet -d '{"name":"Ana"}'   # -> {"message":"Halo, Ana!"}
```
Verifikasi otomatis (bufconn, tanpa port): `go test ./48-grpc-gateway`

## Konsep

```
                       ┌──────────────────┐
REST/JSON client ────► │  HTTP Gateway    │ ──┐
(browser, mobile)      └──────────────────┘   │ (terjemahkan ke gRPC)
                                               ▼
gRPC client ─────────────────────────► ┌──────────────────┐
(service internal)                      │  Greeter (gRPC)  │  <- satu-satunya logika
                                        └──────────────────┘
```
- **Satu sumber logika** (`service/service.go`) — implementasi gRPC.
- **Gateway** (`gateway/gateway.go`) menerima REST, menerjemahkannya ke panggilan gRPC, lalu mengubah balasan gRPC → JSON. Termasuk **memetakan kode gRPC → status HTTP** (`InvalidArgument` → 400, dst — Modul 17).

Test membuktikan: REST `POST /v1/greet {"name":"Ana"}` → gRPC → `{"message":"Halo, Ana!"}`, dan name kosong → gRPC `InvalidArgument` → HTTP **400**.

## Manual vs Generated

Modul ini menulis gateway **manual** agar konsepnya jelas & 100% testable. Di **produksi**, pakai **[grpc-gateway](https://github.com/grpc-ecosystem/grpc-gateway)** yang **men-generate** gateway ini dari **anotasi** di `.proto`:

```protobuf
import "google/api/annotations.proto";

service Greeter {
  rpc SayHello(HelloRequest) returns (HelloReply) {
    option (google.api.http) = { post: "/v1/greet" body: "*" };
  }
}
```
Lalu generate:
```bash
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
protoc --grpc-gateway_out=. --grpc-gateway_opt=module=... greeter.proto
```
Keuntungan versi generated: otomatis menangani path parameter, query string, streaming, & menghasilkan **OpenAPI spec** (Modul 28) sekaligus. Tulis endpoint sekali di proto → dapat gRPC + REST + dokumentasi.

## Kapan pakai?
- API yang dikonsumsi **internal (gRPC)** dan **eksternal/web (REST)** sekaligus.
- Migrasi bertahap REST → gRPC tanpa memutus klien lama.
- Ingin satu kontrak (proto) sebagai sumber kebenaran untuk semua antarmuka.

## Perbandingan pendekatan API di kurikulum ini
| Modul | Pendekatan |
|-------|-----------|
| 12 | REST murni (`net/http`) |
| 13 | REST framework (Fiber) |
| 16, 34 | gRPC (unary, streaming, interceptor) |
| 35 | GraphQL |
| **48** | **gRPC + REST dari satu proto** |

## Latihan
1. Tambah RPC `ListGreetings` + rute REST `GET /v1/greetings`.
2. Tambah path parameter: `GET /v1/greet/{name}` (baca `r.PathValue`).
3. Pasang grpc-gateway sungguhan dengan anotasi `google.api.http` & generate.
4. Tambahkan OpenAPI output (`--openapiv2_out`) & sajikan Swagger UI (Modul 28).
5. Tambah middleware (auth, logging) yang berlaku untuk REST **dan** gRPC.
