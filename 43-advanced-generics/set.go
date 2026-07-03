// Modul 43 — Advanced Generics: type sets, iterator, struktur data generik,
// functional options. Lanjutan Modul 6.
package main

import "iter"

// Set[T] = struktur data generik: himpunan elemen UNIK. Bekerja untuk tipe apa
// pun yang comparable (bisa jadi key map).
type Set[T comparable] struct {
	m map[T]struct{}
}

func NewSet[T comparable](items ...T) *Set[T] {
	s := &Set[T]{m: make(map[T]struct{})}
	for _, it := range items {
		s.Add(it)
	}
	return s
}

func (s *Set[T]) Add(v T)      { s.m[v] = struct{}{} }
func (s *Set[T]) Remove(v T)   { delete(s.m, v) }
func (s *Set[T]) Has(v T) bool { _, ok := s.m[v]; return ok }
func (s *Set[T]) Len() int     { return len(s.m) }

// Union & Intersect mengembalikan Set baru.
func (s *Set[T]) Union(other *Set[T]) *Set[T] {
	out := NewSet[T]()
	for v := range s.m {
		out.Add(v)
	}
	for v := range other.m {
		out.Add(v)
	}
	return out
}

func (s *Set[T]) Intersect(other *Set[T]) *Set[T] {
	out := NewSet[T]()
	for v := range s.m {
		if other.Has(v) {
			out.Add(v)
		}
	}
	return out
}

// All mengembalikan ITERATOR (iter.Seq, Go 1.23+) -> bisa dipakai `for v := range set.All()`.
func (s *Set[T]) All() iter.Seq[T] {
	return func(yield func(T) bool) {
		for v := range s.m {
			if !yield(v) {
				return
			}
		}
	}
}
