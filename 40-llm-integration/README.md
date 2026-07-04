# 40 — Integrasi LLM (Claude) dengan Go

Modul penutup: membangun aplikasi ber-AI dengan **API Claude (Anthropic)** memakai **Go SDK resmi**. Menutup chat, streaming, tool use, dan **RAG** — dengan pola **interface + mock** agar bisa di-test tanpa API key.

Jalankan:
```bash
go run ./40-llm-integration                          # mode DEMO (mock, tanpa key)
ANTHROPIC_API_KEY=sk-... go run ./40-llm-integration # panggil Claude sungguhan
```
Verifikasi otomatis: `go test ./40-llm-integration` (pakai mock — jalan di CI tanpa key)

## 📦 Setup

```bash
go get github.com/anthropics/anthropic-sdk-go
export ANTHROPIC_API_KEY=sk-ant-...   # dari console.anthropic.com
```

### Model yang tersedia (per 2026)
| Model | Konstanta SDK | Untuk |
|-------|---------------|-------|
| **Claude Opus 4.8** | `anthropic.ModelClaudeOpus4_8` | paling mumpuni (default modul ini) |
| Claude Sonnet 4.6 | `anthropic.ModelClaudeSonnet4_6` | seimbang cepat/pintar |
| Claude Haiku 4.5 | `anthropic.ModelClaudeHaiku4_5_20251001` | paling cepat & murah |
| Claude Fable 5 | `anthropic.ModelClaudeFable5` | reasoning terberat |

> Pakai **`claude-opus-4-8`** kecuali ada alasan khusus. String ID sudah lengkap apa adanya — jangan tambahkan sufiks tanggal. Untuk model/kapabilitas terbaru, cek Models API (`client.Models.List`).

## Konsep

### 1. Interface + Mock (kunci agar testable)
Aplikasi bergantung pada interface `Chatter`, bukan SDK langsung (Modul 4 & 29):
```go
type Chatter interface {
    Chat(ctx, system string, messages []Message) (string, error)
}
```
- Produksi → `AnthropicChatter` (memanggil Claude).
- Test → `MockChatter` (balasan tetap, **tanpa API key/jaringan**) → `go test` jalan di CI.

Semua test modul ini memakai mock — cepat, deterministik, gratis.

### 2. Chat dasar
```go
resp, _ := client.Messages.New(ctx, anthropic.MessageNewParams{
    Model:     anthropic.ModelClaudeOpus4_8,
    MaxTokens: 1024,
    System:    []anthropic.TextBlockParam{{Text: "Kamu asisten ramah."}},
    Messages:  []anthropic.MessageParam{anthropic.NewUserMessage(anthropic.NewTextBlock("Halo"))},
})
// respons = beberapa content block; gabungkan yang bertipe TextBlock
```

### 3. Streaming (UX chat)
Token dikirim bertahap → tak menunggu jawaban penuh, mencegah timeout output panjang:
```go
stream := client.Messages.NewStreaming(ctx, params)
message := anthropic.Message{}
for stream.Next() {
    event := stream.Current()
    message.Accumulate(event)
    if d, ok := event.AsAny().(anthropic.ContentBlockDeltaEvent); ok {
        if td, ok := d.Delta.AsAny().(anthropic.TextDelta); ok {
            fmt.Print(td.Text) // tampilkan potongan
        }
    }
}
```

### 4. Tool Use (Claude memanggil fungsimu)
Beri Claude "tools"; ia memutuskan kapan memanggilnya, kamu eksekusi, hasil dikirim balik. SDK Go punya **tool runner** yang mengurus loop-nya:
```go
tool, _ := toolrunner.NewBetaToolFromJSONSchema(
    "get_weather", "Cuaca kota",
    func(ctx context.Context, in struct{ City string `json:"city" jsonschema:"required"` }) (anthropic.BetaToolResultBlockParamContentUnion, error) {
        return anthropic.BetaToolResultBlockParamContentUnion{
            OfText: &anthropic.BetaTextBlockParam{Text: "Cerah 30°C di " + in.City},
        }, nil
    })
runner := client.Beta.Messages.NewToolRunner([]anthropic.BetaTool{tool}, anthropic.BetaToolRunnerParams{...})
msg, _ := runner.RunToCompletion(ctx) // otomatis: panggil API -> eksekusi tool -> ulangi sampai selesai
```
Ini fondasi **agent**: Claude memakai tool (cari DB, panggil API, jalankan kode) untuk menyelesaikan tugas.

### 5. RAG (Retrieval-Augmented Generation)
Agar LLM menjawab berdasar **datamu** (dokumen internal) & mengurangi halusinasi:
```
pertanyaan -> retrieve dokumen relevan -> sisipkan sebagai konteks -> LLM -> jawaban
```
`rag.go` melakukannya: cari dokumen relevan (skor kata sederhana), bangun system prompt yang membatasi jawaban pada konteks, lalu tanya LLM. Test membuktikan retrieval mengambil dokumen yang tepat & menolak saat tak ada yang relevan.

> Produksi: ganti pencarian kata dengan **embedding + vector DB** (pgvector, Pinecone, Qdrant) untuk pencarian semantik.

## Praktik penting
- **Jangan hardcode API key** — pakai env (Modul 19). Jangan commit.
- **Streaming** untuk output panjang / `max_tokens` besar (hindari timeout).
- **`MaxTokens` cukup besar** — kalau kena batas, output terpotong.
- **Tangani error & rate limit** (SDK retry otomatis 429/5xx).
- **Interface + mock** agar test cepat & tak boros kuota.
- **Prompt caching** untuk system prompt besar yang dipakai berulang (hemat biaya).

## Kapan & Di Mana Dipakai
- Chatbot & asisten, ringkasan/ekstraksi/klasifikasi dokumen, RAG atas basis pengetahuan, agent yang memakai tool, code assistant.

## Latihan
1. Ganti model ke Haiku (`ModelClaudeHaiku4_5_20251001`) & bandingkan kecepatan.
2. Tambah tool `get_time` dan biarkan Claude memanggilnya (tool runner).
3. Ganti retrieval RAG dengan pencarian embedding (mis. `hkdf`/kosinus atas vektor dummy).
4. Tambah **structured output** (`output_config.format`) agar balasan berupa JSON tervalidasi.
5. Bungkus `Chatter` dengan retry + circuit breaker (Modul 32) untuk panggilan LLM yang tahan gangguan.

---
🎓 **SELESAI — 40 modul!** Dari `package main` pertama hingga aplikasi ber-AI. Kamu kini punya fondasi lengkap Go: fundamental, concurrency, backend, microservices, production-readiness, distributed systems, dan integrasi LLM. Lihat `README.md` root untuk peta penuh.

## ✅ Solusi Latihan (Pembahasan)

1. **Ganti ke Haiku** — set model ke konstanta Haiku (`claude-haiku-4-5`) di request. Lebih cepat & murah untuk tugas ringan; bandingkan latensi vs Opus.
2. **Tool `get_time`** — daftarkan tool dengan schema kosong; saat Claude memanggilnya, kembalikan `time.Now()`. Tool runner mengulang loop hingga Claude selesai (`stop_reason` bukan `tool_use`).
3. **RAG embedding** — ganti pencarian kata-kunci dengan kosinus similarity atas vektor embedding (dummy/nyata). Ambil top-k dokumen terdekat sebagai konteks.
4. **Structured output** — pakai `output_config.format` (JSON schema) agar balasan tervalidasi jadi struct Go; hindari parsing teks bebas yang rapuh.
5. **`Chatter` + resiliency** — bungkus implementasi `Chatter` dengan retry + circuit breaker (Modul 32) agar panggilan LLM tahan gangguan jaringan/rate-limit. Interface `Chatter` membuat ini mudah (dekorator).
