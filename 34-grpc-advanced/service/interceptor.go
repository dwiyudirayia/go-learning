package service

import (
	"context"
	"log/slog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// ValidToken: token dummy untuk demo auth interceptor.
const ValidToken = "Bearer rahasia"

// --- Interceptor UNARY ---

// LoggingUnaryInterceptor mencatat tiap panggilan unary.
func LoggingUnaryInterceptor(logger *slog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		resp, err := handler(ctx, req)
		if logger != nil {
			logger.Info("grpc unary", slog.String("method", info.FullMethod), slog.Any("err", err))
		}
		return resp, err
	}
}

// AuthUnaryInterceptor menolak panggilan tanpa token valid di metadata.
func AuthUnaryInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	if err := checkAuth(ctx); err != nil {
		return nil, err
	}
	return handler(ctx, req)
}

// --- Interceptor STREAM ---

func LoggingStreamInterceptor(logger *slog.Logger) grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		err := handler(srv, ss)
		if logger != nil {
			logger.Info("grpc stream", slog.String("method", info.FullMethod), slog.Any("err", err))
		}
		return err
	}
}

func AuthStreamInterceptor(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	if err := checkAuth(ss.Context()); err != nil {
		return err
	}
	return handler(srv, ss)
}

// checkAuth memeriksa header "authorization" di metadata gRPC.
func checkAuth(ctx context.Context) error {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Error(codes.Unauthenticated, "metadata tidak ada")
	}
	tokens := md.Get("authorization")
	if len(tokens) == 0 || tokens[0] != ValidToken {
		return status.Error(codes.Unauthenticated, "token tidak valid")
	}
	return nil
}
