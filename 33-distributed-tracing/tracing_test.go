package main

import (
	"context"
	"testing"

	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestTraceHierarchy(t *testing.T) {
	sr := tracetest.NewSpanRecorder()
	tp := newTracerProvider("test", sr)
	defer tp.Shutdown(context.Background())

	if err := HandleOrder(context.Background(), 1); err != nil {
		t.Fatalf("HandleOrder: %v", err)
	}

	spans := sr.Ended()
	// Harus ada 4 span: HandleOrder + 3 langkah.
	if len(spans) != 4 {
		t.Fatalf("jumlah span = %d; want 4", len(spans))
	}

	// Cari span induk & anak-anaknya.
	byName := map[string]int{}
	var parentID, chargeParent string
	for _, s := range spans {
		byName[s.Name()]++
		if s.Name() == "HandleOrder" {
			parentID = s.SpanContext().SpanID().String()
		}
		if s.Name() == "chargePayment" {
			chargeParent = s.Parent().SpanID().String()
		}
	}

	for _, name := range []string{"HandleOrder", "validateOrder", "chargePayment", "saveToDatabase"} {
		if byName[name] != 1 {
			t.Errorf("span %q muncul %d kali; want 1", name, byName[name])
		}
	}

	// chargePayment harus ANAK dari HandleOrder (context propagation).
	if chargeParent != parentID {
		t.Errorf("parent chargePayment = %s; want %s (HandleOrder)", chargeParent, parentID)
	}

	// Semua span dalam SATU trace (trace ID sama).
	traceID := spans[0].SpanContext().TraceID()
	for _, s := range spans {
		if s.SpanContext().TraceID() != traceID {
			t.Errorf("span %q trace ID beda -> bukan satu trace", s.Name())
		}
	}
}

func TestTraceAtribut(t *testing.T) {
	sr := tracetest.NewSpanRecorder()
	tp := newTracerProvider("test", sr)
	defer tp.Shutdown(context.Background())

	_ = HandleOrder(context.Background(), 99)

	for _, s := range sr.Ended() {
		if s.Name() != "HandleOrder" {
			continue
		}
		found := false
		for _, kv := range s.Attributes() {
			if string(kv.Key) == "order.id" && kv.Value.AsInt64() == 99 {
				found = true
			}
		}
		if !found {
			t.Error("span HandleOrder tidak punya atribut order.id=99")
		}
	}
}
