package main

import (
	"slices"
	"testing"
)

// TestFilterLazy menguji komposisi iterator: ambil 5 bilangan genap pertama dari
// deret TAK HINGGA (Hitung) melalui Filter. Membuktikan iterator bersifat malas
// (lazy) — bila tidak, program takkan pernah berhenti.
func TestFilterLazy(t *testing.T) {
	var got []int
	for n := range Filter(Hitung(), func(n int) bool { return n%2 == 0 }) {
		got = append(got, n)
		if len(got) == 5 {
			break // menghentikan iterator tak hingga
		}
	}
	want := []int{0, 2, 4, 6, 8}
	if !slices.Equal(got, want) {
		t.Errorf("Filter genap 5 pertama = %v, mau %v", got, want)
	}
}

// TestSetAll menguji Set generik + iterator All() dikumpulkan & diurutkan.
func TestSetAll(t *testing.T) {
	s := NewSet("go", "rust", "go", "zig", "c")
	got := slices.Sorted(s.All()) // slices.Sorted menerima iter.Seq
	want := []string{"c", "go", "rust", "zig"}
	if !slices.Equal(got, want) {
		t.Errorf("Set.All terurut = %v, mau %v", got, want)
	}
}

// TestBuildOptions menguji functional options generik.
func TestBuildOptions(t *testing.T) {
	srv := Build(
		func(s *Server) { s.Host = "localhost" },
		func(s *Server) { s.Port = 8080 },
	)
	if srv.Host != "localhost" || srv.Port != 8080 {
		t.Errorf("Build = %+v, mau {localhost 8080}", srv)
	}
}
