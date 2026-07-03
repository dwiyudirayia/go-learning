// Jalankan: go run ./44-auth-advanced
// Verifikasi otomatis: go test ./44-auth-advanced
package main

import (
	"fmt"
	"net/http/httptest"
)

func main() {
	fmt.Println("=== 44 — Auth Advanced ===")
	demoRBAC()
	demoABAC()
	demoOAuth()
}

func demoRBAC() {
	fmt.Println("\n-- RBAC (role-based) --")
	rbac := NewRBAC()
	rbac.DefineRole("viewer", PermReadArticle)
	rbac.DefineRole("editor", PermReadArticle, PermWriteArticle)
	rbac.DefineRole("admin", PermReadArticle, PermWriteArticle, PermDeleteArticle, PermManageUsers)

	rbac.AssignRole("ana", "editor")
	rbac.AssignRole("budi", "viewer")

	fmt.Printf("ana (editor) boleh tulis?  %t\n", rbac.Can("ana", PermWriteArticle))
	fmt.Printf("ana (editor) boleh hapus?  %t\n", rbac.Can("ana", PermDeleteArticle))
	fmt.Printf("budi (viewer) boleh tulis? %t\n", rbac.Can("budi", PermWriteArticle))
}

func demoABAC() {
	fmt.Println("\n-- ABAC (attribute-based) --")
	// Kebijakan: boleh jika pemilik ATAU admin ATAU sesama departemen (read saja).
	policy := AnyOf(PolicyOwner, PolicyAdmin, PolicySameDeptRead)

	ana := Subject{ID: "ana", Role: "member", Department: "eng"}
	doc := Resource{OwnerID: "ana", Department: "eng"}

	fmt.Printf("ana edit dokumennya sendiri? %t\n", policy(AccessRequest{ana, doc, "write"}))
	other := Resource{OwnerID: "budi", Department: "eng"}
	fmt.Printf("ana read dokumen tim (non-sensitif)? %t\n", policy(AccessRequest{ana, other, "read"}))
	fmt.Printf("ana write dokumen orang lain? %t\n", policy(AccessRequest{ana, other, "write"}))
}

func demoOAuth() {
	fmt.Println("\n-- OAuth2 (authorization code flow) --")
	provider := NewMockProvider()
	srv := httptest.NewServer(provider.Handler())
	defer srv.Close()

	// Langkah 2: tukar code -> token.
	token, err := ExchangeCode(srv.URL+"/token", "my-app", "s3cret", "auth-code-xyz", "http://localhost/callback")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Printf("dapat access token: %s\n", token)

	// Langkah 3: ambil profil user.
	info, _ := FetchUserInfo(srv.URL+"/userinfo", token)
	fmt.Printf("profil user: email=%v name=%v\n", info["email"], info["name"])
}
