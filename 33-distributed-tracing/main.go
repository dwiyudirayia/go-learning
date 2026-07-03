// Jalankan: go run ./33-distributed-tracing
// Verifikasi otomatis: go test ./33-distributed-tracing
//
// Demo memakai in-memory recorder agar mudah dilihat. Di produksi, ganti
// processor dengan exporter ke Jaeger/Tempo (lihat README).
package main

import (
	"context"
	"fmt"
	"sort"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

func main() {
	fmt.Println("=== 33 — Distributed Tracing (OpenTelemetry) ===")

	// Recorder menangkap span di memori (pengganti exporter untuk demo).
	sr := tracetest.NewSpanRecorder()
	tp := newTracerProvider("order-service", sr)
	defer tp.Shutdown(context.Background())

	// Jalankan operasi ber-trace.
	_ = HandleOrder(context.Background(), 42)

	printTrace(sr.Ended())
}

// printTrace mencetak span sebagai pohon (parent -> child) dengan durasi.
func printTrace(spans []sdktrace.ReadOnlySpan) {
	// Index anak berdasar SpanID induk.
	children := map[trace.SpanID][]sdktrace.ReadOnlySpan{}
	ids := map[trace.SpanID]bool{}
	for _, s := range spans {
		ids[s.SpanContext().SpanID()] = true
	}
	var roots []sdktrace.ReadOnlySpan
	for _, s := range spans {
		p := s.Parent().SpanID()
		if p.IsValid() && ids[p] {
			children[p] = append(children[p], s)
		} else {
			roots = append(roots, s)
		}
	}

	fmt.Println("\ntrace untuk order 42:")
	var walk func(s sdktrace.ReadOnlySpan, depth int)
	walk = func(s sdktrace.ReadOnlySpan, depth int) {
		dur := s.EndTime().Sub(s.StartTime())
		fmt.Printf("%*s%s (%s)\n", depth*2, "", s.Name(), dur.Round(1e6))
		kids := children[s.SpanContext().SpanID()]
		sort.Slice(kids, func(i, j int) bool { return kids[i].StartTime().Before(kids[j].StartTime()) })
		for _, c := range kids {
			walk(c, depth+1)
		}
	}
	for _, r := range roots {
		walk(r, 1)
	}
	fmt.Printf("\ntotal span: %d\n", len(spans))
}
