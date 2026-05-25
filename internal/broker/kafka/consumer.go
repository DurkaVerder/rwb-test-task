package kafka

import (
	"encoding/json"
	"log"

	"github.com/IBM/sarama"
)

type Service interface {
	AddRequest(request string) error
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
			c.logger.Printf("Error unmarshaling message: %v", err)
			continue
		}

		if err := c.service.AddRequest(message.Request); err != nil {
			c.logger.Printf("Error adding request to service: %v", err)
			continue
		}

		sess.MarkMessage(msg, "")
	}
	return nil
}
