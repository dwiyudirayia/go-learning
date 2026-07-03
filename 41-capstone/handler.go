package main

import (
	"errors"
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

// buildApp merangkai seluruh rute + middleware. Dipisah dari main agar bisa
// diuji dengan app.Test (Modul 13).
func buildApp(svc *Service, logger *slog.Logger) *fiber.App {
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if fe, ok := err.(*fiber.Error); ok {
				code = fe.Code
			}
			return c.Status(code).JSON(fiber.Map{"error": err.Error()})
		},
	})
	app.Use(recover.New())

	// Health (Modul 30).
	app.Get("/healthz", func(c *fiber.Ctx) error { return c.SendString("ok") })

	// Auth.
	app.Post("/auth/register", func(c *fiber.Ctx) error {
		var req struct{ Email, Password string }
		if err := c.BodyParser(&req); err != nil || req.Email == "" || len(req.Password) < 6 {
			return fiber.NewError(fiber.StatusBadRequest, "email & password (min 6) wajib")
		}
		u, err := svc.Register(req.Email, req.Password)
		if errors.Is(err, ErrEmailTaken) {
			return fiber.NewError(fiber.StatusConflict, err.Error())
		}
		if err != nil {
			return err
		}
		return c.Status(fiber.StatusCreated).JSON(fiber.Map{"id": u.ID, "email": u.Email})
	})

	app.Post("/auth/login", func(c *fiber.Ctx) error {
		var req struct{ Email, Password string }
		_ = c.BodyParser(&req)
		tok, err := svc.Login(req.Email, req.Password)
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, "kredensial salah")
		}
		return c.JSON(fiber.Map{"token": tok})
	})

	// Buat short link (DIPROTEKSI JWT).
	app.Post("/api/shorten", Protected(), func(c *fiber.Ctx) error {
		var req struct{ URL string }
		if err := c.BodyParser(&req); err != nil || req.URL == "" {
			return fiber.NewError(fiber.StatusBadRequest, "url wajib diisi")
		}
		userID := c.Locals("userID").(int64)
		link, err := svc.Shorten(userID, req.URL)
		if err != nil {
			return err
		}
		logger.Info("link dibuat", slog.String("code", link.Code), slog.Int64("user", userID))
		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"code":     link.Code,
			"short":    "/" + link.Code,
			"long_url": link.URL,
		})
	})

	// Redirect (publik). Cache-aside di balik layar.
	app.Get("/:code", func(c *fiber.Ctx) error {
		url, err := svc.Resolve(c.Context(), c.Params("code"))
		if errors.Is(err, ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "link tidak ditemukan")
		}
		if err != nil {
			return err
		}
		return c.Redirect(url, fiber.StatusFound) // 302
	})

	return app
}
