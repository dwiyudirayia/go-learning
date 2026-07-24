package main

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

// Ini juga contoh IDIOMATIK memakai go-cmp di test sungguhan:
// hitung diff, kalau tidak kosong -> laporkan dengan t.Errorf.
func TestCmpDiffDipakaiDiTest(t *testing.T) {
	want := Produk{ID: 1, Nama: "Kopi", Harga: 100, Tag: []string{"panas"}}
	got := Produk{ID: 1, Nama: "Kopi", Harga: 100, Tag: []string{"panas"}}

	// Pola paling umum: cmp.Diff(want, got); string kosong berarti sama.
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("produk tidak sesuai (-want +got):\n%s", diff)
	}
}

func TestSamaPersis(t *testing.T) {
	a := Produk{ID: 1, Nama: "Kopi", Harga: 100, Tag: []string{"panas", "pahit"}}

	t.Run("identik", func(t *testing.T) {
		b := Produk{ID: 1, Nama: "Kopi", Harga: 100, Tag: []string{"panas", "pahit"}}
		if !SamaPersis(a, b) {
			t.Error("dua produk identik seharusnya sama")
		}
	})

	t.Run("beda harga", func(t *testing.T) {
		b := Produk{ID: 1, Nama: "Kopi", Harga: 90, Tag: []string{"panas", "pahit"}}
		if SamaPersis(a, b) {
			t.Error("harga beda seharusnya tidak sama")
		}
	})

	t.Run("beda isi slice", func(t *testing.T) {
		b := Produk{ID: 1, Nama: "Kopi", Harga: 100, Tag: []string{"panas", "manis"}}
		if SamaPersis(a, b) {
			t.Error("tag beda seharusnya tidak sama")
		}
	})
}

// Diff harus MENYEBUT field yang berbeda (nilai jual utama go-cmp).
func TestDiffMenyebutFieldYangBeda(t *testing.T) {
	a := Produk{ID: 1, Nama: "Kopi", Harga: 100}
	b := Produk{ID: 1, Nama: "Kopi", Harga: 90}

	diff := Perbedaan(a, b)
	if diff == "" {
		t.Fatal("ingin ada perbedaan")
	}
	// Laporan harus menyebut field Harga dan kedua nilainya.
	for _, harus := range []string{"Harga", "100", "90"} {
		if !strings.Contains(diff, harus) {
			t.Errorf("diff tidak menyebut %q:\n%s", harus, diff)
		}
	}
}

func TestDiffKosongSaatSama(t *testing.T) {
	a := Produk{ID: 1, Nama: "Kopi", Harga: 100}
	b := Produk{ID: 1, Nama: "Kopi", Harga: 100}
	if diff := Perbedaan(a, b); diff != "" {
		t.Errorf("dua nilai sama seharusnya menghasilkan diff kosong, dapat:\n%s", diff)
	}
}

// IgnoreFields membuat test tahan terhadap field yang tak relevan (mis. ID dari DB).
func TestIgnoreFields(t *testing.T) {
	a := Produk{ID: 1, Nama: "Kopi", Harga: 100}
	b := Produk{ID: 999, Nama: "Kopi", Harga: 100}

	if SamaPersis(a, b) {
		t.Error("dengan ID beda, perbandingan persis seharusnya gagal")
	}
	if !SamaAbaikanID(a, b) {
		t.Error("dengan ID diabaikan, keduanya seharusnya sama")
	}
}

// Membuktikan klaim di komentar: cmp.Equal PANIC pada field unexported tanpa izin.
func TestUnexportedTanpaIzinPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("cmp.Equal pada field unexported TANPA opsi seharusnya panic")
		}
	}()

	a := NewDompet("Ana", 100)
	b := NewDompet("Ana", 100)
	_ = cmp.Equal(a, b) // sengaja tanpa AllowUnexported -> harus panic
}

func TestUnexportedDenganIzin(t *testing.T) {
	tests := []struct {
		nama string
		a, b dompet
		want bool
	}{
		{"identik", NewDompet("Ana", 100), NewDompet("Ana", 100), true},
		{"beda saldo", NewDompet("Ana", 100), NewDompet("Ana", 200), false},
		{"beda pemilik", NewDompet("Ana", 100), NewDompet("Budi", 100), false},
	}
	for _, tt := range tests {
		t.Run(tt.nama, func(t *testing.T) {
			if got := DompetSama(tt.a, tt.b); got != tt.want {
				t.Errorf("DompetSama = %t, ingin %t", got, tt.want)
			}
		})
	}
}

func TestToleransiFloat(t *testing.T) {
	x := 0.1 + 0.2 // tidak persis 0.3 dalam float64

	// Pembanding operator gagal...
	if x == 0.3 {
		t.Skip("pada arsitektur ini 0.1+0.2 kebetulan == 0.3; toleransi tetap benar")
	}
	// ...tapi EquateApprox memaklumi selisih sangat kecil.
	if !HampirSama(x, 0.3) {
		t.Errorf("0.1+0.2 (%v) seharusnya dianggap ~= 0.3 dengan toleransi", x)
	}

	// Selisih besar tetap dianggap berbeda.
	if HampirSama(0.3, 0.4) {
		t.Error("0.3 dan 0.4 selisihnya terlalu besar untuk dianggap sama")
	}
}

// SortSlices: dua slice berisi elemen sama tapi urutan beda dianggap sama.
func TestEquateDenganSortSlices(t *testing.T) {
	a := []int{3, 1, 2}
	b := []int{1, 2, 3}

	// Tanpa opsi, urutan berbeda -> tidak sama.
	if cmp.Equal(a, b) {
		t.Error("tanpa SortSlices, urutan beda seharusnya tidak sama")
	}
	// Dengan SortSlices, keduanya dianggap sama (berguna saat urutan tak penting).
	sortInt := cmpopts.SortSlices(func(x, y int) bool { return x < y })
	if !cmp.Equal(a, b, sortInt) {
		t.Error("dengan SortSlices, isi sama seharusnya dianggap sama")
	}
}
