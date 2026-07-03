package store

import (
	"errors"
	"fmt"
	"testing"
)

// ------------------------------------------------------------------
// Table-driven test + subtest (t.Run) + kasus error
// ------------------------------------------------------------------
func TestApplyDiscount(t *testing.T) {
	tests := []struct {
		name    string
		price   int
		pct     int
		want    int
		wantErr bool
	}{
		{"diskon 10%", 1000, 10, 900, false},
		{"diskon 0%", 500, 0, 500, false},
		{"diskon 100%", 250, 100, 0, false},
		{"pct negatif", 1000, -5, 0, true},
		{"pct > 100", 1000, 150, 0, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ApplyDiscount(tc.price, tc.pct)

			if tc.wantErr {
				if err == nil {
					t.Fatalf("ApplyDiscount(%d,%d): mengharapkan error, dapat nil", tc.price, tc.pct)
				}
				return // kasus error: tidak perlu cek nilai
			}
			if err != nil {
				t.Fatalf("ApplyDiscount(%d,%d): tak terduga error: %v", tc.price, tc.pct, err)
			}
			if got != tc.want {
				t.Errorf("ApplyDiscount(%d,%d) = %d; want %d", tc.price, tc.pct, got, tc.want)
			}
		})
	}
}

func TestIsValidEmail(t *testing.T) {
	tests := []struct {
		in   string
		want bool
	}{
		{"ana@mail.com", true},
		{"a.b@sub.domain.id", true},
		{"tanpa-at.com", false},
		{"dua@@at.com", false},
		{"@mulaidenganat.com", false},
		{"ana@nodot", false},
		{"ana@titikdiakhir.", false},
	}
	for _, tc := range tests {
		t.Run(tc.in, func(t *testing.T) {
			if got := IsValidEmail(tc.in); got != tc.want {
				t.Errorf("IsValidEmail(%q) = %t; want %t", tc.in, got, tc.want)
			}
		})
	}
}

// ------------------------------------------------------------------
// MOCK: implementasi palsu dari ProductRepo untuk menguji Catalog tanpa DB
// ------------------------------------------------------------------
type fakeRepo struct {
	products map[int]Product
}

func (f fakeRepo) FindByID(id int) (Product, error) {
	p, ok := f.products[id]
	if !ok {
		return Product{}, ErrProductNotFound
	}
	return p, nil
}

func TestCatalog_PriceWithTax(t *testing.T) {
	repo := fakeRepo{products: map[int]Product{
		1: {ID: 1, Name: "Kopi", Price: 20000},
	}}
	cat := NewCatalog(repo)

	t.Run("produk ada + pajak 10%", func(t *testing.T) {
		got, err := cat.PriceWithTax(1, 10)
		if err != nil {
			t.Fatalf("tak terduga error: %v", err)
		}
		if want := 22000; got != want {
			t.Errorf("PriceWithTax(1,10) = %d; want %d", got, want)
		}
	})

	t.Run("produk tidak ada -> ErrProductNotFound", func(t *testing.T) {
		_, err := cat.PriceWithTax(999, 10)
		if !errors.Is(err, ErrProductNotFound) {
			t.Errorf("mengharapkan ErrProductNotFound, dapat %v", err)
		}
	})
}

// ------------------------------------------------------------------
// Benchmark
// ------------------------------------------------------------------
func BenchmarkApplyDiscount(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = ApplyDiscount(1000, 10)
	}
}

// ------------------------------------------------------------------
// Example test: diverifikasi via // Output: dan muncul di go doc
// ------------------------------------------------------------------
func ExampleApplyDiscount() {
	got, _ := ApplyDiscount(1000, 10)
	fmt.Println(got)
	// Output: 900
}
