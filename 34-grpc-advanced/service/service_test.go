package service

import (
	"context"
	"errors"
	"io"
	"net"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"

	mathpb "go-learning/34-grpc-advanced/proto"
)

func setup(t *testing.T) mathpb.MathClient {
	t.Helper()
	lis := bufconn.Listen(1024 * 1024)

	// Rantai interceptor: Auth -> Logging (kiri ke kanan urutannya).
	srv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(AuthUnaryInterceptor, LoggingUnaryInterceptor(nil)),
		grpc.ChainStreamInterceptor(AuthStreamInterceptor, LoggingStreamInterceptor(nil)),
	)
	mathpb.RegisterMathServer(srv, &MathServer{})
	go func() { _ = srv.Serve(lis) }()
	t.Cleanup(srv.Stop)

	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) { return lis.DialContext(ctx) }),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	return mathpb.NewMathClient(conn)
}

// authCtx menyisipkan token valid ke metadata request.
func authCtx(ctx context.Context) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "authorization", ValidToken)
}

func TestClientStreamingSum(t *testing.T) {
	client := setup(t)
	stream, err := client.Sum(authCtx(context.Background()))
	if err != nil {
		t.Fatalf("Sum: %v", err)
	}
	for _, v := range []int64{1, 2, 3, 4, 5} {
		if err := stream.Send(&mathpb.NumberRequest{Value: v}); err != nil {
			t.Fatalf("send: %v", err)
		}
	}
	resp, err := stream.CloseAndRecv()
	if err != nil {
		t.Fatalf("close: %v", err)
	}
	if resp.GetSum() != 15 {
		t.Errorf("sum = %d; want 15", resp.GetSum())
	}
}

func TestBidiEcho(t *testing.T) {
	client := setup(t)
	stream, err := client.Echo(authCtx(context.Background()))
	if err != nil {
		t.Fatalf("Echo: %v", err)
	}

	inputs := []string{"a", "b", "c"}
	for _, in := range inputs {
		if err := stream.Send(&mathpb.EchoMessage{Text: in}); err != nil {
			t.Fatalf("send: %v", err)
		}
		resp, err := stream.Recv()
		if err != nil {
			t.Fatalf("recv: %v", err)
		}
		if resp.GetText() != "echo: "+in {
			t.Errorf("echo = %q; want %q", resp.GetText(), "echo: "+in)
		}
	}
	_ = stream.CloseSend()
	// Setelah CloseSend, server juga menutup -> EOF.
	if _, err := stream.Recv(); !errors.Is(err, io.EOF) {
		t.Errorf("recv terakhir = %v; want EOF", err)
	}
}

func TestDeadline(t *testing.T) {
	client := setup(t)
	// Deadline 50ms, tapi server butuh 500ms -> DeadlineExceeded.
	ctx, cancel := context.WithTimeout(authCtx(context.Background()), 50*time.Millisecond)
	defer cancel()

	_, err := client.SlowUnary(ctx, &mathpb.SlowRequest{DelayMs: 500})
	if status.Code(err) != codes.DeadlineExceeded {
		t.Errorf("code = %v; want DeadlineExceeded", status.Code(err))
	}
}

func TestAuthInterceptorMenolak(t *testing.T) {
	client := setup(t)
	// Tanpa metadata auth -> Unauthenticated (dari interceptor, sebelum handler).
	_, err := client.SlowUnary(context.Background(), &mathpb.SlowRequest{DelayMs: 1})
	if status.Code(err) != codes.Unauthenticated {
		t.Errorf("code = %v; want Unauthenticated", status.Code(err))
	}
}
