# Policy otorisasi (RBAC + ABAC) untuk OPA.
# Path keputusan: data.authz.allow
package authz

import rego.v1

default allow := false

# RBAC: admin boleh apa saja.
allow if input.role == "admin"

# RBAC: editor boleh baca & tulis (tapi tidak hapus).
allow if {
	input.role == "editor"
	input.action in {"baca", "tulis"}
}

# RBAC: viewer hanya boleh baca.
allow if {
	input.role == "viewer"
	input.action == "baca"
}

# ABAC: pemilik resource boleh aksi apa pun atas miliknya sendiri.
allow if input.user == input.owner
