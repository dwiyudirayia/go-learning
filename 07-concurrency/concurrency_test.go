package main

import (
	"testing"
	"time"
)

// Test demonstrasi untuk Modul 07 — concurrency.
// Jalankan: go test -race -v ./07-concurrency
//
// Selalu jalankan test concurrency dengan flag -race untuk mendeteksi data race.

// generate menghasilkan 1..n lewat channel; square mengkuadratkan tiap nilai.
// Digabung, keduanya membentuk pipeline deterministik yang mudah diuji.
func TestGenerateSquarePipeline(t *testing.T) {
	var got []int
	for v := range square(generate(5)) {
		got = append(got, v)
	}
	want := []int{1, 4, 9, 16, 25}
	if len(got) != len(want) {
		t.Fatalf("panjang hasil = %d, mau %d (%v)", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("hasil[%d] = %d, mau %d", i, got[i], want[i])
		}
	}
}

func TestGenerateJumlahNilai(t *testing.T) {
	count := 0
	for range generate(100) {
		count++
	}
	if count != 100 {
		t.Errorf("generate(100) menghasilkan %d nilai, mau 100", count)
	}
}

// Pipeline harus selesai cepat — kalau menggantung (channel tak ditutup),
// test ini akan gagal lewat timeout, bukan hang selamanya.
func TestPipelineSelesai(t *testing.T) {
	done := make(chan struct{})
	go func() {
		for range square(generate(10)) {
		}
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("pipeline tidak selesai dalam 2 detik (channel mungkin tak ditutup)")
	}
}
