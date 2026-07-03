package main

import (
	"context"
	"errors"
	"testing"
)

func TestSagaSukses(t *testing.T) {
	var order []string
	s := NewSaga().
		AddStep(Step{Name: "a", Action: func(context.Context) error { order = append(order, "a"); return nil }}).
		AddStep(Step{Name: "b", Action: func(context.Context) error { order = append(order, "b"); return nil }})

	if err := s.Execute(context.Background()); err != nil {
		t.Fatalf("saga: %v", err)
	}
	if len(order) != 2 || order[0] != "a" || order[1] != "b" {
		t.Errorf("urutan = %v; want [a b]", order)
	}
}

func TestSagaGagalKompensasiTerbalik(t *testing.T) {
	var events []string

	s := NewSaga().
		AddStep(Step{
			Name:       "reserve",
			Action:     func(context.Context) error { events = append(events, "do:reserve"); return nil },
			Compensate: func(context.Context) error { events = append(events, "undo:reserve"); return nil },
		}).
		AddStep(Step{
			Name:       "pay",
			Action:     func(context.Context) error { events = append(events, "do:pay"); return nil },
			Compensate: func(context.Context) error { events = append(events, "undo:pay"); return nil },
		}).
		AddStep(Step{
			Name:   "ship",
			Action: func(context.Context) error { return errors.New("gagal") }, // gagal di sini
		})

	err := s.Execute(context.Background())
	if err == nil {
		t.Fatal("mengharapkan error")
	}

	// Kompensasi harus mundur: undo:pay dulu, baru undo:reserve.
	want := []string{"do:reserve", "do:pay", "undo:pay", "undo:reserve"}
	if len(events) != len(want) {
		t.Fatalf("events = %v; want %v", events, want)
	}
	for i := range want {
		if events[i] != want[i] {
			t.Errorf("events[%d] = %q; want %q", i, events[i], want[i])
		}
	}
}
