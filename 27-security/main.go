// Jalankan: go run ./27-security
// Verifikasi otomatis: go test ./27-security
package main

import (
	"encoding/json"
	"log"
	"net/http"

	"golang.org/x/time/rate"
)

func buildHandler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong\n"))
	})

	// Simulasi login -> kembalikan pasangan token.
	mux.HandleFunc("POST /login", func(w http.ResponseWriter, r *http.Request) {
		access, refresh, err := GenerateTokenPair(1)
		if err != nil {
			http.Error(w, "gagal", http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"access": access, "refresh": refresh})
	})

	// Tukar refresh token dengan access token baru.
	mux.HandleFunc("POST /refresh", func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Refresh string `json:"refresh"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		access, err := RefreshAccessToken(body.Refresh)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"access": access})
	})

	// Rantai middleware: security headers -> rate limit (5 req/detik, burst 5) -> router.
	limiter := newIPRateLimiter(rate.Limit(5), 5)
	return securityHeaders(limiter.middleware(mux))
}

func main() {
	log.Println("server keamanan di http://localhost:8080")
	if err := http.ListenAndServe(":8080", buildHandler()); err != nil {
		log.Fatal(err)
	}
}
