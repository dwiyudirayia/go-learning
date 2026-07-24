// spf13/cobra — membangun aplikasi CLI berstruktur (perintah, sub-perintah, flag).
//
// Jalankan: go run ./libraries/cobra tugas tambah "beli susu" --prioritas tinggi
//
//	go run ./libraries/cobra tugas daftar
//	go run ./libraries/cobra --help
//
// Test:     go test ./libraries/cobra
//
// 🔍 Analogi besar: aplikasi CLI yang rumit itu seperti KANTOR PEMERINTAHAN. Tanpa cobra,
// kamu punya satu loket yang menangani semua urusan lewat tumpukan if-else membaca
// os.Args — antre panjang, petugasnya bingung, dan tak ada papan petunjuk.
//
// cobra menyusunnya jadi POHON LOKET: gedung utama (perintah root) punya beberapa lantai
// (sub-perintah), tiap lantai punya loket sendiri. Persis seperti `git`: "git" adalah
// gedungnya, "git commit" lantainya, "-m" formulir di loket itu. Bonusnya, papan petunjuk
// (--help) dibuatkan otomatis di setiap tingkat.
package main

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"sync"

	"github.com/spf13/cobra"
)

func main() {
	// Kesalahan sudah dicetak cobra sendiri; di sini cukup atur kode keluar.
	if err := NewRootCmd(NewPenyimpanan()).Execute(); err != nil {
		os.Exit(1)
	}
}

// ------------------------------------------------------------------
// Data yang dikelola CLI
// ------------------------------------------------------------------

// Tugas satu butir pekerjaan.
type Tugas struct {
	ID        int
	Judul     string
	Prioritas string
	Selesai   bool
}

// Penyimpanan menyimpan tugas di memori.
//
// 🔍 Analogi: sengaja dibuat sebagai objek yang DISUNTIKKAN ke perintah, bukan variabel
// global. Sama seperti modul 12: perintah yang bergantung pada variabel global mustahil
// diuji secara terpisah, karena satu test akan mewarisi kotoran test sebelumnya.
type Penyimpanan struct {
	mu      sync.Mutex
	berikut int
	data    map[int]Tugas
}

func NewPenyimpanan() *Penyimpanan {
	return &Penyimpanan{berikut: 1, data: make(map[int]Tugas)}
}

func (p *Penyimpanan) Tambah(judul, prioritas string) Tugas {
	p.mu.Lock()
	defer p.mu.Unlock()

	t := Tugas{ID: p.berikut, Judul: judul, Prioritas: prioritas}
	p.data[t.ID] = t
	p.berikut++
	return t
}

func (p *Penyimpanan) Daftar() []Tugas {
	p.mu.Lock()
	defer p.mu.Unlock()

	out := make([]Tugas, 0, len(p.data))
	for _, t := range p.data {
		out = append(out, t)
	}
	// Map di Go urutannya acak — selalu urutkan sebelum ditampilkan,
	// kalau tidak keluaran CLI-mu berubah-ubah tiap dijalankan.
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func (p *Penyimpanan) Selesaikan(id int) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	t, ada := p.data[id]
	if !ada {
		return fmt.Errorf("tugas dengan id %d tidak ditemukan", id)
	}
	t.Selesai = true
	p.data[id] = t
	return nil
}

// ------------------------------------------------------------------
// 1. Perintah root
// ------------------------------------------------------------------

// NewRootCmd membangun seluruh pohon perintah.
//
// 🔍 Analogi: fungsi ini seperti DENAH GEDUNG yang dibangun ulang tiap dipanggil.
// Kenapa tidak dibuat sekali saja sebagai variabel global (seperti banyak contoh cobra
// di internet)? Karena flag menyimpan NILAI TERAKHIR yang diisi. Kalau satu test
// menjalankan "--format json", perintah global itu akan tetap ber-format json di test
// berikutnya — bug yang sangat membingungkan. Membangun ulang = selalu bersih.
func NewRootCmd(p *Penyimpanan) *cobra.Command {
	var format string

	root := &cobra.Command{
		Use:   "tugasku",
		Short: "Pengelola tugas sederhana",
		// Long muncul di "tugasku --help".
		Long: "tugasku adalah contoh CLI untuk memperagakan spf13/cobra.\n" +
			"Perintahnya disusun seperti git: tugasku <perintah> <sub-perintah>.",
		// 🔍 Analogi SilenceUsage: secara bawaan, kalau perintahmu gagal, cobra mencetak
		// SELURUH teks bantuan. Bayangkan salah ketik satu huruf lalu layar dibanjiri
		// manual setebal buku. Dengan ini, pesan salah tetap ringkas — bantuan hanya
		// muncul saat pengguna memang salah memakai perintahnya.
		SilenceUsage: true,
	}

	// 🔍 Analogi persistent flag vs flag biasa:
	//   PersistentFlags = ATURAN SEGEDUNG ("dilarang merokok") — berlaku di perintah ini
	//                     DAN semua sub-perintah di bawahnya.
	//   Flags           = aturan SATU RUANGAN saja.
	root.PersistentFlags().StringVar(&format, "format", "teks",
		"format keluaran: teks atau json")

	root.AddCommand(newTugasCmd(p, &format))
	root.AddCommand(newVersiCmd())
	return root
}

// ------------------------------------------------------------------
// 2. Sub-perintah bertingkat
// ------------------------------------------------------------------

func newTugasCmd(p *Penyimpanan, format *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tugas",
		Short: "Kelola daftar tugas",
	}
	cmd.AddCommand(newTambahCmd(p, format), newDaftarCmd(p, format), newSelesaiCmd(p))
	return cmd
}

func newTambahCmd(p *Penyimpanan, format *string) *cobra.Command {
	var prioritas string

	cmd := &cobra.Command{
		Use:   "tambah <judul>",
		Short: "Tambahkan tugas baru",
		// 🔍 Analogi Args: ini SATPAM JUMLAH ARGUMEN. ExactArgs(1) berarti "wajib tepat
		// satu judul". Tanpa ini kamu harus menulis sendiri cek panjang slice di tiap
		// perintah — dan pasti ada yang lupa.
		Args: cobra.ExactArgs(1),
		// 🔍 Analogi RunE vs Run: RunE MENGEMBALIKAN error, sehingga kode keluar proses
		// jadi 1 dan skrip shell (atau CI) tahu perintahnya gagal. Run biasa menelan
		// kegagalan diam-diam — CLI yang selalu bilang "sukses" itu berbahaya di pipeline.
		RunE: func(cmd *cobra.Command, args []string) error {
			judul := strings.TrimSpace(args[0])
			if judul == "" {
				return fmt.Errorf("judul tugas tidak boleh kosong")
			}
			if prioritas != "rendah" && prioritas != "sedang" && prioritas != "tinggi" {
				return fmt.Errorf("prioritas %q tidak dikenal (pilih: rendah, sedang, tinggi)", prioritas)
			}

			t := p.Tambah(judul, prioritas)
			// cmd.OutOrStdout() — BUKAN fmt.Println langsung!
			// 🔍 Analogi: ini seperti mesin cetak yang bisa diarahkan ke kertas mana pun.
			// Di produksi mengarah ke layar; di test diarahkan ke buffer supaya keluarannya
			// bisa diperiksa. Memakai fmt.Println langsung membuat perintah mustahil diuji.
			tulisTugas(cmd.OutOrStdout(), *format, []Tugas{t})
			return nil
		},
	}

	// Flag lokal: hanya berlaku untuk "tugas tambah".
	// "p" di StringVarP berarti ada bentuk pendeknya: -p tinggi
	cmd.Flags().StringVarP(&prioritas, "prioritas", "p", "sedang",
		"prioritas tugas: rendah, sedang, tinggi")
	return cmd
}

func newDaftarCmd(p *Penyimpanan, format *string) *cobra.Command {
	var hanyaBelum bool

	cmd := &cobra.Command{
		Use:     "daftar",
		Short:   "Tampilkan semua tugas",
		Aliases: []string{"ls", "list"}, // alias: ketikan lebih pendek untuk perintah sama
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			tugas := p.Daftar()
			if hanyaBelum {
				sisa := make([]Tugas, 0, len(tugas))
				for _, t := range tugas {
					if !t.Selesai {
						sisa = append(sisa, t)
					}
				}
				tugas = sisa
			}
			tulisTugas(cmd.OutOrStdout(), *format, tugas)
			return nil
		},
	}

	// Flag boolean: kehadirannya saja sudah berarti true (--belum).
	cmd.Flags().BoolVar(&hanyaBelum, "belum", false, "tampilkan hanya yang belum selesai")
	return cmd
}

func newSelesaiCmd(p *Penyimpanan) *cobra.Command {
	return &cobra.Command{
		Use:   "selesai <id>",
		Short: "Tandai tugas sebagai selesai",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var id int
			if _, err := fmt.Sscanf(args[0], "%d", &id); err != nil {
				return fmt.Errorf("id harus berupa angka, dapat %q", args[0])
			}
			if err := p.Selesaikan(id); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "tugas %d ditandai selesai\n", id)
			return nil
		},
	}
}

func newVersiCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "versi",
		Short: "Tampilkan versi aplikasi",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, _ []string) {
			// Run biasa (tanpa E) pantas dipakai di sini: perintah ini mustahil gagal.
			fmt.Fprintln(cmd.OutOrStdout(), "tugasku v1.0.0")
		},
	}
}

// ------------------------------------------------------------------
// 3. Keluaran
// ------------------------------------------------------------------

// tulisTugas mencetak daftar tugas sesuai format yang diminta.
//
// 🔍 Analogi dua format: CLI yang baik melayani DUA jenis pembaca. Manusia ingin tabel
// rapi; skrip otomatis ingin JSON yang bisa disalurkan ke `jq`. Inilah kenapa alat modern
// seperti `kubectl` dan `gh` selalu punya flag "-o json".
func tulisTugas(w io.Writer, format string, tugas []Tugas) {
	if format == "json" {
		fmt.Fprint(w, "[")
		for i, t := range tugas {
			if i > 0 {
				fmt.Fprint(w, ",")
			}
			fmt.Fprintf(w, `{"id":%d,"judul":%q,"prioritas":%q,"selesai":%t}`,
				t.ID, t.Judul, t.Prioritas, t.Selesai)
		}
		fmt.Fprintln(w, "]")
		return
	}

	if len(tugas) == 0 {
		fmt.Fprintln(w, "(belum ada tugas)")
		return
	}
	for _, t := range tugas {
		tanda := " "
		if t.Selesai {
			tanda = "x"
		}
		fmt.Fprintf(w, "[%s] %d. %s (%s)\n", tanda, t.ID, t.Judul, t.Prioritas)
	}
}

// ------------------------------------------------------------------
// 4. Helper untuk pengujian
// ------------------------------------------------------------------

// Jalankan menjalankan CLI dengan argumen tertentu dan menangkap keluarannya.
//
// 🔍 Analogi: ini seperti MENGETIK PERINTAH DI TERMINAL PALSU lalu memotret layarnya.
// Karena semua perintah menulis ke cmd.OutOrStdout() (bukan langsung ke layar), kita bisa
// mengarahkan "layar"-nya ke buffer dan memeriksa isinya. Inilah cara menguji CLI tanpa
// benar-benar menjalankan proses terpisah.
func Jalankan(p *Penyimpanan, args ...string) (keluaran string, err error) {
	var buf strings.Builder

	root := NewRootCmd(p)
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs(args)

	err = root.Execute()
	return buf.String(), err
}
