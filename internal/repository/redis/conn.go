package redis

import (
	"sync"

	redis "github.com/go-redis/redis/v8"
)

type Redis struct {
	client      *redis.Client
	cleanupMu   sync.Mutex
	lastCleanup int64
}

func NewRedis(addr, password string) *Redis {
	return &Redis{
		client: redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: password,
		}),
	}
}
