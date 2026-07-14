// 🔍 Analogi: folder internal/ itu seperti area "KHUSUS KARYAWAN" di toko. Barang di
// dalamnya hanya boleh dipakai oleh kode di rumah yang sama (example/), tak bisa diambil
// tetangga (proyek/modul lain). Ini DIPAKSA compiler — bukan sekadar sopan santun. Gunanya:
// menyembunyikan detail dapur agar bebas kamu ubah tanpa merusak pengguna dari luar.
// 🔍 Analogi: cmd/ = "PINTU DEPAN" aplikasi (titik start / main), internal/ = "DAPUR" (logika
// inti). Pola ini memisahkan "cara dijalankan" dari "apa yang dikerjakan" — rapi & scalable.

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
