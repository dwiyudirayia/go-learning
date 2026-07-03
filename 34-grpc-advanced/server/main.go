// gRPC server advanced — modul 34. Jalankan: go run ./34-grpc-advanced/server
package main

import (
	"log"
	"log/slog"
	"net"
	"os"

	"google.golang.org/grpc"

	mathpb "go-learning/34-grpc-advanced/proto"
	"go-learning/34-grpc-advanced/service"
)

func main() {
	addr := os.Getenv("GRPC_ADDR")
	if addr == "" {
		addr = ":50053"
	}
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Rantai interceptor untuk unary & stream.
	srv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(service.AuthUnaryInterceptor, service.LoggingUnaryInterceptor(logger)),
		grpc.ChainStreamInterceptor(service.AuthStreamInterceptor, service.LoggingStreamInterceptor(logger)),
	)
	mathpb.RegisterMathServer(srv, &service.MathServer{})

	log.Printf("Math gRPC (advanced) di %s", addr)
	log.Fatal(srv.Serve(lis))
}
