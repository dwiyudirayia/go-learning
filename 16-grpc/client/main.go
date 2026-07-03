// gRPC client untuk Calculator — modul 16.
// Jalankan (server harus hidup dulu): go run ./16-grpc/client
package main

import (
	"context"
	"errors"
	"io"
	"log"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	calcpb "go-learning/16-grpc/proto"
)

func main() {
	addr := os.Getenv("GRPC_ADDR")
	if addr == "" {
		addr = "localhost:50051"
	}

	// Buat koneksi (insecure untuk contoh lokal; produksi pakai TLS).
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("gagal connect: %v", err)
	}
	defer conn.Close()

	client := calcpb.NewCalculatorClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// --- Unary ---
	addResp, err := client.Add(ctx, &calcpb.AddRequest{A: 7, B: 5})
	if err != nil {
		log.Fatalf("Add gagal: %v", err)
	}
	log.Printf("Add(7, 5) = %d", addResp.GetResult())

	// --- Server streaming ---
	stream, err := client.Fibonacci(ctx, &calcpb.FibRequest{N: 10})
	if err != nil {
		log.Fatalf("Fibonacci gagal: %v", err)
	}
	log.Print("Fibonacci(10) = ")
	for {
		resp, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break // stream selesai
		}
		if err != nil {
			log.Fatalf("recv gagal: %v", err)
		}
		log.Printf("  %d", resp.GetValue())
	}
}
