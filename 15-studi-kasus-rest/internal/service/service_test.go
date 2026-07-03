// Latihan 3: unit test SERVICE memakai MOCK repository (tanpa database/GORM).
// Membuktikan logika & otorisasi bisa diuji cepat tanpa I/O.
package service

import (
	"errors"
	"testing"

	"go-learning/15-studi-kasus-rest/internal/model"
	"go-learning/15-studi-kasus-rest/internal/repository"
)

// ---- Mock UserRepository (implementasi interface, in-memory) ----
type fakeUserRepo struct {
	byEmail map[string]*model.User
	byID    map[uint]*model.User
	seq     uint
}

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{byEmail: map[string]*model.User{}, byID: map[uint]*model.User{}}
}

func (f *fakeUserRepo) Create(u *model.User) error {
	f.seq++
	u.ID = f.seq
	f.byEmail[u.Email] = u
	f.byID[u.ID] = u
	return nil
}
func (f *fakeUserRepo) FindByEmail(email string) (*model.User, error) {
	if u, ok := f.byEmail[email]; ok {
		return u, nil
	}
	return nil, repository.ErrNotFound
}
func (f *fakeUserRepo) FindByID(id uint) (*model.User, error) {
	if u, ok := f.byID[id]; ok {
		return u, nil
	}
	return nil, repository.ErrNotFound
}

// ---- Mock TaskRepository ----
type fakeTaskRepo struct {
	tasks map[uint]*model.Task
	seq   uint
}

func newFakeTaskRepo() *fakeTaskRepo { return &fakeTaskRepo{tasks: map[uint]*model.Task{}} }

func (f *fakeTaskRepo) Create(t *model.Task) error {
	f.seq++
	t.ID = f.seq
	f.tasks[t.ID] = t
	return nil
}
func (f *fakeTaskRepo) ListByUser(userID uint) ([]model.Task, error) {
	var out []model.Task
	for _, t := range f.tasks {
		if t.UserID == userID {
			out = append(out, *t)
		}
	}
	return out, nil
}
func (f *fakeTaskRepo) FindByID(id uint) (*model.Task, error) {
	if t, ok := f.tasks[id]; ok {
		return t, nil
	}
	return nil, repository.ErrNotFound
}
func (f *fakeTaskRepo) Update(t *model.Task) error { f.tasks[t.ID] = t; return nil }
func (f *fakeTaskRepo) Delete(t *model.Task) error { delete(f.tasks, t.ID); return nil }

// ---- Test ----

func TestAuthService_RegisterLogin(t *testing.T) {
	svc := NewAuthService(newFakeUserRepo())

	if _, err := svc.Register("Ana", "ana@x.id", "secret123"); err != nil {
		t.Fatalf("register: %v", err)
	}
	// Email sama -> ErrEmailTaken.
	if _, err := svc.Register("Ana2", "ana@x.id", "secret123"); !errors.Is(err, ErrEmailTaken) {
		t.Errorf("register duplikat = %v; want ErrEmailTaken", err)
	}
	// Login benar -> token.
	tok, err := svc.Login("ana@x.id", "secret123")
	if err != nil || tok == "" {
		t.Fatalf("login benar: tok=%q err=%v", tok, err)
	}
	// Login salah -> ErrInvalidCredentials.
	if _, err := svc.Login("ana@x.id", "salah"); !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("login salah = %v; want ErrInvalidCredentials", err)
	}
}

func TestTaskService_Otorisasi(t *testing.T) {
	svc := NewTaskService(newFakeTaskRepo())

	// User 1 buat task.
	task, err := svc.Create(1, "punya user 1")
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	// User 2 mencoba mengubah/menghapus -> ErrForbidden.
	if _, err := svc.SetTitle(2, task.ID, "ubah"); !errors.Is(err, ErrForbidden) {
		t.Errorf("SetTitle user lain = %v; want ErrForbidden", err)
	}
	if err := svc.Delete(2, task.ID); !errors.Is(err, ErrForbidden) {
		t.Errorf("Delete user lain = %v; want ErrForbidden", err)
	}

	// Task tak ada -> ErrNotFound.
	if err := svc.Delete(1, 999); !errors.Is(err, ErrNotFound) {
		t.Errorf("Delete tak ada = %v; want ErrNotFound", err)
	}

	// Pemilik sendiri boleh.
	if _, err := svc.SetTitle(1, task.ID, "judul baru"); err != nil {
		t.Errorf("pemilik SetTitle: %v", err)
	}
}
