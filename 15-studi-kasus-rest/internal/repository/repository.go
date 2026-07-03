// Package repository = lapisan akses data. Service bergantung pada INTERFACE
// di sini (bukan GORM langsung) -> mudah di-mock saat test (Modul 4 & 8).
package repository

import (
	"errors"

	"gorm.io/gorm"

	"go-learning/15-studi-kasus-rest/internal/model"
)

var ErrNotFound = errors.New("data tidak ditemukan")

type UserRepository interface {
	Create(u *model.User) error
	FindByEmail(email string) (*model.User, error)
	FindByID(id uint) (*model.User, error)
}

type TaskRepository interface {
	Create(t *model.Task) error
	ListByUser(userID uint) ([]model.Task, error)
	FindByID(id uint) (*model.Task, error)
	Update(t *model.Task) error
	Delete(t *model.Task) error
}

// ---------- Implementasi GORM ----------

type gormUserRepo struct{ db *gorm.DB }

func NewUserRepository(db *gorm.DB) UserRepository { return &gormUserRepo{db} }

func (r *gormUserRepo) Create(u *model.User) error { return r.db.Create(u).Error }

func (r *gormUserRepo) FindByEmail(email string) (*model.User, error) {
	var u model.User
	err := r.db.Where("email = ?", email).First(&u).Error
	return wrap(&u, err)
}

func (r *gormUserRepo) FindByID(id uint) (*model.User, error) {
	var u model.User
	err := r.db.First(&u, id).Error
	return wrap(&u, err)
}

type gormTaskRepo struct{ db *gorm.DB }

func NewTaskRepository(db *gorm.DB) TaskRepository { return &gormTaskRepo{db} }

func (r *gormTaskRepo) Create(t *model.Task) error { return r.db.Create(t).Error }

func (r *gormTaskRepo) ListByUser(userID uint) ([]model.Task, error) {
	var tasks []model.Task
	err := r.db.Where("user_id = ?", userID).Order("id").Find(&tasks).Error
	return tasks, err
}

func (r *gormTaskRepo) FindByID(id uint) (*model.Task, error) {
	var t model.Task
	err := r.db.First(&t, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &t, err
}

func (r *gormTaskRepo) Update(t *model.Task) error { return r.db.Save(t).Error }
func (r *gormTaskRepo) Delete(t *model.Task) error { return r.db.Delete(t).Error }

// wrap menerjemahkan gorm.ErrRecordNotFound ke ErrNotFound repo kita.
func wrap(u *model.User, err error) (*model.User, error) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return u, nil
}
