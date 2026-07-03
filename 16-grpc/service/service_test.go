package service

import (
	"context"
	"errors"
	"io"
	"net"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"

	calcpb "go-learning/16-grpc/proto"
)

// setup menjalankan server gRPC di atas koneksi IN-MEMORY (bufconn) — tanpa port.
// Ini cara idiomatik menguji gRPC.
func setup(t *testing.T) calcpb.CalculatorClient {
	t.Helper()
	lis := bufconn.Listen(1024 * 1024)

	srv := grpc.NewServer()
	calcpb.RegisterCalculatorServer(srv, &CalculatorServer{})
	go func() { _ = srv.Serve(lis) }()
	t.Cleanup(srv.Stop)

	conn, err := grpc.NewClient(
		"passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return lis.DialContext(ctx)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	return calcpb.NewCalculatorClient(conn)
}

func TestAdd(t *testing.T) {
	client := setup(t)
	resp, err := client.Add(context.Background(), &calcpb.AddRequest{A: 7, B: 5})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	if resp.GetResult() != 12 {
		t.Errorf("Add(7,5) = %d; want 12", resp.GetResult())
	}
}

func TestFibonacci(t *testing.T) {
	client := setup(t)
	stream, err := client.Fibonacci(context.Background(), &calcpb.FibRequest{N: 10})
	if err != nil {
		t.Fatalf("Fibonacci: %v", err)
	}

	var got []int64
	for {
		resp, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			t.Fatalf("recv: %v", err)
		}
		got = append(got, resp.GetValue())
	}

	want := []int64{0, 1, 1, 2, 3, 5, 8, 13, 21, 34}
	if len(got) != len(want) {
		t.Fatalf("jumlah = %d; want %d (%v)", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("index %d = %d; want %d", i, got[i], want[i])
		}
	}
}

// Latihan 1: test Multiply.
func TestMultiply(t *testing.T) {
	client := setup(t)
	resp, err := client.Multiply(context.Background(), &calcpb.AddRequest{A: 6, B: 7})
	if err != nil {
		t.Fatalf("Multiply: %v", err)
	}
	if resp.GetResult() != 42 {
		t.Errorf("Multiply(6,7) = %d; want 42", resp.GetResult())
	}
}

// Latihan 5: Fibonacci(n<0) -> error gRPC ber-kode InvalidArgument.
func TestFibonacci_InvalidArgument(t *testing.T) {
	client := setup(t)
	stream, err := client.Fibonacci(context.Background(), &calcpb.FibRequest{N: -1})
	if err == nil {
		_, err = stream.Recv() // error muncul saat mulai membaca stream
	}
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("code = %v; want InvalidArgument", status.Code(err))
	}
}
