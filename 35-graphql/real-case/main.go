// REAL-CASE Modul 35 (GraphQL) — resolver didukung POSTGRESQL.
//
// Versi advanced/ memakai data in-memory. Versi ini menjalankan query GraphQL
// yang resolver-nya benar-benar MEMBACA dari PostgreSQL. Menunjukkan bagaimana
// GraphQL menjadi lapisan API di atas database nyata.
//
// Auto-skip bila POSTGRES_DSN kosong. Jalankan nyata:
//
//	docker compose -f 35-graphql/real-case/docker-compose.yml up -d
//	POSTGRES_DSN='postgres://postgres:secret@127.0.0.1:5432/appdb' go run ./35-graphql/real-case
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/graphql-go/graphql"
	"github.com/jackc/pgx/v5/pgxpool"
)

// buildSchema membangun skema GraphQL yang resolver-nya MENUTUP (closure) di atas
// pool Postgres, sehingga tiap resolver bisa meng-query DB.
//
// Param:
//   - pool : koneksi pool ke Postgres.
//
// Return skema GraphQL siap dieksekusi, atau error bila definisi tak valid.
func buildSchema(pool *pgxpool.Pool) (graphql.Schema, error) {
	// Tipe objek User (bentuk data yang bisa diminta client).
	userType := graphql.NewObject(graphql.ObjectConfig{
		Name: "User",
		Fields: graphql.Fields{
			"id":    &graphql.Field{Type: graphql.Int},
			"nama":  &graphql.Field{Type: graphql.String},
			"email": &graphql.Field{Type: graphql.String},
		},
	})

	// Query root: dua field -> "users" (semua) & "user(id)" (satu).
	queryType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			// users -> SELECT semua baris.
			"users": &graphql.Field{
				Type: graphql.NewList(userType),
				// Resolve dipanggil GraphQL untuk MENGISI field ini. p.Context
				// membawa context request (deadline dll).
				Resolve: func(p graphql.ResolveParams) (any, error) {
					return queryUsers(p.Context, pool, 0)
				},
			},
			// user(id: Int!) -> SELECT satu baris.
			"user": &graphql.Field{
				Type: userType,
				Args: graphql.FieldConfigArgument{ // argumen query yang wajib
					"id": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.Int)},
				},
				Resolve: func(p graphql.ResolveParams) (any, error) {
					id := p.Args["id"].(int)
					rows, err := queryUsers(p.Context, pool, id)
					if err != nil || len(rows) == 0 {
						return nil, err
					}
					return rows[0], nil
				},
			},
		},
	})

	return graphql.NewSchema(graphql.SchemaConfig{Query: queryType})
}

// queryUsers membaca user dari Postgres. Bila id>0, filter satu user.
//
// Mengembalikan []map[string]any agar resolver default GraphQL bisa memetakan
// tiap field berdasarkan nama kunci (id/nama/email).
func queryUsers(ctx context.Context, pool *pgxpool.Pool, id int) ([]map[string]any, error) {
	sql := `SELECT id, nama, email FROM users`
	args := []any{}
	if id > 0 {
		sql += ` WHERE id=$1`
		args = append(args, id)
	}
	sql += ` ORDER BY id`

	rows, err := pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []map[string]any
	for rows.Next() {
		var uid int
		var nama, email string
		if err := rows.Scan(&uid, &nama, &email); err != nil {
			return nil, err
		}
		out = append(out, map[string]any{"id": uid, "nama": nama, "email": email})
	}
	return out, rows.Err()
}

// seed menyiapkan tabel + data contoh (di produksi: migrasi + data nyata).
func seed(ctx context.Context, pool *pgxpool.Pool) error {
	if _, err := pool.Exec(ctx, `CREATE TABLE IF NOT EXISTS users(
		id BIGSERIAL PRIMARY KEY, nama TEXT NOT NULL, email TEXT NOT NULL)`); err != nil {
		return err
	}
	if _, err := pool.Exec(ctx, `TRUNCATE users RESTART IDENTITY`); err != nil {
		return err
	}
	_, err := pool.Exec(ctx, `INSERT INTO users(nama,email) VALUES
		('Budi','budi@mail.com'), ('Ani','ani@mail.com'), ('Cici','cici@mail.com')`)
	return err
}

func main() {
	dsn := os.Getenv("POSTGRES_DSN")
	if dsn == "" {
		fmt.Println("⏭️  DILEWATI: set POSTGRES_DSN untuk versi nyata.")
		fmt.Println("   docker compose -f 35-graphql/real-case/docker-compose.yml up -d")
		fmt.Println("   POSTGRES_DSN='postgres://postgres:secret@127.0.0.1:5432/appdb' go run ./35-graphql/real-case")
		return
	}
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		panic(err)
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		panic("gagal konek Postgres: " + err.Error())
	}
	if err := seed(ctx, pool); err != nil {
		panic(err)
	}

	schema, err := buildSchema(pool)
	if err != nil {
		panic(err)
	}

	// Client memilih PERSIS field yang diinginkan; resolver query Postgres.
	query := `{ user(id: 2) { nama email } allUsers: users { id nama } }`
	result := graphql.Do(graphql.Params{
		Schema:        schema,
		RequestString: query,
		Context:       ctx, // diteruskan ke resolver -> query DB
	})
	if len(result.Errors) > 0 {
		panic(fmt.Sprint(result.Errors))
	}

	out, _ := json.MarshalIndent(result.Data, "  ", "  ")
	fmt.Println("== hasil query GraphQL (data dari Postgres) ==")
	fmt.Println("  " + string(out))
	fmt.Println("== produksi: gqlgen (type-safe), DataLoader (anti N+1), batasi kompleksitas query ==")
}
