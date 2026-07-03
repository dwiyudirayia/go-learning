package main

import "testing"

// Test korektnes: kedua implementasi harus menghasilkan output sama.
func TestBuildSama(t *testing.T) {
	if BuildSlow(100) != BuildFast(100) {
		t.Error("BuildSlow dan BuildFast menghasilkan output berbeda")
	}
	if len(BuildFast(50)) != 50 {
		t.Error("panjang hasil salah")
	}
}

// Benchmark: jalankan dengan `go test -bench . -benchmem ./26-profiling`.
// Perhatikan kolom "allocs/op" -> BuildFast jauh lebih sedikit.
func BenchmarkBuildSlow(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = BuildSlow(1000)
	}
}

func BenchmarkBuildFast(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = BuildFast(1000)
	}
}

func BenchmarkSumSquares(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = SumSquares(10000)
	}
}
