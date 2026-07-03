// Modul 46 — Integrasi Layanan Pihak Ketiga (Stripe, S3, Email).
//
// POLA KUNCI: definisikan INTERFACE untuk tiap layanan eksternal. Produksi pakai
// implementasi nyata (SDK); test pakai MOCK. Aplikasimu tak terikat vendor &
// bisa diuji tanpa memanggil API sungguhan (Modul 4, 8, 40).
package main

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

// ------------------------------------------------------------------
// 1. Payment (mis. Stripe)
// ------------------------------------------------------------------
type ChargeRequest struct {
	AmountCents int
	Currency    string
	Source      string // token kartu
}
type ChargeResult struct {
	ID      string
	Success bool
}

type PaymentGateway interface {
	Charge(ctx context.Context, req ChargeRequest) (ChargeResult, error)
}

// MockGateway: implementasi palsu untuk test. (Produksi: stripe-go.)
type MockGateway struct {
	Charges []ChargeRequest // menangkap panggilan untuk assert
}

var ErrInvalidAmount = errors.New("amount harus > 0")

func (m *MockGateway) Charge(_ context.Context, req ChargeRequest) (ChargeResult, error) {
	if req.AmountCents <= 0 {
		return ChargeResult{}, ErrInvalidAmount
	}
	m.Charges = append(m.Charges, req)
	return ChargeResult{ID: fmt.Sprintf("ch_mock_%d", len(m.Charges)), Success: true}, nil
}

// ------------------------------------------------------------------
// 2. Storage (mis. AWS S3)
// ------------------------------------------------------------------
type Storage interface {
	Put(ctx context.Context, key string, data []byte) error
	Get(ctx context.Context, key string) ([]byte, error)
}

// InMemoryStorage: pengganti S3 untuk test. (Produksi: aws-sdk-go-v2/s3.)
type InMemoryStorage struct {
	mu sync.Mutex
	m  map[string][]byte
}

func NewInMemoryStorage() *InMemoryStorage { return &InMemoryStorage{m: map[string][]byte{}} }

var ErrObjectNotFound = errors.New("objek tidak ditemukan")

func (s *InMemoryStorage) Put(_ context.Context, key string, data []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	cp := make([]byte, len(data))
	copy(cp, data)
	s.m[key] = cp
	return nil
}
func (s *InMemoryStorage) Get(_ context.Context, key string) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, ok := s.m[key]
	if !ok {
		return nil, ErrObjectNotFound
	}
	return data, nil
}

// ------------------------------------------------------------------
// 3. Email (mis. SendGrid)
// ------------------------------------------------------------------
type Email struct {
	To, Subject, Body string
}
type Emailer interface {
	Send(ctx context.Context, e Email) error
}

// MockEmailer: menyimpan email yang "dikirim" untuk diperiksa test.
type MockEmailer struct {
	Sent []Email
}

func (m *MockEmailer) Send(_ context.Context, e Email) error {
	m.Sent = append(m.Sent, e)
	return nil
}

// ------------------------------------------------------------------
// OrderService: menggabungkan ketiga layanan (dependency injection).
// ------------------------------------------------------------------
type OrderService struct {
	pay     PaymentGateway
	storage Storage
	email   Emailer
}

func NewOrderService(pay PaymentGateway, storage Storage, email Emailer) *OrderService {
	return &OrderService{pay: pay, storage: storage, email: email}
}

// Checkout: bayar -> simpan struk -> kirim email. Semua lewat interface.
func (s *OrderService) Checkout(ctx context.Context, userEmail string, amountCents int) (string, error) {
	res, err := s.pay.Charge(ctx, ChargeRequest{AmountCents: amountCents, Currency: "usd", Source: "tok_visa"})
	if err != nil {
		return "", fmt.Errorf("pembayaran gagal: %w", err)
	}
	receipt := fmt.Sprintf("Struk %s: dibayar %d sen", res.ID, amountCents)
	if err := s.storage.Put(ctx, "receipts/"+res.ID+".txt", []byte(receipt)); err != nil {
		return "", err
	}
	_ = s.email.Send(ctx, Email{To: userEmail, Subject: "Pesanan diterima", Body: receipt})
	return res.ID, nil
}
