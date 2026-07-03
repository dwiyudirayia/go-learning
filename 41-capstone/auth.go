package main

import (
	"errors"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// --- Password hashing (Modul 15/27) ---
func hashPassword(pw string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	return string(b), err
}
func checkPassword(hash, pw string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(pw)) == nil
}

// --- JWT (Modul 15) ---
func jwtSecret() []byte {
	if s := os.Getenv("JWT_SECRET"); s != "" {
		return []byte(s)
	}
	return []byte("capstone-dev-secret-ganti-di-produksi")
}

type claims struct {
	UserID int64 `json:"uid"`
	jwt.RegisteredClaims
}

func generateToken(userID int64) (string, error) {
	c := claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, c).SignedString(jwtSecret())
}

var errInvalidToken = errors.New("token tidak valid")

func parseToken(tok string) (int64, error) {
	var c claims
	t, err := jwt.ParseWithClaims(tok, &c, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errInvalidToken
		}
		return jwtSecret(), nil
	})
	if err != nil || !t.Valid {
		return 0, errInvalidToken
	}
	return c.UserID, nil
}

// Protected: middleware Fiber yang mewajibkan Bearer token valid.
func Protected() fiber.Handler {
	return func(c *fiber.Ctx) error {
		auth := c.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			return fiber.NewError(fiber.StatusUnauthorized, "butuh token Bearer")
		}
		uid, err := parseToken(strings.TrimPrefix(auth, "Bearer "))
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, "token tidak valid")
		}
		c.Locals("userID", uid)
		return c.Next()
	}
}
