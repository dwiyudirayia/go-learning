// REAL-CASE Modul 17 (microservices) — dua service berkomunikasi via gRPC TCP.
//
// Versi advanced/ mensimulasikan panggilan antar-service dgn fungsi in-process.
// Versi ini menjalankan "inventory service" sebagai server gRPC pada PORT TCP
// nyata, lalu "order service" memanggilnya sebagai CLIENT — persis dua proses
// terpisah (hanya saja di satu mesin). Berjalan lokal tanpa infra eksternal.
//
// Jalankan:
//
//	go run ./17-studi-kasus-microservices/real-case
package main

import (
	"context"
	"fmt"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
)

// startInventoryService menjalankan "inventory service" sebagai server gRPC.
//
// Untuk contoh ini kita pakai layanan Health bawaan gRPC sebagai stand-in:
// status SERVING berarti "stok tersedia". Fungsi mengembalikan ALAMAT TCP tempat
// server mendengarkan dan fungsi stop() untuk mematikannya.
//
// Return:
//   - addr : alamat "host:port" yang bisa di-dial client (mis. 127.0.0.1:45xxx)
//   - stop : panggil untuk menghentikan server dengan rapi
func startInventoryService() (addr string, stop func()) {
	// net.Listen("tcp", "127.0.0.1:0") -> minta port BEBAS dari OS (":0").
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	srv := grpc.NewServer()

	// health.NewServer() menyediakan implementasi standar grpc.health.v1.
	h := health.NewServer()
	h.SetServingStatus("stock", healthpb.HealthCheckResponse_SERVING) // "stok" sehat
	healthpb.RegisterHealthServer(srv, h)

	// Serve() memblokir, jadi dijalankan di goroutine terpisah.
	go func() { _ = srv.Serve(lis) }()
	return lis.Addr().String(), srv.GracefulStop
}

// OrderService = "order service" yang BERGANTUNG pada inventory service.
// Ia menyimpan koneksi gRPC ke inventory (client), bukan implementasinya —
// inilah batas antar-service.
type OrderService struct {
	inventory healthpb.HealthClient
}

// NewOrderService membuat client gRPC yang terhubung ke alamat inventory.
//
// Param:
//   - inventoryAddr : alamat TCP inventory service (dari startInventoryService)
//
// Catatan: di produksi pakai TLS/mTLS, BUKAN insecure. grpc.NewClient bersifat
// "lazy" — koneksi TCP nyata baru dibuat saat RPC pertama.
func NewOrderService(inventoryAddr string) *OrderService {
	conn, err := grpc.NewClient(inventoryAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	return &OrderService{inventory: healthpb.NewHealthClient(conn)}
}

// CreateOrder memproses satu order: SEBELUM membuat order, ia memanggil
// inventory service untuk memastikan stok tersedia (panggilan lintas service).
//
// Param:
//   - ctx : membawa DEADLINE. Deadline ini "mengalir" ke panggilan gRPC hilir —
//     bila waktu habis, panggilan ke inventory otomatis dibatalkan (propagasi
//     deadline lintas hop, inti ketahanan microservices).
//   - sku : barang yang dipesan.
//
// Return error bila stok tak tersedia atau inventory tak bisa dihubungi.
func (o *OrderService) CreateOrder(ctx context.Context, sku string) error {
	// Panggilan gRPC NYATA ke inventory service lewat TCP.
	resp, err := o.inventory.Check(ctx, &healthpb.HealthCheckRequest{Service: "stock"})
	if err != nil {
		// status.Code mengurai kode gRPC (mis. DeadlineExceeded) dari error.
		return fmt.Errorf("inventory tak terjangkau (code=%s): %w", status.Code(err), err)
	}
	if resp.Status != healthpb.HealthCheckResponse_SERVING {
		return fmt.Errorf("stok tidak tersedia untuk %s", sku)
	}
	fmt.Printf("  ✅ order %q dibuat (inventory: %s)\n", sku, resp.Status)
	return nil
}

func main() {
	// 1) Nyalakan inventory service (server gRPC) di port TCP nyata.
	invAddr, stopInventory := startInventoryService()
	defer stopInventory()
	fmt.Println("== inventory service di", invAddr, "==")

	// 2) Order service terhubung ke inventory sebagai client.
	order := NewOrderService(invAddr)

	// 3) Order normal: deadline cukup -> sukses.
	fmt.Println("== order dengan deadline sehat ==")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := order.CreateOrder(ctx, "SKU-1"); err != nil {
		fmt.Println("  gagal:", err)
	}

	// 4) Order dengan deadline SANGAT ketat -> panggilan hilir dibatalkan.
	//    Menunjukkan propagasi deadline: order tak menunggu selamanya.
	fmt.Println("== order dengan deadline kelewat ketat ==")
	ctxKetat, cancel2 := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel2()
	time.Sleep(time.Millisecond) // pastikan deadline sudah lewat
	if err := order.CreateOrder(ctxKetat, "SKU-2"); err != nil {
		fmt.Println("  ditolak cepat:", err)
	}

	fmt.Println("== produksi: service discovery (Consul/K8s DNS), mTLS, retry+circuit breaker, tracing ==")
}
