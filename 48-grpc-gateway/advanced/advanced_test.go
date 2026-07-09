package main

import (
	"net/http"
	"testing"

	"google.golang.org/grpc/codes"
)

// TestGrpcKeHTTP menguji pemetaan kode status gRPC -> status HTTP (inti gateway).
func TestGrpcKeHTTP(t *testing.T) {
	kasus := []struct {
		code codes.Code
		want int
	}{
		{codes.OK, http.StatusOK},                        // 200
		{codes.NotFound, http.StatusNotFound},            // 404
		{codes.InvalidArgument, http.StatusBadRequest},   // 400
		{codes.Unauthenticated, http.StatusUnauthorized}, // 401
		{codes.PermissionDenied, http.StatusForbidden},   // 403
		{codes.Internal, http.StatusInternalServerError}, // 500 (default)
		{codes.Unavailable, http.StatusInternalServerError},
	}
	for _, k := range kasus {
		if got := grpcKeHTTP(k.code); got != k.want {
			t.Errorf("grpcKeHTTP(%s) = %d, mau %d", k.code, got, k.want)
		}
	}
}
