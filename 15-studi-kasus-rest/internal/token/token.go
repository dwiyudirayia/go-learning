// Package token membungkus pembuatan & verifikasi JWT (HS256).
package token

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// secret diambil dari env; ada default untuk dev. DI PRODUKSI WAJIB set JWT_SECRET.
func secret() []byte {
	if s := os.Getenv("JWT_SECRET"); s != "" {
		return []byte(s)
	}
	return []byte("dev-secret-ganti-di-produksi")
}

// Claims: data yang kita simpan di dalam token.
type Claims struct {
	UserID uint `json:"uid"`
	jwt.RegisteredClaims
}

// Generate membuat token yang berlaku 24 jam untuk userID tertentu.
func Generate(userID uint) (string, error) {
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(secret())
}

var ErrInvalidToken = errors.New("token tidak valid")

// Parse memverifikasi token & mengembalikan userID di dalamnya.
func Parse(tokenStr string) (uint, error) {
	var claims Claims
	tok, err := jwt.ParseWithClaims(tokenStr, &claims, func(t *jwt.Token) (any, error) {
		// Pastikan algoritma sesuai (cegah serangan "alg=none").
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return secret(), nil
	})
	if err != nil || !tok.Valid {
		return 0, ErrInvalidToken
	}
	return claims.UserID, nil
}
