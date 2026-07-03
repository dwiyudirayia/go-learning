// Modul 33 — Distributed Tracing dengan OpenTelemetry.
//
// TRACE = perjalanan SATU request melintasi banyak fungsi/service. Tiap langkah
// = SPAN (punya nama, durasi, atribut, dan parent). Span dihubungkan lewat
// context.Context -> membentuk pohon (trace) yang bisa divisualisasikan di
// Jaeger/Tempo untuk menemukan langkah yang lambat.
package main

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

// newTracerProvider membuat TracerProvider dengan span processor yang diberikan
// (mis. exporter ke stdout/Jaeger, atau recorder untuk test).
func newTracerProvider(serviceName string, processor sdktrace.SpanProcessor) *sdktrace.TracerProvider {
	res, _ := resource.New(context.Background(),
		resource.WithAttributes(attribute.String("service.name", serviceName)),
	)
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSpanProcessor(processor),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)
	// Propagator: cara meneruskan konteks trace ANTAR SERVICE lewat header HTTP
	// (traceparent). Sehingga span di service B menjadi anak span di service A.
	otel.SetTextMapPropagator(propagation.TraceContext{})
	return tp
}

// tracer mengambil tracer global.
func tracer() trace.Tracer { return otel.Tracer("go-learning/33") }
