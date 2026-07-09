// Teknik Advanced Modul 19 (config) — contoh runnable + penjelasan detail.
//
// Fokus: presedensi Viper (env > file > default) + unmarshal ke struct typed.
// Jalankan:
//
//	go run ./19-config/advanced
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

// Config typed: hindari viper.GetString(...) tersebar di seluruh kode.
// Nilai di-unmarshal sekali ke struct ini lalu dipakai type-safe.
type Config struct {
	Port int `mapstructure:"port"`
	Log  struct {
		Level string `mapstructure:"level"`
	} `mapstructure:"log"`
}

func main() {
	v := viper.New()

	// =====================================================================
	// 1. DEFAULT — lapisan paling bawah (dipakai bila tak ada override)
	// =====================================================================
	v.SetDefault("port", 8080)
	v.SetDefault("log.level", "info")

	// =====================================================================
	// 2. FILE — meng-override default (di sini dari string, seolah config.yaml)
	// =====================================================================
	v.SetConfigType("yaml")
	_ = v.ReadConfig(strings.NewReader("port: 9090\nlog:\n  level: debug\n"))

	// =====================================================================
	// 3. ENV — presedensi TERTINGGI (12-factor: config lewat environment)
	// =====================================================================
	// Prefix APP_ + replacer titik->garis bawah: kunci "log.level" dibaca dari
	// env APP_LOG_LEVEL. AutomaticEnv membuat semua kunci otomatis terbaca env.
	v.SetEnvPrefix("APP")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	os.Setenv("APP_PORT", "3000") // seolah di-set operator saat deploy

	fmt.Println("== presedensi tiap kunci ==")
	fmt.Printf("  port      = %d   (default 8080 -> file 9090 -> ENV 3000 MENANG)\n", v.GetInt("port"))
	fmt.Printf("  log.level = %q (default info -> FILE debug MENANG, tak ada env)\n", v.GetString("log.level"))

	// =====================================================================
	// 4. Unmarshal ke struct + fail-fast validasi saat startup
	// =====================================================================
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		panic(err)
	}
	if cfg.Port < 1 || cfg.Port > 65535 { // validasi -> keluar cepat bila salah
		panic(fmt.Sprintf("port tidak valid: %d", cfg.Port))
	}
	fmt.Println("== struct typed ==")
	fmt.Printf("  %+v\n", cfg)

	// Catatan: SECRET (password DB, API key) JANGAN ditaruh di file config yang
	// ter-commit; ambil dari env/secret manager. Lihat modul 27 (security).
}
