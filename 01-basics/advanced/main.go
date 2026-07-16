// Teknik Advanced Modul 01 (basics) — contoh runnable + penjelasan detail.
//
// Folder ini PACKAGE main terpisah dari materi utama. Jalankan:
//
//	go run ./01-basics/advanced
//
// Tiap bagian di bawah mendemonstrasikan satu teknik "dasar tapi sering luput".
package main

import "fmt"

// ===========================================================================
// 1. iota — generator konstanta berurut untuk ENUM
// ===========================================================================
// Di dalam SATU blok const, iota mulai dari 0 dan bertambah 1 tiap baris.
// Ini cara idiomatik membuat enum tanpa menulis angka manual (dan tanpa risiko
// salah ketik nomor). Menambah anggota baru cukup menyisipkan baris.
//
// 🔍 Analogi: iota itu seperti MESIN NOMOR ANTRIAN — tiap baris const otomatis
// "narik tiket" nomor berikutnya. Mau sisip peserta baru di tengah? Tinggal
// tambah baris, semua nomor di bawahnya bergeser sendiri tanpa diketik ulang.

type Level int

const (
	Debug    Level = iota // 0
	Info                  // 1  (baris tanpa ekspresi meneruskan "Level = iota")
	Warn                  // 2
	ErrorLvl              // 3  (dinamai ErrorLvl, bukan Error, agar tak rancu dgn interface error)
)

// String() membuat Level memenuhi fmt.Stringer. fmt otomatis memanggilnya saat
// mencetak, sehingga kita lihat "INFO" bukan angka 1 — jauh lebih mudah dibaca.
func (l Level) String() string {
	switch l {
	case Debug:
		return "DEBUG"
	case Info:
		return "INFO"
	case Warn:
		return "WARN"
	case ErrorLvl:
		return "ERROR"
	default:
		return fmt.Sprintf("Level(%d)", int(l))
	}
}

// ===========================================================================
// 2. iota + geser bit — BITMASK (satu integer menampung banyak flag on/off)
// ===========================================================================
// 1 << iota menghasilkan 1, 2, 4, 8, ... (satu bit menyala per konstanta).
// Beberapa flag digabung dengan OR (|), dicek dengan AND (&). Pola ini dipakai
// os.OpenFile (O_RDONLY|O_CREATE), izin file Unix, dsb.
//
// 🔍 Analogi: bitmask itu seperti PANEL SAKELAR LAMPU di satu papan — tiap bit
// adalah satu sakelar on/off. OR (|) = menyalakan beberapa sakelar sekaligus,
// AND (&) = mengintip apakah sakelar tertentu sedang menyala.

type Perm uint8

const (
	PRead  Perm = 1 << iota // 1 -> 0b001
	PWrite                  // 2 -> 0b010
	PExec                   // 4 -> 0b100
)

// Has mengecek apakah sebuah flag menyala di dalam himpunan izin.
func (p Perm) Has(flag Perm) bool { return p&flag != 0 }

// ===========================================================================
// 3. Named return value + defer — memodifikasi hasil saat keluar
// ===========================================================================
// Return bernama (hasil, err) sudah jadi variabel di dalam fungsi. defer bisa
// mengubahnya SEBELUM benar-benar dikembalikan. Pola ini dipakai untuk
// mengubah panic menjadi error yang rapi di batas fungsi.
//
// 🔍 Analogi: named return + defer itu seperti PETUGAS QC DI PINTU KELUAR
// pabrik — barang (nilai return) sudah dikemas, tapi sebelum benar-benar
// keluar pintu, QC masih boleh menempel label "cacat" (mengubah err).

func bagiAman(a, b int) (hasil int, err error) {
	defer func() {
		// recover() menangkap panic (mis. pembagian nol) dan kita ubah jadi err.
		if r := recover(); r != nil {
			err = fmt.Errorf("pulih dari panic: %v", r)
		}
	}()
	hasil = a / b // b==0 -> panic -> ditangkap defer di atas
	return        // "naked return": mengembalikan (hasil, err) apa adanya
}

// ===========================================================================
// 4. Labeled break/continue — keluar dari loop bersarang
// ===========================================================================
// Tanpa label, break hanya keluar dari loop terdalam. Label memungkinkan
// keluar langsung dari loop luar — lebih bersih daripada variabel bendera.
//
// 🔍 Analogi: break biasa itu seperti keluar dari SATU RUANGAN saja di gedung
// bertingkat; labeled break seperti TANGGA DARURAT — sekali lompat langsung
// keluar gedung (loop terluar) tanpa melewati tiap lantai satu per satu.

func cariPasangan(target int) (int, int, bool) {
	for i := 1; i <= 9; i++ {
	Inner:
		for j := 1; j <= 9; j++ {
			switch {
			case i*j == target:
				return i, j, true
			case i*j > target:
				break Inner // lompat ke iterasi i berikutnya, bukan sekadar 1 level
			}
		}
	}
	return 0, 0, false
}

// ===========================================================================
// 5. Variable shadowing — jebakan := di scope baru
// ===========================================================================
// := membuat variabel BARU di scope tempat ia ditulis. Bila namanya sama
// dengan variabel luar, yang luar "terbayangi" dan perubahan tak terlihat di
// luar. `go vet` dengan analyzer shadow bisa mendeteksi ini.
//
// 🔍 Analogi: shadowing itu seperti DUA ORANG BERNAMA SAMA di rapat — di dalam
// ruangan kecil kamu memanggil "Budi" yang duduk di situ, bukan Budi di lobi.
// Apa pun yang kamu titipkan ke Budi-dalam tidak pernah sampai ke Budi-luar.

func demoShadowing() {
	x := 1
	{
		x := 100 // variabel BARU, hanya hidup di blok ini (shadow x luar)
		x++
		fmt.Println("  x di dalam blok:", x) // 101
	}
	fmt.Println("  x di luar blok :", x) // tetap 1 — perubahan tadi tak bocor
}

func main() {
	fmt.Println("== 1. iota enum ==")
	for _, l := range []Level{Debug, Info, Warn, ErrorLvl} {
		fmt.Printf("  %d -> %s\n", int(l), l)
	}

	fmt.Println("== 2. bitmask ==")
	izin := PRead | PWrite // gabung dua flag
	fmt.Printf("  izin=0b%03b  bisa baca=%v  bisa eksekusi=%v\n",
		izin, izin.Has(PRead), izin.Has(PExec))

	fmt.Println("== 3. named return + defer recover ==")
	if _, err := bagiAman(10, 0); err != nil {
		fmt.Println("  bagi 10/0 aman:", err)
	}
	h, _ := bagiAman(10, 2)
	fmt.Println("  bagi 10/2 =", h)

	fmt.Println("== 4. labeled break ==")
	if i, j, ok := cariPasangan(12); ok {
		fmt.Printf("  %d x %d = 12\n", i, j)
	}

	fmt.Println("== 5. shadowing ==")
	demoShadowing()
}
