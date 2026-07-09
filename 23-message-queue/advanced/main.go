// Teknik Advanced Modul 23 (message queue) — contoh runnable + penjelasan detail.
//
// Simulasi pure-Go semantik pengiriman pesan: at-least-once + consumer
// IDEMPOTEN (dedup), retry, dan Dead Letter Queue. Jalankan:
//
//	go run ./23-message-queue/advanced
package main

import (
	"errors"
	"fmt"
)

type Pesan struct {
	ID   string // kunci idempotensi (dedup)
	Body string
}

var errPoison = errors.New("pesan rusak permanen")

// handle memproses pesan. "poison" selalu gagal (meniru payload cacat).
func handle(m Pesan) error {
	if m.Body == "poison" {
		return errPoison
	}
	return nil
}

func main() {
	// At-least-once: broker boleh mengirim ulang -> ada DUPLIKAT (mis. m-1 dua kali)
	// dan pesan poison yang selalu gagal.
	inbox := []Pesan{
		{"m-1", "buat pesanan"},
		{"m-2", "kirim email"},
		{"m-1", "buat pesanan"}, // DUPLIKAT (redelivery) -> harus di-skip
		{"m-3", "poison"},       // selalu gagal -> harus masuk DLQ
		{"m-4", "update stok"},
	}

	// =====================================================================
	// 1. Consumer IDEMPOTEN — dedup berdasarkan ID pesan
	// =====================================================================
	// Karena pengiriman at-least-once, konsumen WAJIB aman diproses berulang.
	// Cara termudah: catat ID yang sudah sukses; abaikan bila datang lagi.
	sudahDiproses := map[string]bool{}

	const maxRetry = 2
	var dlq []Pesan // Dead Letter Queue

	for _, m := range inbox {
		if sudahDiproses[m.ID] {
			fmt.Printf("  [skip] %s duplikat -> diabaikan (idempoten)\n", m.ID)
			continue
		}

		// =================================================================
		// 2. RETRY terbatas -> DLQ. Pesan yang selalu gagal tak boleh
		//    memblokir antrean selamanya; setelah maxRetry, buang ke DLQ.
		// =================================================================
		var err error
		for attempt := 1; attempt <= maxRetry; attempt++ {
			if err = handle(m); err == nil {
				break
			}
			fmt.Printf("  [retry] %s gagal (percobaan %d/%d): %v\n", m.ID, attempt, maxRetry, err)
		}

		if err != nil {
			dlq = append(dlq, m) // menyerah -> DLQ untuk inspeksi manual
			fmt.Printf("  [dlq ] %s dipindah ke Dead Letter Queue\n", m.ID)
			continue
		}

		sudahDiproses[m.ID] = true // "ack": tandai sukses
		fmt.Printf("  [ack ] %s: %q\n", m.ID, m.Body)
	}

	fmt.Println("== ringkasan ==")
	fmt.Printf("  sukses diproses: %d pesan unik\n", len(sudahDiproses))
	fmt.Printf("  di DLQ         : %d (%v)\n", len(dlq), dlq)
}
