// Jalankan: go run ./39-cloud-native
// Verifikasi otomatis: go test ./39-cloud-native
package main

import (
	"context"
	"fmt"
)

func main() {
	fmt.Println("=== 39 — Cloud-Native ===")

	// Demo reconcile loop (pola controller Kubernetes).
	fmt.Println("\n-- Reconcile loop (controller pattern) --")
	ps := NewPodSet(3)
	fmt.Printf("desired=3, actual=%d -> Reconcile()\n", ps.Actual())
	ps.Reconcile()
	fmt.Printf("  actual sekarang = %d, jejak: %v\n", ps.Actual(), ps.Events())

	fmt.Println("desired diubah ke 1 (scale down) -> Reconcile()")
	ps.SetDesired(1)
	ps.Reconcile()
	fmt.Printf("  actual sekarang = %d\n", ps.Actual())

	fmt.Println("Reconcile() lagi saat sudah selaras -> tak ada aksi (idempotent)")
	before := len(ps.Events())
	ps.Reconcile()
	fmt.Printf("  jumlah event tetap: %t\n", len(ps.Events()) == before)

	// Demo serverless handler.
	fmt.Println("\n-- Serverless handler --")
	resp, err := HandleOrder(context.Background(), OrderRequest{Item: "keyboard", Qty: 2})
	fmt.Printf("HandleOrder -> %+v err=%v\n", resp, err)
}
