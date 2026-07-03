// Package inventoryserver mengimplementasikan Inventory gRPC service.
// Dipisah agar bisa dipakai oleh binary inventory-service DAN oleh test.
package inventoryserver

import (
	"context"
	"sync"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	invpb "go-learning/17-studi-kasus-microservices/proto"
)

type Server struct {
	invpb.UnimplementedInventoryServer
	mu       sync.Mutex
	products map[int64]*invpb.Product
}

// New membuat server dengan beberapa produk contoh.
func New() *Server {
	return &Server{products: map[int64]*invpb.Product{
		1: {Id: 1, Name: "Keyboard", Stock: 10, Price: 250000},
		2: {Id: 2, Name: "Mouse", Stock: 3, Price: 120000},
		3: {Id: 3, Name: "Monitor", Stock: 0, Price: 1500000},
	}}
}

func (s *Server) GetProduct(ctx context.Context, req *invpb.ProductID) (*invpb.Product, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.products[req.GetId()]
	if !ok {
		// Error gRPC ber-kode -> client bisa memetakannya ke status HTTP.
		return nil, status.Errorf(codes.NotFound, "produk %d tidak ditemukan", req.GetId())
	}
	return p, nil
}

func (s *Server) ReserveStock(ctx context.Context, req *invpb.ReserveRequest) (*invpb.ReserveResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	p, ok := s.products[req.GetProductId()]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "produk %d tidak ditemukan", req.GetProductId())
	}
	if req.GetQty() <= 0 {
		return nil, status.Errorf(codes.InvalidArgument, "qty harus > 0")
	}
	if p.Stock < req.GetQty() {
		return nil, status.Errorf(codes.FailedPrecondition,
			"stok tidak cukup: tersisa %d, diminta %d", p.Stock, req.GetQty())
	}
	p.Stock -= req.GetQty() // kurangi stok (reservasi)
	return &invpb.ReserveResponse{Ok: true, RemainingStock: p.Stock}, nil
}
