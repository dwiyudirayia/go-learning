// gRPC client advanced — modul 34. Jalankan (server hidup dulu):
//
//	go run ./34-grpc-advanced/client
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
	"google.golang.org/grpc/metadata"

	mathpb "go-learning/34-grpc-advanced/proto"
	"go-learning/34-grpc-advanced/service"
)

func main() {
	addr := os.Getenv("GRPC_ADDR")
	if addr == "" {
		addr = "localhost:50053"
	}
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	client := mathpb.NewMathClient(conn)

	// Semua request menyertakan token (auth interceptor).
	ctx := metadata.AppendToOutgoingContext(context.Background(), "authorization", service.ValidToken)

	// Client streaming: kirim 1..5 -> total.
	sum, _ := client.Sum(ctx)
	for i := int64(1); i <= 5; i++ {
		_ = sum.Send(&mathpb.NumberRequest{Value: i})
	}
	res, _ := sum.CloseAndRecv()
	log.Printf("Sum(1..5) = %d", res.GetSum())

	// Bidi: echo 3 pesan.
	echo, _ := client.Echo(ctx)
	go func() {
		for _, s := range []string{"halo", "dunia", "gRPC"} {
			_ = echo.Send(&mathpb.EchoMessage{Text: s})
		}
		_ = echo.CloseSend()
	}()
	for {
		m, err := echo.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Echo -> %s", m.GetText())
	}

	// Deadline: minta 500ms tapi beri 50ms.
	dctx, cancel := context.WithTimeout(ctx, 50*time.Millisecond)
	defer cancel()
	if _, err := client.SlowUnary(dctx, &mathpb.SlowRequest{DelayMs: 500}); err != nil {
		log.Printf("SlowUnary (diharapkan gagal): %v", err)
	}
}
