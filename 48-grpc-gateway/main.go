// Modul 48 — gRPC-Gateway: satu service gRPC diekspos JUGA sebagai REST.
//
// Jalankan: go run ./48-grpc-gateway
//
//	# gRPC di :50054, REST gateway di :8081
//	curl -X POST localhost:8081/v1/greet -d '{"name":"Ana"}'   # -> {"message":"Halo, Ana!"}
//
// Verifikasi otomatis: go test ./48-grpc-gateway
package main

import (
	"log"
	"net"
	"net/http"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"go-learning/48-grpc-gateway/gateway"
	greeterpb "go-learning/48-grpc-gateway/proto"
	"go-learning/48-grpc-gateway/service"
)

// 🔍 Analogi besar gRPC-Gateway: kamu punya satu DAPUR (service gRPC) berisi logika. Masalahnya,
// tamu internal (microservice lain) suka bicara "bahasa gRPC" yang cepat, tapi tamu luar (browser,
// aplikasi mobile, curl) maunya "bahasa REST/JSON" yang umum. Alih-alih memasak dua kali, kamu pasang
// PENERJEMAH di depan (gateway): ia menerima pesanan REST dari luar, menerjemahkannya jadi panggilan
// gRPC ke dapur, lalu menerjemahkan balik jawabannya ke JSON. SATU sumber logika, DUA pintu masuk.
// Hebatnya, penerjemah ini bisa di-generate otomatis dari file .proto yang sama — tak ditulis tangan.
func main() {
	// 1. Jalankan server gRPC (sumber logika tunggal).
	lis, err := net.Listen("tcp", ":50054")
	if err != nil {
		log.Fatal(err)
	}
	gs := grpc.NewServer()
	greeterpb.RegisterGreeterServer(gs, &service.GreeterServer{})
	go func() { _ = gs.Serve(lis) }()
	log.Println("gRPC di :50054")

	// 2. Buat gRPC client yang dipakai gateway.
	conn, err := grpc.NewClient("localhost:50054", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// 3. Jalankan HTTP gateway (menerjemahkan REST -> gRPC).
	srv := &http.Server{
		Addr:              ":8081",
		Handler:           gateway.New(greeterpb.NewGreeterClient(conn)),
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Println("REST gateway di :8081 -> POST /v1/greet")
	log.Fatal(srv.ListenAndServe())
}
