# 📚 Panduan Belajar — Cara Menguasai 48 Modul Ini

Repo ini besar (48 modul). Panduan ini membantumu **fokus**: urutan, cara belajar tiap modul, estimasi waktu, dan apa yang penting. Jangan coba habiskan semua sekaligus — ikuti jalurnya.

---

## 🎯 Cara belajar SATU modul (ulangi untuk tiap modul)

```
1. BACA   README.md modul       (konsep + "Kapan Dipakai")     ~10 menit
2. JALAN  go run ./NN-nama       (lihat output, cocokkan)        ~5 menit
3. BACA   kode (main.go / *.go)  (pahami tiap komentar)          ~15 menit
4. UBAH   coba ubah sesuatu & jalankan lagi (belajar aktif!)     ~10 menit
5. LATIH  kerjakan latihan di README (di jawaban-saya/)          ~20 menit
6. CEK    go test ./NN-nama      (lihat bagaimana diuji)
7. DALAMI go run ./NN-nama/advanced  (teknik lanjutan; bab "🚀 Teknik Advanced")
```
> **Belajar aktif > pasif.** Mengubah kode & melihat efeknya jauh lebih melekat daripada sekadar membaca. Rusak dulu, perbaiki, pahami.

> 💡 **Buntu di sintaks/istilah?** Jangan berhenti — buka **[`docs/`](docs/)** (lihat di bawah) sebagai rujukan cepat, lalu lanjut.

---

## 📎 Rujukan cepat — folder `docs/`

Modul mengajarkan **berurutan**; `docs/` untuk **melihat cepat** kapan pun butuh (tak perlu dihafal). Buka sesuai kebutuhan:

| Buka ini… | …saat |
|-----------|-------|
| [docs/CHEATSHEET.md](docs/CHEATSHEET.md) | lupa sintaks (slice, map, goroutine, generics) |
| [docs/IDIOM.md](docs/IDIOM.md) | ingin menulis kode yang "terasa Go" + *kenapa* |
| [docs/PITFALLS.md](docs/PITFALLS.md) | kode berperilaku aneh / bug halus (typed-nil, aliasing, race) |
| [docs/CONCURRENCY.md](docs/CONCURRENCY.md) | kerja dgn goroutine/channel/context (Fase 2 & modul 38) |
| [docs/TESTING.md](docs/TESTING.md) | menulis/merancang test (table-driven, mock, fuzz, integrasi) — Fase 3 & modul 08/37 |
| [docs/TOOLING.md](docs/TOOLING.md) | pakai `go test/vet/pprof`, modules, build tags |
| [docs/GLOSSARY.md](docs/GLOSSARY.md) | ketemu istilah asing (receiver, method set, GMP, dll) |

**Saran per fase:** Fase 1 → baca **IDIOM** & **CHEATSHEET** santai; tiap kali ketemu bug aneh → **PITFALLS**. Sebelum Fase 2 → baca **CONCURRENCY**. Kapan pun pakai toolchain → **TOOLING**.

---

## 🗺️ Jalur belajar (ikuti berurutan)

### 🟢 WAJIB — fondasi (jangan dilewati)
| Fase | Modul | Fokus | Estimasi |
|------|-------|-------|----------|
| 1 | 01–06 | Sintaks & idiom Go (slice, interface, error, generics) | ~1 minggu |
| 2 | 07 | **Concurrency** (goroutine, channel, context) — inti Go | ~3 hari |
| 3 | 08–10 | Testing, stdlib, struktur proyek | ~3 hari |

**Setelah Fase 1–3, kamu sudah bisa nulis Go dengan benar.** Ini pondasi untuk semua modul lain.

### 🔵 BACKEND — kalau tujuanmu jadi backend engineer
| Fase | Modul | Fokus | Estimasi |
|------|-------|-------|----------|
| 4–5 | 11–15 | CLI, REST (net/http & Fiber), database, **studi kasus JWT** | ~1 minggu |
| 6 | 16–17 | gRPC + **studi kasus microservices** | ~4 hari |
| 7 | 18–21 | Observability, config, graceful shutdown, migrasi | ~4 hari |
| 8 | 22–25 | Cache, message queue, websocket, background jobs | ~1 minggu |
| 9 | 26–30 | Profiling, security, API docs, clean arch, deployment | ~1 minggu |

**⭐ Paling penting di sini: Modul 41 (Capstone)** — bangun sendiri, gabungkan semuanya. Ini yang mematangkan.

### 🟣 LANJUTAN — untuk mendalami / spesialisasi (opsional, sesuai kebutuhan)
| Fase | Modul | Kapan pelajari |
|------|-------|----------------|
| 10 | 31–33 | Saat kerja dengan microservices (saga, resiliency, tracing) |
| 11 | 34–36 | Saat butuh gRPC lanjut, GraphQL, atau SQL type-safe |
| 12 | 37–40 | Saat fokus kualitas (fuzzing), performa, atau bangun app AI |
| 13 | 41–43 | **41 wajib** (capstone). 42–43 untuk jadi "expert" |
| 14 | 44–48 | Hanya bila relevan: auth kompleks, event sourcing, integrasi, TUI/WASM, gRPC-gateway |

---

## 🧭 Rekomendasi jalur berdasar tujuan

- **"Saya mau bisa Go dengan benar"** → Modul 1–10, lalu latihan.
- **"Saya mau jadi backend engineer"** → Modul 1–30 + **41 (capstone)**. Ini paket lengkap.
- **"Saya sudah backend, mau naik level"** → Modul 31–40 + 45.
- **"Saya mau spesialisasi X"** → Modul 44 (auth), 46 (integrasi), 47 (TUI/WASM), dst sesuai kebutuhan.
- **"Saya mau bangun app AI"** → Modul 1–10 + 13 + 15 + **40**.

---

## 💡 Prinsip agar tidak kewalahan

1. **Satu modul per sesi.** Selesaikan (baca+jalan+latih) sebelum lanjut.
2. **Jangan hafal — pahami polanya.** Interface+mock, layered arch, error handling — pola ini berulang di banyak modul.
3. **Latihan itu wajib**, bukan opsional. Tiap modul punya folder `jawaban-saya/` — kerjakan di sana. Untuk mencocokkan: modul **1–9** punya `latihan/solusi.go`; modul **10–48** punya bagian **"Solusi Latihan (Pembahasan)"** / **"Status Solusi Latihan"** di README. Kerjakan dulu, baru intip solusinya.
4. **Kembali & ulang.** Modul lanjutan sering merujuk modul dasar — wajar bolak-balik.
5. **Bangun proyekmu sendiri** setelah Fase 5. Praktik nyata > menonton.
6. **Biasakan** `go test`, `go vet`, `go fmt` — persis kerja Go profesional.

---

## 🔧 Perintah harian
```bash
make run MOD=01-basics    # jalankan modul
make test                 # semua test
make test-race            # + deteksi data race
go run ./NN-nama          # jalankan langsung
go test -v ./NN-nama      # test detail satu modul
```

---

## ✅ Cara tahu kamu sudah paham sebuah modul
- Bisa **menjelaskan** konsepnya ke orang lain dengan kata sendiri.
- Bisa **mengerjakan latihannya** tanpa melihat solusi.
- Bisa **menjawab** "kapan aku pakai ini di proyek nyata?" (lihat bagian "Kapan Dipakai" tiap README).

Kalau tiga hal itu ✔ — lanjut. Kalau belum — ulangi, itu normal.

Selamat belajar! Konsistensi kecil tiap hari > maraton sekali. 🐹
