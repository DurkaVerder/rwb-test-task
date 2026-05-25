package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/DurkaVerder/rwb-test-task/internal/broker/kafka"
	"github.com/DurkaVerder/rwb-test-task/internal/repository/redis"
	"github.com/DurkaVerder/rwb-test-task/internal/service"
	v1 "github.com/DurkaVerder/rwb-test-task/internal/transport/http/v1"
	"github.com/IBM/sarama"
	"github.com/gin-gonic/gin"
)

func main() {
	logger := log.New(os.Stdout, "main: ", log.LstdFlags|log.Lshortfile)

	redisRepository := redis.NewRedis("localhost:6379", "")

	service := service.NewService(redisRepository)

	consumer := kafka.NewConsumer(log.Default(), service)

	handlers := v1.NewHandlers(service)

	cfg := sarama.NewConfig()
	cfg.Consumer.Return.Errors = true
	cfg.Consumer.Offsets.Initial = sarama.OffsetOldest

	brokers := []string{os.Getenv("KAFKA_BROKER")}
	topics := []string{os.Getenv("KAFKA_TOPICS")}

	consumerGroup, err := sarama.NewConsumerGroup(brokers, os.Getenv("KAFKA_GROUP_ID"), cfg)
	if err != nil {
		panic(err)
	}
	defer consumerGroup.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sigchan := make(chan os.Signal, 1)
		signal.Notify(sigchan, os.Interrupt)
		<-sigchan
		logger.Println("Received interrupt signal, shutting down...")
		cancel()
	}()

	go func() {
		for err := range consumerGroup.Errors() {
			logger.Println("Error from consumer group:", err)
		}
	}()

	go func() {
		r := gin.Default()
		v1 := r.Group("/api/v1")
		v1.GET("/top-requests", handlers.GetTopNRequests)

		if err := r.Run(":8080"); err != nil {
			logger.Println("Error starting HTTP server:", err)
		}
	}()

	for {
		if err := consumerGroup.Consume(ctx, topics, consumer); err != nil {
			logger.Println("Error consuming messages:", err)
			break
		}

		if ctx.Err() != nil {
			logger.Println("Context error:", ctx.Err())
			break
		}
	}

}
