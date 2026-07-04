package main

import (
	"encoding/json"
	"strings"
	"testing"
)

// Test demonstrasi untuk Modul 09 — menguji perilaku tag JSON pada struct User.
// Jalankan: go test -v ./09-stdlib
//
// Ini menegaskan kontrak yang mudah rusak tanpa sadar: nama field di JSON,
// omitempty, dan field unexported yang tak boleh bocor.

func TestUserMarshalTags(t *testing.T) {
	u := User{ID: 1, Name: "Ana", Admin: true, pass: "rahasia"}
	b, err := json.Marshal(u)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}
	got := string(b)

	// is_admin: tag mengganti nama field.
	if !strings.Contains(got, `"is_admin":true`) {
		t.Errorf("harusnya ada \"is_admin\":true, dapat: %s", got)
	}
	// email kosong + omitempty -> tak muncul.
	if strings.Contains(got, "email") {
		t.Errorf("email kosong (omitempty) tak boleh muncul, dapat: %s", got)
	}
	// field unexported 'pass' tak boleh bocor ke JSON.
	if strings.Contains(strings.ToLower(got), "rahasia") {
		t.Errorf("field unexported bocor ke JSON! dapat: %s", got)
	}
}

func TestUserUnmarshal(t *testing.T) {
	var u User
	in := `{"id":2,"name":"Budi","email":"budi@mail.id","is_admin":false}`
	if err := json.Unmarshal([]byte(in), &u); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}
	if u.ID != 2 || u.Name != "Budi" || u.Email != "budi@mail.id" {
		t.Errorf("hasil Unmarshal tak sesuai: %+v", u)
	}
}

// Round-trip: marshal lalu unmarshal harus mengembalikan data yang setara
// (untuk field yang diekspor).
func TestUserRoundTrip(t *testing.T) {
	orig := User{ID: 9, Name: "Cici", Email: "cici@mail.id", Admin: true}
	b, _ := json.Marshal(orig)
	var back User
	if err := json.Unmarshal(b, &back); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}
	if back != orig {
		t.Errorf("round-trip berubah: %+v -> %+v", orig, back)
	}
}
