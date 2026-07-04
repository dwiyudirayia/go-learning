package main

import "testing"

// Test demonstrasi untuk Modul 06 — generics.
// Jalankan: go test -v ./06-generics
//
// Satu fungsi generik, banyak tipe. Test membuktikan Map/Filter/Reduce/Sum
// bekerja untuk int maupun string tanpa duplikasi kode.

func TestMax(t *testing.T) {
	if got := Max(3, 7); got != 7 {
		t.Errorf("Max(3,7) = %d, mau 7", got)
	}
	if got := Max("apel", "zebra"); got != "zebra" {
		t.Errorf("Max string = %q, mau \"zebra\"", got)
	}
}

func TestMapFilterReduce(t *testing.T) {
	nums := []int{1, 2, 3, 4, 5}

	// Map: kuadratkan.
	kuadrat := Map(nums, func(n int) int { return n * n })
	wantKuadrat := []int{1, 4, 9, 16, 25}
	if !equalInt(kuadrat, wantKuadrat) {
		t.Errorf("Map kuadrat = %v, mau %v", kuadrat, wantKuadrat)
	}

	// Map lintas-tipe: int -> string panjang berbeda.
	labels := Map(nums, func(n int) bool { return n%2 == 0 })
	if len(labels) != len(nums) {
		t.Errorf("Map harus mempertahankan panjang")
	}

	// Filter: hanya genap.
	genap := Filter(nums, func(n int) bool { return n%2 == 0 })
	if !equalInt(genap, []int{2, 4}) {
		t.Errorf("Filter genap = %v, mau [2 4]", genap)
	}

	// Reduce: jumlah semua.
	total := Reduce(nums, 0, func(acc, n int) int { return acc + n })
	if total != 15 {
		t.Errorf("Reduce jumlah = %d, mau 15", total)
	}
}

func TestSum(t *testing.T) {
	if got := Sum([]int{1, 2, 3}); got != 6 {
		t.Errorf("Sum int = %d, mau 6", got)
	}
	if got := Sum([]float64{1.5, 2.5}); got != 4.0 {
		t.Errorf("Sum float64 = %v, mau 4.0", got)
	}
}

func equalInt(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
