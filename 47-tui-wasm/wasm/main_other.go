//go:build !js || !wasm

// Stub untuk build NON-WASM (host biasa) agar `go build ./...` & `go vet ./...`
// tidak gagal. Kode WASM sesungguhnya ada di main_js.go (build tag `js && wasm`).
package main

func main() {}
