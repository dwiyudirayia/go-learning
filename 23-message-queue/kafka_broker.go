package main

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
)

// KafkaBroker: kafka-go menangani reconnect & retry SECARA INTERNAL di Writer
// (produsen) dan Reader (konsumen). Kita tambahkan supervisor sebagai lapis
// pertahanan tambahan.
//
// Model: topic = topic Kafka. group = consumer group (offset di-commit otomatis).
// group == "" -> dibuat group unik per konsumen (meniru broadcast; di Kafka
// broadcast = tiap konsumen punya group sendiri).
type KafkaBroker struct {
	brokers []string
	writer  *kafka.Writer
	logger  *slog.Logger
}

func NewKafkaBroker(brokersCSV string, logger *slog.Logger) *KafkaBroker {
	brokers := strings.Split(brokersCSV, ",")
	return &KafkaBroker{
		brokers: brokers,
		logger:  logger,
		writer: &kafka.Writer{
			Addr:                   kafka.TCP(brokers...),
			Balancer:               &kafka.LeastBytes{},
			RequiredAcks:           kafka.RequireAll, // tunggu semua replika -> durabilitas
			AllowAutoTopicCreation: true,
			// kafka-go otomatis retry & reconnect di dalam WriteMessages.
		},
	}
}

func (b *KafkaBroker) Publish(ctx context.Context, topic string, data []byte) error {
	return b.writer.WriteMessages(ctx, kafka.Message{Topic: topic, Value: data})
}

func (b *KafkaBroker) Subscribe(ctx context.Context, topic, group string, handler Handler) error {
	groupID := group
	if groupID == "" {
		groupID = fmt.Sprintf("bcast-%d", time.Now().UnixNano()) // broadcast -> group unik
	}

	go superviseConsumer(ctx, "kafka:"+topic, DefaultRetry(), b.logger, func(ctx context.Context) error {
		r := kafka.NewReader(kafka.ReaderConfig{
			Brokers:     b.brokers,
			Topic:       topic,
			GroupID:     groupID, // consumer group -> load balancing + commit offset
			MinBytes:    1,
			MaxBytes:    10e6,
			MaxWait:     500 * time.Millisecond,
			StartOffset: kafka.FirstOffset,
		})
		defer r.Close()

		for {
			// ReadMessage: auto-reconnect di dalam; commit offset otomatis (group).
			m, err := r.ReadMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return nil // shutdown normal
				}
				return err // supervisor akan menyambung ulang
			}
			_ = handler(ctx, m.Value)
		}
	})
	return nil
}

func (b *KafkaBroker) Close() error { return b.writer.Close() }
