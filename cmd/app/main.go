package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/DurkaVerder/rwb-test-task/internal/broker/kafka"
	"github.com/DurkaVerder/rwb-test-task/internal/repository/redis"
	service "github.com/DurkaVerder/rwb-test-task/internal/service/search"
	stoplistService "github.com/DurkaVerder/rwb-test-task/internal/service/stoplist"
	v1 "github.com/DurkaVerder/rwb-test-task/internal/transport/http/v1"
	"github.com/IBM/sarama"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Config struct {
	KafkaBrokers  []string
	KafkaTopics   []string
	KafkaGroupID  string
	RedisAddr     string
	RedisPassword string
	HTTPAddr      string
}

func main() {
	logger := log.New(os.Stdout, "main: ", log.LstdFlags|log.Lshortfile)

	cfg, err := loadConfig()
	if err != nil {
		logger.Println("Configuration error:", err)
		os.Exit(1)
	}

	redisRepository := redis.NewRedisRepository(cfg.RedisAddr, cfg.RedisPassword)

	searchService := service.NewSearchService(redisRepository)
	stopListService := stoplistService.NewStopListService(redisRepository)

	consumer := kafka.NewConsumer(logger, searchService)

	handlers := v1.NewHandlers(searchService)
	stopListHandlers := v1.NewStopListHandlers(stopListService)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	saramaCfg := sarama.NewConfig()
	saramaCfg.Consumer.Return.Errors = true
	saramaCfg.Consumer.Offsets.Initial = sarama.OffsetOldest

	consumerGroup, err := createConsumerGroupWithRetry(ctx, cfg.KafkaBrokers, cfg.KafkaGroupID, saramaCfg, logger)
	if err != nil {
		logger.Println("Error creating consumer group:", err)
		os.Exit(1)
	}
	defer consumerGroup.Close()

	go func() {
		for err := range consumerGroup.Errors() {
			logger.Println("Error from consumer group:", err)
		}
	}()

	router := gin.Default()
	v1Group := router.Group("/api/v1")
	v1Group.GET("/top-requests", handlers.GetTopNQueries)
	v1Group.GET("/stoplist", stopListHandlers.GetStopList)
	v1Group.POST("/stoplist", stopListHandlers.AddStopWord)
	v1Group.DELETE("/stoplist", stopListHandlers.RemoveStopWord)
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	server := &http.Server{
		Addr:    cfg.HTTPAddr,
		Handler: router,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Println("Error starting HTTP server:", err)
		}
	}()

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Println("Error shutting down HTTP server:", err)
		}
	}()

	for {
		if err := consumerGroup.Consume(ctx, cfg.KafkaTopics, consumer); err != nil {
			logger.Println("Error consuming messages:", err)
			break
		}

		if ctx.Err() != nil {
			logger.Println("Context error:", ctx.Err())
			break
		}
	}

}

func createConsumerGroupWithRetry(ctx context.Context, brokers []string, groupID string, cfg *sarama.Config, logger *log.Logger) (sarama.ConsumerGroup, error) {
	backoff := 1 * time.Second
	maxBackoff := 30 * time.Second

	for {
		consumerGroup, err := sarama.NewConsumerGroup(brokers, groupID, cfg)
		if err == nil {
			return consumerGroup, nil
		}

		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		logger.Printf("Error creating consumer group: %v. Retrying in %s...", err, backoff)

		select {
		case <-time.After(backoff):
		case <-ctx.Done():
			return nil, ctx.Err()
		}

		if backoff < maxBackoff {
			backoff *= 2
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
		}
	}
}

func loadConfig() (Config, error) {
	brokers := splitCSV(os.Getenv("KAFKA_BROKER"))
	if len(brokers) == 0 {
		return Config{}, fmt.Errorf("KAFKA_BROKER is required")
	}

	topics := splitCSV(os.Getenv("KAFKA_TOPICS"))
	if len(topics) == 0 {
		return Config{}, fmt.Errorf("KAFKA_TOPICS is required")
	}

	groupID := strings.TrimSpace(os.Getenv("KAFKA_GROUP_ID"))
	if groupID == "" {
		return Config{}, fmt.Errorf("KAFKA_GROUP_ID is required")
	}

	redisAddr := strings.TrimSpace(os.Getenv("REDIS_ADDR"))
	if redisAddr == "" {
		return Config{}, fmt.Errorf("REDIS_ADDR is required")
	}

	httpAddr := strings.TrimSpace(os.Getenv("HTTP_ADDR"))
	if httpAddr == "" {
		httpAddr = ":8080"
	}

	return Config{
		KafkaBrokers:  brokers,
		KafkaTopics:   topics,
		KafkaGroupID:  groupID,
		RedisAddr:     redisAddr,
		RedisPassword: os.Getenv("REDIS_PASSWORD"),
		HTTPAddr:      httpAddr,
	}, nil
}

func splitCSV(value string) []string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}

	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}
