// Jalankan:
//
//	go run ./19-config                                  # pakai default
//	go run ./19-config config.example.yaml              # pakai file
//	APP_PORT=9090 APP_ENV=staging go run ./19-config    # override via env
//
// Verifikasi otomatis: go test ./19-config
package main

import (
	"fmt"
	"os"
)

func main() {
	// Argumen pertama (opsional) = path file config.
	path := ""
	if len(os.Args) > 1 {
		path = os.Args[1]
	}

	cfg, err := Load(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, "config error:", err)
		os.Exit(1)
	}

	fmt.Printf("AppName        : %s\n", cfg.AppName)
	fmt.Printf("Env            : %s\n", cfg.Env)
	fmt.Printf("Port           : %d\n", cfg.Port)
	fmt.Printf("Database.DSN   : %s\n", cfg.Database.DSN)
	fmt.Printf("Database.Conns : %d\n", cfg.Database.MaxOpenConns)
	secret := "(kosong)"
	if cfg.JWT.Secret != "" {
		secret = "(diset, disembunyikan)"
	}
	fmt.Printf("JWT.Secret     : %s\n", secret)
}
