# 35 — GraphQL

GraphQL: **satu endpoint**, client meminta **persis** field yang dibutuhkan (tak lebih, tak kurang). Alternatif REST (Modul 12–13) untuk API dengan relasi kompleks & kebutuhan client beragam.

Jalankan:
```bash
go run ./35-graphql
# POST http://localhost:8080/graphql  body: {"query":"{ authors { name books { title } } }"}
```
Verifikasi otomatis: `go test ./35-graphql`

> Modul ini pakai [`graphql-go/graphql`](https://github.com/graphql-go/graphql) (skema programatik, tanpa codegen). Untuk **gqlgen** (schema-first + codegen, paling populer), lihat bagian bawah.

## GraphQL vs REST

| | REST | GraphQL |
|-|------|---------|
| Endpoint | banyak (`/authors`, `/books`) | **satu** (`/graphql`) |
| Data | server tentukan bentuk | **client** minta field spesifik |
| Over/under-fetching | umum | dihindari |
| Versioning | `/v1`, `/v2` | evolusi skema (deprecate field) |
| Caching | mudah (HTTP) | lebih rumit |

Client minta `{ authors { name books { title } } }` → dapat **persis** itu, dalam satu request (tak perlu `/authors` lalu N× `/books`).

## Konsep

### Schema & Types
```go
authorType := graphql.NewObject(graphql.ObjectConfig{
    Name: "Author",
    Fields: graphql.Fields{
        "name":  &graphql.Field{Type: graphql.String},
        "books": &graphql.Field{Type: graphql.NewList(bookType), Resolve: ...},
    },
})
```
Skema **ketat & tervalidasi**: query field yang tak ada → error (test membuktikan). Ini keunggulan besar — kontrak jelas & self-documenting.

### Resolver
Fungsi yang mengambil data untuk sebuah field. `p.Source` = objek induk, `p.Args` = argumen.
```go
Resolve: func(p graphql.ResolveParams) (any, error) {
    a := p.Source.(Author)
    return store.booksByAuthor(a.ID), nil  // resolver relasi
}
```

### Query vs Mutation
- **Query** = baca (`{ authors { ... } }`).
- **Mutation** = tulis (`mutation { addBook(title: "X", authorId: 1) { id } }`).

## ⚠️ Masalah N+1 & DataLoader

Field relasi (`author.books`) memanggil resolver **per author** → 1 query authors + N query books = **N+1** (bunuh performa). Solusi: **DataLoader** — kumpulkan semua `authorID` dalam satu tick, ambil sekaligus (batch), lalu bagikan.
```
// Tanpa DataLoader: SELECT books WHERE author=1; author=2; ... (N query)
// Dengan DataLoader: SELECT books WHERE author IN (1,2,...) (1 query)
```
Library: [`graph-gophers/dataloader`](https://github.com/graph-gophers/dataloader).

## Alternatif: gqlgen (schema-first, populer di produksi)
```bash
# 1. tulis schema.graphqls, 2. konfigurasi gqlgen.yml, 3. generate:
go run github.com/99designs/gqlgen generate
# gqlgen membuat kode + interface resolver yang kamu isi -> type-safe penuh.
```
gqlgen unggul untuk proyek besar (type-safe, codegen). graphql-go (modul ini) lebih ringan & tanpa build step.

## Kapan & Di Mana Dipakai
- API dengan banyak jenis client (web, mobile) yang butuh bentuk data berbeda.
- Data ber-relasi kompleks (graph), agar client tak over-fetch.
- BFF (Backend-for-Frontend).

## Latihan
1. Tambah query `author(id: Int!)` untuk mengambil satu author.
2. Tambah mutation `deleteBook(id: Int!)`.
3. Tambah field `Book.author` (relasi balik) — hati-hati siklus resolver.
4. Terapkan **DataLoader** untuk field `books` (atasi N+1).
5. Migrasikan ke **gqlgen** (schema-first) dan bandingkan pengalamannya.
