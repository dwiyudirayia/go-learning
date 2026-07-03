// Package middleware berisi middleware Fiber, termasuk proteksi JWT.
package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"go-learning/15-studi-kasus-rest/internal/token"
)

// LocalUserID: key untuk menyimpan userID hasil verifikasi token di context.
const LocalUserID = "userID"

// Protected memverifikasi header Authorization: Bearer <token>.
// Bila valid, userID disimpan di c.Locals untuk dipakai handler.
func Protected() fiber.Handler {
	return func(c *fiber.Ctx) error {
		auth := c.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			return fiber.NewError(fiber.StatusUnauthorized, "butuh token Bearer")
		}
		raw := strings.TrimPrefix(auth, "Bearer ")

		userID, err := token.Parse(raw)
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, "token tidak valid")
		}
		c.Locals(LocalUserID, userID)
		return c.Next()
	}
}

// UserID mengambil userID yang sudah diverifikasi dari context.
func UserID(c *fiber.Ctx) uint {
	id, _ := c.Locals(LocalUserID).(uint)
	return id
}
