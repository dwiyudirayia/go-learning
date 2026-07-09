# 14 — Database: `database/sql` & GORM

Jalankan:
```bash
go run ./14-database
```
Modul ini memakai **SQLite pure-Go** (`glebarez/sqlite`, tanpa cgo) agar jalan di mana saja tanpa server DB. Konsepnya sama untuk PostgreSQL/MySQL — cukup ganti driver + DSN.

## Bagian A — `database/sql` (mentah)

Paket standar; kamu menulis SQL sendiri. Kontrol penuh, nol "magic".
```go
db, _ := sql.Open("sqlite", path)   // (postgres: "pgx"/"postgres", mysql: "mysql")
defer db.Close()

db.Exec("INSERT INTO users(name,email) VALUES(?,?)", name, email) // placeholder!
db.QueryRow("SELECT name FROM users WHERE id=?", id).Scan(&name)  // 1 baris
rows, _ := db.Query("SELECT ...")                                 // banyak baris
defer rows.Close()
for rows.Next() { rows.Scan(&a, &b) }
```

**Wajib diperhatikan:**
- **Selalu pakai placeholder `?`** (atau `$1` di Postgres) — JANGAN string-concat → mencegah **SQL injection**.
- `QueryRow(...).Scan(...)` → cek `sql.ErrNoRows` untuk "tidak ada".
- `Query` → **wajib `rows.Close()`** dan sebaiknya cek `rows.Err()`.
- **Transaksi**: `tx, _ := db.Begin()` → `tx.Exec(...)` → `tx.Commit()` / `tx.Rollback()`.
- Untuk konteks/timeout: pakai varian `...Context` (`db.QueryContext(ctx, ...)`).

## Bagian B — GORM (ORM)

ORM populer: memetakan struct ↔ tabel, mengurangi boilerplate.
```go
type User struct {
	gorm.Model                 // ID, CreatedAt, UpdatedAt, DeletedAt
	Name  string
	Email string `gorm:"uniqueIndex"`
	Posts []Post               // relasi has-many
}

db, _ := gorm.Open(sqlite.Open(path), &gorm.Config{})
db.AutoMigrate(&User{}, &Post{})       // buat/ubah tabel dari struct

db.Create(&user)                       // INSERT (+ relasi sekaligus)
db.First(&u, id)                       // SELECT ... LIMIT 1
db.Where("name = ?", "Ana").Find(&us)  // SELECT banyak
db.Preload("Posts").First(&u, id)      // JOIN/muat relasi
db.Model(&u).Update("Name", "X")       // UPDATE
db.Delete(&u)                          // SOFT delete (isi DeletedAt)
db.Unscoped().Find(&all)               // termasuk yang soft-deleted
```

**Konsep kunci:**
- **`gorm.Model`** memberi `ID` + timestamp + **soft delete** (`DeletedAt`). Baris tak benar-benar dihapus; query normal menyembunyikannya. `Unscoped()` untuk melihat semua.
- **AutoMigrate** enak untuk dev; di produksi banyak tim pakai migrasi eksplisit (mis. `golang-migrate`, atau file SQL) agar terkontrol & reversible.
- **Preload** menghindari masalah **N+1 query**.

## `database/sql` vs GORM

| | database/sql | GORM |
|-|--------------|------|
| Kendali SQL | penuh, manual | otomatis (bisa raw juga) |
| Boilerplate | banyak | sedikit |
| Kurva belajar | SQL | API GORM |
| Cocok untuk | query kompleks, performa kritis | CRUD cepat, model kaya relasi |

Banyak tim **mencampur**: GORM untuk CRUD umum, `database/sql`/raw untuk query berat.

## Ganti ke PostgreSQL (produksi)
```go
import "gorm.io/driver/postgres"
dsn := "host=localhost user=app password=secret dbname=app port=5432 sslmode=disable"
db, _ := gorm.Open(postgres.Open(dsn), &gorm.Config{})
```
Struktur kode (repository, model) **tidak berubah** — hanya driver + DSN.

## Kapan & Di Mana Dipakai
- Hampir semua backend butuh persistensi. Pola: `repository` membungkus akses DB, `service` memakai repository lewat **interface** (Modul 4) → mudah di-test dengan mock (Modul 8).

## Latihan
1. Tambah model `Category` dan relasi `Post belongs-to Category`.
2. Tulis fungsi `repository` (interface + implementasi GORM) untuk `User` (Create/Get/List/Delete).
3. Buat query raw `database/sql` dengan `JOIN` untuk menghitung jumlah post per user.
4. Ganti soft delete jadi hard delete dengan `Unscoped().Delete(...)`.
5. Tambah timeout memakai `db.WithContext(ctx)` (GORM) / `QueryContext` (raw).

> Repository + interface ini dipakai di **Modul 15 (studi kasus REST + JWT)**.

## ✅ Status Solusi Latihan
Latihan **1, 3, 4 sudah diselesaikan** di fungsi `latihanDemo()` (main.go): model `Category` belongs-to, query raw JOIN, dan hard delete (`Unscoped`). Latihan 2 & 5 (repository interface, context timeout) sebagai tantangan lanjutan.

---

## 🚀 Teknik Advanced (Level Up)
> 💻 **Contoh runnable + komentar detail** untuk teknik di bawah ada di folder [`advanced/`](advanced). Jalankan: `go run ./14-database/advanced`


- **Connection pool tuning** — `db.SetMaxOpenConns`, `SetMaxIdleConns`, `SetConnMaxLifetime`. Salah setel → connection exhaustion atau koneksi basi.
- **Selalu `...Context`** — `QueryContext`/`ExecContext` agar query ikut batal saat request cancel.
- **Transaksi aman** — pola `tx, _ := db.Begin(); defer tx.Rollback(); ...; tx.Commit()` (rollback jadi no-op setelah commit).
- **`sql.Null*` & custom `Scanner`/`Valuer`** — tangani NULL & tipe kustom (mis. JSON column, enum).
- **Cegah SQL injection** — SELALU parameterized (`?`/`$1`), jangan string concat.
- **GORM lanjutan** — `Preload` (hindari N+1), `Select` kolom spesifik, hooks (`BeforeCreate`), soft delete, batch insert `CreateInBatches`. Awas N+1 tersembunyi.
- **Naik level** — type-safe query via **sqlc** [[36-sqlc-advanced-db]]; skema via migrasi [[21-migrations]].
