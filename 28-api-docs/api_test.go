package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestOpenAPISpec(t *testing.T) {
	h := buildHandler()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/openapi.json", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("/openapi.json = %d; want 200", rec.Code)
	}
	// Spec harus JSON valid & memuat info OpenAPI.
	var spec map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &spec); err != nil {
		t.Fatalf("spec bukan JSON valid: %v", err)
	}
	if spec["openapi"] == nil {
		t.Error("spec tidak punya field 'openapi'")
	}
	paths, ok := spec["paths"].(map[string]any)
	if !ok || paths["/books"] == nil {
		t.Error("spec tidak mendokumentasikan path /books")
	}
}

func TestErrorFormatKonsisten(t *testing.T) {
	h := buildHandler()

	// POST tanpa field wajib -> error envelope terstruktur.
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("POST", "/api/v1/books", strings.NewReader(`{"title":""}`)))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d; want 400", rec.Code)
	}
	var e APIError
	if err := json.Unmarshal(rec.Body.Bytes(), &e); err != nil {
		t.Fatalf("error bukan JSON: %v", err)
	}
	if e.Error.Code != "VALIDATION_ERROR" {
		t.Errorf("error.code = %q; want VALIDATION_ERROR", e.Error.Code)
	}
	if e.Error.Message == "" {
		t.Error("error.message kosong")
	}
}

func TestVersionedCRUD(t *testing.T) {
	h := buildHandler()

	// POST valid -> 201
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("POST", "/api/v1/books",
		strings.NewReader(`{"title":"Go","author":"RP"}`)))
	if rec.Code != http.StatusCreated {
		t.Fatalf("POST = %d; want 201", rec.Code)
	}

	// GET -> berisi 1
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/api/v1/books", nil))
	var books []Book
	_ = json.Unmarshal(rec.Body.Bytes(), &books)
	if len(books) != 1 {
		t.Errorf("jumlah buku = %d; want 1", len(books))
	}
}
