// inventory-service: gRPC server (internal). Modul 17.
// Jalankan: go run ./17-studi-kasus-microservices/inventory-service
package main

import (
	"log"
	"net"
	"os"

	"google.golang.org/grpc"

	"go-learning/17-studi-kasus-microservices/internal/inventoryserver"
	invpb "go-learning/17-studi-kasus-microservices/proto"
)

func main() {
	addr := os.Getenv("GRPC_ADDR")
	if addr == "" {
		addr = ":50052"
	}
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}

	s := grpc.NewServer()
	invpb.RegisterInventoryServer(s, inventoryserver.New())

	log.Printf("inventory-service (gRPC) di %s", addr)
	log.Fatal(s.Serve(lis))
}
