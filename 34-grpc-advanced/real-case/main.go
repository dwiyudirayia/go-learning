// REAL-CASE Modul 34 (gRPC advanced) — CHAIN interceptor + auth di atas TCP nyata.
//
// Versi advanced/ memakai bufconn. Versi ini menjalankan server gRPC pada PORT
// TCP nyata dengan RANTAI interceptor (recovery -> auth -> logging), lalu client
// men-dial lewat TCP dengan metadata token. Berjalan lokal.
//
// Jalankan:
//
//	go run ./34-grpc-advanced/real-case
package main

import (
	"context"
	"fmt"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// recoveryInterceptor = lapisan TERLUAR: menangkap panic dari handler/interceptor
// lain agar server tak crash, mengubahnya menjadi error gRPC Internal.
//
// Tanda tangan interceptor unary server (wajib persis begini):
//
//	func(ctx, req, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp, err)
func recoveryInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = status.Errorf(codes.Internal, "panic dipulihkan: %v", r)
		}
	}()
	return handler(ctx, req)
}

// authInterceptor memeriksa token pada metadata "authorization". Bila absen/salah,
// tolak dengan codes.Unauthenticated SEBELUM handler dijalankan.
func authInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	// metadata.FromIncomingContext mengambil header yang dikirim client.
	md, _ := metadata.FromIncomingContext(ctx)
	tokens := md.Get("authorization")
	if len(tokens) == 0 || tokens[0] != "bearer rahasia" {
		return nil, status.Error(codes.Unauthenticated, "token tidak valid/absen")
	}
	return handler(ctx, req)
}

// loggingInterceptor mencatat method & durasi tiap RPC (lapisan TERDALAM di rantai).
func loggingInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	mulai := time.Now()
	resp, err := handler(ctx, req)
	fmt.Printf("  [log] %s -> err=%v (%s)\n", info.FullMethod, err, time.Since(mulai).Round(time.Microsecond))
	return resp, err
}

// startServer menjalankan server gRPC ber-rantai interceptor pada TCP nyata.
//
// grpc.ChainUnaryInterceptor(a, b, c) menjalankan a -> b -> c -> handler.
// Urutan penting: recovery paling luar agar menangkap panic dari yang lain.
//
// Return: alamat TCP + fungsi stop.
func startServer() (addr string, stop func()) {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	srv := grpc.NewServer(grpc.ChainUnaryInterceptor(
		recoveryInterceptor,
		authInterceptor,
		loggingInterceptor,
	))
	h := health.NewServer()
	h.SetServingStatus("db", healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(srv, h)
	go func() { _ = srv.Serve(lis) }()
	return lis.Addr().String(), srv.GracefulStop
}

// dial membuat client gRPC ke alamat TCP (insecure hanya untuk demo lokal).
func dial(addr string) healthpb.HealthClient {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	return healthpb.NewHealthClient(conn)
}

func main() {
	addr, stop := startServer()
	defer stop()
	fmt.Println("== server gRPC (chain interceptor) di", addr, "==")
	client := dial(addr)

	// 1) TANPA token -> ditolak authInterceptor sebelum mencapai handler.
	fmt.Println("== panggilan tanpa token ==")
	ctx1, c1 := context.WithTimeout(context.Background(), 2*time.Second)
	defer c1()
	_, err := client.Check(ctx1, &healthpb.HealthCheckRequest{Service: "db"})
	fmt.Printf("  hasil: code=%s\n", status.Code(err))

	// 2) DENGAN token -> lolos semua interceptor.
	// metadata.AppendToOutgoingContext menyisipkan header ke request keluar.
	fmt.Println("== panggilan dengan token ==")
	ctx2 := metadata.AppendToOutgoingContext(context.Background(), "authorization", "bearer rahasia")
	ctx2, c2 := context.WithTimeout(ctx2, 2*time.Second)
	defer c2()
	resp, err := client.Check(ctx2, &healthpb.HealthCheckRequest{Service: "db"})
	if err != nil {
		panic(err)
	}
	fmt.Println("  status service 'db':", resp.Status)

	fmt.Println("== produksi: retry via service config, LB Envoy/xDS, keepalive, deadline propagation ==")
}
