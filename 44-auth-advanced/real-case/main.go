// REAL-CASE Modul 44 (auth) — otorisasi via OPA (Open Policy Agent) eksternal.
//
// Versi advanced/ menaruh aturan RBAC/ABAC sebagai if-else di kode Go. Di
// produksi, aturan sering dipindah ke POLICY ENGINE terpisah (OPA) yang ditulis
// dalam Rego — bisa diubah tanpa deploy ulang aplikasi, diuji & diaudit sendiri.
// Aplikasi hanya mengirim "input" (siapa, aksi, resource) dan menerima keputusan.
//
// Auto-skip bila OPA_URL kosong. Jalankan nyata:
//
//	docker compose -f 44-auth-advanced/real-case/docker-compose.yml up -d
//	OPA_URL=http://127.0.0.1:8181 go run ./44-auth-advanced/real-case
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

// decide mengirim input ke OPA dan mengembalikan keputusan boolean.
// Endpoint data OPA: POST /v1/data/<package>/<rule> dengan body {"input": {...}}.
func decide(ctx context.Context, baseURL string, input map[string]any) (bool, error) {
	body, _ := json.Marshal(map[string]any{"input": input})
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/v1/data/authz/allow", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	var out struct {
		Result bool `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return false, err
	}
	return out.Result, nil
}

func main() {
	base := os.Getenv("OPA_URL")
	if base == "" {
		fmt.Println("⏭️  DILEWATI: set OPA_URL untuk versi nyata.")
		fmt.Println("   docker compose -f 44-auth-advanced/real-case/docker-compose.yml up -d")
		fmt.Println("   OPA_URL=http://127.0.0.1:8181 go run ./44-auth-advanced/real-case")
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Aturan ada di policies/authz.rego (di-load OPA). Aplikasi tak tahu detail
	// aturan — hanya mengirim fakta & memakai keputusan.
	skenario := []map[string]any{
		{"user": "budi", "role": "editor", "action": "tulis", "owner": "budi"},
		{"user": "budi", "role": "editor", "action": "hapus", "owner": "ani"},
		{"user": "cici", "role": "admin", "action": "hapus", "owner": "ani"},
		{"user": "ani", "role": "viewer", "action": "hapus", "owner": "ani"},
	}

	fmt.Println("== otorisasi via OPA (Rego) ==")
	for _, s := range skenario {
		ok, err := decide(ctx, base, s)
		if err != nil {
			panic("gagal panggil OPA: " + err.Error())
		}
		fmt.Printf("  user=%s role=%s action=%s owner=%s -> allow=%v\n",
			s["user"], s["role"], s["action"], s["owner"], ok)
	}
	fmt.Println("  (ubah policies/authz.rego tanpa deploy ulang aplikasi)")
}
