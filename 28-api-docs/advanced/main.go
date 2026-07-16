// Teknik Advanced Modul 28 (API docs) — contoh runnable + penjelasan detail.
//
// Pendekatan SPEC-FIRST: definisikan skema, hasilkan dokumen OpenAPI, lalu
// VALIDASI request terhadap skema yang sama (dokumen = perilaku). Jalankan:
//
//	go run ./28-api-docs/advanced
package main

import (
	"encoding/json"
	"fmt"
)

// Skema field minimal (subset OpenAPI) — SATU sumber kebenaran untuk dokumen
// DAN validasi, sehingga keduanya tak pernah menyimpang (drift).
//
// 🔍 Analogi: satu sumber kebenaran itu seperti RESEP MASTER di dapur pusat —
// buku menu untuk pelanggan (dokumen OpenAPI) dan instruksi koki (validasi)
// dicetak dari resep yang SAMA. Kalau menu dan dapur ditulis terpisah,
// cepat atau lambat pelanggan memesan hidangan yang dapurnya sudah beda
// (dokumentasi drift).
type Field struct {
	Nama  string `json:"name"`
	Tipe  string `json:"type"`
	Wajib bool   `json:"required"`
}

type Endpoint struct {
	Method string  `json:"method"`
	Path   string  `json:"path"`
	Fields []Field `json:"-"` // dipakai internal utk validasi & bangun spec
}

func main() {
	ep := Endpoint{
		Method: "POST", Path: "/users",
		Fields: []Field{
			{"email", "string", true},
			{"nama", "string", true},
			{"umur", "integer", false},
		},
	}

	// =====================================================================
	// 1. Hasilkan dokumen OpenAPI dari skema (code -> spec)
	// =====================================================================
	fmt.Println("== 1. OpenAPI (dihasilkan dari skema) ==")
	fmt.Println(indent(buildOpenAPI(ep)))

	// =====================================================================
	// 2. Validasi request memakai SKEMA YANG SAMA (spec-driven validation)
	// =====================================================================
	// Karena validasi memakai definisi field yang sama dengan dokumen, mustahil
	// dokumen dan perilaku berbeda.
	//
	// 🔍 Analogi: spec-driven validation itu seperti PENJAGA GERBANG yang
	// membaca peraturan dari papan pengumuman YANG SAMA dengan yang dibaca
	// pengunjung — mustahil pengunjung mengikuti aturan yang dipajang tapi
	// tetap ditolak karena penjaga memakai aturan versi lain.
	fmt.Println("== 2. validasi request thd skema ==")
	cek(ep, map[string]any{"email": "a@b.com", "nama": "Budi", "umur": 20})          // valid
	cek(ep, map[string]any{"email": "a@b.com"})                                      // "nama" hilang
	cek(ep, map[string]any{"email": "a@b.com", "nama": "Budi", "umur": "dua puluh"}) // tipe salah

	// Catatan: di produksi pakai generator matang (swag/oapi-codegen) + middleware
	// validasi dari spec; versikan API (/v1) & umumkan deprecation via header Sunset.
}

func buildOpenAPI(ep Endpoint) map[string]any {
	props := map[string]any{}
	var required []string
	for _, f := range ep.Fields {
		props[f.Nama] = map[string]any{"type": f.Tipe}
		if f.Wajib {
			required = append(required, f.Nama)
		}
	}
	return map[string]any{
		"openapi": "3.0.0",
		"paths": map[string]any{
			ep.Path: map[string]any{
				lower(ep.Method): map[string]any{
					"requestBody": map[string]any{
						"content": map[string]any{
							"application/json": map[string]any{
								"schema": map[string]any{
									"type": "object", "properties": props, "required": required,
								},
							},
						},
					},
				},
			},
		},
	}
}

// cek memvalidasi payload: field wajib ada + tipe cocok. Kembalikan 200/400.
func cek(ep Endpoint, payload map[string]any) {
	var masalah []string
	for _, f := range ep.Fields {
		v, ada := payload[f.Nama]
		if !ada {
			if f.Wajib {
				masalah = append(masalah, "field wajib '"+f.Nama+"' hilang")
			}
			continue
		}
		if !tipeCocok(f.Tipe, v) {
			masalah = append(masalah, fmt.Sprintf("field '%s' harus %s", f.Nama, f.Tipe))
		}
	}
	if len(masalah) == 0 {
		fmt.Printf("  200 OK    %v\n", payload)
	} else {
		fmt.Printf("  400 Bad   %v -> %v\n", payload, masalah)
	}
}

func tipeCocok(tipe string, v any) bool {
	switch tipe {
	case "string":
		_, ok := v.(string)
		return ok
	case "integer":
		switch v.(type) {
		case int, int64, float64: // JSON number -> float64
			return true
		}
		return false
	}
	return true
}

func indent(v any) string { b, _ := json.MarshalIndent(v, "  ", "  "); return "  " + string(b) }
func lower(s string) string {
	out := []rune(s)
	for i, r := range out {
		if r >= 'A' && r <= 'Z' {
			out[i] = r + 32
		}
	}
	return string(out)
}
