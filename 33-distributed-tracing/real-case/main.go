// REAL-CASE Modul 33 (tracing) — EXPORTER OpenTelemetry NYATA (stdout).
//
// Versi advanced/ memakai tracetest (in-memory, untuk assertion di test). Versi
// ini memakai exporter SUNGGUHAN: span diserialisasi & "dikirim" (di sini ke
// stdout sebagai JSON, jadi jalan lokal tanpa backend). Untuk produksi, tukar
// exporter ke OTLP (otlptracehttp) yang mengirim ke Jaeger/Tempo/Grafana —
// KODE tracing-nya sama persis, hanya exporter yang berganti.
//
// Jalankan:
//
//	go run ./33-distributed-tracing/real-case
//
// Produksi (OTLP -> Jaeger):
//
//	docker run -d -p 4318:4318 -p 16686:16686 jaegertracing/all-in-one
//	# ganti exporter -> otlptracehttp.New(ctx), set OTEL_EXPORTER_OTLP_ENDPOINT
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func main() {
	ctx := context.Background()

	// EXPORTER nyata: menulis span sebagai JSON ke stdout.
	exp, err := stdouttrace.New(stdouttrace.WithWriter(os.Stdout))
	if err != nil {
		panic(err)
	}

	// Resource = identitas service yang menempel di setiap span (service.name dst).
	res := resource.NewSchemaless(attribute.String("service.name", "demo-service"))

	// TracerProvider dengan BATCHER (buffer & kirim batch) + resource.
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(res),
	)
	defer func() { _ = tp.Shutdown(ctx) }()
	tracer := tp.Tracer("demo")

	// Span parent + child (relasi dibawa lewat context; lintas proses via traceparent).
	ctx, root := tracer.Start(ctx, "HandleRequest")
	root.SetAttributes(attribute.String("http.method", "GET"), attribute.String("http.route", "/checkout"))

	_, db := tracer.Start(ctx, "db.query")
	db.SetAttributes(attribute.String("db.statement", "SELECT * FROM cart"))
	time.Sleep(2 * time.Millisecond)
	db.End()

	_, ext := tracer.Start(ctx, "call.inventory")
	ext.SetAttributes(attribute.String("peer.service", "inventory"))
	time.Sleep(3 * time.Millisecond)
	ext.End()

	root.End()

	fmt.Fprintln(os.Stderr, "== span diekspor ke stdout (JSON) di bawah ==")
	// Paksa exporter mengirim buffer sebelum program keluar.
	if err := tp.ForceFlush(ctx); err != nil {
		panic(err)
	}
}
