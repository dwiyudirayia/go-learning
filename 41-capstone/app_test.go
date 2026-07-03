package main

import (
	"database/sql"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"

	_ "modernc.org/sqlite"
)

// testApp merangkai app dengan SQLite temp + Redis in-memory (tanpa infra).
func testApp(t *testing.T) *fiber.App {
	t.Helper()
	db, err := sql.Open("sqlite", filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	if err := setupSchema(db); err != nil {
		t.Fatalf("schema: %v", err)
	}

	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis: %v", err)
	}
	t.Cleanup(mr.Close)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { rdb.Close() })

	svc := NewService(NewStore(db), NewCache(rdb))
	return buildApp(svc, slog.New(slog.NewTextHandler(io.Discard, nil)))
}

func req(t *testing.T, app *fiber.App, method, path, body, token string) *http.Response {
	t.Helper()
	var r io.Reader
	if body != "" {
		r = strings.NewReader(body)
	}
	httpReq := httptest.NewRequest(method, path, r)
	if body != "" {
		httpReq.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		httpReq.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := app.Test(httpReq, -1)
	if err != nil {
		t.Fatalf("%s %s: %v", method, path, err)
	}
	return resp
}

func TestCapstoneFullFlow(t *testing.T) {
	app := testApp(t)

	// Register -> 201.
	if r := req(t, app, "POST", "/auth/register", `{"email":"a@b.c","password":"secret123"}`, ""); r.StatusCode != http.StatusCreated {
		t.Fatalf("register = %d; want 201", r.StatusCode)
	}
	// Login -> token.
	r := req(t, app, "POST", "/auth/login", `{"email":"a@b.c","password":"secret123"}`, "")
	var lr struct{ Token string }
	b, _ := io.ReadAll(r.Body)
	_ = json.Unmarshal(b, &lr)
	if lr.Token == "" {
		t.Fatal("token kosong")
	}

	// Shorten tanpa token -> 401.
	if r := req(t, app, "POST", "/api/shorten", `{"url":"https://go.dev"}`, ""); r.StatusCode != http.StatusUnauthorized {
		t.Errorf("shorten tanpa token = %d; want 401", r.StatusCode)
	}

	// Shorten dengan token -> 201 + code.
	r = req(t, app, "POST", "/api/shorten", `{"url":"https://go.dev"}`, lr.Token)
	if r.StatusCode != http.StatusCreated {
		t.Fatalf("shorten = %d; want 201", r.StatusCode)
	}
	var sr struct{ Code string }
	b, _ = io.ReadAll(r.Body)
	_ = json.Unmarshal(b, &sr)
	if sr.Code == "" {
		t.Fatal("code kosong")
	}

	// Redirect -> 302 + Location.
	r = req(t, app, "GET", "/"+sr.Code, "", "")
	if r.StatusCode != http.StatusFound {
		t.Fatalf("redirect = %d; want 302", r.StatusCode)
	}
	if loc := r.Header.Get("Location"); loc != "https://go.dev" {
		t.Errorf("Location = %q; want https://go.dev", loc)
	}

	// Redirect kedua -> harus cache HIT (tetap 302 & benar).
	r = req(t, app, "GET", "/"+sr.Code, "", "")
	if r.Header.Get("Location") != "https://go.dev" {
		t.Error("redirect kedua (cache) salah")
	}

	// Kode tak ada -> 404.
	if r := req(t, app, "GET", "/tidakada", "", ""); r.StatusCode != http.StatusNotFound {
		t.Errorf("kode tak ada = %d; want 404", r.StatusCode)
	}
}
