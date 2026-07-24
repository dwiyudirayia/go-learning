package main

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestParseUUID(t *testing.T) {
	tests := []struct {
		nama    string
		input   string
		wantErr bool
	}{
		{"format baku dengan tanda hubung", "550e8400-e29b-41d4-a716-446655440000", false},
		{"tanpa tanda hubung tetap diterima", "550e8400e29b41d4a716446655440000", false},
		{"format URN diterima", "urn:uuid:550e8400-e29b-41d4-a716-446655440000", false},
		{"teks sembarang ditolak", "bukan-uuid", true},
		{"string kosong ditolak", "", true},
		{"kurang satu karakter ditolak", "550e8400-e29b-41d4-a716-44665544000", true},
	}

	for _, tt := range tests {
		t.Run(tt.nama, func(t *testing.T) {
			_, err := ParseUUID(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("ParseUUID(%q) = nil error, ingin error", tt.input)
				}
				// Sentinel harus bisa dikenali lewat errors.Is walau sudah dibungkus.
				if !errors.Is(err, ErrUUIDTidakValid) {
					t.Errorf("error tidak membungkus ErrUUIDTidakValid: %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseUUID(%q) error tak terduga: %v", tt.input, err)
			}
		})
	}
}

func TestIsValid(t *testing.T) {
	if !IsValid(uuid.New().String()) {
		t.Error("UUID hasil uuid.New() seharusnya valid")
	}
	if IsValid("halo dunia") {
		t.Error("teks sembarang seharusnya tidak valid")
	}
}

func TestVersionOf(t *testing.T) {
	v4 := NewV4()
	if got, err := VersionOf(v4); err != nil || got != 4 {
		t.Errorf("VersionOf(v4) = %d, %v; ingin 4, nil", got, err)
	}

	v7, err := NewV7()
	if err != nil {
		t.Fatalf("NewV7 gagal: %v", err)
	}
	if got, err := VersionOf(v7); err != nil || got != 7 {
		t.Errorf("VersionOf(v7) = %d, %v; ingin 7, nil", got, err)
	}
}

// Inti nilai jual v7: ID yang dibuat belakangan selalu lebih besar secara leksikografis.
func TestV7SelaluUrut(t *testing.T) {
	ids, err := GenerateV7Batch(50)
	if err != nil {
		t.Fatalf("GenerateV7Batch gagal: %v", err)
	}
	if !IsSorted(ids) {
		t.Error("UUID v7 seharusnya urut menaik sesuai waktu pembuatan")
	}
}

func TestIsNil(t *testing.T) {
	var kosong uuid.UUID
	if !IsNil(kosong) {
		t.Error("zero value UUID seharusnya dianggap Nil")
	}
	if IsNil(uuid.New()) {
		t.Error("UUID hasil generate seharusnya tidak Nil")
	}
}

func TestHitungUnik(t *testing.T) {
	a, b := uuid.New(), uuid.New()
	if got := HitungUnik([]uuid.UUID{a, b, a, b, a}); got != 2 {
		t.Errorf("HitungUnik = %d, ingin 2", got)
	}
	if got := HitungUnik(nil); got != 0 {
		t.Errorf("HitungUnik(nil) = %d, ingin 0", got)
	}
}

// MustParse panic pada input kotor — pembungkus kita harus mengubahnya jadi error.
func TestMustParseAmanMenangkapPanic(t *testing.T) {
	if _, err := MustParseAman("jelas-bukan-uuid"); err == nil {
		t.Error("ingin error dari input tidak valid, dapat nil")
	}
	if _, err := MustParseAman(uuid.New().String()); err != nil {
		t.Errorf("input valid seharusnya tanpa error, dapat: %v", err)
	}
}
