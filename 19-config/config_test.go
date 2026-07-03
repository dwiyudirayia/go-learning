package main

import "testing"

func TestDefault(t *testing.T) {
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("load default: %v", err)
	}
	if cfg.Port != 8080 || cfg.Env != "dev" || cfg.AppName != "myapp" {
		t.Errorf("default tak sesuai: %+v", cfg)
	}
}

func TestEnvOverride(t *testing.T) {
	// t.Setenv otomatis di-reset setelah test.
	t.Setenv("APP_PORT", "9090")
	t.Setenv("APP_ENV", "staging")
	t.Setenv("APP_DATABASE_DSN", "postgres://x")

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if cfg.Port != 9090 {
		t.Errorf("Port = %d; want 9090 (dari env)", cfg.Port)
	}
	if cfg.Env != "staging" {
		t.Errorf("Env = %q; want staging", cfg.Env)
	}
	if cfg.Database.DSN != "postgres://x" {
		t.Errorf("DSN = %q; want dari env (nested)", cfg.Database.DSN)
	}
}

func TestFile(t *testing.T) {
	cfg, err := Load("config.example.yaml")
	if err != nil {
		t.Fatalf("load file: %v", err)
	}
	if cfg.AppName != "task-manager" {
		t.Errorf("AppName = %q; want dari file", cfg.AppName)
	}
	if cfg.Database.MaxOpenConns != 25 {
		t.Errorf("MaxOpenConns = %d; want 25 (dari file)", cfg.Database.MaxOpenConns)
	}
}

func TestValidasiGagal(t *testing.T) {
	// env tidak valid.
	t.Setenv("APP_ENV", "produksi-typo")
	if _, err := Load(""); err == nil {
		t.Error("mengharapkan error untuk env tidak valid")
	}
}

func TestProduksiButuhSecret(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	// tanpa secret -> harus gagal
	if _, err := Load(""); err == nil {
		t.Error("production tanpa secret harus gagal")
	}
	// dengan secret kuat -> lolos
	t.Setenv("APP_JWT_SECRET", "rahasia-yang-cukup-panjang")
	if _, err := Load(""); err != nil {
		t.Errorf("production dengan secret kuat harusnya lolos: %v", err)
	}
}
