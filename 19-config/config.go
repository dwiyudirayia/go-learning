// Modul 19 — Config: memuat konfigurasi dari default + file + environment,
// dengan validasi. Memakai Viper.
package main

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config: seluruh konfigurasi aplikasi dalam satu struct (bukan variabel global tersebar).
type Config struct {
	AppName  string         `mapstructure:"app_name"`
	Env      string         `mapstructure:"env"` // dev | staging | production
	Port     int            `mapstructure:"port"`
	Database DatabaseConfig `mapstructure:"database"`
	JWT      JWTConfig      `mapstructure:"jwt"`
}

type DatabaseConfig struct {
	DSN          string `mapstructure:"dsn"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
}

type JWTConfig struct {
	Secret string `mapstructure:"secret"`
}

// 🔍 Analogi besar: memuat config berlapis itu seperti ATURAN BERPAKAIAN bertingkat.
//  1. default   = seragam standar pabrik ("kalau tak diatur, pakai ini").
//  2. file      = aturan kantor cabang (menimpa seragam pabrik).
//  3. env var   = instruksi bos langsung hari ini (PALING menang, menimpa semuanya).
// Lapisan lebih tinggi menimpa yang rendah. Prinsip "12-factor": config lewat ENVIRONMENT,
// bukan di-hardcode — jadi 1 binary sama bisa jalan di dev/staging/produksi hanya dgn ganti env.

// Load memuat config dengan prioritas (dari rendah ke tinggi):
//  1. nilai default (di kode)
//  2. file config (bila path diberikan & ada)
//  3. environment variable (paling menang) -> praktik 12-factor
//
// Env override: APP_PORT=9090, APP_DATABASE_DSN=..., APP_JWT_SECRET=...
func Load(configPath string) (*Config, error) {
	v := viper.New()

	// 1. Default.
	v.SetDefault("app_name", "myapp")
	v.SetDefault("env", "dev")
	v.SetDefault("port", 8080)
	v.SetDefault("database.dsn", "app.db")
	v.SetDefault("database.max_open_conns", 10)
	v.SetDefault("jwt.secret", "")

	// 2. File (opsional). Diabaikan bila tidak ada.
	if configPath != "" {
		v.SetConfigFile(configPath)
		if err := v.ReadInConfig(); err != nil {
			if _, notFound := err.(viper.ConfigFileNotFoundError); !notFound {
				// Error selain "file tak ada" (mis. YAML rusak) tetap dilaporkan.
				return nil, fmt.Errorf("baca config: %w", err)
			}
		}
	}

	// 3. Environment. Prefix APP_, dan "database.dsn" -> "APP_DATABASE_DSN".
	v.SetEnvPrefix("APP")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// 🔍 Analogi: "fail fast" itu seperti CEK KELENGKAPAN sebelum pesawat lepas landas. Lebih baik
// aplikasi menolak START dengan pesan jelas ("jwt.secret wajib di production") daripada telanjur
// jalan lalu meledak di tengah melayani pengguna. Validasi config di awal = mencegah bencana diam-diam.
// Validate memastikan config masuk akal SEBELUM aplikasi jalan (fail fast).
func (c *Config) Validate() error {
	switch c.Env {
	case "dev", "staging", "production":
	default:
		return fmt.Errorf("env tidak valid: %q (harus dev/staging/production)", c.Env)
	}
	if c.Port < 1 || c.Port > 65535 {
		return fmt.Errorf("port tidak valid: %d", c.Port)
	}
	// Aturan khusus produksi: secret WAJIB diisi & kuat.
	if c.Env == "production" && len(c.JWT.Secret) < 16 {
		return fmt.Errorf("jwt.secret wajib >= 16 karakter di production")
	}
	return nil
}
