package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"go-learning/11-cli-cobra/internal/store"
)

// Latihan 3: subcommand `stats` menampilkan ringkasan tugas.
var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Tampilkan jumlah total/selesai/belum",
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := store.Load(filePath)
		if err != nil {
			return err
		}
		total, done, pending := s.Stats()
		fmt.Printf("total=%d  selesai=%d  belum=%d\n", total, done, pending)
		return nil
	},
}

func init() { rootCmd.AddCommand(statsCmd) }
