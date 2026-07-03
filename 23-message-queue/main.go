// Jalankan: go run ./23-message-queue
//
// Pilih broker via env (default in-memory, tak butuh server):
//
//	BROKER=inmemory go run ./23-message-queue          # default
//	BROKER=nats     go run ./23-message-queue          # NATS embedded
//	BROKER=rabbitmq RABBITMQ_URL=amqp://guest:guest@localhost:5672/ go run ./23-message-queue
//	BROKER=kafka    KAFKA_BROKERS=localhost:9092 go run ./23-message-queue
//
// Verifikasi otomatis: go test ./23-message-queue
package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"sync"
	"time"
)

func main() {
	fmt.Println("=== 23 — Message Queue (multi-broker + resiliensi) ===")
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelWarn}))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	broker, cleanup := selectBroker(os.Getenv("BROKER"), logger)
	defer cleanup()
	defer broker.Close()

	// Konsumen (queue group "workers" -> load balancing).
	var wg sync.WaitGroup
	wg.Add(3)
	_ = broker.Subscribe(ctx, "orders.created", "workers", func(_ context.Context, data []byte) error {
		fmt.Printf("  [worker] terima: %s\n", data)
		wg.Done()
		return nil
	})
	time.Sleep(100 * time.Millisecond) // beri waktu subscribe siap

	fmt.Println("publisher mengirim 3 event (dengan retry)...")
	for i := 1; i <= 3; i++ {
		msg := []byte(fmt.Sprintf("order-%d", i))
		if err := PublishWithRetry(ctx, broker, "orders.created", msg, logger); err != nil {
			log.Fatal(err)
		}
	}
	wg.Wait()

	// Demo skema resiliensi (hanya untuk in-memory: bisa simulasi putus koneksi).
	if im, ok := broker.(*InMemoryBroker); ok {
		demoReconnect(ctx, im, logger)
	}
	fmt.Println("selesai.")
}

// demoReconnect mensimulasikan broker MATI saat publish, lalu HIDUP lagi.
// PublishWithRetry menahan pesan (retry+backoff) sampai broker pulih.
func demoReconnect(ctx context.Context, im *InMemoryBroker, logger *slog.Logger) {
	fmt.Println("\n-- Demo resiliensi: broker 'mati' lalu 'hidup' lagi --")

	got := make(chan string, 1)
	_ = im.Subscribe(ctx, "penting", "", func(_ context.Context, data []byte) error {
		got <- string(data)
		return nil
	})
	time.Sleep(50 * time.Millisecond)

	im.SetConnected(false) // koneksi putus
	fmt.Println("  broker MATI, publisher mencoba mengirim (akan retry)...")

	done := make(chan struct{})
	go func() {
		_ = PublishWithRetry(ctx, im, "penting", []byte("pesan-penting"), logger)
		close(done)
	}()

	time.Sleep(300 * time.Millisecond)
	im.SetConnected(true) // koneksi pulih
	fmt.Println("  broker HIDUP lagi -> pesan tertunda akhirnya terkirim")

	<-done
	select {
	case msg := <-got:
		fmt.Printf("  konsumen menerima: %q (tidak hilang!)\n", msg)
	case <-time.After(2 * time.Second):
		fmt.Println("  (timeout)")
	}
}

// selectBroker membuat broker sesuai env. cleanup dipanggil saat selesai.
func selectBroker(name string, logger *slog.Logger) (Broker, func()) {
	switch name {
	case "nats":
		ns, err := StartEmbeddedNATS()
		if err != nil {
			log.Fatal(err)
		}
		b, err := NewNATSBroker(ns.ClientURL(), logger)
		if err != nil {
			log.Fatal(err)
		}
		return b, ns.Shutdown
	case "rabbitmq":
		b, err := NewRabbitMQBroker(getenv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"), logger)
		if err != nil {
			log.Fatal(err)
		}
		return b, func() {}
	case "kafka":
		return NewKafkaBroker(getenv("KAFKA_BROKERS", "localhost:9092"), logger), func() {}
	default:
		fmt.Println("(memakai broker in-memory; set BROKER=nats|rabbitmq|kafka untuk lainnya)")
		return NewInMemoryBroker(), func() {}
	}
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
