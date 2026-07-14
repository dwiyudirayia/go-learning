package main

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"unicode"
)

// 🔍 Analogi besar RAG: LLM itu seperti murid PANDAI tapi PELUPA soal fakta spesifik perusahaanmu —
// kalau ditanya hal yang tak diketahuinya, ia bisa "mengarang percaya diri" (halusinasi). RAG =
// UJIAN OPEN-BOOK: sebelum murid menjawab, kita CARIKAN dulu halaman buku yang relevan (retrieve),
// selipkan ke soal ("jawab HANYA berdasar teks ini"), baru ia menjawab. Hasilnya berbasis datamu
// (dokumen internal, DB) & jujur bilang "tidak tahu" kalau jawabannya tak ada. Mengurangi karangan.
//
// 🔍 Catatan: retrieve di sini pakai cocok-kata sederhana demi kejelasan. Di produksi, "kemiripan
// makna" dicari pakai EMBEDDING + vector DB — supaya "uang kembali" cocok dengan "refund" walau beda kata.

// RAG (Retrieval-Augmented Generation): sebelum bertanya ke LLM, AMBIL dokumen
// relevan dari basis pengetahuanmu, lalu sisipkan sebagai KONTEKS. Ini membuat
// LLM menjawab berdasar datamu (dokumen internal, DB) & mengurangi halusinasi.
//
//	pertanyaan -> retrieve dokumen relevan -> augment prompt -> LLM -> jawaban

type Doc struct {
	Title   string
	Content string
}

type RAG struct {
	docs    []Doc
	chatter Chatter // bergantung pada interface -> bisa mock saat test
}

func NewRAG(chatter Chatter, docs []Doc) *RAG {
	return &RAG{docs: docs, chatter: chatter}
}

// retrieve memilih k dokumen paling relevan dengan skor kata sederhana.
// (Produksi: pakai embedding + vector DB untuk pencarian semantik.)
func (r *RAG) retrieve(query string, k int) []Doc {
	// Pecah pada non-huruf/angka -> tanda baca ("refund?") tak merusak pencocokan.
	words := strings.FieldsFunc(strings.ToLower(query), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})
	type scored struct {
		doc   Doc
		score int
	}
	var ranked []scored
	for _, d := range r.docs {
		text := strings.ToLower(d.Title + " " + d.Content)
		score := 0
		for _, w := range words {
			if strings.Contains(text, w) {
				score++
			}
		}
		if score > 0 {
			ranked = append(ranked, scored{d, score})
		}
	}
	sort.SliceStable(ranked, func(i, j int) bool { return ranked[i].score > ranked[j].score })

	out := make([]Doc, 0, k)
	for i := 0; i < len(ranked) && i < k; i++ {
		out = append(out, ranked[i].doc)
	}
	return out
}

// Answer menjalankan alur RAG: retrieve -> augment -> tanya LLM.
func (r *RAG) Answer(ctx context.Context, query string) (string, error) {
	docs := r.retrieve(query, 3)
	if len(docs) == 0 {
		return "Maaf, tidak ada informasi relevan di basis pengetahuan.", nil
	}

	// Bangun system prompt yang membatasi jawaban HANYA pada konteks.
	var ctxBuilder strings.Builder
	ctxBuilder.WriteString("Jawab pertanyaan HANYA berdasarkan konteks berikut. ")
	ctxBuilder.WriteString("Jika jawabannya tidak ada di konteks, katakan tidak tahu.\n\nKONTEKS:\n")
	for i, d := range docs {
		fmt.Fprintf(&ctxBuilder, "[%d] %s: %s\n", i+1, d.Title, d.Content)
	}

	return r.chatter.Chat(ctx, ctxBuilder.String(), []Message{
		{Role: "user", Text: query},
	})
}
