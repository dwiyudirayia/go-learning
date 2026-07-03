package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const appVersion = "1.0.0"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Tampilkan versi aplikasi",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("tasks versi %s\n", appVersion)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
