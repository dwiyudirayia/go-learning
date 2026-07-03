package main

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	"go-learning/48-grpc-gateway/gateway"
	greeterpb "go-learning/48-grpc-gateway/proto"
	"go-learning/48-grpc-gateway/service"
)

// setup: gRPC server di atas bufconn + client + gateway HTTP handler (tanpa port).
func setup(t *testing.T) http.Handler {
	t.Helper()
	lis := bufconn.Listen(1024 * 1024)
	gs := grpc.NewServer()
	greeterpb.RegisterGreeterServer(gs, &service.GreeterServer{})
	go func() { _ = gs.Serve(lis) }()
	t.Cleanup(gs.Stop)

	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) { return lis.DialContext(ctx) }),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	return gateway.New(greeterpb.NewGreeterClient(conn))
}

func TestGatewayRESTtoGRPC(t *testing.T) {
	h := setup(t)

	// Request REST/JSON -> diteruskan sebagai gRPC -> balasan JSON.
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("POST", "/v1/greet", strings.NewReader(`{"name":"Ana"}`)))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d; want 200", rec.Code)
	}
	var resp map[string]string
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp["message"] != "Halo, Ana!" {
		t.Errorf("message = %q; want 'Halo, Ana!'", resp["message"])
	}
}

func TestGatewayMapErrorGRPCtoHTTP(t *testing.T) {
	h := setup(t)

	// name kosong -> gRPC InvalidArgument -> HTTP 400 (pemetaan error lintas protokol).
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("POST", "/v1/greet", strings.NewReader(`{"name":""}`)))
	if rec.Code != http.StatusBadRequest {
		t.Errorf("name kosong = %d; want 400", rec.Code)
	}
}
