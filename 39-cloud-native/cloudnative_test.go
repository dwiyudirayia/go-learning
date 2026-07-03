package main

import (
	"context"
	"testing"
)

func TestReconcileScaleUp(t *testing.T) {
	ps := NewPodSet(3)
	ps.Reconcile()
	if ps.Actual() != 3 {
		t.Errorf("actual = %d; want 3", ps.Actual())
	}
}

func TestReconcileScaleDown(t *testing.T) {
	ps := NewPodSet(5)
	ps.Reconcile() // naik ke 5
	ps.SetDesired(2)
	ps.Reconcile() // turun ke 2
	if ps.Actual() != 2 {
		t.Errorf("actual = %d; want 2", ps.Actual())
	}
}

func TestReconcileIdempotent(t *testing.T) {
	ps := NewPodSet(3)
	ps.Reconcile()
	n := len(ps.Events())
	ps.Reconcile() // sudah selaras -> tak ada aksi
	if len(ps.Events()) != n {
		t.Errorf("Reconcile saat selaras menambah %d event; want 0", len(ps.Events())-n)
	}
}

func TestHandleOrderSukses(t *testing.T) {
	resp, err := HandleOrder(context.Background(), OrderRequest{Item: "keyboard", Qty: 2})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if resp.Total != 500000 {
		t.Errorf("total = %d; want 500000", resp.Total)
	}
}

func TestHandleOrderValidasi(t *testing.T) {
	if _, err := HandleOrder(context.Background(), OrderRequest{Item: "keyboard", Qty: 0}); err == nil {
		t.Error("qty 0 harusnya error")
	}
	if _, err := HandleOrder(context.Background(), OrderRequest{Item: "unknown", Qty: 1}); err == nil {
		t.Error("item tak dikenal harusnya error")
	}
}
