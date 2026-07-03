// Jalankan: go run ./46-service-integrations
// Verifikasi otomatis: go test ./46-service-integrations
package main

import (
	"context"
	"fmt"
	"time"
)

func main() {
	fmt.Println("=== 46 — Service Integrations ===")
	ctx := context.Background()

	// Semua layanan eksternal diwakili MOCK -> jalan tanpa API/kredensial.
	pay := &MockGateway{}
	storage := NewInMemoryStorage()
	email := &MockEmailer{}
	svc := NewOrderService(pay, storage, email)

	fmt.Println("\n-- Checkout (bayar -> simpan struk -> email) --")
	id, err := svc.Checkout(ctx, "ana@mail.id", 25000)
	if err != nil {
		fmt.Println("error:", err)
	} else {
		fmt.Printf("checkout sukses, charge=%s\n", id)
		receipt, _ := storage.Get(ctx, "receipts/"+id+".txt")
		fmt.Printf("struk tersimpan: %s\n", receipt)
		fmt.Printf("email terkirim ke: %s (subjek: %q)\n", email.Sent[0].To, email.Sent[0].Subject)
	}

	fmt.Println("\n-- Webhook verification --")
	secret := "whsec_rahasia"
	payload := []byte(`{"event":"payment.succeeded","id":"ch_123"}`)
	now := time.Now()
	header := SignPayload(secret, now.Unix(), payload) // dibuat oleh pengirim (Stripe)

	if err := VerifyWebhook(secret, header, payload, 5*time.Minute, now); err == nil {
		fmt.Println("webhook ASLI -> terverifikasi ✔")
	}
	// Payload dipalsukan -> verifikasi gagal.
	if err := VerifyWebhook(secret, header, []byte(`{"event":"HACKED"}`), 5*time.Minute, now); err != nil {
		fmt.Printf("webhook DIPALSUKAN -> ditolak: %v\n", err)
	}
}
