package main

import (
	"context"
	"errors"
	"fmt"
)

// SERVERLESS (FaaS): kamu hanya menulis FUNGSI handler; platform (AWS Lambda,
// Google Cloud Functions) yang mengurus server, skala, & lifecycle.
//
// Handler = fungsi murni yang mudah di-test. Untuk deploy ke AWS Lambda, bungkus:
//
//	import "github.com/aws/aws-lambda-go/lambda"
//	func main() { lambda.Start(HandleOrder) }
//
// Lihat README untuk cara deploy.

type OrderRequest struct {
	Item string `json:"item"`
	Qty  int    `json:"qty"`
}

type OrderResponse struct {
	OrderID string `json:"order_id"`
	Total   int    `json:"total"`
	Message string `json:"message"`
}

var hargaSatuan = map[string]int{"keyboard": 250000, "mouse": 120000}

// HandleOrder adalah handler serverless — fungsi murni, tanpa server/global state.
func HandleOrder(ctx context.Context, req OrderRequest) (OrderResponse, error) {
	if req.Qty <= 0 {
		return OrderResponse{}, errors.New("qty harus > 0")
	}
	harga, ok := hargaSatuan[req.Item]
	if !ok {
		return OrderResponse{}, fmt.Errorf("item tidak dikenal: %q", req.Item)
	}
	return OrderResponse{
		OrderID: fmt.Sprintf("ord-%s-%d", req.Item, req.Qty),
		Total:   harga * req.Qty,
		Message: "order diterima",
	}, nil
}
