// REAL-CASE Modul 23 (message queue) — KAFKA SUNGGUHAN (segmentio/kafka-go).
//
// Broker in-memory (advanced/) hilang saat proses mati. Kafka = commit log
// terdistribusi & persisten: pesan tersimpan, bisa di-replay, consumer group
// membagi partisi. Ini producer + consumer nyata terhadap broker Kafka.
//
// Auto-skip bila KAFKA_BROKERS kosong. Jalankan nyata:
//
//	docker compose -f 23-message-queue/real-case/docker-compose.yml up -d
//	KAFKA_BROKERS=127.0.0.1:9092 go run ./23-message-queue/real-case
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
)

func main() {
	brokersEnv := os.Getenv("KAFKA_BROKERS")
	if brokersEnv == "" {
		fmt.Println("⏭️  DILEWATI: set KAFKA_BROKERS untuk versi nyata.")
		fmt.Println("   docker compose -f 23-message-queue/real-case/docker-compose.yml up -d")
		fmt.Println("   KAFKA_BROKERS=127.0.0.1:9092 go run ./23-message-queue/real-case")
		return
	}
	brokers := strings.Split(brokersEnv, ",")
	const topic = "orders"
	ctx := context.Background()

	// PRODUCER: tulis beberapa pesan. AllowAutoTopicCreation memudahkan demo.
	w := &kafka.Writer{
		Addr:                   kafka.TCP(brokers...),
		Topic:                  topic,
		Balancer:               &kafka.LeastBytes{},
		AllowAutoTopicCreation: true,
	}
	defer w.Close()

	fmt.Println("== produce ke Kafka ==")
	msgs := make([]kafka.Message, 3)
	for i := range msgs {
		msgs[i] = kafka.Message{
			Key:   []byte(fmt.Sprintf("order-%d", i+1)),
			Value: []byte(fmt.Sprintf(`{"id":%d,"status":"created"}`, i+1)),
		}
	}
	writeCtx, cancelW := context.WithTimeout(ctx, 10*time.Second)
	defer cancelW()
	if err := w.WriteMessages(writeCtx, msgs...); err != nil {
		panic("gagal produce: " + err.Error())
	}
	fmt.Printf("  %d pesan terkirim ke topik %q\n", len(msgs), topic)

	// CONSUMER: consumer group membaca dari awal. GroupID = otomatis commit offset
	// -> restart lanjut dari posisi terakhir (at-least-once).
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     brokers,
		Topic:       topic,
		GroupID:     "demo-consumer",
		StartOffset: kafka.FirstOffset,
	})
	defer r.Close()

	fmt.Println("== consume dari Kafka ==")
	for i := 0; i < len(msgs); i++ {
		readCtx, cancelR := context.WithTimeout(ctx, 10*time.Second)
		m, err := r.ReadMessage(readCtx)
		cancelR()
		if errors.Is(err, context.DeadlineExceeded) {
			fmt.Println("  timeout menunggu pesan")
			break
		}
		if err != nil {
			panic(err)
		}
		fmt.Printf("  partisi %d offset %d: key=%s value=%s\n", m.Partition, m.Offset, m.Key, m.Value)
	}
	fmt.Println("  catatan: Kafka menjamin urutan PER-PARTISI; pilih key partisi per-aggregate.")
}
