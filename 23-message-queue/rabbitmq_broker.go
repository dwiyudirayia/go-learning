package main

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// RabbitMQBroker mendemokan RECONNECT MANUAL — AMQP tidak auto-reconnect, jadi
// kita pantau notifikasi "connection closed" lalu menyambung ulang.
//
// Model topik: satu exchange FANOUT per topic.
//   - group == "" -> queue eksklusif per konsumen (broadcast).
//   - group != "" -> satu queue bersama untuk grup (load balancing / work queue).
type RabbitMQBroker struct {
	url    string
	logger *slog.Logger

	mu   sync.Mutex
	conn *amqp.Connection
	ch   *amqp.Channel
}

func NewRabbitMQBroker(url string, logger *slog.Logger) (*RabbitMQBroker, error) {
	b := &RabbitMQBroker{url: url, logger: logger}
	if err := b.connect(); err != nil {
		return nil, err
	}
	return b, nil
}

func (b *RabbitMQBroker) connect() error {
	conn, err := amqp.Dial(b.url)
	if err != nil {
		return err
	}
	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return err
	}
	b.mu.Lock()
	b.conn, b.ch = conn, ch
	b.mu.Unlock()
	return nil
}

func (b *RabbitMQBroker) channel() *amqp.Channel {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.ch
}

func (b *RabbitMQBroker) Publish(ctx context.Context, topic string, data []byte) error {
	// Retry publish: bila channel putus, sambung ulang lalu kirim.
	cfg := RetryConfig{MaxAttempts: 5, BaseDelay: 100 * time.Millisecond, MaxDelay: 2 * time.Second}
	return Retry(ctx, cfg, func() error {
		ch := b.channel()
		if ch == nil || ch.IsClosed() {
			if err := b.connect(); err != nil {
				return err
			}
			ch = b.channel()
		}
		if err := ch.ExchangeDeclare(topic, "fanout", true, false, false, false, nil); err != nil {
			return err
		}
		return ch.PublishWithContext(ctx, topic, "", false, false, amqp.Publishing{
			ContentType:  "application/octet-stream",
			Body:         data,
			DeliveryMode: amqp.Persistent, // pesan bertahan bila broker restart
		})
	})
}

func (b *RabbitMQBroker) Subscribe(ctx context.Context, topic, group string, handler Handler) error {
	// Supervisor: sambung -> konsumsi -> bila putus, backoff & sambung ulang.
	go superviseConsumer(ctx, "rabbitmq:"+topic, DefaultRetry(), b.logger, func(ctx context.Context) error {
		conn, err := amqp.Dial(b.url)
		if err != nil {
			return err
		}
		defer conn.Close()
		ch, err := conn.Channel()
		if err != nil {
			return err
		}
		defer ch.Close()

		if err := ch.ExchangeDeclare(topic, "fanout", true, false, false, false, nil); err != nil {
			return err
		}

		// Fair dispatch: jangan kirim pesan baru ke konsumen yang masih sibuk.
		_ = ch.Qos(1, 0, false)

		queueName := group
		exclusive := group == "" // broadcast -> queue eksklusif per konsumen
		q, err := ch.QueueDeclare(queueName, !exclusive, exclusive, exclusive, false, nil)
		if err != nil {
			return err
		}
		if err := ch.QueueBind(q.Name, "", topic, false, nil); err != nil {
			return err
		}

		msgs, err := ch.Consume(q.Name, "", false /*autoAck=false -> ack manual*/, exclusive, false, false, nil)
		if err != nil {
			return err
		}

		// Notifikasi bila koneksi ditutup -> keluar -> supervisor menyambung ulang.
		closed := conn.NotifyClose(make(chan *amqp.Error, 1))

		for {
			select {
			case <-ctx.Done():
				return nil
			case err := <-closed:
				if err != nil {
					return err
				}
				return errors.New("koneksi rabbitmq ditutup")
			case d, ok := <-msgs:
				if !ok {
					return errors.New("channel consume ditutup")
				}
				if herr := handler(ctx, d.Body); herr != nil {
					_ = d.Nack(false, true) // requeue -> diproses ulang (at-least-once)
				} else {
					_ = d.Ack(false) // sukses -> hapus dari queue
				}
			}
		}
	})
	return nil
}

func (b *RabbitMQBroker) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.ch != nil {
		_ = b.ch.Close()
	}
	if b.conn != nil {
		return b.conn.Close()
	}
	return nil
}
