package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-playground/validator/v10"
)

// Fiber punya app.Test(req) untuk menguji tanpa membuka port sungguhan.
func newTestApp() *server {
	s := &server{store: NewBookStore(), validate: validator.New()}
	return s
}

func TestFiber_CreateAndGet(t *testing.T) {
	s := newTestApp()
	app := buildApp(s)

	// POST valid -> 201
	req := httptest.NewRequest("POST", "/api/books",
		strings.NewReader(`{"title":"Go","author":"RP","year":2015}`))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("POST status = %d; want 201", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	var created Book
	_ = json.Unmarshal(body, &created)
	if created.ID == 0 || created.Title != "Go" {
		t.Fatalf("hasil create tak sesuai: %+v", created)
	}

	// GET /api/books/1 -> 200
	resp, _ = app.Test(httptest.NewRequest("GET", "/api/books/1", nil))
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET status = %d; want 200", resp.StatusCode)
	}
}

func TestFiber_Validation(t *testing.T) {
	app := buildApp(newTestApp())

	// title terlalu pendek (min=2) + author kosong -> 400
	req := httptest.NewRequest("POST", "/api/books",
		strings.NewReader(`{"title":"G","author":""}`))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("validasi status = %d; want 400", resp.StatusCode)
	}
}

func TestFiber_NotFound(t *testing.T) {
	app := buildApp(newTestApp())

	resp, _ := app.Test(httptest.NewRequest("GET", "/api/books/999", nil))
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("GET tak ada = %d; want 404", resp.StatusCode)
	}

	resp, _ = app.Test(httptest.NewRequest("GET", "/api/books/abc", nil))
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("GET id non-angka = %d; want 400", resp.StatusCode)
	}
}

// Latihan 1: test update.
func TestFiber_Update(t *testing.T) {
	s := newTestApp()
	s.store.Create(Book{Title: "awal", Author: "a", Year: 2020})
	app := buildApp(s)

	req := httptest.NewRequest("PUT", "/api/books/1",
		strings.NewReader(`{"title":"baru","author":"b","year":2021}`))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("PUT = %d; want 200", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	var b Book
	_ = json.Unmarshal(body, &b)
	if b.Title != "baru" {
		t.Errorf("title = %q; want baru", b.Title)
	}
}

// Latihan 3: test filter ?author=.
func TestFiber_Filter(t *testing.T) {
	s := newTestApp()
	s.store.Create(Book{Title: "Go", Author: "Rob Pike", Year: 2015})
	s.store.Create(Book{Title: "C", Author: "Dennis Ritchie", Year: 1978})
	app := buildApp(s)

	resp, _ := app.Test(httptest.NewRequest("GET", "/api/books?author=pike", nil))
	body, _ := io.ReadAll(resp.Body)
	var books []Book
	_ = json.Unmarshal(body, &books)
	if len(books) != 1 || books[0].Author != "Rob Pike" {
		t.Errorf("filter author=pike -> %+v; want 1 (Rob Pike)", books)
	}
}

func TestFiber_Delete(t *testing.T) {
	s := newTestApp()
	s.store.Create(Book{Title: "hapus aku", Author: "x", Year: 2020})
	app := buildApp(s)

	resp, _ := app.Test(httptest.NewRequest("DELETE", "/api/books/1", nil))
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("DELETE status = %d; want 204", resp.StatusCode)
	}
	resp, _ = app.Test(httptest.NewRequest("GET", "/api/books/1", nil))
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("GET setelah delete = %d; want 404", resp.StatusCode)
	}
}
