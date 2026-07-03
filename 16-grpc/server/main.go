// gRPC server untuk Calculator — modul 16.
// Jalankan: go run ./16-grpc/server   (dengar di :50051)
package main

import (
	"log"
	"net"
	"os"

	"google.golang.org/grpc"

	calcpb "go-learning/16-grpc/proto"
	"go-learning/16-grpc/service"
)

func main() {
	addr := os.Getenv("GRPC_ADDR")
	if addr == "" {
		addr = ":50051"
	}

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("gagal listen: %v", err)
	}

	// Pasang interceptor logging (latihan 4).
	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(service.LoggingUnaryInterceptor))
	// Daftarkan implementasi service ke server gRPC.
	calcpb.RegisterCalculatorServer(grpcServer, &service.CalculatorServer{})

	log.Printf("gRPC server berjalan di %s", addr)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("gagal serve: %v", err)
	}
}
