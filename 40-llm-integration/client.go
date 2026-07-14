// Modul 40 — Integrasi LLM (Claude) dengan Go SDK resmi.
//
// Pola kunci: definisikan INTERFACE (Chatter) sehingga logika aplikasi tak
// terikat SDK, mudah di-MOCK saat test (tanpa API key), dan bisa ditukar.
package main

import (
	"context"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
)

// Message = satu giliran percakapan (role "user" atau "assistant").
type Message struct {
	Role string // "user" | "assistant"
	Text string
}

// 🔍 Analogi besar: kenapa bikin interface Chatter, bukan pakai SDK Anthropic langsung? Sama seperti
// modul 08 & 29 — kita menaruh "COLOKAN standar". Aplikasi cuma tahu "ada sesuatu yang bisa Chat()".
// Di produksi colokannya diisi Claude asli; saat TEST diisi MockChatter (balasan palsu) — jadi test
// jalan cepat, gratis, tanpa API key & tanpa internet. Bonus: kalau mau ganti provider LLM, cukup
// buat implementasi Chatter baru; logika RAG & aplikasi tak tersentuh sama sekali.
//
// 🔍 Analogi system prompt vs messages: "system" itu seperti ARAHAN SUTRADARA ke aktor ("kamu CS yang
// sopan, jawab singkat"); "messages" itu naskah dialog bergantian (user <-> assistant). "token" =
// potongan kata; MaxTokens membatasi panjang jawaban (seperti batas halaman). Streaming = jawaban
// muncul mengalir kata-demi-kata (UX chat enak) alih-alih menunggu semuanya jadi lalu muncul sekaligus.

// Chatter = abstraksi LLM. Aplikasi memakai interface ini, bukan SDK langsung.
type Chatter interface {
	Chat(ctx context.Context, system string, messages []Message) (string, error)
}

// ------------------------------------------------------------------
// Implementasi NYATA memakai Anthropic Go SDK
// ------------------------------------------------------------------

type AnthropicChatter struct {
	client anthropic.Client
	model  anthropic.Model
}

// NewAnthropicChatter membuat client. API key dibaca dari env ANTHROPIC_API_KEY.
// Default model: Claude Opus 4.8 (paling mumpuni; ganti bila perlu).
func NewAnthropicChatter() *AnthropicChatter {
	return &AnthropicChatter{
		client: anthropic.NewClient(), // membaca ANTHROPIC_API_KEY
		model:  anthropic.ModelClaudeOpus4_8,
	}
}

func toAnthropicMessages(msgs []Message) []anthropic.MessageParam {
	out := make([]anthropic.MessageParam, 0, len(msgs))
	for _, m := range msgs {
		if m.Role == "assistant" {
			out = append(out, anthropic.NewAssistantMessage(anthropic.NewTextBlock(m.Text)))
		} else {
			out = append(out, anthropic.NewUserMessage(anthropic.NewTextBlock(m.Text)))
		}
	}
	return out
}

// Chat mengirim percakapan ke Claude dan mengembalikan teks balasannya.
func (c *AnthropicChatter) Chat(ctx context.Context, system string, msgs []Message) (string, error) {
	params := anthropic.MessageNewParams{
		Model:     c.model,
		MaxTokens: 1024,
		Messages:  toAnthropicMessages(msgs),
	}
	// System prompt = instruksi/peran. Diletakkan di field terpisah.
	if system != "" {
		params.System = []anthropic.TextBlockParam{{Text: system}}
	}

	resp, err := c.client.Messages.New(ctx, params)
	if err != nil {
		return "", err
	}

	// Respons berisi beberapa content block; gabungkan yang bertipe teks.
	var sb strings.Builder
	for _, block := range resp.Content {
		if t, ok := block.AsAny().(anthropic.TextBlock); ok {
			sb.WriteString(t.Text)
		}
	}
	return sb.String(), nil
}

// Stream mendemokan STREAMING: token dikirim bertahap lewat callback onDelta.
// Streaming penting untuk UX chat (tak menunggu seluruh jawaban selesai) dan
// mencegah timeout pada output panjang.
func (c *AnthropicChatter) Stream(ctx context.Context, system string, msgs []Message, onDelta func(string)) (string, error) {
	params := anthropic.MessageNewParams{
		Model:     c.model,
		MaxTokens: 1024,
		Messages:  toAnthropicMessages(msgs),
	}
	if system != "" {
		params.System = []anthropic.TextBlockParam{{Text: system}}
	}

	stream := c.client.Messages.NewStreaming(ctx, params)
	message := anthropic.Message{}
	for stream.Next() {
		event := stream.Current()
		_ = message.Accumulate(event) // rakit pesan lengkap
		if delta, ok := event.AsAny().(anthropic.ContentBlockDeltaEvent); ok {
			if td, ok := delta.Delta.AsAny().(anthropic.TextDelta); ok {
				onDelta(td.Text) // kirim potongan teks ke pemanggil
			}
		}
	}
	if err := stream.Err(); err != nil {
		return "", err
	}
	var sb strings.Builder
	for _, block := range message.Content {
		if t, ok := block.AsAny().(anthropic.TextBlock); ok {
			sb.WriteString(t.Text)
		}
	}
	return sb.String(), nil
}

// ------------------------------------------------------------------
// Implementasi MOCK untuk test (tanpa API key & tanpa jaringan)
// ------------------------------------------------------------------

type MockChatter struct {
	Reply       string    // balasan yang akan dikembalikan
	GotSystem   string    // menangkap system prompt yang dikirim (untuk assert)
	GotMessages []Message // menangkap pesan yang dikirim
}

func (m *MockChatter) Chat(ctx context.Context, system string, msgs []Message) (string, error) {
	m.GotSystem = system
	m.GotMessages = msgs
	return m.Reply, nil
}
