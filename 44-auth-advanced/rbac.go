// Modul 44 — Auth Advanced: RBAC, ABAC, OAuth2/OIDC flow, session.
// Lanjutan Modul 15 & 27. rbac.go = Role-Based Access Control.
package main

// 🔍 Analogi besar RBAC: seperti KARTU AKSES kantor berdasar JABATAN. "Editor" boleh masuk ruang
// tulis-artikel; "Admin" boleh masuk semua ruang termasuk kelola-user. Kamu tak memberi izin per
// orang satu-satu (ribet & rawan salah) — kamu beri orang sebuah PERAN, dan peran itu sudah membawa
// paket izinnya. Karyawan baru? cukup beri peran "Editor", langsung dapat semua izin editor. RBAC
// menjawab "kamu SIAPA (perannya apa)?". Cocok saat izin bisa dikelompokkan rapi per jabatan.
// RBAC: hak akses ditentukan oleh PERAN (role) yang dimiliki user. Peran punya
// sekumpulan izin (permission). Cek: "apakah user boleh melakukan X?".

type Permission string

const (
	PermReadArticle   Permission = "article:read"
	PermWriteArticle  Permission = "article:write"
	PermDeleteArticle Permission = "article:delete"
	PermManageUsers   Permission = "user:manage"
)

type RBAC struct {
	rolePerms map[string]map[Permission]bool // role -> set izin
	userRoles map[string][]string            // user -> daftar role
}

func NewRBAC() *RBAC {
	return &RBAC{
		rolePerms: make(map[string]map[Permission]bool),
		userRoles: make(map[string][]string),
	}
}

func (r *RBAC) DefineRole(role string, perms ...Permission) {
	set := make(map[Permission]bool, len(perms))
	for _, p := range perms {
		set[p] = true
	}
	r.rolePerms[role] = set
}

func (r *RBAC) AssignRole(user, role string) {
	r.userRoles[user] = append(r.userRoles[user], role)
}

// Can memeriksa apakah user punya izin tertentu (lewat salah satu perannya).
func (r *RBAC) Can(user string, perm Permission) bool {
	for _, role := range r.userRoles[user] {
		if r.rolePerms[role][perm] {
			return true
		}
	}
	return false
}
