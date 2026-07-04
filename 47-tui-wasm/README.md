# 47 — TUI & WebAssembly

Dua cara memakai Go di luar server: **TUI** (aplikasi terminal interaktif) dengan Bubble Tea, dan **WebAssembly** (Go jalan di browser).

## Bagian A — TUI dengan Bubble Tea

Jalankan (butuh terminal sungguhan):
```bash
go run ./47-tui-wasm     # counter interaktif: +/- naik-turun, r reset, q keluar
```
Verifikasi otomatis: `go test ./47-tui-wasm`

### Arsitektur Elm (Model-Update-View)
[Bubble Tea](https://github.com/charmbracelet/bubbletea) memakai pola dari bahasa Elm:
```
        ┌──────────────────────────────────┐
        │  Model (state aplikasi)          │
        │     ▲                    │        │
        │     │ state baru         │ render │
        │  Update(msg) ◄─pesan─ View() ──► terminal
        └──────────────────────────────────┘
```
```go
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() { case "+": m.count++ ; case "q": return m, tea.Quit }
    }
    return m, nil  // state baru
}
```

### Kenapa mudah di-test?
`Update` **murni**: `(state lama, pesan) → state baru`. Bisa diuji tanpa terminal — cukup panggil `Update` dengan `tea.KeyMsg` & periksa state. Test membuktikan counter naik/turun, tak negatif, reset, & quit — **semua tanpa TTY**.

Ekosistem: [Bubbles](https://github.com/charmbracelet/bubbles) (komponen: input, list, table), [Lip Gloss](https://github.com/charmbracelet/lipgloss) (styling). Dipakai `gh` dashboard, `glow`, dll.

## Bagian B — WebAssembly (Go di browser)

Go bisa dikompilasi ke **WASM** & jalan di browser (mis. logika berat, game, tool client-side).

### Build & jalankan
```bash
# 1. Kompilasi ke WASM
GOOS=js GOARCH=wasm go build -o 47-tui-wasm/wasm/main.wasm ./47-tui-wasm/wasm

# 2. Salin runtime loader (Go 1.24+ ada di lib/wasm; sebelumnya misc/wasm)
cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" 47-tui-wasm/wasm/

# 3. Sajikan & buka di browser
cd 47-tui-wasm/wasm && python3 -m http.server 8080   # http://localhost:8080
```

### Ekspor fungsi Go ke JavaScript
```go
import "syscall/js"
func add(_ js.Value, args []js.Value) any { return args[0].Int() + args[1].Int() }
func main() {
    js.Global().Set("goAdd", js.FuncOf(add)) // -> dipanggil dari JS: goAdd(2,3)
    select {}                                 // tahan agar fungsi tetap hidup
}
```
`syscall/js` menjembatani Go ↔ JavaScript/DOM. `index.html` memuat `main.wasm` via `WebAssembly.instantiateStreaming` lalu memanggil `goAdd`/`goGreet`.

### Catatan build (yang penting dipahami)
- File `main_js.go` diberi tag `//go:build js && wasm` → hanya ikut saat build WASM.
- File `main_other.go` (`//go:build !js || !wasm`) = **stub** agar `go build ./...` di host biasa tak gagal (kalau tidak, direktori hanya berisi file js/wasm → "build constraints exclude all Go files").
- Ini pola umum kode yang platform-specific.

**Trade-off WASM:** binari besar (modul ini ~2MB, bisa dikecilkan dengan TinyGo), tak akses DOM seclangsung tanpa `syscall/js`. Cocok untuk komputasi client-side, bukan pengganti JS untuk UI ringan.

## Kapan & Di Mana Dipakai
- **TUI**: dev tool, installer, dashboard terminal, CLI interaktif (lengkapi Cobra Modul 11).
- **WASM**: jalankan logika Go yang sudah ada di browser, game, image/crypto processing client-side, plugin.

## Latihan
1. Ubah counter jadi **todo list** (Bubbles `list`/`textinput`).
2. Tambah `View` berwarna dengan Lip Gloss.
3. Tambah fungsi WASM `goReverse(s)` & panggil dari `index.html`.
4. Bangun WASM dengan **TinyGo** & bandingkan ukuran binari.
5. Buat TUI yang memanggil REST API (Modul 13) & menampilkan hasilnya.

## ✅ Solusi Latihan (Pembahasan)

1. **Todo list** — ganti model counter dengan slice item + Bubbles `textinput` (tambah item) & `list` (tampil). `Update` menangani key Enter (tambah) & `d` (hapus).
2. **View berwarna** — `lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true).Render(s)` di dalam `View()`. Styling deklaratif.
3. **WASM `goReverse`** — daftarkan fungsi: `js.Global().Set("goReverse", js.FuncOf(func(_ js.Value, args []js.Value) any { return reverse(args[0].String()) }))`; panggil dari `index.html`. Build `GOOS=js GOARCH=wasm`.
4. **TinyGo** — `tinygo build -o main.wasm -target wasm ./...`; binari jauh lebih kecil dari toolchain Go standar (cocok untuk web), dengan trade-off dukungan stdlib terbatas.
5. **TUI panggil REST** — dalam `Update`, jalankan `tea.Cmd` yang melakukan HTTP GET (Modul 13) secara async; hasilnya kembali sebagai `tea.Msg` → render di `View`. Jangan blok event loop.
