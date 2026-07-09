# 11 — CLI dengan Cobra

[Cobra](https://github.com/spf13/cobra) adalah library CLI paling populer di Go (dipakai `kubectl`, `docker`, `gh`, `hugo`). Ia menyediakan subcommand, flag, help & autocompletion otomatis.

Coba:
```bash
go run ./11-cli-cobra add "belajar cobra"
go run ./11-cli-cobra list
go run ./11-cli-cobra done 1
go run ./11-cli-cobra list --all
go run ./11-cli-cobra version
go run ./11-cli-cobra --help          # help otomatis
```
> Task disimpan ke file JSON (default `TMPDIR/golearn-tasks.json`). Ganti dengan `--file path.json`.

## Konsep Cobra

### Struktur
Konvensi proyek Cobra: `main.go` tipis → `cmd/` berisi tiap perintah. Modul ini:
```
11-cli-cobra/
├── main.go                 # panggil cmd.Execute()
├── cmd/
│   ├── root.go             # rootCmd + persistent flag --file
│   ├── add.go  list.go  done.go  version.go
└── internal/store/store.go # persistensi JSON (logika, terpisah dari CLI)
```

### `cobra.Command`
```go
var addCmd = &cobra.Command{
	Use:   "add [teks]",       // sintaks pemakaian (muncul di help)
	Short: "Tambah tugas baru",
	Args:  cobra.MinimumNArgs(1), // validasi argumen
	RunE:  func(cmd *cobra.Command, args []string) error { ... },
}
rootCmd.AddCommand(addCmd)  // daftarkan subcommand
```
- **`Run` vs `RunE`** — pakai `RunE` agar bisa `return error`; Cobra mencetak error & set exit code. Idiomatik untuk CLI produksi.
- **`Args`** — validator bawaan: `ExactArgs(n)`, `MinimumNArgs(n)`, `NoArgs`, dll.

### Flag: persistent vs local
```go
rootCmd.PersistentFlags().StringVar(&filePath, "file", def, "...") // untuk root + SEMUA subcommand
listCmd.Flags().BoolVar(&showAll, "all", false, "...")             // hanya untuk 'list'
```

### Help & error otomatis
Cobra otomatis menghasilkan `--help`, pesan usage saat argumen salah, dan `did you mean` untuk typo — tanpa kode tambahan.

## Cobra vs `flag` (stdlib)
- **`flag` (stdlib, lihat Modul 10)** — cukup untuk CLI 1–2 flag sederhana.
- **Cobra** — saat butuh **subcommand** bertingkat, banyak flag, help rapi, atau CLI yang tumbuh besar.
- Sering dipasangkan dengan **Viper** (`spf13/viper`) untuk config dari file/env/flag sekaligus.

## Kapan & Di Mana Dipakai (Studi Kasus Nyata)

| Kasus | Contoh |
|-------|--------|
| Tool internal / DevOps | migrator DB (`app migrate up`), seeder, backup |
| CLI produk | `gh`, `kubectl`, `docker` — semua Cobra |
| Worker/one-off command | `app worker`, `app import file.csv` |
| Multi-binary via `cmd/` | `cmd/server`, `cmd/cli`, `cmd/migrate` satu repo |

**Pola nyata:** aplikasi backend sering punya beberapa entry point — server HTTP **dan** CLI admin (migrasi, buat user admin, dsb). Cobra menyatukannya rapi.

## Latihan
1. Tambah subcommand `rm [id]` untuk menghapus tugas.
2. Tambah flag `--priority` (int) pada `add`, simpan & tampilkan di `list`.
3. Tambah subcommand `stats` yang menampilkan jumlah total/selesai/belum.
4. Tambah test untuk `internal/store` (Add, MarkDone, ErrNotFound) — tanpa menyentuh Cobra.

Kunci jawaban: kembangkan sendiri (arsitekturnya sudah disiapkan) — atau minta saya buatkan.

## ✅ Status Solusi Latihan
Latihan **1–4 sudah diselesaikan di kode ini**: `rm` (cmd/rm.go), `--priority` (cmd/add.go), `stats` (cmd/stats.go), dan test store (internal/store/store_test.go). Jalankan `go test ./11-cli-cobra/...`.

---

## 🚀 Teknik Advanced (Level Up)
> 💻 **Contoh runnable + komentar detail** untuk teknik di bawah ada di folder [`advanced/`](advanced). Jalankan: `go run ./11-cli-cobra/advanced`


- **Persistent vs local flags** — `PersistentFlags()` diwarisi subcommand; `Flags()` lokal. `MarkFlagRequired`, flag groups (`MarkFlagsRequiredTogether`).
- **`RunE` bukan `Run`** — kembalikan error agar exit code benar; `os.Exit` hanya di `main`, tidak di dalam command.
- **`cmd.Context()`** — propagasi context (mis. dari `signal.NotifyContext`) ke seluruh command untuk cancel bersih.
- **Viper binding** — `viper.BindPFlag` gabungkan flag + env + file config dengan presedensi jelas. Lihat [[19-config]].
- **Shell completion** — Cobra generate completion bash/zsh/fish otomatis (`cmd completion zsh`).
- **`PreRunE`/`PostRunE`** — validasi/setup sebelum eksekusi (mis. buka koneksi DB, cek auth).
- **Exit code semantik** — 0 sukses, non-0 gagal; petakan jenis error → kode berbeda bila perlu.
