package main

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// HandleOrder mensimulasikan operasi bisnis yang terdiri dari beberapa langkah.
// Tiap langkah membuat SPAN anak, sehingga trace membentuk pohon:
//
//	HandleOrder
//	├── validateOrder
//	├── chargePayment   (span anak dari HandleOrder)
//	└── saveToDatabase
func HandleOrder(ctx context.Context, orderID int) error {
	// Span induk untuk seluruh operasi.
	ctx, span := tracer().Start(ctx, "HandleOrder")
	defer span.End()
	// Atribut = metadata untuk memfilter/mencari trace nanti.
	span.SetAttributes(attribute.Int("order.id", orderID))

	if err := validateOrder(ctx, orderID); err != nil {
		span.SetStatus(codes.Error, err.Error()) // tandai trace gagal
		return err
	}
	if err := chargePayment(ctx, orderID); err != nil {
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	saveToDatabase(ctx, orderID)
	span.SetStatus(codes.Ok, "")
	return nil
}

func validateOrder(ctx context.Context, orderID int) error {
	_, span := tracer().Start(ctx, "validateOrder")
	defer span.End()
	time.Sleep(2 * time.Millisecond)
	return nil
}

func chargePayment(ctx context.Context, orderID int) error {
	// Span ini punya parent = span HandleOrder (lewat ctx).
	_, span := tracer().Start(ctx, "chargePayment")
	defer span.End()
	span.SetAttributes(attribute.String("payment.provider", "stripe"))
	time.Sleep(5 * time.Millisecond) // langkah paling lambat -> terlihat di trace
	return nil
}

func saveToDatabase(ctx context.Context, orderID int) {
	_, span := tracer().Start(ctx, "saveToDatabase")
	defer span.End()
	span.SetAttributes(attribute.String("db.system", "postgresql"))
	time.Sleep(1 * time.Millisecond)
}
