package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"go-learning/11-cli-cobra/internal/store"
)

var doneCmd = &cobra.Command{
	Use:   "done [id]",
	Short: "Tandai tugas selesai",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("id harus angka: %q", args[0])
		}
		s, err := store.Load(filePath)
		if err != nil {
			return err
		}
		if err := s.MarkDone(id); err != nil {
			return err
		}
		fmt.Printf("tugas #%d ditandai selesai\n", id)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(doneCmd)
}
