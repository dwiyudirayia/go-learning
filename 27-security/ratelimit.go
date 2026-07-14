// Modul 27 — Security: rate limiting, security headers, refresh token.
package main

import (
	"net"
	"net/http"
	"sync"

	"golang.org/x/time/rate"
)

// 🔍 Analogi besar "token bucket": bayangkan tiap IP punya EMBER berisi koin. Tiap request
// memakai 1 koin. Ember diisi ulang pelan-pelan (r koin/detik) dan muat maksimal b koin (burst).
// Selama masih ada koin -> boleh lewat. Ember kosong -> tolak (429). Ini membiarkan "ledakan
// singkat" wajar (b koin sekaligus) tapi menahan pengguna yang menembak terus-menerus (brute-force/DoS).
// Analogi lain: pintu tol dgn kuota — beberapa mobil boleh lewat cepat, tapi tak boleh membanjiri.

// ipRateLimiter membatasi laju request PER IP memakai algoritma token bucket.
// Mencegah abuse/brute-force/DoS ringan.
type ipRateLimiter struct {
	mu       sync.Mutex
	limiters map[string]*rate.Limiter
	r        rate.Limit // token per detik
	b        int        // burst (kapasitas ember)
}

func newIPRateLimiter(r rate.Limit, b int) *ipRateLimiter {
	return &ipRateLimiter{limiters: make(map[string]*rate.Limiter), r: r, b: b}
}

func (l *ipRateLimiter) get(ip string) *rate.Limiter {
	l.mu.Lock()
	defer l.mu.Unlock()
	lim, ok := l.limiters[ip]
	if !ok {
		lim = rate.NewLimiter(l.r, l.b)
		l.limiters[ip] = lim
	}
	return lim
}

// middleware menolak request dengan 429 bila IP melampaui batas.
func (l *ipRateLimiter) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := clientIP(r)
		if !l.get(ip).Allow() {
			http.Error(w, "terlalu banyak request", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
