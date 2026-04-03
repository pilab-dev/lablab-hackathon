package messaging

import (
	"context"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/rs/zerolog/log"
)

// NATSClient wraps the NATS connection and JetStream context
type NATSClient struct {
	conn *nats.Conn
	js   jetstream.JetStream
}

// NewNATSClient connects to a NATS server and initializes JetStream
func NewNATSClient(url string) (*NATSClient, error) {
	if url == "" {
		url = nats.DefaultURL // nats://localhost:4222
	}

	conn, err := nats.Connect(url,
		nats.RetryOnFailedConnect(true),
		nats.MaxReconnects(10),
		nats.ReconnectWait(2*time.Second),
		nats.DisconnectErrHandler(func(_ *nats.Conn, err error) {
			log.Error().Err(err).Msg("NATS disconnected")
		}),
		nats.ReconnectHandler(func(_ *nats.Conn) {
			log.Info().Msg("NATS reconnected")
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	js, err := jetstream.New(conn)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to create JetStream context: %w", err)
	}

	log.Info().Str("url", conn.ConnectedUrl()).Msg("Connected to NATS")
	return &NATSClient{conn: conn, js: js}, nil
}

// Publish sends a message to a NATS Core subject (hot path — fire and forget)
func (c *NATSClient) Publish(subject string, data []byte) error {
	return c.conn.Publish(subject, data)
}

// Subscribe registers a callback for messages on a NATS Core subject
func (c *NATSClient) Subscribe(subject string, handler func(msg *nats.Msg)) (*nats.Subscription, error) {
	sub, err := c.conn.Subscribe(subject, handler)
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to %s: %w", subject, err)
	}
	return sub, nil
}

// EnsureStream creates a JetStream stream if it does not already exist
func (c *NATSClient) EnsureStream(ctx context.Context, name string, subjects []string, storage jetstream.StorageType) error {
	stream, err := c.js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:     name,
		Subjects: subjects,
		Storage:  storage,
	})
	if err != nil {
		return fmt.Errorf("failed to create stream %s: %w", name, err)
	}

	info, _ := stream.Info(ctx)
	log.Info().Str("stream", info.Config.Name).Int("messages", int(info.State.Msgs)).Msg("JetStream stream ready")
	return nil
}

// PublishPersisted publishes a message to a JetStream subject (cold path — persisted)
func (c *NATSClient) PublishPersisted(ctx context.Context, subject string, data []byte) error {
	_, err := c.js.Publish(ctx, subject, data)
	if err != nil {
		return fmt.Errorf("failed to publish persisted message to %s: %w", subject, err)
	}
	return nil
}

// ConsumePersisted creates a durable consumer and calls handler for each message
func (c *NATSClient) ConsumePersisted(ctx context.Context, stream, durable string, handler func(msg jetstream.Msg)) error {
	consumer, err := c.js.CreateOrUpdateConsumer(ctx, stream, jetstream.ConsumerConfig{
		Durable:   durable,
		AckPolicy: jetstream.AckExplicitPolicy,
	})
	if err != nil {
		return fmt.Errorf("failed to create consumer %s: %w", durable, err)
	}

	consumeCtx, err := consumer.Consume(func(msg jetstream.Msg) {
		handler(msg)
	})
	if err != nil {
		return fmt.Errorf("failed to start consuming: %w", err)
	}

	go func() {
		<-ctx.Done()
		consumeCtx.Stop()
		log.Info().Str("consumer", durable).Msg("Consumer stopped")
	}()

	log.Info().Str("consumer", durable).Str("stream", stream).Msg("JetStream consumer started")
	return nil
}

// Close gracefully shuts down the NATS connection
func (c *NATSClient) Close() {
	if c.conn != nil {
		c.conn.Close()
		log.Info().Msg("NATS connection closed")
	}
}
