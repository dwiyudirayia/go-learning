# 44 — Auth Advanced (RBAC, ABAC, OAuth2)

Lanjutan Modul 15 (JWT) & 27 (security). Otorisasi tingkat lanjut: **RBAC**, **ABAC**, dan **OAuth2/OIDC** (login "with Google/GitHub").

Jalankan:
```bash
go run ./44-auth-advanced
go test ./44-auth-advanced
```

## Autentikasi vs Otorisasi
- **Autentikasi** (Modul 15): *siapa* kamu? (login → JWT).
- **Otorisasi** (modul ini): *boleh apa* kamu? (RBAC/ABAC).

## 1. RBAC — Role-Based Access Control

Hak akses ditentukan **peran**. Peran punya sekumpulan izin.
```
user ──punya──► role(s) ──punya──► permission(s)
ana ─► editor ─► {article:read, article:write}
```
```go
rbac.DefineRole("editor", PermReadArticle, PermWriteArticle)
rbac.AssignRole("ana", "editor")
rbac.Can("ana", PermWriteArticle)  // true
rbac.Can("ana", PermDeleteArticle) // false
```
Sederhana & umum. Cocok saat hak akses berpola tetap (viewer/editor/admin).

## 2. ABAC — Attribute-Based Access Control

Keputusan berdasar **atribut** (subjek, resource, aksi, konteks) — lebih fleksibel dari RBAC. Cocok untuk aturan seperti *"pemilik boleh edit dokumennya"* atau *"hanya jam kerja"*.
```go
policy := AnyOf(PolicyOwner, PolicyAdmin, PolicySameDeptRead)
policy(AccessRequest{Subject: ana, Resource: doc, Action: "write"})
```
Test membuktikan: pemilik boleh write, sesama dept boleh read non-sensitif, tapi non-pemilik tak boleh write & dokumen sensitif terlindungi.

| | RBAC | ABAC |
|-|------|------|
| Basis | peran | atribut/kebijakan |
| Fleksibilitas | rendah | tinggi |
| Kompleksitas | rendah | sedang–tinggi |
| Cocok | hak akses berpola | aturan kontekstual/dinamis |

Banyak sistem memakai **keduanya**: RBAC untuk kasar, ABAC untuk aturan halus (kepemilikan, tenant, waktu).

## 3. OAuth2 / OIDC — "Login with X"

Alur **Authorization Code** (yang paling aman untuk web):
```
1. User -> redirect ke provider (Google) -> login & setuju
2. Provider -> redirect balik ke app-mu dengan ?code=...
3. Backend-mu: tukar code + client_secret -> ACCESS TOKEN   (server-to-server, rahasia)
4. Backend-mu: pakai token -> ambil profil user (/userinfo)
```
Modul ini mengimplementasikan **langkah 3 & 4** (sisi backend) + provider yang di-mock (`httptest`):
```go
token, _ := ExchangeCode(tokenURL, clientID, clientSecret, code, redirectURI)
info, _  := FetchUserInfo(userinfoURL, token) // {sub, email, name}
```
- **OIDC** = OAuth2 + lapisan identitas (ID token berformat JWT berisi `sub`, `email`).
- **PKCE** wajib untuk SPA/mobile (client tanpa secret) — tambahan `code_verifier`/`code_challenge`.
- Produksi: pakai library [`golang.org/x/oauth2`](https://pkg.go.dev/golang.org/x/oauth2) + `coreos/go-oidc`.

## 4. Session vs JWT (sekilas)
- **JWT** (Modul 15): stateless, disimpan di client, tak bisa dicabut sebelum expiry.
- **Session**: state di server (Redis/DB), session ID di cookie, **bisa dicabut** kapan saja. Cocok untuk web tradisional & logout instan.

## Multi-tenancy (sekilas)
Satu aplikasi melayani banyak organisasi (tenant). Setiap data diberi `tenant_id`; otorisasi wajib mengecek `subject.tenant == resource.tenant` (bisa jadi salah satu Policy ABAC). Cegah kebocoran data antar-tenant.

## Kapan & Di Mana Dipakai
- SaaS multi-tenant, aplikasi dengan peran beragam, login sosial (Google/GitHub), API B2B.

## Latihan
1. Tambah `PolicyBusinessHours` (izin hanya jam 9–17) & gabung ke ABAC.
2. Tambah middleware Fiber (Modul 13) yang menolak request bila `rbac.Can` false.
3. Tambah `tenant_id` ke Subject/Resource + policy multi-tenancy.
4. Implementasikan session store (Redis, Modul 22) + logout (hapus session).
5. Integrasikan `golang.org/x/oauth2` dengan provider GitHub sungguhan.

## ✅ Solusi Latihan (Pembahasan)

1. **`PolicyBusinessHours`** — policy ABAC yang mengembalikan izin hanya bila `time.Now().Hour()` di 9..17. Gabung dengan policy lain (semua harus lolos = AND).
2. **Middleware Fiber** — `if !rbac.Can(subject, action, resource) { return c.SendStatus(403) }` sebelum handler (Modul 13). Otorisasi terpusat.
3. **`tenant_id` multi-tenancy** — tambahkan `TenantID` ke `Subject` & `Resource`; policy tolak bila `subject.TenantID != resource.TenantID`. Isolasi data antar-tenant.
4. **Session store Redis** — simpan session di Redis (Modul 22) dengan TTL; logout = `DEL session:<id>`. Lebih mudah revoke dibanding JWT stateless.
5. **OAuth2 GitHub nyata** — `oauth2.Config{ClientID, ClientSecret, Endpoint: github.Endpoint, Scopes:["read:user"]}`; alur: redirect → callback tukar `code` jadi token → panggil API GitHub. (Modul memakai mock provider via httptest agar bisa diuji tanpa kredensial.)

---

## 🚀 Teknik Advanced (Level Up)
> 💻 **Contoh runnable + komentar detail** untuk teknik di bawah ada di folder [`advanced/`](advanced). Jalankan: `go run ./44-auth-advanced/advanced`


- **Model otorisasi** — **RBAC** (peran), **ABAC** (atribut/policy), **ReBAC** (relasi, ala Zanzibar/SpiceDB). Pilih sesuai kompleksitas.
- **OAuth2 / OIDC** — Authorization Code + **PKCE** untuk klien publik; OIDC tambah identitas (ID token). Jangan pakai implicit flow (usang).
- **JWT vs opaque token** — JWT stateless (sulit di-revoke) vs opaque (butuh introspection, mudah revoke). Kombinasi: access JWT pendek + refresh dengan rotasi.
- **Policy engine** — OPA/Rego atau Casbin untuk aturan otorisasi terpusat & teruji, bukan `if role==...` tersebar.
- **Session & refresh rotation** — deteksi reuse token (indikasi pencurian) → revoke seluruh keluarga.
- **Least privilege & audit** — cakupan minimal + log keputusan otorisasi.
