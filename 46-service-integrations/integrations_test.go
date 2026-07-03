package main

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestCheckoutFlow(t *testing.T) {
	pay := &MockGateway{}
	storage := NewInMemoryStorage()
	email := &MockEmailer{}
	svc := NewOrderService(pay, storage, email)

	id, err := svc.Checkout(context.Background(), "ana@mail.id", 25000)
	if err != nil {
		t.Fatalf("checkout: %v", err)
	}
	// Semua efek samping terverifikasi via mock — tanpa API sungguhan.
	if len(pay.Charges) != 1 || pay.Charges[0].AmountCents != 25000 {
		t.Errorf("charge tak sesuai: %+v", pay.Charges)
	}
	if _, err := storage.Get(context.Background(), "receipts/"+id+".txt"); err != nil {
		t.Errorf("struk tak tersimpan: %v", err)
	}
	if len(email.Sent) != 1 || email.Sent[0].To != "ana@mail.id" {
		t.Errorf("email tak terkirim: %+v", email.Sent)
	}
}

func TestCheckoutBayarGagal(t *testing.T) {
	svc := NewOrderService(&MockGateway{}, NewInMemoryStorage(), &MockEmailer{})
	// Amount 0 -> payment gagal -> checkout gagal.
	if _, err := svc.Checkout(context.Background(), "a@b.c", 0); err == nil {
		t.Error("amount 0 harusnya gagal")
	}
}

func TestStorage(t *testing.T) {
	s := NewInMemoryStorage()
	_ = s.Put(context.Background(), "k", []byte("data"))
	got, _ := s.Get(context.Background(), "k")
	if string(got) != "data" {
		t.Errorf("get = %q; want data", got)
	}
	if _, err := s.Get(context.Background(), "tidakada"); !errors.Is(err, ErrObjectNotFound) {
		t.Error("key tak ada harusnya ErrObjectNotFound")
	}
}

func TestWebhookVerification(t *testing.T) {
	secret := "whsec_x"
	payload := []byte(`{"amount":100}`)
	now := time.Unix(1_700_000_000, 0)
	header := SignPayload(secret, now.Unix(), payload)

	// Valid.
	if err := VerifyWebhook(secret, header, payload, 5*time.Minute, now); err != nil {
		t.Errorf("webhook asli ditolak: %v", err)
	}
	// Payload dipalsukan.
	if err := VerifyWebhook(secret, header, []byte(`{"amount":999999}`), 5*time.Minute, now); !errors.Is(err, ErrBadSignature) {
		t.Error("payload palsu harus ditolak")
	}
	// Secret salah.
	if err := VerifyWebhook("secret-salah", header, payload, 5*time.Minute, now); !errors.Is(err, ErrBadSignature) {
		t.Error("secret salah harus ditolak")
	}
	// Terlalu lama (replay).
	later := now.Add(10 * time.Minute)
	if err := VerifyWebhook(secret, header, payload, 5*time.Minute, later); !errors.Is(err, ErrTooOld) {
		t.Error("webhook lama harus ditolak (anti-replay)")
	}
}
