// Teknik Advanced Modul 48 (gRPC-Gateway) — contoh runnable + penjelasan detail.
//
// Gateway REST -> gRPC: handler HTTP memanggil layanan gRPC (Health via bufconn)
// dan MEMETAKAN kode status gRPC -> status HTTP. Jalankan:
//
//	go run ./48-grpc-gateway/advanced
package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

// ===========================================================================
// Pemetaan kode gRPC -> status HTTP (inti sebuah gateway). Konsumen REST tak
// perlu tahu soal gRPC; mereka menerima semantik HTTP yang lazim.
//
// 🔍 Analogi: gateway itu seperti RESEPSIONIS HOTEL BILINGUAL — tamu bicara
// bahasa umum (REST/JSON), staf dapur bicara bahasa internal (gRPC/protobuf).
// Resepsionis menerjemahkan dua arah, termasuk nada jawabannya: "kamar tak
// ada" versi dapur (codes.NotFound) disampaikan ke tamu dengan sopan santun
// yang mereka pahami (HTTP 404).
// ===========================================================================
func grpcKeHTTP(c codes.Code) int {
	switch c {
	case codes.OK:
		return http.StatusOK
	case codes.NotFound:
		return http.StatusNotFound
	case codes.InvalidArgument:
		return http.StatusBadRequest
	case codes.Unauthenticated:
		return http.StatusUnauthorized
	case codes.PermissionDenied:
		return http.StatusForbidden
	default:
		return http.StatusInternalServerError
	}
}

func main() {
	// --- Backend gRPC (Health) di atas bufconn ---
	lis := bufconn.Listen(1024 * 1024)
	srv := grpc.NewServer()
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
	grpcClient := healthpb.NewHealthClient(conn)

	// =====================================================================
	// GATEWAY: handler HTTP menerjemahkan REST -> panggilan gRPC -> respons JSON.
	// Path /health/{service} dipetakan ke RPC Health.Check(service).
	// =====================================================================
	gateway := http.NewServeMux()
	gateway.HandleFunc("GET /health/{service}", func(w http.ResponseWriter, r *http.Request) {
		svc := r.PathValue("service")
		resp, err := grpcClient.Check(r.Context(), &healthpb.HealthCheckRequest{Service: svc})
		if err != nil {
			st, _ := status.FromError(err)
			w.WriteHeader(grpcKeHTTP(st.Code())) // kode gRPC -> status HTTP
			fmt.Fprintf(w, `{"error":%q,"grpc_code":%q}`, st.Message(), st.Code())
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"service":%q,"status":%q}`, svc, resp.Status)
	})

	httpSrv := httptest.NewServer(gateway)
	defer httpSrv.Close()

	get := func(path string) string {
		resp, err := http.Get(httpSrv.URL + path)
		if err != nil {
			return "err: " + err.Error()
		}
		defer resp.Body.Close()
		b, _ := io.ReadAll(resp.Body)
		return fmt.Sprintf("%d %s", resp.StatusCode, b)
	}

	fmt.Println("== REST -> gRPC gateway ==")
	fmt.Println("  GET /health/db    ->", get("/health/db"))    // 200 (SERVING)
	fmt.Println("  GET /health/cache ->", get("/health/cache")) // 404 (NotFound gRPC -> 404 HTTP)

	// Catatan: gateway sungguhan (grpc-gateway) meng-GENERATE kode ini dari
	// anotasi google.api.http di .proto; streaming server-side dipetakan ke
	// chunked/SSE, sedangkan bidi-streaming tak natural di REST.
}
