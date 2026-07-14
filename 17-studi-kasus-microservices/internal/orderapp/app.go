// Package orderapp membangun HTTP API (Fiber) untuk order-service.
// Ia adalah CLIENT gRPC ke inventory-service — mendemokan komunikasi antar service.
package orderapp

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	invpb "go-learning/17-studi-kasus-microservices/proto"
)

// 🔍 Analogi besar: microservices itu seperti FRANCHISE dengan banyak toko spesialis, bukan
// satu supermarket raksasa. order-service (toko pemesanan) tak menyimpan stok sendiri; ia
// "menelepon" (gRPC) inventory-service (gudang) untuk cek & pesan stok. Untung: tiap toko bisa
// dikembangkan/di-scale/di-deploy terpisah. Rugi: kini ada JARINGAN di antara mereka — bisa
// lambat/putus — makanya butuh timeout, retry, dan pemetaan error (lihat fungsi di bawah).

// BuildApp menerima InventoryClient (interface hasil generate) — sehingga di
// test bisa diisi client bufconn, di produksi diisi koneksi gRPC sungguhan.
func BuildApp(inv invpb.InventoryClient) *fiber.App {
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if fe, ok := err.(*fiber.Error); ok {
				code = fe.Code
			}
			return c.Status(code).JSON(fiber.Map{"error": err.Error()})
		},
	})

	// Latihan 5: health check — untuk liveness probe / load balancer.
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	// GET /products/:id -> teruskan ke inventory via gRPC.
	app.Get("/products/:id", func(c *fiber.Ctx) error {
		id, err := c.ParamsInt("id")
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "id harus angka")
		}
		p, err := inv.GetProduct(c.Context(), &invpb.ProductID{Id: int64(id)})
		if err != nil {
			return grpcToHTTP(err)
		}
		return c.JSON(p)
	})

	// POST /orders {product_id, qty} -> reservasi stok lewat inventory, buat order.
	app.Post("/orders", func(c *fiber.Ctx) error {
		var req struct {
			ProductID int64 `json:"product_id"`
			Qty       int64 `json:"qty"`
		}
		if err := c.BodyParser(&req); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "body tidak valid")
		}

		// Panggil inventory-service (gRPC) dengan timeout + retry (latihan 3).
		res, err := reserveWithRetry(c.Context(), inv, &invpb.ReserveRequest{
			ProductId: req.ProductID,
			Qty:       req.Qty,
		})
		if err != nil {
			return grpcToHTTP(err) // stok kurang -> 409, produk tak ada -> 404, dst
		}

		// Ambil detail produk untuk hitung total.
		p, err := inv.GetProduct(c.Context(), &invpb.ProductID{Id: req.ProductID})
		if err != nil {
			return grpcToHTTP(err)
		}

		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"status":          "confirmed",
			"product":         p.GetName(),
			"qty":             req.Qty,
			"total":           p.GetPrice() * req.Qty,
			"remaining_stock": res.GetRemainingStock(),
		})
	})

	return app
}

// reserveWithRetry (latihan 3) memanggil ReserveStock dengan batas waktu 2 detik
// dan mencoba ulang (maks 3x) HANYA bila error transien (codes.Unavailable),
// mis. inventory-service sedang restart. Error bisnis (stok kurang) TIDAK diretry.
// 🔍 Analogi: retry itu seperti MENELEPON ULANG saat nada sibuk. Tapi hati-hati: kita hanya
// menelepon ulang bila "salurannya" yang bermasalah (codes.Unavailable ~ service sedang restart).
// Kalau jawabannya "stok memang habis" (error bisnis), menelepon ulang percuma — malah bahaya
// (bisa memesan dobel). "backoff" = jeda sejenak antar percobaan agar tak membanjiri server.
func reserveWithRetry(parent context.Context, inv invpb.InventoryClient, req *invpb.ReserveRequest) (*invpb.ReserveResponse, error) {
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		ctx, cancel := context.WithTimeout(parent, 2*time.Second)
		res, err := inv.ReserveStock(ctx, req)
		cancel()
		if err == nil {
			return res, nil
		}
		lastErr = err
		if status.Code(err) != codes.Unavailable {
			return nil, err // bukan transien -> jangan retry
		}
		time.Sleep(50 * time.Millisecond) // backoff singkat sebelum coba lagi
	}
	return nil, lastErr
}

// 🔍 Analogi: grpcToHTTP itu PENERJEMAH ANTAR-BAHASA-ERROR. Gudang bicara "bahasa gRPC"
// (codes.NotFound), tapi pelanggan web mengerti "bahasa HTTP" (404). Fungsi ini menerjemahkan
// tiap kode agar pelanggan luar menerima status yang benar & bermakna, bukan 500 generik.
// grpcToHTTP memetakan kode error gRPC -> status HTTP yang sesuai.
func grpcToHTTP(err error) error {
	switch status.Code(err) {
	case codes.NotFound:
		return fiber.NewError(fiber.StatusNotFound, status.Convert(err).Message())
	case codes.InvalidArgument:
		return fiber.NewError(fiber.StatusBadRequest, status.Convert(err).Message())
	case codes.FailedPrecondition:
		return fiber.NewError(fiber.StatusConflict, status.Convert(err).Message())
	default:
		return fiber.NewError(fiber.StatusBadGateway, "inventory service error: "+status.Convert(err).Message())
	}
}
