package main

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func testHandler() http.Handler {
	// Logger yang membuang output agar test senyap.
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	return buildHandler(logger, newRegistry())
}

func TestHelloDanMetrics(t *testing.T) {
	h := testHandler()

	// Panggil /hello beberapa kali.
	for i := 0; i < 3; i++ {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest("GET", "/hello?name=Ana", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("/hello status = %d; want 200", rec.Code)
		}
	}
	// Sekali /boom (status 500).
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/boom", nil))
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("/boom status = %d; want 500", rec.Code)
	}

	// Ambil /metrics dan pastikan counter tercatat.
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/metrics", nil))
	body := rec.Body.String()

	if !strings.Contains(body, "http_requests_total") {
		t.Error("metrics tidak memuat http_requests_total")
	}
	// Harus ada entri untuk /hello dengan status 200.
	if !strings.Contains(body, `http_requests_total{method="GET",path="GET /hello",status="200"} 3`) {
		t.Errorf("counter /hello tidak sesuai; body:\n%s", body)
	}
	if !strings.Contains(body, "http_request_duration_seconds") {
		t.Error("metrics tidak memuat histogram durasi")
	}
}
