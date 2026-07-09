// Teknik Advanced Modul 15 (studi kasus REST) — contoh runnable + penjelasan detail.
//
// Menyoroti arsitektur BERLAPIS + pengujian dengan MOCK repository + pemetaan
// error domain -> status HTTP + hashing bcrypt. Jalankan:
//
//	go run ./15-studi-kasus-rest/advanced
package main

import (
	"errors"
	"fmt"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

// ===========================================================================
// Error domain (sentinel) — service bicara dalam bahasa domain, bukan HTTP.
// ===========================================================================
var (
	ErrValidasi = errors.New("input tidak valid")
	ErrDuplikat = errors.New("email sudah terdaftar")
	ErrAuth     = errors.New("kredensial salah")
)

type User struct {
	Email  string
	HashPw string // TIDAK pernah simpan password plaintext
}

// ===========================================================================
// 1. Lapisan REPOSITORY sebagai INTERFACE -> mudah di-mock saat test
// ===========================================================================
// Service bergantung pada abstraksi ini, bukan implementasi konkret. Di produksi
// diisi repo Postgres/SQLite; di test diisi mock in-memory (lihat memRepo).
type UserRepo interface {
	Simpan(u User) error
	CariByEmail(email string) (User, bool)
}

type memRepo struct{ data map[string]User }

func newMemRepo() *memRepo { return &memRepo{data: map[string]User{}} }

func (r *memRepo) Simpan(u User) error {
	if _, ada := r.data[u.Email]; ada {
		return ErrDuplikat
	}
	r.data[u.Email] = u
	return nil
}
func (r *memRepo) CariByEmail(e string) (User, bool) { u, ok := r.data[e]; return u, ok }

// ===========================================================================
// 2. Lapisan SERVICE — aturan bisnis murni (tak tahu HTTP)
// ===========================================================================
type AuthService struct{ repo UserRepo }

func (s *AuthService) Register(email, pw string) error {
	if email == "" || len(pw) < 6 {
		return fmt.Errorf("%w: email wajib & password minimal 6 karakter", ErrValidasi)
	}
	// bcrypt: hash lambat & bergaram. MinCost dipakai agar demo cepat; di
	// produksi pilih cost lebih tinggi (mis. 12) sesuai anggaran latensi.
	hash, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.MinCost)
	if err != nil {
		return err
	}
	return s.repo.Simpan(User{Email: email, HashPw: string(hash)})
}

func (s *AuthService) Login(email, pw string) error {
	u, ok := s.repo.CariByEmail(email)
	if !ok {
		return ErrAuth // JANGAN bocorkan "email tak ada" vs "password salah"
	}
	// CompareHashAndPassword = perbandingan tahan timing-attack.
	if bcrypt.CompareHashAndPassword([]byte(u.HashPw), []byte(pw)) != nil {
		return ErrAuth
	}
	return nil
}

// ===========================================================================
// 3. Lapisan HANDLER (batas HTTP) — SATU tempat memetakan error -> status
// ===========================================================================
func statusUntuk(err error) int {
	switch {
	case err == nil:
		return http.StatusOK
	case errors.Is(err, ErrValidasi):
		return http.StatusBadRequest // 400
	case errors.Is(err, ErrDuplikat):
		return http.StatusConflict // 409
	case errors.Is(err, ErrAuth):
		return http.StatusUnauthorized // 401
	default:
		return http.StatusInternalServerError // 500
	}
}

func main() {
	// Rakit dependency (composition root): repo mock -> service.
	svc := &AuthService{repo: newMemRepo()}

	type skenario struct {
		nama    string
		jalanin func() error
	}
	langkah := []skenario{
		{"register valid", func() error { return svc.Register("budi@mail.com", "rahasia123") }},
		{"register duplikat", func() error { return svc.Register("budi@mail.com", "lainlain") }},
		{"register password pendek", func() error { return svc.Register("ani@mail.com", "123") }},
		{"login benar", func() error { return svc.Login("budi@mail.com", "rahasia123") }},
		{"login password salah", func() error { return svc.Login("budi@mail.com", "salah") }},
		{"login email asing", func() error { return svc.Login("hantu@mail.com", "apa saja") }},
	}

	fmt.Println("STATUS  SKENARIO                     PESAN")
	for _, s := range langkah {
		err := s.jalanin()
		pesan := "ok"
		if err != nil {
			pesan = err.Error()
		}
		fmt.Printf("%-7d %-28s %s\n", statusUntuk(err), s.nama, pesan)
	}
}
