package main

import (
	"context"
	"time"
)

// Scheduler menjalankan sebuah fungsi secara berkala (seperti cron sederhana).
// Untuk ekspresi cron sungguhan ("0 */5 * * *"), pakai github.com/robfig/cron.
type Scheduler struct {
	interval time.Duration
	task     func()
}

func NewScheduler(interval time.Duration, task func()) *Scheduler {
	return &Scheduler{interval: interval, task: task}
}

// Run menjalankan task tiap interval sampai ctx dibatalkan (graceful, Modul 20).
func (s *Scheduler) Run(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			s.task()
		case <-ctx.Done():
			return // berhenti rapi
		}
	}
}
