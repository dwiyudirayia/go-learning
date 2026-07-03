package orderapp

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	"go-learning/17-studi-kasus-microservices/internal/inventoryserver"
	invpb "go-learning/17-studi-kasus-microservices/proto"
)

// wire menyalakan inventory-service (gRPC) di atas bufconn + order-service (Fiber)
// yang memakainya sebagai client. Menguji interaksi DUA service in-process.
func wire(t *testing.T) *fiber.App {
	t.Helper()
	lis := bufconn.Listen(1024 * 1024)
	gs := grpc.NewServer()
	invpb.RegisterInventoryServer(gs, inventoryserver.New())
	go func() { _ = gs.Serve(lis) }()
	t.Cleanup(gs.Stop)

	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return lis.DialContext(ctx)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	return BuildApp(invpb.NewInventoryClient(conn))
}

func req(t *testing.T, app *fiber.App, method, path, body string) *http.Response {
	t.Helper()
	var r io.Reader
	if body != "" {
		r = strings.NewReader(body)
	}
	httpReq := httptest.NewRequest(method, path, r)
	if body != "" {
		httpReq.Header.Set("Content-Type", "application/json")
	}
	resp, err := app.Test(httpReq, -1)
	if err != nil {
		t.Fatalf("%s %s: %v", method, path, err)
	}
	return resp
}

func TestGetProduct(t *testing.T) {
	app := wire(t)
	resp := req(t, app, "GET", "/products/1", "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /products/1 = %d; want 200", resp.StatusCode)
	}
	// produk tak ada -> 404 (dipetakan dari codes.NotFound)
	resp = req(t, app, "GET", "/products/99", "")
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("GET /products/99 = %d; want 404", resp.StatusCode)
	}
}

func TestOrderSukses(t *testing.T) {
	app := wire(t)
	resp := req(t, app, "POST", "/orders", `{"product_id":1,"qty":2}`)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("POST /orders = %d; want 201", resp.StatusCode)
	}
	var out map[string]any
	b, _ := io.ReadAll(resp.Body)
	_ = json.Unmarshal(b, &out)
	if out["status"] != "confirmed" {
		t.Errorf("status = %v; want confirmed", out["status"])
	}
	// Keyboard harga 250000 x 2 = 500000
	if out["total"].(float64) != 500000 {
		t.Errorf("total = %v; want 500000", out["total"])
	}
}

func TestOrderStokKurang(t *testing.T) {
	app := wire(t)
	// Monitor (id=3) stok 0 -> reservasi gagal -> 409 Conflict.
	resp := req(t, app, "POST", "/orders", `{"product_id":3,"qty":1}`)
	if resp.StatusCode != http.StatusConflict {
		t.Errorf("order stok kurang = %d; want 409", resp.StatusCode)
	}
}

func TestOrderProdukTakAda(t *testing.T) {
	app := wire(t)
	resp := req(t, app, "POST", "/orders", `{"product_id":999,"qty":1}`)
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("order produk tak ada = %d; want 404", resp.StatusCode)
	}
}

// Latihan 5: test health check.
func TestHealth(t *testing.T) {
	app := wire(t)
	resp := req(t, app, "GET", "/health", "")
	if resp.StatusCode != http.StatusOK {
		t.Errorf("GET /health = %d; want 200", resp.StatusCode)
	}
}
