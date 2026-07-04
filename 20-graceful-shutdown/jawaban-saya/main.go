// Workspace latihan Modul 20 — graceful-shutdown.
// Tulis JAWABANMU di sini. Jalankan: go run ./20-graceful-shutdown/jawaban-saya
//
// Untuk latihan yang MENGUBAH kode modul (mis. menambah endpoint/metrik),
// salin file terkait ke folder ini lalu ubah, ATAU ubah langsung di modul &
// catat perubahanmu di komentar bawah. Bandingkan dgn "Solusi Latihan" di README.
//
// Checklist latihan:
//
//	[ ] 1. db.Close() setelah Shutdown
//	[ ] 2. Background worker (WaitGroup) ditunggu saat shutdown
//	[ ] 3. Terapkan ke server Fiber Modul 13
//	[ ] 4. Readiness /readyz -> 503 setelah sinyal
//	[ ] 5. Uji dengan kill -TERM <pid>
package main

import "fmt"

func main() {
	fmt.Println("TODO: kerjakan latihan Modul 20 di sini")
}
