# 17 — 📦 Studi Kasus: Microservices (Fiber + gRPC + Docker)

Dua service yang berkomunikasi — menutup seluruh kurikulum. Mendemokan pola produksi paling umum: **REST di edge, gRPC di internal**.

```
        Client (curl/browser)
              │ HTTP/JSON
              ▼
     ┌─────────────────┐        gRPC         ┌────────────────────┐
     │  order-service  │ ──────────────────► │ inventory-service  │
     │   (Fiber/HTTP)  │  ReserveStock/Get   │     (gRPC)         │
     └─────────────────┘                     └────────────────────┘
       edge, publik                            internal, cepat
```

- **inventory-service** — gRPC server, mengelola stok produk (in-memory).
- **order-service** — HTTP API (Fiber) yang jadi **client gRPC** ke inventory. Menerima order, cek & reservasi stok lewat inventory, balas konfirmasi.

## Jalankan

### Opsi A — dua proses lokal
```bash
# terminal 1
go run ./17-studi-kasus-microservices/inventory-service      # gRPC :50052

# terminal 2
go run ./17-studi-kasus-microservices/order-service          # HTTP :3000
```
```bash
curl localhost:3000/products/1
curl -X POST localhost:3000/orders -H 'Content-Type: application/json' -d '{"product_id":1,"qty":2}'
```

### Opsi B — Docker Compose (satu perintah)
```bash
cd 17-studi-kasus-microservices
docker compose up --build
# order-service terbuka di localhost:3000, inventory hanya internal
```

### Verifikasi otomatis (tanpa proses/port, bufconn)
```bash
go test ./17-studi-kasus-microservices/...
```

## Konsep kunci

### 1. Komunikasi antar service
`order-service` memegang **`InventoryClient`** (interface hasil generate). Di produksi diisi koneksi gRPC sungguhan (`grpc.NewClient`), di test diisi client **bufconn** — kode `BuildApp` tidak berubah. Ini dependency injection lintas service.

### 2. Pemetaan error lintas protokol
inventory mengembalikan **kode gRPC** yang bermakna; order menerjemahkannya ke **status HTTP**:
| gRPC code (inventory) | HTTP (order) | Kondisi |
|-----------------------|--------------|---------|
| `NotFound` | 404 | produk tak ada |
| `FailedPrecondition` | 409 | stok tidak cukup |
| `InvalidArgument` | 400 | qty ≤ 0 |
| lainnya | 502 Bad Gateway | inventory bermasalah |

### 3. Service discovery via DNS Compose
Di `docker-compose.yml`, order memakai `INVENTORY_ADDR=inventory:50052` — nama `inventory` di-resolve oleh DNS internal Docker Compose. Di Kubernetes polanya sama (nama Service).

### 4. Image kecil & aman (multi-stage build)
`Dockerfile` membangun binari statis (`CGO_ENABLED=0`) lalu menyalinnya ke image **distroless** (tanpa shell/OS penuh) → kecil & permukaan serangan minimal.

### 5. Konfigurasi via environment (12-factor)
Semua alamat/port dari env (`GRPC_ADDR`, `INVENTORY_ADDR`, `PORT`) — tak ada yang di-hardcode. Sama untuk lokal, Docker, maupun cloud.

## Struktur
```
17-studi-kasus-microservices/
├── proto/inventory.proto (+ generated)
├── internal/
│   ├── inventoryserver/   # impl gRPC (dipakai binary & test)
│   └── orderapp/          # Fiber app + gRPC client (+ integration test bufconn)
├── inventory-service/  main.go  Dockerfile
├── order-service/      main.go  Dockerfile
└── docker-compose.yml
```

## Yang membuat ini "produksi-minded"
- Interface untuk dependency (testable, tukar implementasi).
- Error terstruktur & dipetakan lintas protokol.
- Config via env; image minimal; service terisolasi.
- Integration test in-process (bufconn) — cepat, jalan di CI tanpa Docker.

## Latihan (proyek penutup)
1. Tambah **payment-service** (gRPC) yang dipanggil order-service setelah reservasi stok.
2. Tambah **interceptor** gRPC untuk logging + request ID lintas service.
3. Tambah retry + timeout (`context`) saat order memanggil inventory.
4. Tambah database (Modul 14) pada inventory-service (ganti in-memory).
5. Tambah **health check** endpoint & `depends_on: condition: service_healthy` di compose.
6. Deploy ke Kubernetes (Deployment + Service untuk masing-masing).

---
🎓 **Selamat!** Kamu telah menuntaskan seluruh kurikulum: dari sintaks dasar hingga sistem microservices. Lihat `README.md` root untuk peta lengkapnya.

## ✅ Status Solusi Latihan
Latihan **3 & 5 sudah diselesaikan**: timeout + retry pada panggilan gRPC (`reserveWithRetry`) dan health check (`GET /health`). Latihan 1, 2, 4, 6 (payment-service, interceptor, database, K8s) sebagai proyek lanjutan.
