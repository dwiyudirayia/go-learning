package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"go-learning/15-studi-kasus-rest/internal/model"
)

// testApp membuat app dengan DB SQLite in-memory UNIK per test (isolasi penuh,
// ID mulai dari 1). Nama DB diambil dari t.Name() agar tiap test terpisah.
func testApp(t *testing.T) *fiber.App {
	t.Helper()
	dsn := "file:" + t.Name() + "?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&model.User{}, &model.Task{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return buildApp(db)
}

// loginToken helper: register + login, kembalikan token.
func loginToken(t *testing.T, app *fiber.App, email string) string {
	t.Helper()
	do(t, app, "POST", "/auth/register", `{"name":"U","email":"`+email+`","password":"secret123"}`, "")
	resp := do(t, app, "POST", "/auth/login", `{"email":"`+email+`","password":"secret123"}`, "")
	var lr struct{ Token string }
	b, _ := io.ReadAll(resp.Body)
	_ = json.Unmarshal(b, &lr)
	return lr.Token
}

func do(t *testing.T, app *fiber.App, method, path, body, token string) *http.Response {
	t.Helper()
	var r io.Reader
	if body != "" {
		r = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, r)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := app.Test(req, -1) // -1 = tanpa timeout
	if err != nil {
		t.Fatalf("request %s %s: %v", method, path, err)
	}
	return resp
}

func TestFullFlow(t *testing.T) {
	app := testApp(t)

	// 1. Register -> 201
	resp := do(t, app, "POST", "/auth/register", `{"name":"Ana","email":"ana@mail.id","password":"secret123"}`, "")
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("register = %d; want 201", resp.StatusCode)
	}

	// 2. Register email sama -> 409
	resp = do(t, app, "POST", "/auth/register", `{"name":"Ana2","email":"ana@mail.id","password":"secret123"}`, "")
	if resp.StatusCode != http.StatusConflict {
		t.Errorf("register duplikat = %d; want 409", resp.StatusCode)
	}

	// 3. Login salah password -> 401
	resp = do(t, app, "POST", "/auth/login", `{"email":"ana@mail.id","password":"salah"}`, "")
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("login salah = %d; want 401", resp.StatusCode)
	}

	// 4. Login benar -> token
	resp = do(t, app, "POST", "/auth/login", `{"email":"ana@mail.id","password":"secret123"}`, "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("login = %d; want 200", resp.StatusCode)
	}
	var lr struct{ Token string }
	body, _ := io.ReadAll(resp.Body)
	_ = json.Unmarshal(body, &lr)
	if lr.Token == "" {
		t.Fatal("token kosong")
	}

	// 5. Akses /tasks TANPA token -> 401
	resp = do(t, app, "GET", "/tasks/", "", "")
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("tasks tanpa token = %d; want 401", resp.StatusCode)
	}

	// 6. Create task DENGAN token -> 201
	resp = do(t, app, "POST", "/tasks/", `{"title":"belajar auth"}`, lr.Token)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create task = %d; want 201", resp.StatusCode)
	}
	var task model.Task
	body, _ = io.ReadAll(resp.Body)
	_ = json.Unmarshal(body, &task)
	if task.ID == 0 || task.UserID == 0 {
		t.Fatalf("task tak sesuai: %+v", task)
	}

	// 7. List task -> berisi 1
	resp = do(t, app, "GET", "/tasks/", "", lr.Token)
	body, _ = io.ReadAll(resp.Body)
	var list []model.Task
	_ = json.Unmarshal(body, &list)
	if len(list) != 1 {
		t.Errorf("jumlah task = %d; want 1", len(list))
	}

	// 8. Tandai selesai
	resp = do(t, app, "PATCH", "/tasks/1/done", "", lr.Token)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("done = %d; want 200", resp.StatusCode)
	}

	// 9. Hapus -> 204
	resp = do(t, app, "DELETE", "/tasks/1", "", lr.Token)
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("delete = %d; want 204", resp.StatusCode)
	}
}

func TestOtorisasiAntarUser(t *testing.T) {
	app := testApp(t)

	// User A buat task; ambil ID-nya dari response (jangan hardcode).
	tokenA := loginToken(t, app, "a@x.id")
	resp := do(t, app, "POST", "/tasks/", `{"title":"punya A"}`, tokenA)
	var task model.Task
	b, _ := io.ReadAll(resp.Body)
	_ = json.Unmarshal(b, &task)

	// User B login, lalu coba hapus task milik A -> 403 Forbidden.
	tokenB := loginToken(t, app, "b@x.id")
	resp = do(t, app, "DELETE", "/tasks/"+itoa(task.ID), "", tokenB)
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("B hapus task A = %d; want 403", resp.StatusCode)
	}
}

func itoa(u uint) string { return strconv.FormatUint(uint64(u), 10) }
