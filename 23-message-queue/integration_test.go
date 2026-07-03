package main

import (
	"context"
	"os"
	"testing"
	"time"
)

// Integration test RabbitMQ. Otomatis DI-SKIP kecuali env RABBITMQ_URL diset.
// Contoh menjalankan (butuh RabbitMQ berjalan):
//
//	docker run -d --rm -p 5672:5672 rabbitmq:3
//	RABBITMQ_URL=amqp://guest:guest@localhost:5672/ go test -run TestRabbitMQ ./23-message-queue
func TestRabbitMQIntegration(t *testing.T) {
	url := os.Getenv("RABBITMQ_URL")
	if url == "" {
		t.Skip("set RABBITMQ_URL untuk menjalankan integration test RabbitMQ")
	}

	b, err := NewRabbitMQBroker(url, nil)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer b.Close()
	roundTrip(t, b)
}

// Integration test Kafka. Otomatis DI-SKIP kecuali env KAFKA_BROKERS diset.
// Contoh:
//
//	docker run -d --rm -p 9092:9092 apache/kafka:latest
//	KAFKA_BROKERS=localhost:9092 go test -run TestKafka ./23-message-queue
func TestKafkaIntegration(t *testing.T) {
	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		t.Skip("set KAFKA_BROKERS untuk menjalankan integration test Kafka")
	}

	b := NewKafkaBroker(brokers, nil)
	defer b.Close()
	roundTrip(t, b)
}

// roundTrip: publish 1 pesan, pastikan konsumen menerimanya (kontrak Broker sama
// untuk semua implementasi).
func roundTrip(t *testing.T, b Broker) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	got := make(chan string, 1)
	if err := b.Subscribe(ctx, "test-topic", "test-group", func(_ context.Context, data []byte) error {
		select {
		case got <- string(data):
		default:
		}
		return nil
	}); err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	time.Sleep(2 * time.Second) // beri waktu konsumen siap (Kafka perlu waktu join group)

	if err := b.Publish(ctx, "test-topic", []byte("halo")); err != nil {
		t.Fatalf("publish: %v", err)
	}

	select {
	case m := <-got:
		if m != "halo" {
			t.Errorf("terima %q; want halo", m)
		}
	case <-ctx.Done():
		t.Fatal("timeout menunggu pesan")
	}
}
