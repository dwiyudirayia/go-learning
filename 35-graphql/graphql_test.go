package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/graphql-go/graphql"
)

func schema(t *testing.T) graphql.Schema {
	t.Helper()
	sc, err := newSchema(newStore())
	if err != nil {
		t.Fatalf("schema: %v", err)
	}
	return sc
}

func TestQueryAuthors(t *testing.T) {
	sc := schema(t)
	// Client minta HANYA field yang dibutuhkan.
	r := graphql.Do(graphql.Params{Schema: sc, RequestString: `{ authors { id name } }`})
	if len(r.Errors) > 0 {
		t.Fatalf("errors: %v", r.Errors)
	}
	data := r.Data.(map[string]any)
	authors := data["authors"].([]any)
	if len(authors) != 2 {
		t.Errorf("jumlah author = %d; want 2", len(authors))
	}
}

func TestQueryRelasiBooks(t *testing.T) {
	sc := schema(t)
	// Relasi author -> books (resolver bersarang).
	r := graphql.Do(graphql.Params{Schema: sc, RequestString: `{ authors { name books { title } } }`})
	if len(r.Errors) > 0 {
		t.Fatalf("errors: %v", r.Errors)
	}
	// Rob Pike (id 1) punya 2 buku.
	data := r.Data.(map[string]any)
	for _, a := range data["authors"].([]any) {
		m := a.(map[string]any)
		if m["name"] == "Rob Pike" {
			books := m["books"].([]any)
			if len(books) != 2 {
				t.Errorf("Rob Pike punya %d buku; want 2", len(books))
			}
		}
	}
}

func TestMutationAddBook(t *testing.T) {
	sc := schema(t)
	q := `mutation { addBook(title: "Buku Baru", authorId: 1) { id title authorId } }`
	r := graphql.Do(graphql.Params{Schema: sc, RequestString: q})
	if len(r.Errors) > 0 {
		t.Fatalf("errors: %v", r.Errors)
	}
	book := r.Data.(map[string]any)["addBook"].(map[string]any)
	if book["title"] != "Buku Baru" {
		t.Errorf("title = %v; want Buku Baru", book["title"])
	}
}

func TestHTTPHandler(t *testing.T) {
	h := graphQLHandler(schema(t))
	body := `{"query":"{ authors { name } }"}`
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/graphql", strings.NewReader(body))
	h(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d; want 200", rec.Code)
	}
	var resp struct {
		Data map[string]any `json:"data"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.Data["authors"] == nil {
		t.Error("response tak berisi authors")
	}
}

// Query yang salah field -> error validasi (kekuatan GraphQL: skema ketat).
func TestQueryFieldTakDikenal(t *testing.T) {
	sc := schema(t)
	r := graphql.Do(graphql.Params{Schema: sc, RequestString: `{ authors { tidakAda } }`})
	if len(r.Errors) == 0 {
		t.Error("mengharapkan error untuk field tak dikenal")
	}
}
