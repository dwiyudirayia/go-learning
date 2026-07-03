package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"go-learning/11-cli-cobra/internal/store"
)

var addPriority int // latihan 2: nilai flag --priority

var addCmd = &cobra.Command{
	Use:   "add [teks tugas]",
	Short: "Tambah tugas baru",
	Args:  cobra.MinimumNArgs(1), // validasi: minimal 1 argumen
	// RunE (bukan Run) agar bisa mengembalikan error -> Cobra cetak & set exit code.
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := store.Load(filePath)
		if err != nil {
			return err
		}
		// Gabungkan semua argumen jadi satu teks tugas.
		text := args[0]
		for _, a := range args[1:] {
			text += " " + a
		}
		id, err := s.AddWithPriority(text, addPriority)
		if err != nil {
			return err
		}
		fmt.Printf("ditambahkan #%d (prioritas %d): %s\n", id, addPriority, text)
		return nil
	},
}

func init() {
	// Latihan 2: flag lokal --priority untuk perintah add.
	addCmd.Flags().IntVar(&addPriority, "priority", 0, "prioritas tugas (angka lebih besar = lebih penting)")
	rootCmd.AddCommand(addCmd) // daftarkan sebagai subcommand root
}
