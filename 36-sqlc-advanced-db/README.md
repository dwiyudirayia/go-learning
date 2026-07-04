# 36 — sqlc (SQL Type-Safe) + Connection Pool + Transaksi

Lanjutan Modul 14. **sqlc** menghasilkan kode Go **type-safe** dari SQL biasa — kamu tulis SQL, sqlc generate fungsi Go-nya (tanpa `Scan` manual, tanpa string SQL tersebar, dicek saat kompilasi). Plus tuning **connection pool** & pola **transaksi**.

Jalankan:
```bash
go run ./36-sqlc-advanced-db
```
Verifikasi otomatis: `go test ./36-sqlc-advanced-db`

## 📦 Instalasi & Generate sqlc

```bash
# Pasang CLI (pure Go)
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# Generate kode dari SQL (dijalankan di folder modul):
cd 36-sqlc-advanced-db
sqlc generate     # membaca sqlc.yaml -> menghasilkan internal/sqlcdb/*.go
```

## Cara kerja sqlc

### 1. Tulis skema & query dalam SQL biasa
`db/schema.sql`:
```sql
CREATE TABLE authors (id INTEGER PRIMARY KEY, name TEXT NOT NULL);
```
`db/query.sql` (anotasi `-- name:` memberi nama & jenis hasil):
```sql
-- name: CreateAuthor :one
INSERT INTO authors (name) VALUES (?) RETURNING *;

-- name: ListAuthors :many
SELECT * FROM authors ORDER BY id;
```
`:one` → satu baris, `:many` → slice, `:exec` → tanpa hasil, `:execresult` → LastInsertId.

### 2. Konfigurasi `sqlc.yaml`
```yaml
version: "2"
sql:
  - engine: "sqlite"        # atau postgresql / mysql
    schema: "db/schema.sql"
    queries: "db/query.sql"
    gen: { go: { package: "sqlcdb", out: "internal/sqlcdb", emit_json_tags: true } }
```

### 3. Pakai kode hasil generate (type-safe!)
```go
author, err := store.CreateAuthor(ctx, "Rob Pike")  // author bertipe sqlcdb.Author
// tanpa rows.Scan, tanpa salah ketik nama kolom -> dicek compiler
```
Ganti query SQL → regenerate → kalau kode Go tak lagi cocok, **gagal kompilasi** (bukan error runtime). Ini keunggulan besar dibanding `database/sql` mentah (Modul 14) & lebih ringan dari GORM.

## Connection Pool (`OpenDB`)

`*sql.DB` adalah **pool koneksi**, bukan satu koneksi. Setel di produksi:
```go
db.SetMaxOpenConns(25)                 // maks koneksi bersamaan (jaga DB tak kewalahan)
db.SetMaxIdleConns(5)                  // koneksi idle disimpan (hemat handshake)
db.SetConnMaxLifetime(5 * time.Minute) // daur ulang (hindari koneksi basi di balik LB)
db.SetConnMaxIdleTime(time.Minute)
```
Salah setel = sumber masalah performa umum (koneksi habis, atau terlalu banyak membebani DB).

## Transaksi (pola `execTx`)

Pola idiomatik: fungsi yang menerima closure, commit bila sukses, rollback bila error/panic.
```go
func (s *Store) execTx(ctx, fn func(*sqlcdb.Queries) error) error {
    tx, _ := s.db.BeginTx(ctx, nil)
    q := s.Queries.WithTx(tx)   // sqlc: Queries yang beroperasi di dalam tx
    if err := fn(q); err != nil { tx.Rollback(); return err }
    return tx.Commit()
}
```
`WithTx` (dihasilkan sqlc) membuat query berjalan dalam transaksi. Test membuktikan: gagal di tengah → **rollback** → tak ada data tersimpan.

## sqlc vs GORM vs database/sql
| | database/sql (14) | GORM (14) | **sqlc (ini)** |
|-|-------------------|-----------|----------------|
| Cara | SQL manual + Scan | ORM (magic) | SQL + codegen |
| Type-safe | ❌ | sebagian | ✅ penuh |
| Boilerplate | banyak | sedikit | sedikit |
| Kontrol SQL | penuh | terbatas | penuh |

Banyak tim modern memilih sqlc: kontrol SQL penuh **plus** keamanan tipe.

## Kapan & Di Mana Dipakai
- Proyek yang ingin SQL eksplisit + keamanan tipe (tanpa "magic" ORM).
- Query kompleks (JOIN, CTE, window function) yang sulit di ORM.

## Latihan
1. Tambah query `-- name: UpdateAuthorName :exec` lalu regenerate & pakai.
2. Tambah `-- name: SearchAuthors :many` dengan `WHERE name LIKE ?`.
3. Ganti engine ke `postgresql` di `sqlc.yaml` (perlu sintaks `$1` bukan `?`).
4. Uji pool: turunkan `SetMaxOpenConns(1)` & jalankan query paralel — amati antrean.
5. Tambah transaksi transfer (kurangi A, tambah B) yang rollback bila saldo kurang.

## ✅ Solusi Latihan (Pembahasan)

1. **`UpdateAuthorName :exec`** — tulis di `query.sql`:
   ```sql
   -- name: UpdateAuthorName :exec
   UPDATE authors SET name = ? WHERE id = ?;
   ```
   `sqlc generate` → method `UpdateAuthorName(ctx, arg)`.
2. **`SearchAuthors :many`** — `WHERE name LIKE ?`; panggil dengan `"%"+q+"%"`. `:many` menghasilkan `[]Author`.
3. **Ganti ke PostgreSQL** — di `sqlc.yaml` set `engine: postgresql`; placeholder jadi `$1,$2` (bukan `?`), pakai driver `pgx`. Query type-safe tetap.
4. **Uji pool** — `db.SetMaxOpenConns(1)` lalu jalankan banyak query di goroutine paralel → mereka mengantre (serial). Bukti pool tuning berpengaruh nyata (Modul 26 untuk ukur).
5. **Transaksi transfer** — `WithTx`: `UPDATE ... balance-=amt WHERE id=A`; cek saldo; `UPDATE ... balance+=amt WHERE id=B`. Bila saldo kurang → `return err` → `tx.Rollback()` (defer). Atomik.
