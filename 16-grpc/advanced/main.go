// Teknik Advanced Modul 16 (gRPC) — contoh runnable + penjelasan detail.
//
// Memakai layanan gRPC Health bawaan (tanpa perlu menulis .proto) + bufconn
// (transport in-memory). Mendemokan interceptor, metadata, deadline, dan
// pemetaan status code. Jalankan:
//
//	go run ./16-grpc/advanced
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
	lis := bufconn.Listen(1024 * 1024) // "jaringan" in-memory, tanpa port TCP

	// =====================================================================
	// 1. Server-side UNARY INTERCEPTOR — middleware untuk semua RPC
	// =====================================================================
	// Interceptor membungkus tiap panggilan: logging, auth, recovery, metrics.
	// Di sini kita baca metadata yang dikirim klien lalu teruskan ke handler.
	logInterceptor := func(ctx context.Context, req any,
		info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		md, _ := metadata.FromIncomingContext(ctx)
		fmt.Printf("  [server] %s | metadata client=%v\n", info.FullMethod, md.Get("client"))
		return handler(ctx, req)
	}

	srv := grpc.NewServer(grpc.UnaryInterceptor(logInterceptor))

	// Health service bawaan gRPC: kita tandai service "db" SERVING.
	hsrv := health.NewServer()
	hsrv.SetServingStatus("db", healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(srv, hsrv)

	go func() { _ = srv.Serve(lis) }()
	defer srv.Stop()

	// =====================================================================
	// 2. Client-side interceptor + dial via bufconn
	// =====================================================================
	// WithContextDialer mengarahkan koneksi ke listener bufconn (bukan DNS).
	// Interceptor klien menyisipkan metadata "client" ke setiap request.
	dialer := func(ctx context.Context, _ string) (net.Conn, error) {
		return lis.DialContext(ctx)
	}
	cliInterceptor := func(ctx context.Context, method string, req, reply any,
		cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		ctx = metadata.AppendToOutgoingContext(ctx, "client", "demo-cli")
		return invoker(ctx, method, req, reply, cc, opts...)
	}
	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(dialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()), // aman: bufconn lokal
		grpc.WithUnaryInterceptor(cliInterceptor),
	)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	client := healthpb.NewHealthClient(conn)

	// =====================================================================
	// 3. Panggilan dengan DEADLINE (selalu batasi waktu RPC)
	// =====================================================================
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	fmt.Println("== 1&3. RPC sukses dengan metadata + deadline ==")
	resp, err := client.Check(ctx, &healthpb.HealthCheckRequest{Service: "db"})
	if err != nil {
		panic(err)
	}
	fmt.Println("  status service 'db':", resp.Status) // SERVING

	// =====================================================================
	// 4. Pemetaan STATUS CODE gRPC pada error
	// =====================================================================
	// Meminta service yang tak terdaftar -> server balas codes.NotFound.
	// status.FromError mengurai kode + pesan terstruktur (bukan sekadar string).
	fmt.Println("== 4. status code pada error ==")
	_, err = client.Check(ctx, &healthpb.HealthCheckRequest{Service: "cache"})
	st, _ := status.FromError(err)
	fmt.Printf("  service 'cache' -> code=%s (NotFound? %v)\n",
		st.Code(), st.Code() == codes.NotFound)
}
