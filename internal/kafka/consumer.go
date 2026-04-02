package kafka

import (
	"context"
	"log/slog"

	"github.com/segmentio/kafka-go"
)

type MessageHandler func(ctx context.Context, msg kafka.Message) error

type Consumer struct {
	reader *kafka.Reader
}

func NewConsumer(brokers []string, groupID string, topics []string) *Consumer {
	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:        brokers,
			GroupID:        groupID,
			GroupTopics:    topics,
			MinBytes:       1,
			MaxBytes:       10e6,
			CommitInterval: 0,
			StartOffset:    kafka.FirstOffset,
		}),
	}
}

func (c *Consumer) Consume(ctx context.Context, handler MessageHandler) error {
	for {
		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			slog.Error("fetch message failed", "err", err)
			continue
		}

		if err := handler(ctx, msg); err != nil {
			slog.Error("handler failed", "topic", msg.Topic, "offset", msg.Offset, "err", err)
			continue
		}

		if err := c.reader.CommitMessages(ctx, msg); err != nil {
			slog.Error("commit offset failed", "err", err)
		}
	}
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
