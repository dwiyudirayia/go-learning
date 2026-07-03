//go:build js && wasm

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
