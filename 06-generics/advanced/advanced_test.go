package main

import (
	"slices"
	"testing"
)

// TestMax menguji fungsi generik Max untuk beberapa tipe di dalam constraint.
func TestMax(t *testing.T) {
	if got := Max(3, 7); got != 7 {
		t.Errorf("Max(3,7) = %d, mau 7", got)
	}
	if got := Max(2.5, 1.5); got != 2.5 {
		t.Errorf("Max(2.5,1.5) = %v, mau 2.5", got)
	}
	// Tipe kustom berbasis float64 (Celsius) juga harus bekerja (~float64).
	if got := Max[Celsius](20, 25); got != 25 {
		t.Errorf("Max[Celsius](20,25) = %v, mau 25", got)
	}
}

// TestSet menguji Set generik: dedup, Has, dan Len.
func TestSet(t *testing.T) {
	s := NewSet("go", "rust", "go", "c") // "go" duplikat
	if s.Len() != 3 {
		t.Errorf("Len = %d, mau 3 (duplikat dihitung sekali)", s.Len())
	}
	if !s.Has("rust") {
		t.Error("Has(rust) = false, mau true")
	}
	if s.Has("zig") {
		t.Error("Has(zig) = true, mau false")
	}
}

// TestMap menguji transformasi generik []int -> []string (tipe input & output beda).
func TestMap(t *testing.T) {
	got := Map([]int{1, 2, 3}, func(n int) string {
		return string(rune('a' + n - 1)) // 1->"a", 2->"b", 3->"c"
	})
	want := []string{"a", "b", "c"}
	if !slices.Equal(got, want) {
		t.Errorf("Map = %v, mau %v", got, want)
	}
}
