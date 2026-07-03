# 19 — Konfigurasi (Viper) & 12-Factor

Aplikasi produksi tidak boleh punya nilai ter-hardcode (port, DSN, secret). Konfigurasi harus bisa berubah **tanpa rebuild**. Modul ini pakai [Viper](https://github.com/spf13/viper) dengan prinsip **12-factor**.

Jalankan:
```bash
go run ./19-config                                # default
go run ./19-config config.example.yaml            # dari file
APP_PORT=9090 APP_ENV=staging go run ./19-config  # override via env
```
Verifikasi otomatis: `go test ./19-config`

## Prioritas sumber config (rendah → tinggi)

```
default (kode)  <  file (config.yaml)  <  environment variable
```
Env **paling menang** — inti 12-factor: config lewat environment, sehingga image/binary yang sama jalan di dev, staging, dan produksi hanya dengan mengganti env.

```go
v.SetDefault("port", 8080)          // 1. default
v.SetConfigFile("config.yaml"); v.ReadInConfig() // 2. file (opsional)
v.SetEnvPrefix("APP")
v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
v.AutomaticEnv()                    // 3. env: APP_PORT, APP_DATABASE_DSN, ...
v.Unmarshal(&cfg)                   // isi struct
```

## Config sebagai satu struct

Kumpulkan semua config di **satu struct** (bukan variabel global tersebar). Tag `mapstructure` memetakan key → field, termasuk **nested** (`database.dsn` → `APP_DATABASE_DSN`).
```go
type Config struct {
    Port     int            `mapstructure:"port"`
    Database DatabaseConfig `mapstructure:"database"`
}
```

## Validasi = fail fast

Cek config **sebelum** aplikasi mulai. Lebih baik crash saat start dengan pesan jelas daripada error misterius jam 3 pagi.
```go
if c.Env == "production" && len(c.JWT.Secret) < 16 {
    return fmt.Errorf("jwt.secret wajib >= 16 karakter di production")
}
```

## Aturan penting

- **Secret JANGAN di file yang di-commit.** Isi via env (`APP_JWT_SECRET`) atau secret manager (Vault, AWS Secrets Manager, K8s Secret). File `.example.yaml` hanya template dengan nilai kosong.
- **Fail fast**: validasi + `os.Exit(1)` bila config salah.
- Untuk **CLI** (Modul 11), Viper berpadu dengan Cobra: flag > env > file > default.

## Kapan & Di Mana Dipakai
- Setiap service: port, DSN database, alamat Redis/queue, secret, feature flag, log level.
- Deploy multi-environment (dev/staging/prod) dengan **satu** binary/image.

## Latihan
1. Tambah field `LogLevel` dan validasi nilainya (debug/info/warn/error).
2. Tambah `RedisAddr` + default `localhost:6379`.
3. Dukung format file `.json` selain `.yaml` (Viper mendeteksi dari ekstensi).
4. Tambah `Watch` (Viper `WatchConfig`) untuk reload config saat file berubah.
5. Integrasikan config ini ke server Modul 15 (ganti `os.Getenv` manual).
