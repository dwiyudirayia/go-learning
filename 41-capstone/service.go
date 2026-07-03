package main

import (
	"context"
	"crypto/rand"
	"errors"
)

var (
	ErrEmailTaken   = errors.New("email sudah terdaftar")
	ErrInvalidCreds = errors.New("email atau password salah")
)

// Service = lapisan logika bisnis (Modul 29). Memakai Store (data) + Cache.
type Service struct {
	store *Store
	cache *Cache
}

func NewService(store *Store, cache *Cache) *Service {
	return &Service{store: store, cache: cache}
}

func (s *Service) Register(email, password string) (User, error) {
	if _, err := s.store.GetUserByEmail(email); err == nil {
		return User{}, ErrEmailTaken
	}
	hash, err := hashPassword(password)
	if err != nil {
		return User{}, err
	}
	return s.store.CreateUser(email, hash)
}

func (s *Service) Login(email, password string) (string, error) {
	u, err := s.store.GetUserByEmail(email)
	if err != nil || !checkPassword(u.PasswordHash, password) {
		return "", ErrInvalidCreds
	}
	return generateToken(u.ID)
}

const base62 = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func randCode(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	for i := range b {
		b[i] = base62[int(b[i])%len(base62)]
	}
	return string(b)
}

// Shorten membuat kode pendek untuk sebuah URL.
func (s *Service) Shorten(userID int64, url string) (Link, error) {
	code := randCode(6)
	return s.store.CreateLink(code, url, userID)
}

// Resolve menerapkan pola CACHE-ASIDE (Modul 22):
//  1. cek cache; kalau ada -> pakai (cepat).
//  2. kalau miss -> baca DB, isi cache.
//
// (Produksi: hitung klik asinkron via worker queue/channel — Modul 25 — agar
// redirect tak menunggu tulis DB. Di sini sinkron demi kesederhanaan & test.)
func (s *Service) Resolve(ctx context.Context, code string) (string, error) {
	if url, ok, err := s.cache.GetURL(ctx, code); err == nil && ok {
		_ = s.store.IncrementClicks(code)
		return url, nil
	}

	link, err := s.store.GetLinkByCode(code)
	if err != nil {
		return "", err
	}
	s.cache.SetURL(ctx, code, link.URL)
	_ = s.store.IncrementClicks(code)
	return link.URL, nil
}
