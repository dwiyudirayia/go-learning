// Package handler = lapisan HTTP (Fiber). Tugasnya: parse request, panggil
// service, petakan error bisnis -> status HTTP, tulis response.
package handler

import (
	"errors"

	"github.com/gofiber/fiber/v2"

	"go-learning/15-studi-kasus-rest/internal/middleware"
	"go-learning/15-studi-kasus-rest/internal/service"
)

type Handler struct {
	auth  *service.AuthService
	tasks *service.TaskService
}

func New(auth *service.AuthService, tasks *service.TaskService) *Handler {
	return &Handler{auth: auth, tasks: tasks}
}

// Register memasang semua rute ke app.
func (h *Handler) Register(app *fiber.App) {
	app.Post("/auth/register", h.register)
	app.Post("/auth/login", h.login)

	// Endpoint /me terproteksi (latihan 1).
	app.Get("/me", middleware.Protected(), h.me)

	// Grup task DIPROTEKSI JWT.
	tasks := app.Group("/tasks", middleware.Protected())
	tasks.Get("/", h.listTasks)
	tasks.Post("/", h.createTask)
	tasks.Patch("/:id", h.updateTask) // latihan 2: ubah judul
	tasks.Patch("/:id/done", h.doneTask)
	tasks.Delete("/:id", h.deleteTask)
}

// me (latihan 1): profil user dari token.
func (h *Handler) me(c *fiber.Ctx) error {
	u, err := h.auth.GetByID(middleware.UserID(c))
	if err != nil {
		return mapErr(err)
	}
	return c.JSON(u)
}

// updateTask (latihan 2): ubah judul task milik user.
func (h *Handler) updateTask(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "id harus angka")
	}
	var req createTaskReq
	if err := c.BodyParser(&req); err != nil || req.Title == "" {
		return fiber.NewError(fiber.StatusBadRequest, "title wajib diisi")
	}
	t, err := h.tasks.SetTitle(middleware.UserID(c), uint(id), req.Title)
	if err != nil {
		return mapErr(err)
	}
	return c.JSON(t)
}

// ---------- Auth ----------

type registerReq struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *Handler) register(c *fiber.Ctx) error {
	var req registerReq
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "body tidak valid")
	}
	if req.Email == "" || len(req.Password) < 6 {
		return fiber.NewError(fiber.StatusBadRequest, "email wajib & password minimal 6 karakter")
	}
	u, err := h.auth.Register(req.Name, req.Email, req.Password)
	if err != nil {
		return mapErr(err)
	}
	return c.Status(fiber.StatusCreated).JSON(u)
}

type loginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *Handler) login(c *fiber.Ctx) error {
	var req loginReq
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "body tidak valid")
	}
	tok, err := h.auth.Login(req.Email, req.Password)
	if err != nil {
		return mapErr(err)
	}
	return c.JSON(fiber.Map{"token": tok})
}

// ---------- Task (userID dari token via middleware) ----------

func (h *Handler) listTasks(c *fiber.Ctx) error {
	tasks, err := h.tasks.List(middleware.UserID(c))
	if err != nil {
		return mapErr(err)
	}
	return c.JSON(tasks)
}

type createTaskReq struct {
	Title string `json:"title"`
}

func (h *Handler) createTask(c *fiber.Ctx) error {
	var req createTaskReq
	if err := c.BodyParser(&req); err != nil || req.Title == "" {
		return fiber.NewError(fiber.StatusBadRequest, "title wajib diisi")
	}
	t, err := h.tasks.Create(middleware.UserID(c), req.Title)
	if err != nil {
		return mapErr(err)
	}
	return c.Status(fiber.StatusCreated).JSON(t)
}

func (h *Handler) doneTask(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "id harus angka")
	}
	t, err := h.tasks.SetDone(middleware.UserID(c), uint(id), true)
	if err != nil {
		return mapErr(err)
	}
	return c.JSON(t)
}

func (h *Handler) deleteTask(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "id harus angka")
	}
	if err := h.tasks.Delete(middleware.UserID(c), uint(id)); err != nil {
		return mapErr(err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// mapErr menerjemahkan error bisnis -> fiber.Error dengan status yang tepat.
func mapErr(err error) error {
	switch {
	case errors.Is(err, service.ErrEmailTaken):
		return fiber.NewError(fiber.StatusConflict, err.Error()) // 409
	case errors.Is(err, service.ErrInvalidCredentials):
		return fiber.NewError(fiber.StatusUnauthorized, err.Error()) // 401
	case errors.Is(err, service.ErrForbidden):
		return fiber.NewError(fiber.StatusForbidden, err.Error()) // 403
	case errors.Is(err, service.ErrNotFound):
		return fiber.NewError(fiber.StatusNotFound, err.Error()) // 404
	default:
		return fiber.NewError(fiber.StatusInternalServerError, "kesalahan server")
	}
}
