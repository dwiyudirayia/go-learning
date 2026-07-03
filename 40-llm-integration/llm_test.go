package main

import (
	"context"
	"strings"
	"testing"
)

// Semua test memakai MockChatter -> tak butuh API key / jaringan -> jalan di CI.

func TestChatMemakaiMock(t *testing.T) {
	mock := &MockChatter{Reply: "jawaban mock"}
	got, err := mock.Chat(context.Background(), "kamu asisten", []Message{
		{Role: "user", Text: "halo"},
	})
	if err != nil {
		t.Fatalf("chat: %v", err)
	}
	if got != "jawaban mock" {
		t.Errorf("reply = %q; want 'jawaban mock'", got)
	}
	// Mock menangkap apa yang dikirim -> kita bisa memverifikasi perilaku aplikasi.
	if mock.GotSystem != "kamu asisten" {
		t.Errorf("system = %q", mock.GotSystem)
	}
}

func TestRAGRetrieveDokumenRelevan(t *testing.T) {
	mock := &MockChatter{Reply: "3-5 hari kerja"}
	rag := NewRAG(mock, []Doc{
		{Title: "Refund", Content: "Refund diproses 3-5 hari kerja."},
		{Title: "Pengiriman", Content: "Pengiriman 2-4 hari."},
		{Title: "Garansi", Content: "Garansi 1 tahun."},
	})

	ans, err := rag.Answer(context.Background(), "berapa lama refund?")
	if err != nil {
		t.Fatalf("answer: %v", err)
	}
	if ans != "3-5 hari kerja" {
		t.Errorf("jawaban = %q", ans)
	}

	// Kunci RAG: system prompt yang dikirim ke LLM HARUS memuat dokumen relevan
	// (Refund), dan idealnya bukan yang tak relevan.
	if !strings.Contains(mock.GotSystem, "Refund") {
		t.Errorf("konteks tak memuat dokumen Refund:\n%s", mock.GotSystem)
	}
	if strings.Contains(mock.GotSystem, "Garansi") {
		t.Errorf("konteks seharusnya tak memuat dokumen tak relevan (Garansi)")
	}
}

func TestRAGTanpaDokumenRelevan(t *testing.T) {
	mock := &MockChatter{Reply: "harusnya tak dipakai"}
	rag := NewRAG(mock, []Doc{{Title: "Refund", Content: "Refund 3-5 hari."}})

	ans, _ := rag.Answer(context.Background(), "resep rendang padang")
	if !strings.Contains(ans, "tidak ada informasi") {
		t.Errorf("tanpa dokumen relevan harusnya menolak; dapat: %q", ans)
	}
}

// Pastikan implementasi nyata memenuhi interface (compile-time check).
var _ Chatter = (*AnthropicChatter)(nil)
var _ Chatter = (*MockChatter)(nil)
