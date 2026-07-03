// Package service = implementasi Greeter gRPC (satu-satunya sumber logika).
package service

import (
	"context"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	greeterpb "go-learning/48-grpc-gateway/proto"
)

type GreeterServer struct {
	greeterpb.UnimplementedGreeterServer
}

func (s *GreeterServer) SayHello(ctx context.Context, req *greeterpb.HelloRequest) (*greeterpb.HelloReply, error) {
	if req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "name wajib diisi")
	}
	return &greeterpb.HelloReply{Message: fmt.Sprintf("Halo, %s!", req.GetName())}, nil
}
