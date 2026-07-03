// Jalankan:
//
//	go run ./40-llm-integration              # mode DEMO (mock, tanpa API key)
//	ANTHROPIC_API_KEY=sk-... go run ./40-llm-integration   # panggil Claude sungguhan
//
// Verifikasi otomatis: go test ./40-llm-integration
package main

import (
	"context"
	"fmt"
	"os"
)

func main() {
	fmt.Println("=== 40 — Integrasi LLM (Claude) ===")
	ctx := context.Background()

	// Pilih implementasi: Claude sungguhan bila ada API key, mock bila tidak.
	var chatter Chatter
	if os.Getenv("ANTHROPIC_API_KEY") != "" {
		fmt.Println("(memakai Claude sungguhan via ANTHROPIC_API_KEY)")
		chatter = NewAnthropicChatter()
	} else {
		fmt.Println("(tak ada ANTHROPIC_API_KEY -> memakai MOCK; set env untuk memanggil Claude asli)")
		chatter = &MockChatter{Reply: "Halo! (ini balasan mock — set ANTHROPIC_API_KEY untuk jawaban asli)"}
	}

	// 1. Chat sederhana.
	reply, err := chatter.Chat(ctx, "Kamu asisten yang ramah dan ringkas.", []Message{
		{Role: "user", Text: "Sebutkan 3 kelebihan Go untuk backend."},
	})
	if err != nil {
		fmt.Println("error:", err)
	} else {
		fmt.Printf("\n[chat] %s\n", reply)
	}

	// 2. RAG: jawab dari basis pengetahuan.
	kb := []Doc{
		{Title: "Refund", Content: "Refund diproses 3-5 hari kerja ke metode pembayaran asal."},
		{Title: "Pengiriman", Content: "Pengiriman standar 2-4 hari, ekspres 1 hari."},
		{Title: "Garansi", Content: "Semua produk bergaransi 1 tahun."},
	}
	rag := NewRAG(chatter, kb)
	ans, _ := rag.Answer(ctx, "Berapa lama proses refund?")
	fmt.Printf("\n[RAG] pertanyaan: 'Berapa lama proses refund?'\n[RAG] jawaban : %s\n", ans)

	// 3. Streaming (hanya jika Claude sungguhan; mock tak mengimplementasikan Stream).
	if ac, ok := chatter.(*AnthropicChatter); ok {
		fmt.Print("\n[stream] ")
		_, _ = ac.Stream(ctx, "", []Message{{Role: "user", Text: "Tulis satu kalimat motivasi."}},
			func(chunk string) { fmt.Print(chunk) })
		fmt.Println()
	}
}
