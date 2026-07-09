// REAL-CASE Modul 40 (LLM) — panggilan NYATA ke Claude via anthropic-sdk-go.
//
// Versi advanced/ memakai MockChatter (tanpa API key). Versi ini memanggil API
// Anthropic sungguhan. Model default: claude-opus-4-8 (paling mutakhir).
//
// Auto-skip bila ANTHROPIC_API_KEY kosong. Jalankan nyata:
//
//	export ANTHROPIC_API_KEY=sk-ant-...
//	go run ./40-llm-integration/real-case
package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
)

func main() {
	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		fmt.Println("⏭️  DILEWATI: set ANTHROPIC_API_KEY untuk memanggil Claude sungguhan.")
		fmt.Println("   export ANTHROPIC_API_KEY=sk-ant-...")
		fmt.Println("   go run ./40-llm-integration/real-case")
		return
	}
	ctx := context.Background()

	// NewClient membaca ANTHROPIC_API_KEY dari environment.
	client := anthropic.NewClient()

	// Panggilan Messages: pilih model, batas token, dan daftar pesan.
	params := anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeOpus4_8, // model default terbaru
		MaxTokens: 256,
		System:    []anthropic.TextBlockParam{{Text: "Jawab singkat dalam Bahasa Indonesia."}},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock("Sebutkan 3 kelebihan Go untuk backend, satu kalimat.")),
		},
	}

	resp, err := client.Messages.New(ctx, params)
	if err != nil {
		panic("panggilan Claude gagal: " + err.Error())
	}

	// Respons = kumpulan content block; gabungkan yang bertipe teks.
	var sb strings.Builder
	for _, block := range resp.Content {
		if t, ok := block.AsAny().(anthropic.TextBlock); ok {
			sb.WriteString(t.Text)
		}
	}
	fmt.Println("== jawaban Claude (claude-opus-4-8) ==")
	fmt.Println(sb.String())
	fmt.Printf("\n[token: input=%d output=%d]\n", resp.Usage.InputTokens, resp.Usage.OutputTokens)

	// Untuk output/inout panjang -> STREAMING (client.Messages.NewStreaming) agar
	// tak kena timeout. Untuk RAG -> ambil konteks dari VECTOR DB (pgvector/Qdrant)
	// lalu suntikkan ke prompt. Untuk hemat biaya -> prompt caching.
}
