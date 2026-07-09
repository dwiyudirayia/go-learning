// Teknik Advanced Modul 34 (gRPC advanced) — contoh runnable + penjelasan detail.
//
// CHAIN interceptor (recovery -> auth -> logging) + deadline, via bufconn &
// layanan Health bawaan (tanpa .proto). Jalankan:
//
//	go run ./34-grpc-advanced/advanced
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
	"google.golang.org/grpc/test/bufconn"
)

func main() {
	lis := bufconn.Listen(1024 * 1024)

	// =====================================================================
	// CHAIN interceptor: dieksekusi berurutan membungkus handler.
	// Urutan penting -> recovery TERLUAR agar menangkap panic dari yang lain.
	// =====================================================================
	recovery := func(ctx context.Context, req any, info *grpc.UnaryServerInfo, h grpc.UnaryHandler) (resp any, err error) {
		defer func() {
			if r := recover(); r != nil {
				err = status.Errorf(codes.Internal, "panic dipulihkan: %v", r)
			}
		}()
		return h(ctx, req)
	}
	auth := func(ctx context.Context, req any, info *grpc.UnaryServerInfo, h grpc.UnaryHandler) (any, error) {
		md, _ := metadata.FromIncomingContext(ctx)
		tokens := md.Get("authorization")
		if len(tokens) == 0 || tokens[0] != "bearer rahasia" {
			return nil, status.Error(codes.Unauthenticated, "token tidak valid/absen")
		}
		return h(ctx, req)
	}
	logging := func(ctx context.Context, req any, info *grpc.UnaryServerInfo, h grpc.UnaryHandler) (any, error) {
		mulai := time.Now()
		resp, err := h(ctx, req)
		fmt.Printf("  [log] %s -> err=%v (%s)\n", info.FullMethod, err, time.Since(mulai).Round(time.Microsecond))
		return resp, err
	}

	srv := grpc.NewServer(grpc.ChainUnaryInterceptor(recovery, auth, logging))
	hsrv := health.NewServer()
	hsrv.SetServingStatus("db", healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(srv, hsrv)
	go func() { _ = srv.Serve(lis) }()
	defer srv.Stop()

	dialer := func(ctx context.Context, _ string) (net.Conn, error) { return lis.DialContext(ctx) }
	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(dialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	client := healthpb.NewHealthClient(conn)

	// --- Panggilan 1: TANPA token -> ditolak auth interceptor ---
	fmt.Println("== 1. tanpa token ==")
	ctx1, c1 := context.WithTimeout(context.Background(), 2*time.Second) // deadline
	defer c1()
	_, err = client.Check(ctx1, &healthpb.HealthCheckRequest{Service: "db"})
	st, _ := status.FromError(err)
	fmt.Printf("  hasil: code=%s (Unauthenticated? %v)\n", st.Code(), st.Code() == codes.Unauthenticated)

	// --- Panggilan 2: DENGAN token -> lolos semua interceptor ---
	fmt.Println("== 2. dengan token ==")
	ctx2 := metadata.AppendToOutgoingContext(context.Background(), "authorization", "bearer rahasia")
	ctx2, c2 := context.WithTimeout(ctx2, 2*time.Second)
	defer c2()
	resp, err := client.Check(ctx2, &healthpb.HealthCheckRequest{Service: "db"})
	if err != nil {
		panic(err)
	}
	fmt.Println("  hasil: status service 'db' =", resp.Status)
}
