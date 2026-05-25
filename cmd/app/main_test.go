package main

import (
	"reflect"
	"testing"
)

func TestSplitCSV(t *testing.T) {
	result := splitCSV(" a, b ,, c ")
	expected := []string{"a", "b", "c"}
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("expected %v, got %v", expected, result)
	}
}

func TestLoadConfigDefaults(t *testing.T) {
	t.Setenv("KAFKA_BROKER", "k1:9092, k2:9092")
	t.Setenv("KAFKA_TOPICS", "search-requests")
	t.Setenv("KAFKA_GROUP_ID", "top-requests")
	t.Setenv("REDIS_ADDR", "redis:6379")
	t.Setenv("REDIS_PASSWORD", "secret")
	t.Setenv("HTTP_ADDR", "")

	cfg, err := loadConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.HTTPAddr != ":8080" {
		t.Fatalf("expected default HTTP_ADDR, got %s", cfg.HTTPAddr)
	}

	if cfg.RedisPassword != "secret" {
		t.Fatalf("expected Redis password to be set")
	}
}

func TestLoadConfigMissingBroker(t *testing.T) {
	t.Setenv("KAFKA_BROKER", "")
	t.Setenv("KAFKA_TOPICS", "search-requests")
	t.Setenv("KAFKA_GROUP_ID", "top-requests")
	t.Setenv("REDIS_ADDR", "redis:6379")

	_, err := loadConfig()
	if err == nil {
		t.Fatalf("expected error when broker is missing")
	}
}
