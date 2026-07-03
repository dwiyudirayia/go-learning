package main

import (
	"testing"
	"time"
)

func TestSet(t *testing.T) {
	s := NewSet("a", "b", "a") // "a" ganda -> tetap unik
	if s.Len() != 2 {
		t.Errorf("len = %d; want 2 (unik)", s.Len())
	}
	if !s.Has("a") || s.Has("z") {
		t.Error("Has salah")
	}

	a := NewSet(1, 2, 3)
	b := NewSet(2, 3, 4)
	if a.Union(b).Len() != 4 {
		t.Errorf("union = %d; want 4", a.Union(b).Len())
	}
	if a.Intersect(b).Len() != 2 {
		t.Errorf("intersect = %d; want 2", a.Intersect(b).Len())
	}
}

func TestSetIterator(t *testing.T) {
	s := NewSet(10, 20, 30)
	sum := 0
	for v := range s.All() { // range-over-func
		sum += v
	}
	if sum != 60 {
		t.Errorf("sum via iterator = %d; want 60", sum)
	}
}

func TestIteratorChain(t *testing.T) {
	// Filter genap dari 0..9, kuadratkan, kumpulkan.
	genap := Filter(Count(10), func(n int) bool { return n%2 == 0 })
	sq := Map(genap, func(n int) int { return n * n })
	got := Collect(sq)
	want := []int{0, 4, 16, 36, 64}
	if len(got) != len(want) {
		t.Fatalf("got %v; want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("index %d = %d; want %d", i, got[i], want[i])
		}
	}
}

func TestIteratorLazyBreak(t *testing.T) {
	// Iterator harus BERHENTI saat konsumen break (lazy) — tak menghitung semua.
	calls := 0
	seq := Map(Count(1000000), func(n int) int { calls++; return n })
	taken := 0
	for range seq {
		taken++
		if taken == 3 {
			break
		}
	}
	if calls > 5 { // hanya sedikit yang dievaluasi, bukan sejuta
		t.Errorf("iterator tidak lazy: %d evaluasi", calls)
	}
}

func TestFunctionalOptions(t *testing.T) {
	def := NewServer()
	if def.Port != 8080 || def.TLS {
		t.Errorf("default salah: %+v", def)
	}
	custom := NewServer(WithPort(9090), WithTLS(), WithTimeout(time.Second))
	if custom.Port != 9090 || !custom.TLS || custom.Timeout != time.Second {
		t.Errorf("custom salah: %+v", custom)
	}
	// Host tak diubah -> tetap default.
	if custom.Host != "localhost" {
		t.Errorf("host = %q; want default localhost", custom.Host)
	}
}
