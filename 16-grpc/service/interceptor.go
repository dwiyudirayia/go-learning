package service

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
)

// LoggingUnaryInterceptor (latihan 4): middleware gRPC yang mencatat tiap
// panggilan unary — analog middleware HTTP, tapi untuk gRPC.
// Pasang: grpc.NewServer(grpc.UnaryInterceptor(service.LoggingUnaryInterceptor)).
func LoggingUnaryInterceptor(
	ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (any, error) {
	start := time.Now()
	resp, err := handler(ctx, req) // teruskan ke handler asli
	log.Printf("gRPC %s selesai dalam %s (err=%v)", info.FullMethod, time.Since(start), err)
	return resp, err
}
