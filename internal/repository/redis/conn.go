package redis

import (
	"sync"

	redis "github.com/go-redis/redis/v8"
)

type RedisRepository struct {
	client      *redis.Client
	cleanupMu   sync.Mutex
	lastCleanup int64
}

func NewRedisRepository(addr, password string) *RedisRepository {
	return &RedisRepository{
		client: redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: password,
		}),
	}
}
