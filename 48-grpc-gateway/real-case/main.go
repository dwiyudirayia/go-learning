// REAL-CASE Modul 48 (gRPC-Gateway) — HTTP -> gRPC di atas TCP nyata.
//
// Versi advanced/ memakai bufconn. Versi ini: backend gRPC berjalan pada PORT
// TCP nyata, dan gateway HTTP (server terpisah) menerjemahkan REST -> gRPC serta
// memetakan kode status gRPC -> status HTTP. Berjalan lokal.
//
// Jalankan:
//
//	go run ./48-grpc-gateway/real-case
package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
)

// grpcKeHTTP memetakan kode status gRPC ke status HTTP yang setara. Ini INTI
// sebuah gateway: konsumen REST menerima semantik HTTP yang lazim (404/401/…)
// tanpa perlu tahu soal gRPC.
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

// startBackend menjalankan backend gRPC (Health) pada TCP nyata.
// Return alamat TCP + fungsi stop.
func startBackend() (addr string, stop func()) {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	srv := grpc.NewServer()
	h := health.NewServer()
	h.SetServingStatus("db", healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(srv, h)
	go func() { _ = srv.Serve(lis) }()
	return lis.Addr().String(), srv.GracefulStop
}

// newGateway membangun handler HTTP yang menjadi GATEWAY ke backend gRPC.
//
// Param:
//   - grpcClient : client gRPC ke backend.
//
// Rute "GET /health/{service}" -> memanggil RPC Check(service) -> membalas JSON
// dengan status HTTP hasil pemetaan.
func newGateway(grpcClient healthpb.HealthClient) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health/{service}", func(w http.ResponseWriter, r *http.Request) {
		svc := r.PathValue("service") // ambil segmen wildcard {service}

		// Teruskan context request (membawa deadline) ke panggilan gRPC hilir.
		resp, err := grpcClient.Check(r.Context(), &healthpb.HealthCheckRequest{Service: svc})
		if err != nil {
			// status.FromError mengurai kode gRPC dari error -> dipetakan ke HTTP.
			st, _ := status.FromError(err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(grpcKeHTTP(st.Code()))
			fmt.Fprintf(w, `{"error":%q,"grpc_code":%q}`, st.Message(), st.Code())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"service":%q,"status":%q}`, svc, resp.Status)
	})
	return mux
}

func main() {
	// 1) Backend gRPC di TCP nyata.
	backendAddr, stop := startBackend()
	defer stop()
	fmt.Println("== backend gRPC di", backendAddr, "==")

	// 2) Client gRPC untuk dipakai gateway.
	conn, err := grpc.NewClient(backendAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// 3) Gateway HTTP (server terpisah) di depan backend gRPC.
	gw := httptest.NewServer(newGateway(healthpb.NewHealthClient(conn)))
	defer gw.Close()
	fmt.Println("== gateway HTTP di", gw.URL, "==")

	// get memanggil gateway HTTP dan mengembalikan ringkasan status+body.
	get := func(path string) string {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, gw.URL+path, nil)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return "err: " + err.Error()
		}
		defer resp.Body.Close()
		b, _ := io.ReadAll(resp.Body)
		return fmt.Sprintf("%d %s", resp.StatusCode, b)
	}

	fmt.Println("== REST -> gRPC ==")
	fmt.Println("  GET /health/db    ->", get("/health/db"))    // 200 SERVING
	fmt.Println("  GET /health/cache ->", get("/health/cache")) // 404 (NotFound gRPC -> 404 HTTP)
	fmt.Println("== produksi: grpc-gateway meng-GENERATE ini dari anotasi google.api.http di .proto ==")
}
