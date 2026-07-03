// Package service = lapisan logika bisnis. Tidak tahu soal HTTP maupun GORM;
// hanya bergantung pada interface repository.
package service

import (
	"errors"

	"golang.org/x/crypto/bcrypt"

	"go-learning/15-studi-kasus-rest/internal/model"
	"go-learning/15-studi-kasus-rest/internal/repository"
	"go-learning/15-studi-kasus-rest/internal/token"
)

// Sentinel error bisnis -> nanti dipetakan ke status HTTP oleh handler.
var (
	ErrEmailTaken         = errors.New("email sudah terdaftar")
	ErrInvalidCredentials = errors.New("email atau password salah")
	ErrNotFound           = errors.New("tidak ditemukan")
	ErrForbidden          = errors.New("tidak punya akses")
)

// ---------------- Auth ----------------

type AuthService struct{ users repository.UserRepository }

func NewAuthService(users repository.UserRepository) *AuthService {
	return &AuthService{users: users}
}

func (s *AuthService) Register(name, email, password string) (*model.User, error) {
	// Cek email unik.
	if _, err := s.users.FindByEmail(email); err == nil {
		return nil, ErrEmailTaken
	}
	// Hash password (JANGAN simpan plaintext).
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	u := &model.User{Name: name, Email: email, PasswordHash: string(hash)}
	if err := s.users.Create(u); err != nil {
		return nil, err
	}
	return u, nil
}

// GetByID (latihan 1) mengambil profil user untuk endpoint /me.
func (s *AuthService) GetByID(id uint) (*model.User, error) {
	u, err := s.users.FindByID(id)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrNotFound
	}
	return u, err
}

// Login memverifikasi kredensial & mengembalikan JWT.
func (s *AuthService) Login(email, password string) (string, error) {
	u, err := s.users.FindByEmail(email)
	if err != nil {
		return "", ErrInvalidCredentials // jangan bocorkan "email tak ada"
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) != nil {
		return "", ErrInvalidCredentials
	}
	return token.Generate(u.ID)
}

// ---------------- Task (semua operasi terikat pemilik) ----------------

type TaskService struct{ tasks repository.TaskRepository }

func NewTaskService(tasks repository.TaskRepository) *TaskService {
	return &TaskService{tasks: tasks}
}

func (s *TaskService) Create(userID uint, title string) (*model.Task, error) {
	t := &model.Task{Title: title, UserID: userID}
	if err := s.tasks.Create(t); err != nil {
		return nil, err
	}
	return t, nil
}

func (s *TaskService) List(userID uint) ([]model.Task, error) {
	return s.tasks.ListByUser(userID)
}

// ownedTask memastikan task ada DAN milik user tsb (otorisasi).
func (s *TaskService) ownedTask(userID, id uint) (*model.Task, error) {
	t, err := s.tasks.FindByID(id)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if t.UserID != userID {
		return nil, ErrForbidden // task orang lain
	}
	return t, nil
}

func (s *TaskService) SetDone(userID, id uint, done bool) (*model.Task, error) {
	t, err := s.ownedTask(userID, id)
	if err != nil {
		return nil, err
	}
	t.Done = done
	if err := s.tasks.Update(t); err != nil {
		return nil, err
	}
	return t, nil
}

// SetTitle (latihan 2) mengubah judul task milik user (dengan cek kepemilikan).
func (s *TaskService) SetTitle(userID, id uint, title string) (*model.Task, error) {
	t, err := s.ownedTask(userID, id)
	if err != nil {
		return nil, err
	}
	t.Title = title
	if err := s.tasks.Update(t); err != nil {
		return nil, err
	}
	return t, nil
}

func (s *TaskService) Delete(userID, id uint) error {
	t, err := s.ownedTask(userID, id)
	if err != nil {
		return err
	}
	return s.tasks.Delete(t)
}
