package main

import (
	"runtime"
	"testing"
)

func TestUrutanFieldMempengaruhiUkuran(t *testing.T) {
	// GoodStruct harus lebih kecil dari BadStruct berkat urutan field.
	if goodSize() >= badSize() {
		t.Errorf("goodSize=%d badSize=%d; good harus < bad", goodSize(), badSize())
	}
	// Di arsitektur 64-bit: Bad=24, Good=16.
	if badSize() != 24 {
		t.Errorf("badSize = %d; want 24 (64-bit)", badSize())
	}
	if goodSize() != 16 {
		t.Errorf("goodSize = %d; want 16 (64-bit)", goodSize())
	}
}

func TestRuntimeFacts(t *testing.T) {
	if runtime.GOMAXPROCS(0) < 1 {
		t.Error("GOMAXPROCS harus >= 1")
	}
	if runtime.NumGoroutine() < 1 {
		t.Error("minimal ada 1 goroutine (test ini sendiri)")
	}
}

// Benchmark membuktikan: nilai yang lolos ke heap (escape) lebih lambat
// & mengalokasi, sedangkan yang tetap di stack tidak.
// Jalankan: go test -bench . -benchmem ./42-go-internals
func BenchmarkEscapeToHeap(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = escapeToHeap() // alokasi heap
	}
}

func BenchmarkStayOnStack(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = stayOnStack() // tanpa alokasi
	}
}
