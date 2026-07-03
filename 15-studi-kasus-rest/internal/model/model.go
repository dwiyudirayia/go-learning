// Package model berisi entity domain (dipetakan ke tabel oleh GORM).
package model

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Name         string `json:"name"`
	Email        string `json:"email" gorm:"uniqueIndex"`
	PasswordHash string `json:"-"` // tanda "-" -> TIDAK pernah ikut JSON
	Tasks        []Task `json:"tasks,omitempty"`
}

type Task struct {
	gorm.Model
	Title  string `json:"title"`
	Done   bool   `json:"done"`
	UserID uint   `json:"user_id"`
}
