package main

import (
	"net/http/httptest"
	"testing"
)

func TestRBAC(t *testing.T) {
	rbac := NewRBAC()
	rbac.DefineRole("editor", PermReadArticle, PermWriteArticle)
	rbac.DefineRole("admin", PermReadArticle, PermWriteArticle, PermDeleteArticle)
	rbac.AssignRole("ana", "editor")
	rbac.AssignRole("c: root", "admin")

	if !rbac.Can("ana", PermWriteArticle) {
		t.Error("editor harus boleh write")
	}
	if rbac.Can("ana", PermDeleteArticle) {
		t.Error("editor TAK boleh delete")
	}
	if !rbac.Can("c: root", PermDeleteArticle) {
		t.Error("admin harus boleh delete")
	}
	if rbac.Can("nobody", PermReadArticle) {
		t.Error("user tanpa role tak boleh apa-apa")
	}
}

func TestABAC(t *testing.T) {
	policy := AnyOf(PolicyOwner, PolicyAdmin, PolicySameDeptRead)
	ana := Subject{ID: "ana", Role: "member", Department: "eng"}

	// Pemilik -> boleh write.
	if !policy(AccessRequest{ana, Resource{OwnerID: "ana"}, "write"}) {
		t.Error("pemilik harus boleh write")
	}
	// Non-pemilik, sesama dept, read non-sensitif -> boleh.
	if !policy(AccessRequest{ana, Resource{OwnerID: "x", Department: "eng"}, "read"}) {
		t.Error("sesama dept harus boleh read non-sensitif")
	}
	// Non-pemilik write -> ditolak.
	if policy(AccessRequest{ana, Resource{OwnerID: "x", Department: "eng"}, "write"}) {
		t.Error("non-pemilik tak boleh write")
	}
	// Dokumen sensitif -> read ditolak (bukan pemilik/admin).
	if policy(AccessRequest{ana, Resource{OwnerID: "x", Department: "eng", Sensitive: true}, "read"}) {
		t.Error("dokumen sensitif tak boleh dibaca sesama dept")
	}
	// Admin -> boleh apa pun.
	admin := Subject{ID: "boss", Role: "admin"}
	if !policy(AccessRequest{admin, Resource{OwnerID: "x", Sensitive: true}, "delete"}) {
		t.Error("admin harus boleh")
	}
}

func TestOAuthFlow(t *testing.T) {
	provider := NewMockProvider()
	srv := httptest.NewServer(provider.Handler())
	defer srv.Close()

	// Code valid -> token.
	token, err := ExchangeCode(srv.URL+"/token", "my-app", "s3cret", "auth-code-xyz", "http://cb")
	if err != nil {
		t.Fatalf("exchange: %v", err)
	}
	if token == "" {
		t.Fatal("token kosong")
	}

	// Token -> profil.
	info, err := FetchUserInfo(srv.URL+"/userinfo", token)
	if err != nil {
		t.Fatalf("userinfo: %v", err)
	}
	if info["email"] != "ana@mail.id" {
		t.Errorf("email = %v; want ana@mail.id", info["email"])
	}
}

func TestOAuthCodeSalah(t *testing.T) {
	provider := NewMockProvider()
	srv := httptest.NewServer(provider.Handler())
	defer srv.Close()

	// Code salah -> gagal.
	if _, err := ExchangeCode(srv.URL+"/token", "my-app", "s3cret", "code-palsu", "http://cb"); err == nil {
		t.Error("code salah harusnya gagal")
	}
	// Secret salah -> gagal.
	if _, err := ExchangeCode(srv.URL+"/token", "my-app", "salah", "auth-code-xyz", "http://cb"); err == nil {
		t.Error("secret salah harusnya gagal")
	}
}
