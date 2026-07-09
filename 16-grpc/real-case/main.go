// REAL-CASE Modul 16 (gRPC) — server & client di atas TCP SUNGGUHAN (bukan bufconn).
//
// bufconn (advanced/) hebat untuk test in-memory. Versi ini menjalankan server
// gRPC pada PORT TCP nyata dan client men-dial alamat itu — seperti dua service
// terpisah. Berjalan lokal tanpa infra eksternal (localhost).
//
// Jalankan:
//
//	go run ./16-grpc/real-case
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
)

func main() {
	// =====================================================================
	// SERVER: dengarkan pada TCP port nyata (127.0.0.1:0 = port bebas).
	// =====================================================================
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	srv := grpc.NewServer()
	hsrv := health.NewServer()
	hsrv.SetServingStatus("db", healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(srv, hsrv)
	go func() { _ = srv.Serve(lis) }()
	defer srv.GracefulStop()

	alamat := lis.Addr().String()
	fmt.Println("== server gRPC berjalan di", alamat, "==")

	// =====================================================================
	// CLIENT: dial alamat TCP nyata. Di produksi pakai TLS/mTLS, BUKAN insecure.
	// =====================================================================
	conn, err := grpc.NewClient(alamat, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	client := healthpb.NewHealthClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	resp, err := client.Check(ctx, &healthpb.HealthCheckRequest{Service: "db"})
	if err != nil {
		panic(err)
	}
	fmt.Println("  Check(db) via TCP ->", resp.Status)
	fmt.Println("  (produksi: TLS/mTLS, load balancing via Envoy/xDS, health check & reflection)")
}
