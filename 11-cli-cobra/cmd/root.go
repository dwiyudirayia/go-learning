// Package cmd berisi definisi perintah CLI (pola standar proyek Cobra).
package cmd

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// filePath diisi dari persistent flag --file, dipakai semua subcommand.
var filePath string

// rootCmd adalah perintah dasar: dijalankan tanpa subcommand hanya menampilkan help.
var rootCmd = &cobra.Command{
	Use:   "tasks",
	Short: "tasks — CLI manajemen tugas sederhana (contoh Cobra)",
	Long: `tasks adalah CLI contoh untuk modul 11.
Menyimpan daftar tugas ke file JSON. Subcommand: add, list, done, version.`,
}

// Execute dipanggil dari main. Mengembalikan error agar main yang memutuskan exit code.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Persistent flag: berlaku untuk root dan SEMUA subcommand.
	defaultPath := filepath.Join(os.TempDir(), "golearn-tasks.json")
	rootCmd.PersistentFlags().StringVar(&filePath, "file", defaultPath, "path file penyimpanan task")
}
