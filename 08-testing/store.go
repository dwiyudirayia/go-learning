// Package store adalah kode contoh yang DIUJI di modul 08 — Testing.
// Lihat store_test.go untuk contoh unit test, table-driven, mock, benchmark, example.
package store

import (
	"errors"
	"strings"
)

// ApplyDiscount mengurangi price sebesar pct persen (0..100).
func ApplyDiscount(price, pct int) (int, error) {
	if pct < 0 || pct > 100 {
		return 0, errors.New("persen diskon harus 0..100")
	}
	return price - price*pct/100, nil
}

// 🔍 Analogi: fungsi seperti ini "murni" — input sama selalu output sama, tanpa efek samping.
// Fungsi murni itu PALING MUDAH DIUJI: seperti soal matematika yang jawabannya pasti, jadi
// tinggal cocokkan "IsValidEmail('a@b.com') harus true". Inilah yang diuji table-driven di _test.go.
// IsValidEmail cek sederhana: tepat satu '@' dan ada '.' setelahnya.
func IsValidEmail(s string) bool {
	at := strings.IndexByte(s, '@')
	if at <= 0 || strings.Count(s, "@") != 1 {
		return false
	}
	domain := s[at+1:]
	return strings.Contains(domain, ".") && !strings.HasSuffix(domain, ".")
}

// ------------------------------------------------------------------
// Contoh service + interface untuk demonstrasi MOCKING
// ------------------------------------------------------------------

type Product struct {
	ID    int
	Name  string
	Price int
}

// ErrProductNotFound dipakai repo saat produk tak ada (sentinel).
var ErrProductNotFound = errors.New("produk tidak ditemukan")

// 🔍 Analogi besar: kenapa Catalog bergantung pada INTERFACE (ProductRepo), bukan DB langsung?
// Bayangkan stopkontak. Catalog cuma butuh "colokan" yang bisa FindByID. Di dunia nyata
// kamu colok "database asli"; saat menguji, kamu colok "database boongan (mock)" yang isinya
// sudah kamu atur. Karena stopkontaknya sama, Catalog tak sadar bedanya — inilah kenapa
// interface bikin kode MUDAH DIUJI tanpa perlu database sungguhan yang lambat & ribet.

// ProductRepo adalah dependency Catalog. Di produksi diisi implementasi DB;
// di test diisi implementasi palsu (mock).
type ProductRepo interface {
	FindByID(id int) (Product, error)
}

type Catalog struct {
	repo ProductRepo
}

func NewCatalog(repo ProductRepo) *Catalog {
	return &Catalog{repo: repo}
}

// PriceWithTax mengambil produk lewat repo lalu menambahkan pajak taxPct persen.
func (c *Catalog) PriceWithTax(id, taxPct int) (int, error) {
	p, err := c.repo.FindByID(id)
	if err != nil {
		return 0, err // teruskan error dari repo (mis. ErrProductNotFound)
	}
	return p.Price + p.Price*taxPct/100, nil
}
