package main

// ABAC: Attribute-Based Access Control. Keputusan berdasar ATRIBUT (subjek,
// sumber daya, aksi, konteks) — lebih fleksibel dari RBAC. Cocok untuk aturan
// seperti "pemilik boleh mengedit dokumennya sendiri" atau "hanya jam kerja".

type Subject struct {
	ID         string
	Role       string
	Department string
}

type Resource struct {
	OwnerID    string
	Department string
	Sensitive  bool
}

type AccessRequest struct {
	Subject  Subject
	Resource Resource
	Action   string // "read" | "write" | "delete"
}

// Policy = fungsi yang memutuskan izin dari atribut. Bisa dirangkai.
type Policy func(AccessRequest) bool

// contoh kebijakan yang bisa digabung:

// PolicyOwner: pemilik boleh melakukan apa pun pada sumber dayanya.
func PolicyOwner(req AccessRequest) bool {
	return req.Subject.ID == req.Resource.OwnerID
}

// PolicyAdmin: admin boleh apa pun.
func PolicyAdmin(req AccessRequest) bool {
	return req.Subject.Role == "admin"
}

// PolicySameDeptRead: satu departemen boleh MEMBACA (bukan menulis) yang non-sensitif.
func PolicySameDeptRead(req AccessRequest) bool {
	return req.Action == "read" &&
		!req.Resource.Sensitive &&
		req.Subject.Department == req.Resource.Department
}

// AnyOf: izinkan bila SALAH SATU policy mengizinkan.
func AnyOf(policies ...Policy) Policy {
	return func(req AccessRequest) bool {
		for _, p := range policies {
			if p(req) {
				return true
			}
		}
		return false
	}
}
