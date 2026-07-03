// Package service berisi implementasi CalculatorServer (logika gRPC).
// Dipisah dari main agar bisa dipakai server sungguhan DAN test (bufconn).
package service

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	calcpb "go-learning/16-grpc/proto"
)

// CalculatorServer mengimplementasikan service Calculator dari proto.
// UnimplementedCalculatorServer disematkan untuk forward-compatibility
// (bila proto menambah method, kode tetap kompilasi).
type CalculatorServer struct {
	calcpb.UnimplementedCalculatorServer
}

// Add — RPC unary: satu request, satu response.
func (s *CalculatorServer) Add(ctx context.Context, req *calcpb.AddRequest) (*calcpb.AddResponse, error) {
	return &calcpb.AddResponse{Result: req.GetA() + req.GetB()}, nil
}

// Multiply — RPC unary (latihan 1).
func (s *CalculatorServer) Multiply(ctx context.Context, req *calcpb.AddRequest) (*calcpb.AddResponse, error) {
	return &calcpb.AddResponse{Result: req.GetA() * req.GetB()}, nil
}

// Fibonacci — RPC server-streaming: kirim n bilangan Fibonacci satu per satu.
func (s *CalculatorServer) Fibonacci(req *calcpb.FibRequest, stream grpc.ServerStreamingServer[calcpb.FibResponse]) error {
	// Latihan 5: kembalikan error gRPC ber-kode saat input tak valid.
	if req.GetN() < 0 {
		return status.Errorf(codes.InvalidArgument, "n tidak boleh negatif: %d", req.GetN())
	}
	var a, b int64 = 0, 1
	for i := int32(0); i < req.GetN(); i++ {
		if err := stream.Send(&calcpb.FibResponse{Value: a}); err != nil {
			return err
		}
		a, b = b, a+b
	}
	return nil
}
