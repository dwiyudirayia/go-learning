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

// 🔍 Analogi besar: gRPC itu cara dua program "menelepon" satu sama lain lewat jaringan,
// tapi terasa seperti memanggil fungsi biasa. File .proto = "KONTRAK/MENU RESMI" yang disepakati
// klien & server (nama method, bentuk request/response). Dari kontrak itu, tool meng-GENERATE
// kode (*.pb.go) untuk kedua sisi — jadi tak ada salah paham format. Dibanding REST/JSON, gRPC
// lebih cepat (biner via HTTP/2) & tipe-nya ketat. Cocok untuk komunikasi antar-microservice.

// 🔍 Analogi: menyematkan UnimplementedCalculatorServer itu seperti memakai ADAPTOR MASA DEPAN.
// Kalau nanti kontrak menambah method baru yang belum kamu tulis, kode tetap kompilasi (method
// baru punya default "belum diimplementasi") — server lama tak langsung rusak. Aman untuk evolusi.

// CalculatorServer mengimplementasikan service Calculator dari proto.
// UnimplementedCalculatorServer disematkan untuk forward-compatibility
// (bila proto menambah method, kode tetap kompilasi).
type CalculatorServer struct {
	calcpb.UnimplementedCalculatorServer
}

// 🔍 Analogi: RPC "unary" itu seperti SMS tanya-jawab: satu pertanyaan (request), satu balasan
// (response). Sedangkan "server-streaming" (lihat Fibonacci) seperti berlangganan siaran:
// satu permintaan, lalu server mengirim BANYAK balasan mengalir satu per satu.
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
		// 🔍 Analogi: kode status gRPC (InvalidArgument, NotFound, ...) itu seperti KODE POS ERROR
		// standar — sama seperti HTTP punya 400/404/500. Klien bisa bereaksi tepat berdasar kodenya,
		// bukan menebak dari teks pesan. InvalidArgument ~ "permintaanmu salah" (mirip HTTP 400).
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
