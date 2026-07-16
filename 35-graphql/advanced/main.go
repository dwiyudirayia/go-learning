// Teknik Advanced Modul 35 (GraphQL) — contoh runnable + penjelasan detail.
//
// Skema + eksekusi query, dan mendemonstrasikan masalah N+1 (yang diatasi
// DataLoader/batching). Jalankan:
//
//	go run ./35-graphql/advanced
package main

import (
	"encoding/json"
	"fmt"

	"github.com/graphql-go/graphql"
)

type User struct {
	ID   int
	Name string
}

func main() {
	allUsers := []User{{1, "Budi"}, {2, "Ani"}, {3, "Cici"}}

	// Penghitung "akses DB" untuk memvisualkan N+1.
	dbCalls := 0

	// Tipe Post (anak dari User).
	postType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Post",
		Fields: graphql.Fields{
			"title": &graphql.Field{Type: graphql.String},
		},
	})

	// Tipe User dengan field "posts" yang di-resolve PER user.
	userType := graphql.NewObject(graphql.ObjectConfig{
		Name: "User",
		Fields: graphql.Fields{
			"id":   &graphql.Field{Type: graphql.Int},
			"name": &graphql.Field{Type: graphql.String},
			"posts": &graphql.Field{
				Type: graphql.NewList(postType),
				// MASALAH N+1: resolver ini dipanggil SEKALI PER user. Untuk 3
				// user -> 3 query terpisah (plus 1 query users = N+1). DataLoader
				// mengumpulkan semua id lalu meng-query SEKALI (batch).
				//
				// 🔍 Analogi: N+1 itu seperti belanja titipan 3 tetangga dengan
				// PERGI KE PASAR 3 KALI — sekali per tetangga. DataLoader =
				// kumpulkan dulu semua daftar titipan, pergi ke pasar SEKALI,
				// pulang bagikan sesuai pemesan (batch by id).
				Resolve: func(p graphql.ResolveParams) (any, error) {
					u := p.Source.(User)
					dbCalls++ // tiap panggilan = satu "hit DB"
					return []map[string]any{
						{"title": fmt.Sprintf("Artikel oleh %s", u.Name)},
					}, nil
				},
			},
		},
	})

	// Query root.
	queryType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"users": &graphql.Field{
				Type: graphql.NewList(userType),
				Resolve: func(p graphql.ResolveParams) (any, error) {
					return allUsers, nil
				},
			},
		},
	})

	schema, err := graphql.NewSchema(graphql.SchemaConfig{Query: queryType})
	if err != nil {
		panic(err)
	}

	// =====================================================================
	// Eksekusi query — client memilih PERSIS field yang diinginkan.
	// =====================================================================
	// 🔍 Analogi: REST itu seperti MENU PAKET di restoran cepat saji — dapat
	// semua isi paket walau cuma mau kentangnya. GraphQL seperti PRASMANAN:
	// ambil persis lauk yang diinginkan (field), tak lebih. Konsekuensinya
	// dapur harus siap kombinasi apa pun — termasuk pesanan aneh yang mahal
	// (query dalam/kompleks perlu dibatasi).
	query := `{ users { name posts { title } } }`
	res := graphql.Do(graphql.Params{Schema: schema, RequestString: query})
	if len(res.Errors) > 0 {
		panic(res.Errors)
	}

	out, _ := json.MarshalIndent(res.Data, "  ", "  ")
	fmt.Println("== hasil query ==")
	fmt.Println(" ", string(out))

	fmt.Println("== N+1 ==")
	fmt.Printf("  resolver 'posts' dipanggil %d kali untuk %d user (inilah N+1)\n", dbCalls, len(allUsers))
	fmt.Println("  solusi: DataLoader mengumpulkan id -> 1 query batch;")
	fmt.Println("  plus batasi kedalaman/kompleksitas query agar tak bisa di-DoS.")
}
