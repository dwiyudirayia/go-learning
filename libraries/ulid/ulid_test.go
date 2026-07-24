package main

import (
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/oklog/ulid/v2"
)

func TestBaruMenghasilkanFormatBenar(t *testing.T) {
	id := Baru()
	s := id.String()

	// ULID selalu 26 karakter Base32 Crockford.
	if len(s) != 26 {
		t.Errorf("panjang = %d, ingin 26", len(s))
	}
	// Tak boleh mengandung tanda hubung (beda dari UUID).
	for _, r := range s {
		if r == '-' {
			t.Error("ULID tidak boleh mengandung tanda hubung")
		}
	}
}

func TestBaruSelaluUnik(t *testing.T) {
	set := make(map[ulid.ULID]struct{})
	for range 1000 {
		id := Baru()
		if _, ada := set[id]; ada {
			t.Fatalf("ULID kembar dihasilkan: %s", id)
		}
		set[id] = struct{}{}
	}
}

// Sifat inti: urut waktu = urut leksikografis.
func TestUrutWaktuSamaDenganUrutTeks(t *testing.T) {
	mulai := time.Date(2026, 7, 24, 10, 0, 0, 0, time.UTC)
	ids := BuatBerurutan(mulai, 10)

	teks := make([]string, len(ids))
	for i, id := range ids {
		teks[i] = id.String()
	}

	// Salin lalu urutkan sebagai string; harus tetap sama dengan urutan pembuatan.
	terurut := slices.Clone(teks)
	slices.Sort(terurut)

	if !slices.Equal(teks, terurut) {
		t.Error("ULID yang dibuat berurutan waktu seharusnya sudah urut secara teks")
	}
}

func TestWaktuDariMembacaKembaliStempel(t *testing.T) {
	tests := []struct {
		nama  string
		waktu time.Time
	}{
		{"pagi", time.Date(2026, 7, 24, 8, 0, 0, 0, time.UTC)},
		{"tengah hari", time.Date(2026, 7, 24, 12, 30, 45, 0, time.UTC)},
		{"lampau", time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)},
	}

	for _, tt := range tests {
		t.Run(tt.nama, func(t *testing.T) {
			id := BaruPadaWaktu(tt.waktu)
			got := WaktuDari(id)
			// ULID beresolusi milidetik, jadi bandingkan pada milidetik.
			want := tt.waktu.Truncate(time.Millisecond)
			if !got.Equal(want) {
				t.Errorf("WaktuDari = %s, ingin %s", got, want)
			}
		})
	}
}

func TestParseULID(t *testing.T) {
	tests := []struct {
		nama    string
		input   string
		wantErr bool
	}{
		{"ULID sah", "01ARZ3NDEKTSV4RRFFQ69G5FAV", false},
		{"terlalu pendek", "01ARZ3NDEK", true},
		{"terlalu panjang", "01ARZ3NDEKTSV4RRFFQ69G5FAVXXXX", true},
		{"karakter terlarang I (ketat)", "01ARZ3NDEKTSV4RRFFQ69G5FAI", true},
		{"karakter terlarang U (ketat)", "01ARZ3NDEKTSV4RRFFQ69G5FAU", true},
		{"tanda seru (ketat)", "01ARZ3NDEKTSV4RRFFQ69G5FA!", true},
		{"tanda hubung (ketat)", "01ARZ3-DEKTSV4RRFFQ69G5FAV", true},
		{"overflow (awalan 8)", "81ARZ3NDEKTSV4RRFFQ69G5FAV", true},
		{"kosong", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.nama, func(t *testing.T) {
			_, err := ParseULID(tt.input)
			if tt.wantErr {
				if !errors.Is(err, ErrULIDTidakValid) {
					t.Errorf("ParseULID(%q) error = %v, ingin ErrULIDTidakValid", tt.input, err)
				}
				return
			}
			if err != nil {
				t.Errorf("ParseULID(%q) error tak terduga: %v", tt.input, err)
			}
		})
	}
}

// Pulang-pergi: String() lalu Parse() harus mengembalikan ULID yang sama.
func TestPulangPergi(t *testing.T) {
	for range 100 {
		asli := Baru()
		kembali, err := ParseULID(asli.String())
		if err != nil {
			t.Fatalf("Parse gagal: %v", err)
		}
		if asli != kembali {
			t.Fatalf("pulang-pergi rusak: %s != %s", asli, kembali)
		}
	}
}

func TestValid(t *testing.T) {
	if !Valid(Baru().String()) {
		t.Error("ULID hasil generate seharusnya valid")
	}
	if Valid("bukan-ulid") {
		t.Error("teks sembarang seharusnya tidak valid")
	}
}

// Mendokumentasikan jebakan: ulid.Parse (longgar) menerima karakter aneh yang
// ulid.ParseStrict (dan Valid milik kita) tolak. Ini alasan kita pakai ParseStrict.
func TestParseLonggarVsKetat(t *testing.T) {
	// Panjang benar (26) tapi mengandung '!' — jelas bukan ULID sah.
	kotor := "01ARZ3NDEKTSV4RRFFQ69G5FA!"

	// Parse longgar MELOLOSKANNYA (perilaku bawaan oklog/ulid).
	if _, err := ulid.Parse(kotor); err != nil {
		t.Logf("catatan: pada versi ini ulid.Parse menolak %q (err=%v)", kotor, err)
	}
	// Tapi fungsi Valid kita (ParseStrict) harus menolaknya.
	if Valid(kotor) {
		t.Errorf("Valid(%q) = true, ingin false — harus memakai ParseStrict", kotor)
	}
}

// Inti MonotonicEntropy: ID dalam milidetik SAMA tetap urut naik.
func TestMonotonicUrutDalamMilidetikSama(t *testing.T) {
	p := NewPembuatMonotonic()
	// Waktu identik untuk SEMUA ID — memaksa bagian stempel waktu sama persis.
	sama := time.Date(2026, 7, 24, 10, 0, 0, 0, time.UTC)

	const n = 100
	ids := make([]ulid.ULID, n)
	for i := range ids {
		id, err := p.Baru(sama)
		if err != nil {
			t.Fatalf("Baru gagal: %v", err)
		}
		ids[i] = id
	}

	// Semuanya harus benar-benar naik (bukan cuma tak-turun).
	for i := 1; i < n; i++ {
		if ids[i].Compare(ids[i-1]) <= 0 {
			t.Fatalf("ID ke-%d (%s) tidak > ID sebelumnya (%s) — monotonic gagal",
				i, ids[i], ids[i-1])
		}
	}

	// Dan semuanya tetap unik.
	set := make(map[ulid.ULID]struct{}, n)
	for _, id := range ids {
		if _, ada := set[id]; ada {
			t.Fatalf("ID monotonic kembar: %s", id)
		}
		set[id] = struct{}{}
	}
}

func TestUrutMenaik(t *testing.T) {
	mulai := time.Date(2026, 7, 24, 10, 0, 0, 0, time.UTC)
	// Sengaja dibuat terbalik urutannya untuk memastikan UrutMenaik benar-benar mengurutkan.
	ids := []ulid.ULID{
		BaruPadaWaktu(mulai.Add(3 * time.Second)),
		BaruPadaWaktu(mulai.Add(1 * time.Second)),
		BaruPadaWaktu(mulai.Add(2 * time.Second)),
	}

	got := UrutMenaik(ids)
	if !slices.IsSorted(got) {
		t.Errorf("UrutMenaik menghasilkan %v yang tidak terurut", got)
	}
	if len(got) != 3 {
		t.Errorf("jumlah = %d, ingin 3", len(got))
	}
}
