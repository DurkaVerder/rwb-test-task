package kafka

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/DurkaVerder/rwb-test-task/internal/metrics"
	"github.com/IBM/sarama"
)

type Service interface {
	AddQuery(ctx context.Context, query string, at time.Time) error
}

type Message struct {
	Request string `json:"request"`
}

type Consumer struct {
	logger  *log.Logger
	service Service
}

func NewConsumer(logger *log.Logger, service Service) *Consumer {
	return &Consumer{
		logger:  logger,
		service: service,
	}
}

func (c *Consumer) Setup(sess sarama.ConsumerGroupSession) error {
	return nil
}
func (c *Consumer) Cleanup(sess sarama.ConsumerGroupSession) error {
	return nil
}

func (c *Consumer) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		var message Message
		if err := json.Unmarshal(msg.Value, &message); err != nil {
			metrics.IncQueryIngestError()
			c.logger.Printf("Error unmarshaling message: %v", err)
			continue
		}

		eventTime := msg.Timestamp
		if eventTime.IsZero() {
			eventTime = time.Now()
		}

		consumeCtx := sess.Context()
		if consumeCtx == nil {
			consumeCtx = context.Background()
		}

		if err := c.service.AddQuery(consumeCtx, message.Request, eventTime); err != nil {
			metrics.IncQueryIngestError()
			c.logger.Printf("Error adding query to service: %v", err)
			continue
		}
		metrics.IncQueryIngested()

		sess.MarkMessage(msg, "")
	}
	return nil
}
