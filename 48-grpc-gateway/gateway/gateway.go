// Package gateway menerjemahkan REST/JSON <-> gRPC. Klien REST bicara HTTP,
// gateway meneruskannya sebagai panggilan gRPC ke service yang sama.
//
// Modul ini menulis gateway MANUAL agar konsepnya jelas & mudah di-test. Di
// produksi, gunakan protoc-gen-grpc-gateway yang MEN-GENERATE gateway ini dari
// anotasi di .proto (lihat README).
package gateway

import (
	"encoding/json"
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	greeterpb "go-learning/48-grpc-gateway/proto"
)

// New membuat HTTP handler yang memakai gRPC client (in-process atau jaringan).
func New(client greeterpb.GreeterClient) http.Handler {
	mux := http.NewServeMux()

	// REST: POST /v1/greet {"name":"Ana"} -> gRPC SayHello -> JSON.
	mux.HandleFunc("POST /v1/greet", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Name string `json:"name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "JSON tidak valid"})
			return
		}

		// Terjemahkan HTTP -> panggilan gRPC.
		reply, err := client.SayHello(r.Context(), &greeterpb.HelloRequest{Name: req.Name})
		if err != nil {
			writeJSON(w, grpcToHTTP(err), map[string]string{"error": status.Convert(err).Message()})
			return
		}

		// Terjemahkan balasan gRPC -> JSON HTTP.
		writeJSON(w, http.StatusOK, map[string]string{"message": reply.GetMessage()})
	})

	return mux
}

// grpcToHTTP memetakan kode gRPC -> status HTTP (Modul 17).
func grpcToHTTP(err error) int {
	switch status.Code(err) {
	case codes.InvalidArgument:
		return http.StatusBadRequest
	case codes.NotFound:
		return http.StatusNotFound
	case codes.Unauthenticated:
		return http.StatusUnauthorized
	default:
		return http.StatusInternalServerError
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
