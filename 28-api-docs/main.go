// Modul 28 — API Docs: OpenAPI/Swagger, versioning, format error konsisten.
//
// Jalankan: go run ./28-api-docs
//
//	http://localhost:8080/docs           -> Swagger UI (dokumentasi interaktif)
//	http://localhost:8080/openapi.json   -> spesifikasi OpenAPI
//	http://localhost:8080/api/v1/books   -> API (berversi)
//
// Verifikasi otomatis: go test ./28-api-docs
package main

import (
	_ "embed" // untuk //go:embed
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
)

type Book struct {
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Author string `json:"author"`
}

// APIError: format error KONSISTEN untuk seluruh API. Klien selalu tahu bentuknya.
type APIError struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func writeError(w http.ResponseWriter, status int, code, msg string) {
	var e APIError
	e.Error.Code = code
	e.Error.Message = msg
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(e)
}

type store struct {
	mu    sync.Mutex
	books []Book
	seq   int64
}

//go:embed openapi.json
var openapiSpec []byte

func buildHandler() http.Handler {
	s := &store{}
	mux := http.NewServeMux()

	// API BERVERSI: /api/v1/... -> saat ada breaking change, buat /api/v2 tanpa
	// merusak klien lama.
	mux.HandleFunc("GET /api/v1/books", func(w http.ResponseWriter, r *http.Request) {
		s.mu.Lock()
		defer s.mu.Unlock()
		writeJSON(w, http.StatusOK, s.books)
	})

	mux.HandleFunc("POST /api/v1/books", func(w http.ResponseWriter, r *http.Request) {
		var b Book
		if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_JSON", "body bukan JSON valid")
			return
		}
		if b.Title == "" || b.Author == "" {
			writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "title & author wajib diisi")
			return
		}
		b.ID = int(atomic.AddInt64(&s.seq, 1))
		s.mu.Lock()
		s.books = append(s.books, b)
		s.mu.Unlock()
		writeJSON(w, http.StatusCreated, b)
	})

	// Spesifikasi OpenAPI (mesin & tools membacanya).
	mux.HandleFunc("GET /openapi.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(openapiSpec)
	})

	// Swagger UI: dokumentasi interaktif (memuat spec dari /openapi.json via CDN).
	mux.HandleFunc("GET /docs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(swaggerHTML))
	})

	return mux
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

const swaggerHTML = `<!doctype html><html><head>
<link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist/swagger-ui.css">
</head><body><div id="ui"></div>
<script src="https://unpkg.com/swagger-ui-dist/swagger-ui-bundle.js"></script>
<script>SwaggerUIBundle({url:'/openapi.json',dom_id:'#ui'});</script>
</body></html>`

func main() {
	log.Println("Swagger UI: http://localhost:8080/docs")
	if err := http.ListenAndServe(":8080", buildHandler()); err != nil {
		log.Fatal(err)
	}
}
