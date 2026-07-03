// Package store menyimpan daftar task ke file JSON (persistensi sederhana).
package store

import (
	"encoding/json"
	"errors"
	"os"
)

type Task struct {
	ID       int    `json:"id"`
	Text     string `json:"text"`
	Done     bool   `json:"done"`
	Priority int    `json:"priority,omitempty"` // latihan 2: prioritas (0 = normal)
}

// Store membungkus path file + daftar task di memori.
type Store struct {
	path  string
	Tasks []Task `json:"tasks"`
}

var ErrNotFound = errors.New("task tidak ditemukan")

// Load membaca task dari file. Bila file belum ada, mulai dari kosong.
func Load(path string) (*Store, error) {
	s := &Store{path: path}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return s, nil // file baru pertama kali
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, &s.Tasks); err != nil {
		return nil, err
	}
	return s, nil
}

// save menulis kembali ke file (pretty JSON).
func (s *Store) save() error {
	data, err := json.MarshalIndent(s.Tasks, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0o644)
}

// Add menambah task baru dan mengembalikan ID-nya.
func (s *Store) Add(text string) (int, error) {
	return s.AddWithPriority(text, 0)
}

// AddWithPriority (latihan 2) menambah task dengan prioritas tertentu.
func (s *Store) AddWithPriority(text string, priority int) (int, error) {
	id := s.nextID()
	s.Tasks = append(s.Tasks, Task{ID: id, Text: text, Priority: priority})
	return id, s.save()
}

// MarkDone menandai task selesai berdasarkan ID.
func (s *Store) MarkDone(id int) error {
	for i := range s.Tasks {
		if s.Tasks[i].ID == id {
			s.Tasks[i].Done = true
			return s.save()
		}
	}
	return ErrNotFound
}

// Remove (latihan 1) menghapus task berdasarkan ID.
func (s *Store) Remove(id int) error {
	for i := range s.Tasks {
		if s.Tasks[i].ID == id {
			s.Tasks = append(s.Tasks[:i], s.Tasks[i+1:]...) // hapus elemen ke-i
			return s.save()
		}
	}
	return ErrNotFound
}

// Stats (latihan 3) mengembalikan jumlah total, selesai, dan belum selesai.
func (s *Store) Stats() (total, done, pending int) {
	total = len(s.Tasks)
	for _, t := range s.Tasks {
		if t.Done {
			done++
		}
	}
	pending = total - done
	return
}

func (s *Store) nextID() int {
	max := 0
	for _, t := range s.Tasks {
		if t.ID > max {
			max = t.ID
		}
	}
	return max + 1
}
