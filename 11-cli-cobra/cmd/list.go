package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"go-learning/11-cli-cobra/internal/store"
)

var showAll bool

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Tampilkan daftar tugas",
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := store.Load(filePath)
		if err != nil {
			return err
		}
		if len(s.Tasks) == 0 {
			fmt.Println("(belum ada tugas)")
			return nil
		}
		for _, t := range s.Tasks {
			if t.Done && !showAll {
				continue // sembunyikan yang selesai kecuali --all
			}
			mark := "[ ]"
			if t.Done {
				mark = "[x]"
			}
			fmt.Printf("%s #%d %s\n", mark, t.ID, t.Text)
		}
		return nil
	},
}

func init() {
	// Local flag: hanya untuk perintah list ini.
	listCmd.Flags().BoolVar(&showAll, "all", false, "tampilkan juga tugas yang selesai")
	rootCmd.AddCommand(listCmd)
}
