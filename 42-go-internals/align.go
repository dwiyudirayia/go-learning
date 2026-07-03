// Modul 42 — Go Internals: scheduler, GC, memory model, escape analysis.
package main

import "unsafe"

// MEMORY ALIGNMENT & PADDING: compiler menyisipkan byte kosong (padding) agar
// tiap field selaras dengan batas ukurannya. URUTAN field memengaruhi ukuran.

// Boros: bool(1)+padding(7)+int64(8)+bool(1)+padding(7) = 24 byte.
type BadStruct struct {
	a bool  // 1 byte, lalu 7 padding agar b selaras 8
	b int64 // 8
	c bool  // 1, lalu 7 padding di akhir
}

// Rapi: field besar dulu -> int64(8)+bool(1)+bool(1)+padding(6) = 16 byte.
type GoodStruct struct {
	b int64 // 8
	a bool  // 1
	c bool  // 1  (+6 padding)
}

func badSize() uintptr  { return unsafe.Sizeof(BadStruct{}) }
func goodSize() uintptr { return unsafe.Sizeof(GoodStruct{}) }

// escapeToHeap mengembalikan *int -> nilai "lolos" ke heap (dikelola GC).
// Cek: go build -gcflags='-m' ./42-go-internals  -> "escapes to heap".
func escapeToHeap() *int {
	x := 42
	return &x // alamat lokal dikembalikan -> harus di heap
}

// stayOnStack: nilai tidak lolos -> dialokasikan di stack (murah, tanpa GC).
func stayOnStack() int {
	x := 42
	return x
}
