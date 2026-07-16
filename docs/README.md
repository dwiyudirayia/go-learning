# 📚 docs/ — Rujukan Cepat Go

Kumpulan dokumen rujukan **ringkas & idiomatik** sebagai **pendamping** modul `01`–`48`. Modul mengajarkan konsep secara berurutan; folder ini untuk **melihat cepat** saat kamu lupa sintaks, mencari pola, atau menghindari jebakan.

| Dokumen | Isi | Kapan dibuka |
|---------|-----|--------------|
| [IDIOM.md](IDIOM.md) | Idiom inti Go + *kenapa*-nya | saat ingin menulis kode yang "terasa Go" |
| [CHEATSHEET.md](CHEATSHEET.md) | Sintaks padat (deklarasi, slice, map, goroutine, generics) | saat lupa sintaks |
| [MEMBACA-TIPE.md](MEMBACA-TIPE.md) | Alur 3 langkah memahami tipe parameter yang tak dikenal | saat fungsi butuh tipe asing (`io.Reader`? `rate.Limit`?) |
| [PITFALLS.md](PITFALLS.md) | Jebakan umum + solusinya | saat kode "aneh" / bug halus |
| [CONCURRENCY.md](CONCURRENCY.md) | Pola konkurensi + aturan main | saat kerja dgn goroutine/channel |
| [TESTING.md](TESTING.md) | Strategi test: table-driven, mock, fuzz, benchmark, integrasi | saat menulis/merancang test |
| [TOOLING.md](TOOLING.md) | Perintah `go`, modules, test, pprof, build tags | saat pakai toolchain |
| [GLOSSARY.md](GLOSSARY.md) | Definisi istilah Go (ID) | saat ketemu istilah asing |

## Cara pakai

1. **Belajar** ikut urutan modul (lihat [`../LEARNING.md`](../LEARNING.md)).
2. **Rujuk** dokumen di sini saat butuh — jangan dihafal sekaligus.
3. **Praktik** di `NN-nama/jawaban-saya/`, lihat teknik lanjut di `NN-nama/advanced/`, dan stack produksi di `NN-nama/real-case/` (peta di [`../REAL-CASE-STACKS.md`](../REAL-CASE-STACKS.md)).

> Semua contoh kode di sini bisa langsung disalin & dicoba. Go **1.26.4**.
