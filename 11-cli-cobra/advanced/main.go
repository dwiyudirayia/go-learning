// Teknik Advanced Modul 11 (CLI Cobra) — contoh runnable + penjelasan detail.
//
// Demo ini MENJALANKAN command secara programatik (SetArgs+Execute) agar bisa
// dilihat hasilnya tanpa mengetik di terminal. Jalankan:
//
//	go run ./11-cli-cobra/advanced
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func main() {
	var verbose bool // flag PERSISTENT: diwarisi semua subcommand
	var ulang int    // flag LOKAL: hanya milik subcommand "sapa"

	root := &cobra.Command{
		Use:           "app",
		Short:         "Demo teknik Cobra lanjutan",
		SilenceUsage:  true, // jangan cetak usage panjang saat error (kita tangani sendiri)
		SilenceErrors: true,
	}
	// PersistentFlags terpasang di root -> otomatis tersedia di subcommand.
	root.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "tampilkan log detail")

	sapa := &cobra.Command{
		Use:  "sapa [nama]",
		Args: cobra.ExactArgs(1), // validasi otomatis: WAJIB tepat 1 argumen
		// PreRunE berjalan SEBELUM RunE — tempat validasi/persiapan (mis. buka DB).
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if ulang < 1 {
				return fmt.Errorf("--ulang harus >= 1 (diberi %d)", ulang)
			}
			return nil
		},
		// RunE (bukan Run) -> kembalikan error agar exit code benar.
		RunE: func(cmd *cobra.Command, args []string) error {
			if verbose {
				fmt.Println("  [verbose] menyapa", args[0], "sebanyak", ulang, "kali")
			}
			fmt.Println("  " + strings.TrimSpace(strings.Repeat("Halo "+args[0]+"! ", ulang)))
			return nil
		},
	}
	sapa.Flags().IntVarP(&ulang, "ulang", "u", 1, "berapa kali menyapa")
	root.AddCommand(sapa)

	// --- Skenario 1: pemakaian normal (flag lokal + persistent) ---
	fmt.Println("== 1. sapa Budi -u 2 -v ==")
	jalankan(root, "sapa", "Budi", "-u", "2", "-v")

	// --- Skenario 2: argumen kurang -> ExactArgs(1) menolak ---
	fmt.Println("== 2. sapa (tanpa nama) -> error validasi args ==")
	jalankan(root, "sapa")

	// --- Skenario 3: PreRunE menolak nilai flag tak valid ---
	fmt.Println("== 3. sapa Ani -u 0 -> PreRunE menolak ==")
	jalankan(root, "sapa", "Ani", "-u", "0")
}

// jalankan menyetel argumen lalu Execute, mencetak error bila ada.
func jalankan(root *cobra.Command, args ...string) {
	root.SetArgs(args)
	root.SetOut(os.Stdout)
	if err := root.Execute(); err != nil {
		fmt.Println("  error:", err)
	}
}
