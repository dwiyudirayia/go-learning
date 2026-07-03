// order-service: HTTP API (Fiber) yang memanggil inventory-service via gRPC. Modul 17.
// Jalankan (inventory-service harus hidup):
//
//	go run ./17-studi-kasus-microservices/order-service
package main

import (
	"log"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"go-learning/17-studi-kasus-microservices/internal/orderapp"
	invpb "go-learning/17-studi-kasus-microservices/proto"
)

func main() {
	// Alamat inventory-service (di Docker: nama service "inventory:50052").
	invAddr := os.Getenv("INVENTORY_ADDR")
	if invAddr == "" {
		invAddr = "localhost:50052"
	}

	// Koneksi gRPC ke inventory-service.
	conn, err := grpc.NewClient(invAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("connect inventory: %v", err)
	}
	defer conn.Close()
	invClient := invpb.NewInventoryClient(conn)

	app := orderapp.BuildApp(invClient)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	log.Printf("order-service (HTTP) di http://localhost:%s -> inventory %s", port, invAddr)
	log.Fatal(app.Listen(":" + port))
}
