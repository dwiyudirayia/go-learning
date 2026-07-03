package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthProbes(t *testing.T) {
	h := buildHandler()

	for _, path := range []string{"/healthz", "/readyz"} {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest("GET", path, nil))
		if rec.Code != http.StatusOK {
			t.Errorf("%s = %d; want 200", path, rec.Code)
		}
	}
}

func TestReadinessToggle(t *testing.T) {
	h := buildHandler()

	// Simulasi shutdown: tandai tidak siap.
	ready.Store(false)
	defer ready.Store(true)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/readyz", nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("/readyz saat not-ready = %d; want 503", rec.Code)
	}
	// Liveness harus tetap OK (proses masih hidup).
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/healthz", nil))
	if rec.Code != http.StatusOK {
		t.Errorf("/healthz = %d; want 200", rec.Code)
	}
}

func TestVersion(t *testing.T) {
	h := buildHandler()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/version", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("/version = %d; want 200", rec.Code)
	}
}
