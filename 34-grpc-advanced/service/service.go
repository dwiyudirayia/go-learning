// Package service mengimplementasikan Math gRPC (streaming + deadline).
package service

import (
	"context"
	"errors"
	"io"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	mathpb "go-learning/34-grpc-advanced/proto"
)

type MathServer struct {
	mathpb.UnimplementedMathServer
}

// Sum — CLIENT STREAMING: baca aliran angka sampai client selesai (io.EOF),
// lalu balas satu total dengan SendAndClose.
func (s *MathServer) Sum(stream grpc.ClientStreamingServer[mathpb.NumberRequest, mathpb.SumResponse]) error {
	var total int64
	for {
		req, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			return stream.SendAndClose(&mathpb.SumResponse{Sum: total})
		}
		if err != nil {
			return err
		}
		total += req.GetValue()
	}
}

// Echo — BIDIRECTIONAL STREAMING: untuk tiap pesan masuk, kirim balasan.
// Client & server streaming berjalan bersamaan.
func (s *MathServer) Echo(stream grpc.BidiStreamingServer[mathpb.EchoMessage, mathpb.EchoMessage]) error {
	for {
		msg, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			return nil // client menutup arah kirimnya
		}
		if err != nil {
			return err
		}
		if err := stream.Send(&mathpb.EchoMessage{Text: "echo: " + msg.GetText()}); err != nil {
			return err
		}
	}
}

// SlowUnary — menghormati DEADLINE/cancellation dari client.
func (s *MathServer) SlowUnary(ctx context.Context, req *mathpb.SlowRequest) (*mathpb.SlowResponse, error) {
	select {
	case <-time.After(time.Duration(req.GetDelayMs()) * time.Millisecond):
		return &mathpb.SlowResponse{Result: "selesai"}, nil
	case <-ctx.Done():
		// Client memberi deadline lebih pendek dari delay -> batalkan pekerjaan.
		return nil, status.Error(codes.DeadlineExceeded, "melebihi deadline client")
	}
}
