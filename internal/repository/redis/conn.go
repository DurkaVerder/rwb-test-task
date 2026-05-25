package redis

import redis "github.com/go-redis/redis/v8"

type Redis struct {
	client *redis.Client
}

func NewRedis(addr, password string) *Redis {
	return &Redis{
		client: redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: password,
		}),
	}
}
