# 21 — Database Migrations dengan golang-migrate

**Migrasi** = perubahan skema database yang **berversi, terlacak, & bisa di-rollback**. Modul ini memakai [**golang-migrate**](https://github.com/golang-migrate/migrate) — library migrasi paling populer di ekosistem Go (mendukung Postgres, MySQL, SQLite, MongoDB, dll).

Jalankan (memakai SQLite pure-Go, tanpa cgo):
```bash
go run ./21-migrations
```
Verifikasi otomatis: `go test ./21-migrations`

---

## 📦 Instalasi

### 1. Sebagai library (dipakai di kode Go)
```bash
go get github.com/golang-migrate/migrate/v4
```
Plus driver yang kamu butuhkan (di-import di kode, lihat di bawah):
- Source (dari mana file migrasi dibaca): `source/iofs` (embed), `source/file`.
- Database: `database/sqlite` (pure-Go), `database/postgres`, `database/mysql`, dst.

> Repo ini sudah menambahkannya ke `go.mod` — tak perlu apa-apa lagi.

### 2. Sebagai CLI (untuk membuat & menjalankan migrasi dari terminal)

**Opsi A — `go install`** (pilih driver via build tag):
```bash
# SQLite pure-Go + PostgreSQL:
go install -tags 'sqlite postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
# binari terpasang di $(go env GOPATH)/bin  -> pastikan ada di PATH
migrate -version
```
Build tag menentukan driver yang di-*compile*: `sqlite` (pure-Go), `sqlite3` (cgo/mattn), `postgres`, `mysql`, `mongodb`, dll.

**Opsi B — binari prebuilt / package manager**:
```bash
# macOS
brew install golang-migrate
# Linux (contoh)
curl -L https://github.com/golang-migrate/migrate/releases/latest/download/migrate.linux-amd64.tar.gz | tar xvz
# Docker
docker run --rm -v $(pwd)/migrations:/migrations migrate/migrate --help
```

---

## 🚀 Penggunaan CLI

### Membuat file migrasi baru
```bash
migrate create -ext sql -dir migrations -seq create_users
# menghasilkan:
#   migrations/000001_create_users.up.sql    <- isi: CREATE TABLE ...
#   migrations/000001_create_users.down.sql  <- isi: DROP TABLE ...
```
`-seq` memberi nomor urut (000001, 000002, ...). Isi `.up.sql` (maju) & `.down.sql` (mundur) manual.

### Menjalankan migrasi
```bash
DB="sqlite://app.db"          # atau: postgres://user:pass@localhost:5432/db?sslmode=disable

migrate -database "$DB" -path migrations up        # terapkan SEMUA yang belum jalan
migrate -database "$DB" -path migrations up 1       # naik 1 langkah
migrate -database "$DB" -path migrations down 1     # rollback 1 langkah
migrate -database "$DB" -path migrations version    # versi saat ini
migrate -database "$DB" -path migrations goto 2     # pindah ke versi tertentu
migrate -database "$DB" -path migrations force 2    # tandai versi (lihat "dirty" di bawah)
```

### URL database per jenis
| DB | URL |
|----|-----|
| SQLite (pure-Go) | `sqlite://app.db` |
| PostgreSQL | `postgres://user:pass@host:5432/dbname?sslmode=disable` |
| MySQL | `mysql://user:pass@tcp(host:3306)/dbname` |

---

## 🧩 Penggunaan sebagai Library (kode modul ini)

File migrasi di-*embed* ke binary lalu dijalankan dari kode — cocok untuk menjalankan migrasi saat start / lewat perintah CLI aplikasimu sendiri.

```go
//go:embed migrations/*.sql
var migrationsFS embed.FS

func newMigrator(db *sql.DB) (*migrate.Migrate, error) {
    src, _ := iofs.New(migrationsFS, "migrations")               // source: embed
    driver, _ := sqlitemigrate.WithInstance(db, &sqlitemigrate.Config{}) // db driver
    return migrate.NewWithInstance("iofs", src, "sqlite", driver)
}
```
Modul ini menyediakan tiga fungsi:
```go
Up(db)      // m.Up()      — terapkan semua migrasi (ErrNoChange ditangani -> idempotent)
Down(db)    // m.Steps(-1) — rollback satu langkah
Version(db) // m.Version() — versi saat ini + status dirty
```
Output demo:
```
setelah Up  -> versi: 3
jalankan Up lagi -> tidak ada perubahan (idempotent)
setelah Down -> versi: 2
```

> ⚠️ Jangan panggil `m.Close()` bila masih ingin memakai `db` — `Close()` ikut menutup koneksi database di baliknya.

---

## 📁 Konvensi file migrasi

golang-migrate membaca versi dari nama file:
```
{versi}_{judul}.up.sql     # maju
{versi}_{judul}.down.sql   # mundur (rollback)
```
Contoh di `migrations/`:
```
001_create_users.up.sql / .down.sql
002_create_posts.up.sql / .down.sql
003_add_users_active.up.sql / .down.sql
```
Tabel `schema_migrations` (dibuat otomatis) melacak versi yang sudah diterapkan.

## ⚠️ Status "dirty"

Kalau sebuah migrasi **gagal di tengah**, golang-migrate menandai versi sebagai **dirty** & menolak jalan lagi (mencegah kerusakan). Perbaiki manual, lalu:
```bash
migrate -database "$DB" -path migrations force <versi_yang_benar>
```

## 🏭 Praktik produksi
- Jalankan migrasi **sebelum** aplikasi start (init container / job / step CI-CD), bukan diam-diam saat runtime.
- Tiap migrasi harus **kecil & reversible**. Sertakan selalu `.down.sql`.
- Untuk data besar, hati-hati migrasi yang mengunci tabel (lakukan bertahap).
- Simpan file migrasi di version control bersama kode.

## Kapan & Di Mana Dipakai
- Setiap proyek dengan database, terutama begitu ada >1 environment atau >1 developer.

## Latihan
1. Pasang CLI (`go install -tags 'sqlite' ...`) lalu buat migrasi `004_create_comments` dengan `migrate create`.
2. Jalankan `up`, `version`, `down 1` dari CLI terhadap `sqlite://demo.db`.
3. Tambah perintah pada CLI aplikasimu (Cobra, Modul 11): `app migrate up` yang memanggil `Up(db)`.
4. Ganti target ke PostgreSQL: import `database/postgres`, pakai URL `postgres://...`.
5. Simulasikan migrasi gagal → amati status **dirty** → perbaiki dengan `force`.
