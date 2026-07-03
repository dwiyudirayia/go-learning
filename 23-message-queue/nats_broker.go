package main

import (
	"context"
	"log/slog"
	"time"

	natsserver "github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
)

// StartEmbeddedNATS menjalankan server NATS di dalam proses (untuk demo/test).
func StartEmbeddedNATS() (*natsserver.Server, error) {
	ns, err := natsserver.NewServer(&natsserver.Options{Port: -1, NoLog: true, NoSigs: true})
	if err != nil {
		return nil, err
	}
	go ns.Start()
	if !ns.ReadyForConnections(5 * time.Second) {
		return nil, ErrDisconnected
	}
	return ns, nil
}

// NATSBroker: resiliensi ditangani BUILT-IN oleh client NATS (auto-reconnect).
type NATSBroker struct {
	nc     *nats.Conn
	logger *slog.Logger
}

func NewNATSBroker(url string, logger *slog.Logger) (*NATSBroker, error) {
	nc, err := nats.Connect(url,
		nats.MaxReconnects(-1),             // reconnect tak terbatas
		nats.ReconnectWait(1*time.Second),  // jeda antar percobaan
		nats.ReconnectBufSize(8*1024*1024), // buffer pesan selama terputus
		nats.DisconnectErrHandler(func(_ *nats.Conn, err error) {
			if logger != nil {
				logger.Warn("NATS terputus", slog.Any("err", err))
			}
		}),
		nats.ReconnectHandler(func(c *nats.Conn) {
			if logger != nil {
				logger.Info("NATS tersambung ulang", slog.String("url", c.ConnectedUrl()))
			}
		}),
	)
	if err != nil {
		return nil, err
	}
	return &NATSBroker{nc: nc, logger: logger}, nil
}

func (b *NATSBroker) Publish(ctx context.Context, topic string, data []byte) error {
	return b.nc.Publish(topic, data)
}

func (b *NATSBroker) Subscribe(ctx context.Context, topic, group string, handler Handler) error {
	cb := func(msg *nats.Msg) { _ = handler(ctx, msg.Data) }

	var (
		sub *nats.Subscription
		err error
	)
	if group != "" {
		sub, err = b.nc.QueueSubscribe(topic, group, cb) // load balancing
	} else {
		sub, err = b.nc.Subscribe(topic, cb) // broadcast
	}
	if err != nil {
		return err
	}
	go func() { <-ctx.Done(); _ = sub.Unsubscribe() }()
	return nil
}

func (b *NATSBroker) Close() error {
	b.nc.Close()
	return nil
}
