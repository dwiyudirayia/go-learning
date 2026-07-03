package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSecurityHeaders(t *testing.T) {
	h := buildHandler()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/ping", nil))

	want := map[string]string{
		"X-Content-Type-Options": "nosniff",
		"X-Frame-Options":        "DENY",
	}
	for k, v := range want {
		if got := rec.Header().Get(k); got != v {
			t.Errorf("header %s = %q; want %q", k, got, v)
		}
	}
}

func TestRateLimit(t *testing.T) {
	h := buildHandler()

	// burst = 5 -> 5 request pertama lolos, ke-6 kena 429.
	var last int
	for i := 0; i < 6; i++ {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/ping", nil)
		req.RemoteAddr = "10.0.0.1:12345" // IP sama
		h.ServeHTTP(rec, req)
		last = rec.Code
	}
	if last != http.StatusTooManyRequests {
		t.Errorf("request ke-6 = %d; want 429", last)
	}

	// IP berbeda tidak terpengaruh.
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/ping", nil)
	req.RemoteAddr = "10.0.0.2:9999"
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("IP berbeda = %d; want 200", rec.Code)
	}
}

func TestRefreshFlow(t *testing.T) {
	access, refresh, err := GenerateTokenPair(1)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	// Refresh token valid -> dapat access token baru.
	newAccess, err := RefreshAccessToken(refresh)
	if err != nil || newAccess == "" {
		t.Fatalf("refresh gagal: %v", err)
	}

	// Memakai ACCESS token untuk refresh harus DITOLAK.
	if _, err := RefreshAccessToken(access); err == nil {
		t.Error("access token seharusnya tidak bisa dipakai untuk refresh")
	}
}

func TestRefreshEndpoint(t *testing.T) {
	h := buildHandler()

	// Login -> ambil refresh token.
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("POST", "/login", nil))
	var pair struct{ Access, Refresh string }
	_ = json.Unmarshal(rec.Body.Bytes(), &pair)

	// Tukar refresh -> access baru.
	rec = httptest.NewRecorder()
	body := `{"refresh":"` + pair.Refresh + `"}`
	h.ServeHTTP(rec, httptest.NewRequest("POST", "/refresh", strings.NewReader(body)))
	if rec.Code != http.StatusOK {
		t.Fatalf("/refresh = %d; want 200", rec.Code)
	}
}
