// Teknik Advanced Modul 40 (LLM integration) — contoh runnable + penjelasan detail.
//
// Pola inti aplikasi LLM: interface + MOCK (uji tanpa API key/biaya), TOOL-USE
// loop, dan RAG (retrieval). Jalankan:
//
//	go run ./40-llm-integration/advanced
//
// Di produksi, Chatter diimplementasikan oleh anthropic-sdk-go:
//
//	client := anthropic.NewClient() // baca ANTHROPIC_API_KEY dari env
//	resp, _ := client.Messages.New(ctx, anthropic.MessageNewParams{
//	    Model:     anthropic.Model("claude-opus-4-8"), // model default terbaru
//	    MaxTokens: 1024,
//	    Thinking:  anthropic.ThinkingConfigParamOfAdaptive(), // adaptive thinking
//	    Messages:  msgs,
//	})
//
// Untuk input/output panjang, pakai STREAMING (client.Messages.NewStreaming) +
// helper .Accumulate/finalMessage agar tak kena timeout.
package main

import (
	"context"
	"fmt"
	"strings"
)

// ===========================================================================
// 1. Abstraksi Chatter + MockChatter (uji tanpa jaringan/biaya)
// ===========================================================================
// Kode aplikasi bergantung pada interface, bukan SDK konkret. Saat test, isi
// dengan MockChatter yang membalas skrip -> deterministik, gratis, cepat.
//
// 🔍 Analogi: MockChatter itu seperti LAWAN MAIN LATIHAN DRAMA yang membaca
// naskah — dialognya sudah pasti, gratis, dan bisa diulang kapan pun. Aktor
// sungguhan (API Claude) baru dipakai saat pentas (produksi); latihan tiap
// hari dengannya mahal dan jawabannya berubah-ubah.
type Chatter interface {
	Chat(ctx context.Context, prompt string) (string, error)
}

type MockChatter struct {
	balasan []string
	i       int
}

func (m *MockChatter) Chat(_ context.Context, _ string) (string, error) {
	if m.i >= len(m.balasan) {
		return "", fmt.Errorf("mock kehabisan balasan")
	}
	r := m.balasan[m.i]
	m.i++
	return r, nil
}

// ===========================================================================
// 2. TOOL-USE loop — model minta tool, kode eksekusi, hasil dikembalikan
// ===========================================================================
// Alur: kirim prompt -> model balas entah JAWABAN atau permintaan TOOL. Bila
// tool, kode menjalankan fungsi lokal lalu mengumpan hasilnya kembali, ulangi
// sampai model memberi jawaban final. Batasi iterasi agar tak loop selamanya.
//
// 🔍 Analogi: tool-use itu seperti DOKTER DAN LABORATORIUM — dokter (model)
// tak bisa mengambil darah sendiri; ia menulis surat rujukan lab ("TOOL:
// cek darah"), petugas lab (kode kita) menjalankannya dan menyerahkan hasil,
// lalu dokter membaca hasilnya untuk menegakkan diagnosis (jawaban final).
// Batas iterasi = mencegah dokter meminta tes baru tanpa henti.
func toolUseLoop(ctx context.Context, llm Chatter, tools map[string]func(string) string) string {
	prompt := "Cuaca di Jakarta?"
	for langkah := 0; langkah < 5; langkah++ { // batas iterasi (guard)
		jawab, _ := llm.Chat(ctx, prompt)
		if strings.HasPrefix(jawab, "TOOL:") {
			// format skrip: "TOOL:namaTool:argumen"
			bagian := strings.SplitN(strings.TrimPrefix(jawab, "TOOL:"), ":", 2)
			nama, arg := bagian[0], bagian[1]
			hasil := tools[nama](arg)
			fmt.Printf("  [tool] %s(%q) -> %q\n", nama, arg, hasil)
			prompt = "hasil tool: " + hasil // umpan balik ke model
			continue
		}
		return jawab // jawaban final
	}
	return "(melebihi batas iterasi tool)"
}

// ===========================================================================
// 3. RAG — Retrieval Augmented Generation (grounding dengan dokumen)
// ===========================================================================
// Ambil potongan dokumen paling relevan (retrieval) lalu SUNTIKKAN ke prompt
// sebagai konteks. Kualitas retrieval lebih menentukan daripada ukuran prompt.
//
// 🔍 Analogi: RAG itu seperti UJIAN OPEN-BOOK — daripada menyuruh siswa
// (model) menjawab dari hafalan (bisa ngarang/halusinasi), pengawas
// (retrieval) membukakan halaman buku yang TEPAT di meja. Kuncinya bukan
// membawa seluruh perpustakaan (prompt raksasa), tapi membuka halaman yang
// benar.
func retrieve(korpus []string, query string) string {
	for _, dok := range korpus { // di produksi: embedding + vektor similarity
		if strings.Contains(strings.ToLower(dok), strings.ToLower(query)) {
			return dok
		}
	}
	return ""
}

func main() {
	ctx := context.Background()

	// --- Tool-use ---
	fmt.Println("== 2. tool-use loop ==")
	llm := &MockChatter{balasan: []string{
		"TOOL:cuaca:Jakarta",        // langkah 1: model minta tool
		"Cuaca Jakarta cerah 32°C.", // langkah 2: jawaban final setelah dapat hasil tool
	}}
	tools := map[string]func(string) string{
		"cuaca": func(kota string) string { return kota + ": cerah, 32C" },
	}
	fmt.Println("  jawaban:", toolUseLoop(ctx, llm, tools))

	// --- RAG ---
	fmt.Println("== 3. RAG ==")
	korpus := []string{
		"Kebijakan refund: dana kembali dalam 7 hari kerja.",
		"Jam operasional: Senin-Jumat 09.00-17.00.",
	}
	ctxDoc := retrieve(korpus, "refund")
	promptRAG := "Berdasar konteks berikut, jawab pertanyaan.\nKonteks: " + ctxDoc + "\nTanya: berapa lama refund?"
	ragLLM := &MockChatter{balasan: []string{"Refund diproses dalam 7 hari kerja."}}
	jawab, _ := ragLLM.Chat(ctx, promptRAG)
	fmt.Printf("  konteks terpilih: %q\n", ctxDoc)
	fmt.Printf("  jawaban ter-grounding: %s\n", jawab)
}
