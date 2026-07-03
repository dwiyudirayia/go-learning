// Package rest = ADAPTER (driving): menerjemahkan HTTP <-> use case.
// Ini "menggerakkan" aplikasi dari luar. Bisa diganti gRPC/CLI tanpa mengubah core.
package rest

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"go-learning/29-clean-architecture/internal/domain"
	"go-learning/29-clean-architecture/internal/service"
)

type Handler struct {
	svc *service.NoteService
}

func New(svc *service.NoteService) *Handler { return &Handler{svc: svc} }

func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /notes", h.list)
	mux.HandleFunc("POST /notes", h.create)
	mux.HandleFunc("GET /notes/{id}", h.get)
	return mux
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	notes, err := h.svc.List()
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, 200, notes)
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var in struct{ Title, Body string }
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeJSON(w, 400, map[string]string{"error": "JSON tidak valid"})
		return
	}
	n, err := h.svc.Create(in.Title, in.Body)
	if errors.Is(err, domain.ErrEmptyTitle) {
		writeJSON(w, 400, map[string]string{"error": err.Error()})
		return
	}
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, 201, n)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeJSON(w, 400, map[string]string{"error": "id harus angka"})
		return
	}
	n, err := h.svc.Get(id)
	if errors.Is(err, domain.ErrNotFound) {
		writeJSON(w, 404, map[string]string{"error": err.Error()})
		return
	}
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, 200, n)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
