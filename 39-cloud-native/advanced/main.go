// Teknik Advanced Modul 39 (cloud native) — contoh runnable + penjelasan detail.
//
// Pola RECONCILE LOOP ala Kubernetes controller: konvergenkan actual -> desired
// secara IDEMPOTEN & level-triggered. Jalankan:
//
//	go run ./39-cloud-native/advanced
package main

import "fmt"

// Desired = keadaan yang DIINGINKAN (deklaratif, mis. dari manifest/CRD).
type Desired struct{ Replicas int }

// Cluster = keadaan AKTUAL yang diamati.
type Cluster struct{ running int }

// ===========================================================================
// reconcile menghitung selisih desired vs actual lalu bertindak untuk menutup
// gap. Sifat penting:
//   - LEVEL-TRIGGERED: bekerja dari SELISIH keadaan, bukan dari event tunggal.
//     Jadi tahan terhadap event yang hilang/ganda.
//   - IDEMPOTEN: bila sudah sesuai, memanggilnya lagi TIDAK melakukan apa-apa.
//
// Inilah inti semua controller/operator Kubernetes.
// ===========================================================================
func reconcile(c *Cluster, d Desired) (aksi int) {
	for c.running < d.Replicas { // kurang -> scale up
		c.running++
		fmt.Printf("  + start pod (%d/%d)\n", c.running, d.Replicas)
		aksi++
	}
	for c.running > d.Replicas { // lebih -> scale down
		c.running--
		fmt.Printf("  - stop pod (%d/%d)\n", c.running, d.Replicas)
		aksi++
	}
	return aksi
}

func main() {
	c := &Cluster{running: 0}

	fmt.Println("== reconcile #1: desired=3 (dari 0) ==")
	reconcile(c, Desired{Replicas: 3})

	fmt.Println("== reconcile #2: desired=3 (sudah sesuai) -> IDEMPOTEN ==")
	if n := reconcile(c, Desired{Replicas: 3}); n == 0 {
		fmt.Println("  tidak ada aksi (actual == desired)")
	}

	fmt.Println("== reconcile #3: desired=1 -> scale down ==")
	reconcile(c, Desired{Replicas: 1})

	fmt.Println("== reconcile #4: satu pod 'mati' tak terduga -> loop menyembuhkan ==")
	c.running = 0 // simulasi drift: pod hilang di luar kendali kita
	reconcile(c, Desired{Replicas: 1})

	// Catatan ekosistem: paketkan dengan HELM (template chart + values per-env),
	// jalankan sebagai operator (CRD + controller-runtime), atau sebagai fungsi
	// SERVERLESS (stateless, waspada cold-start). Selalu hormati SIGTERM & grace
	// period saat di K8s (lihat modul 20 & 30).
}
