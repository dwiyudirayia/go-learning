package main

import "iter"

// ITERATOR (Go 1.23+): fungsi bertipe iter.Seq[T] = func(yield func(T) bool).
// Bisa dipakai langsung dengan `for v := range seq` (range-over-func).
// Keunggulan: LAZY (dihitung saat diminta) & bisa dirangkai (Map/Filter) tanpa
// membuat slice antara.

// Count menghasilkan 0..n-1 secara lazy.
func Count(n int) iter.Seq[int] {
	return func(yield func(int) bool) {
		for i := 0; i < n; i++ {
			if !yield(i) { // yield=false -> konsumen berhenti (mis. break)
				return
			}
		}
	}
}

// Filter meneruskan hanya elemen yang lolos predikat — tetap lazy.
func Filter[T any](seq iter.Seq[T], keep func(T) bool) iter.Seq[T] {
	return func(yield func(T) bool) {
		for v := range seq {
			if keep(v) && !yield(v) {
				return
			}
		}
	}
}

// Map mengubah tiap elemen (bisa ganti tipe T -> U) — lazy.
func Map[T, U any](seq iter.Seq[T], f func(T) U) iter.Seq[U] {
	return func(yield func(U) bool) {
		for v := range seq {
			if !yield(f(v)) {
				return
			}
		}
	}
}

// Collect mengumpulkan iterator menjadi slice (materialisasi).
func Collect[T any](seq iter.Seq[T]) []T {
	var out []T
	for v := range seq {
		out = append(out, v)
	}
	return out
}
