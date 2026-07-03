# 16 — gRPC (Protobuf, Unary & Streaming)

gRPC adalah RPC berkinerja tinggi dari Google: kontrak didefinisikan di **Protocol Buffers** (`.proto`), lalu di-*generate* jadi kode server & client yang type-safe. Cepat (HTTP/2, biner), cocok untuk **komunikasi antar-microservice**.

## Prasyarat (sekali pasang)
```bash
# compiler protobuf
sudo apt-get install -y protobuf-compiler      # (Ubuntu/WSL)
# plugin Go
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
# pastikan $(go env GOPATH)/bin ada di PATH
```

## Jalankan
```bash
# terminal 1 — server
go run ./16-grpc/server            # :50051

# terminal 2 — client
go run ./16-grpc/client
```
Verifikasi otomatis (tanpa port, pakai bufconn): `go test ./16-grpc/service`

## Alur kerja gRPC

### 1. Definisikan kontrak (`proto/calculator.proto`)
```proto
service Calculator {
  rpc Add(AddRequest) returns (AddResponse);              // unary
  rpc Fibonacci(FibRequest) returns (stream FibResponse); // server-streaming
}
message AddRequest { int64 a = 1; int64 b = 2; }
```
Nomor field (`= 1`, `= 2`) adalah **tag biner** — jangan diubah setelah dipakai (kompatibilitas).

### 2. Generate kode
```bash
protoc --go_out=. --go_opt=module=go-learning \
       --go-grpc_out=. --go-grpc_opt=module=go-learning \
       16-grpc/proto/calculator.proto
```
Menghasilkan `calculator.pb.go` (message) & `calculator_grpc.pb.go` (client+server stub). **File ini di-generate — jangan diedit manual.**

### 3. Implementasikan server (`service/service.go`)
```go
type CalculatorServer struct{ calcpb.UnimplementedCalculatorServer } // wajib embed
func (s *CalculatorServer) Add(ctx, req) (*AddResponse, error) { ... }
func (s *CalculatorServer) Fibonacci(req, stream) error { stream.Send(...) }
```
`UnimplementedCalculatorServer` disematkan agar tetap kompilasi bila proto menambah method (forward-compatible).

### 4. Client memakai stub
```go
conn, _ := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
client := calcpb.NewCalculatorClient(conn)
resp, _ := client.Add(ctx, &calcpb.AddRequest{A: 7, B: 5})   // unary
stream, _ := client.Fibonacci(ctx, &calcpb.FibRequest{N: 10}) // streaming
for { resp, err := stream.Recv(); if err == io.EOF { break }; ... }
```

## Empat tipe RPC
| Tipe | Pola | Contoh |
|------|------|--------|
| Unary | 1 req → 1 resp | `Add` |
| Server streaming | 1 req → banyak resp | `Fibonacci`, feed data |
| Client streaming | banyak req → 1 resp | upload potongan |
| Bidirectional | banyak ↔ banyak | chat, realtime |

## gRPC vs REST
| | REST/JSON | gRPC |
|-|-----------|------|
| Format | teks (JSON) | biner (protobuf), lebih kecil/cepat |
| Kontrak | opsional (OpenAPI) | wajib (`.proto`), type-safe |
| Streaming | terbatas | native (4 mode) |
| Browser | langsung | perlu grpc-web/proxy |
| Cocok | API publik, web | **antar-microservice**, internal, low-latency |

## Testing dengan bufconn
`service/service_test.go` menjalankan server gRPC di atas koneksi **in-memory** (`bufconn`) — tak perlu buka port. Ini cara standar unit-test gRPC: cepat & deterministik.

## Kapan Dipakai
Komunikasi internal antar service (order-service ↔ payment-service), sistem low-latency, streaming data. Untuk API yang dikonsumsi browser/publik, REST sering lebih praktis — banyak sistem memakai **keduanya** (REST di edge, gRPC di dalam).

## Latihan
1. Tambah RPC unary `Multiply`.
2. Tambah RPC **client-streaming** `Sum(stream Number) returns (Total)`.
3. Tambah RPC **bidirectional** sederhana (echo).
4. Tambah interceptor (middleware gRPC) untuk logging tiap panggilan.
5. Kembalikan error gRPC yang benar dengan `status.Errorf(codes.InvalidArgument, ...)` saat input tak valid.

> Studi kasus 2 service berkomunikasi via gRPC ada di **Modul 17**.

## ✅ Status Solusi Latihan
Latihan **1, 4, 5 sudah diselesaikan**: RPC `Multiply`, interceptor logging (service/interceptor.go), dan error `InvalidArgument` (status codes). Latihan 2 & 3 (client-streaming, bidirectional) sebagai tantangan.
