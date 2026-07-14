//go:build js && wasm

// 🔍 Analogi besar WASM (WebAssembly): biasanya browser cuma paham JavaScript. WASM itu seperti
// "PASPOR" yang membolehkan kode dari bahasa lain (Go, Rust, C++) ikut JALAN DI DALAM BROWSER dengan
// kecepatan mendekati native. Jadi logika Go yang sama bisa dipakai di server DAN di halaman web.
// Di sini fungsi Go (add, greet) "dititipkan" ke dunia JavaScript (window) agar tombol di halaman
// bisa memanggilnya. "//go:build js && wasm" = tag yang bilang "file ini KHUSUS saat menyasar browser".
// 'select{}' kosong = "tahan program tetap hidup" agar fungsinya terus siap dipanggil dari halaman.
// Program ini dikompilasi ke WebAssembly & jalan DI BROWSER.
// Build: GOOS=js GOARCH=wasm go build -o main.wasm ./47-tui-wasm/wasm
package main

import "syscall/js"

// add & greet diekspor ke JavaScript -> bisa dipanggil dari halaman web.
func add(_ js.Value, args []js.Value) any {
	return args[0].Int() + args[1].Int()
}

func greet(_ js.Value, args []js.Value) any {
	return "Halo, " + args[0].String() + "! Ini dihitung oleh Go di dalam browser (WASM)."
}

func main() {
	// Daftarkan fungsi Go ke objek global JS (window).
	js.Global().Set("goAdd", js.FuncOf(add))
	js.Global().Set("goGreet", js.FuncOf(greet))

	// Tahan program agar fungsi tetap tersedia untuk dipanggil dari JS.
	select {}
}
