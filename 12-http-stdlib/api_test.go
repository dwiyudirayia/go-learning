package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// newTestServer membuat server dengan store kosong + router-nya.
func newTestServer() http.Handler {
	return (&server{store: NewBookStore()}).routes()
}

func TestCreateAndGet(t *testing.T) {
	srv := newTestServer()

	// POST /books
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/books", strings.NewReader(`{"title":"Go","author":"RP"}`))
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("POST status = %d; want 201", rec.Code)
	}
	var created Book
	_ = json.Unmarshal(rec.Body.Bytes(), &created)
	if created.ID == 0 || created.Title != "Go" {
		t.Fatalf("hasil create tak sesuai: %+v", created)
	}

	// GET /books/{id}
	rec = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/books/1", nil)
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET status = %d; want 200", rec.Code)
	}
}

func TestValidationDanNotFound(t *testing.T) {
	srv := newTestServer()

	// POST tanpa title -> 400
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, httptest.NewRequest("POST", "/books", strings.NewReader(`{"author":"x"}`)))
	if rec.Code != http.StatusBadRequest {
		t.Errorf("POST tanpa title = %d; want 400", rec.Code)
	}

	// GET id tak ada -> 404
	rec = httptest.NewRecorder()
	srv.ServeHTTP(rec, httptest.NewRequest("GET", "/books/999", nil))
	if rec.Code != http.StatusNotFound {
		t.Errorf("GET tak ada = %d; want 404", rec.Code)
	}

	// GET id bukan angka -> 400
	rec = httptest.NewRecorder()
	srv.ServeHTTP(rec, httptest.NewRequest("GET", "/books/abc", nil))
	if rec.Code != http.StatusBadRequest {
		t.Errorf("GET id non-angka = %d; want 400", rec.Code)
	}
}

// Latihan 1: test endpoint search.
func TestSearch(t *testing.T) {
	srv := newTestServer()
	srv.ServeHTTP(httptest.NewRecorder(),
		httptest.NewRequest("POST", "/books", strings.NewReader(`{"title":"Go","author":"Rob Pike"}`)))
	srv.ServeHTTP(httptest.NewRecorder(),
		httptest.NewRequest("POST", "/books", strings.NewReader(`{"title":"C","author":"Dennis Ritchie"}`)))

	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, httptest.NewRequest("GET", "/books/search?author=pike", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("search status = %d; want 200", rec.Code)
	}
	var hasil []Book
	_ = json.Unmarshal(rec.Body.Bytes(), &hasil)
	if len(hasil) != 1 || hasil[0].Author != "Rob Pike" {
		t.Errorf("search author=pike -> %+v; want 1 (Rob Pike)", hasil)
	}
}

// Latihan 4: test pagination.
func TestPagination(t *testing.T) {
	srv := newTestServer()
	for i := 0; i < 5; i++ {
		srv.ServeHTTP(httptest.NewRecorder(),
			httptest.NewRequest("POST", "/books", strings.NewReader(`{"title":"buku","author":"x"}`)))
	}
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, httptest.NewRequest("GET", "/books?limit=2&offset=0", nil))
	var page []Book
	_ = json.Unmarshal(rec.Body.Bytes(), &page)
	if len(page) != 2 {
		t.Errorf("limit=2 -> %d item; want 2", len(page))
	}
}

func TestUpdateDanDelete(t *testing.T) {
	srv := newTestServer()

	// buat dulu
	srv.ServeHTTP(httptest.NewRecorder(),
		httptest.NewRequest("POST", "/books", strings.NewReader(`{"title":"awal","author":"a"}`)))

	// PUT /books/1
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, httptest.NewRequest("PUT", "/books/1", strings.NewReader(`{"title":"baru","author":"b"}`)))
	if rec.Code != http.StatusOK {
		t.Fatalf("PUT status = %d; want 200", rec.Code)
	}
	var updated Book
	_ = json.Unmarshal(rec.Body.Bytes(), &updated)
	if updated.Title != "baru" {
		t.Errorf("title setelah update = %q; want baru", updated.Title)
	}

	// DELETE /books/1 -> 204
	rec = httptest.NewRecorder()
	srv.ServeHTTP(rec, httptest.NewRequest("DELETE", "/books/1", nil))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("DELETE status = %d; want 204", rec.Code)
	}

	// GET setelah delete -> 404
	rec = httptest.NewRecorder()
	srv.ServeHTTP(rec, httptest.NewRequest("GET", "/books/1", nil))
	if rec.Code != http.StatusNotFound {
		t.Errorf("GET setelah delete = %d; want 404", rec.Code)
	}
}
