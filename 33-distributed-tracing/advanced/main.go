// Teknik Advanced Modul 33 (distributed tracing) — contoh runnable + penjelasan detail.
//
// OpenTelemetry dengan in-memory SpanRecorder (tracetest) -> membuat span
// parent/child + atribut, lalu MEMERIKSA hasilnya tanpa backend nyata. Jalankan:
//
//	go run ./33-distributed-tracing/advanced
package main

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

func main() {
	// =====================================================================
	// SpanRecorder = exporter in-memory untuk test/verifikasi (tanpa Jaeger).
	// =====================================================================
	sr := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(sr))
	defer tp.Shutdown(context.Background())
	tracer := tp.Tracer("demo-service")

	// =====================================================================
	// 1. Span PARENT/CHILD via propagasi context
	// =====================================================================
	// tracer.Start mengembalikan ctx baru berisi span aktif. Meneruskan ctx itu
	// ke fungsi berikutnya membuat span di dalamnya menjadi ANAK -> membentuk
	// pohon trace. Lintas proses, relasi ini dibawa header W3C `traceparent`.
	ctx, root := tracer.Start(context.Background(), "HandleRequest")
	root.SetAttributes(
		attribute.String("http.method", "GET"),
		attribute.String("http.route", "/checkout"),
	)

	// child 1: query DB
	_, dbSpan := tracer.Start(ctx, "db.query")
	dbSpan.SetAttributes(attribute.String("db.statement", "SELECT * FROM cart"))
	dbSpan.End()

	// child 2: panggil service lain (yang membuat CUCU span)
	callDownstream(ctx, tracer)

	root.End()

	// =====================================================================
	// 2. Memeriksa span yang terekam (nama, relasi parent, atribut)
	// =====================================================================
	fmt.Println("== span terekam ==")
	spans := sr.Ended()
	nama := map[trace.SpanID]string{}
	for _, s := range spans {
		nama[s.SpanContext().SpanID()] = s.Name()
	}
	for _, s := range spans {
		parent := "(root)"
		if s.Parent().IsValid() {
			if n, ok := nama[s.Parent().SpanID()]; ok {
				parent = n
			}
		}
		fmt.Printf("  %-16s parent=%-16s trace=%s…\n",
			s.Name(), parent, s.SpanContext().TraceID().String()[:8])
	}
	fmt.Println("  (semua span berbagi TraceID yang sama = satu perjalanan request)")
}

// callDownstream membuat span untuk panggilan keluar + satu span anak di
// dalamnya, semua mewarisi trace dari ctx (parent).
func callDownstream(ctx context.Context, tracer trace.Tracer) {
	ctx, span := tracer.Start(ctx, "call.inventory")
	defer span.End()
	span.SetAttributes(attribute.String("peer.service", "inventory"))

	_, child := tracer.Start(ctx, "inventory.check")
	child.SetAttributes(attribute.Int("sku.count", 3))
	child.End()
}
