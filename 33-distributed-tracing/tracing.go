// Modul 33 — Distributed Tracing dengan OpenTelemetry.
//
// 🔍 Analogi besar: distributed tracing itu seperti PELACAKAN PAKET (resi). Satu request
// (paket) melewati banyak pos: API gateway -> service order -> service inventory -> DB. TRACE =
// riwayat lengkap perjalanan paket; tiap SPAN = satu stempel pos ("tiba di gudang 10:01, keluar
// 10:03"). Karena tiap span mencatat durasi, kamu bisa langsung lihat "oh, macetnya di gudang
// inventory 2 detik". Tanpa tracing, request lambat di sistem microservice = misteri tak terpecahkan.

// 🔍 Analogi propagation: "traceparent" itu seperti NOMOR RESI yang ditempel & ikut terus di tiap
// pos. Berkat itu, stempel di service B tahu ia bagian dari perjalanan yang sama dengan service A —
// sehingga semua span tersambung jadi SATU pohon utuh, bukan potongan terpisah tak berhubungan.

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
