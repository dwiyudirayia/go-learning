package main

import "testing"

// TestRBAC menguji keputusan role-based (tabel role x aksi -> boleh?).
func TestRBAC(t *testing.T) {
	kasus := []struct {
		role, aksi string
		want       bool
	}{
		{"admin", "hapus", true},   // admin boleh apa saja
		{"editor", "baca", true},   // editor: baca & tulis
		{"editor", "tulis", true},  //
		{"editor", "hapus", false}, // editor TIDAK boleh hapus
		{"viewer", "baca", true},   // viewer hanya baca
		{"viewer", "tulis", false}, //
		{"tamu", "baca", false},    // role tak dikenal -> tolak
	}
	for _, k := range kasus {
		if got := rbacBoleh(k.role, k.aksi); got != k.want {
			t.Errorf("rbacBoleh(%q,%q) = %v, mau %v", k.role, k.aksi, got, k.want)
		}
	}
}

// TestABAC menguji keputusan attribute-based (kepemilikan + peran admin).
func TestABAC(t *testing.T) {
	dok := Dokumen{Pemilik: "ani"}
	kasus := []struct {
		nama string
		subj Subjek
		want bool
	}{
		{"pemilik boleh", Subjek{User: "ani", Role: "viewer"}, true},
		{"admin boleh walau bukan pemilik", Subjek{User: "cici", Role: "admin"}, true},
		{"bukan pemilik & bukan admin -> tolak", Subjek{User: "budi", Role: "editor"}, false},
	}
	for _, k := range kasus {
		if got := abacBolehEdit(k.subj, dok); got != k.want {
			t.Errorf("%s: abacBolehEdit(%+v) = %v, mau %v", k.nama, k.subj, got, k.want)
		}
	}
}
