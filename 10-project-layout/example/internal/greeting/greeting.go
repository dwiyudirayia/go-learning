// Package greeting berisi logika bisnis (privat karena berada di internal/).
// Hanya kode di dalam .../example/ yang boleh meng-import package ini —
// aturan ini DIPAKSA oleh compiler Go untuk semua path yang mengandung /internal/.
package greeting

import "strings"

// Greet menghasilkan sapaan. shout=true membuatnya huruf kapital.
func Greet(name string, shout bool) string {
	if name == "" {
		name = "dunia"
	}
	msg := "Halo, " + name + "!"
	if shout {
		msg = strings.ToUpper(msg)
	}
	return msg
}
