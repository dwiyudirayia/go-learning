// Jalankan: go run ./35-graphql
//
//	POST http://localhost:8080/graphql  body: {"query":"{ authors { name } }"}
//
// Verifikasi otomatis: go test ./35-graphql
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/graphql-go/graphql"
)

// graphQLHandler mengeksekusi query GraphQL dari body request.
func graphQLHandler(schema graphql.Schema) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Query     string         `json:"query"`
			Variables map[string]any `json:"variables"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)

		result := graphql.Do(graphql.Params{
			Schema:         schema,
			RequestString:  req.Query,
			VariableValues: req.Variables,
		})
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(result)
	}
}

func main() {
	fmt.Println("=== 35 — GraphQL ===")
	schema, err := newSchema(newStore())
	if err != nil {
		log.Fatal(err)
	}

	// Contoh: client minta HANYA nama author + judul bukunya (tak lebih).
	demo := `{ authors { name books { title } } }`
	result := graphql.Do(graphql.Params{Schema: schema, RequestString: demo})
	out, _ := json.MarshalIndent(result.Data, "", "  ")
	fmt.Printf("query: %s\nhasil:\n%s\n\n", demo, out)

	http.HandleFunc("/graphql", graphQLHandler(schema))
	log.Println("server GraphQL di http://localhost:8080/graphql")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
