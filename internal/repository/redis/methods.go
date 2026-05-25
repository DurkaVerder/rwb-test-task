package redis

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	redis "github.com/go-redis/redis/v8"
)

const (
	countsKey              = "requests:counts"
	bucketsKey             = "requests:buckets"
	bucketPrefix           = "requests:bucket:"
	windowSeconds          = int64(300)
	bucketTTLSeconds       = 360
	cleanupIntervalSeconds = int64(1)
)

var cleanupScript = redis.NewScript(`
local countsKey = KEYS[1]
local bucketsKey = KEYS[2]
local bucketKey = KEYS[3]

local entries = redis.call('HGETALL', bucketKey)
for i = 1, #entries, 2 do
	local member = entries[i]
	local count = tonumber(entries[i + 1])
	local newScore = redis.call('ZINCRBY', countsKey, -count, member)
	if tonumber(newScore) <= 0 then
		redis.call('ZREM', countsKey, member)
	end
end

redis.call('DEL', bucketKey)
redis.call('ZREM', bucketsKey, bucketKey)
return 1
`)

func (r *RedisRepository) GetTopNQueries(ctx context.Context, n int) ([]string, error) {
	ctx = ensureContext(ctx)
	if n <= 0 {
		return nil, errors.New("n must be positive")
	}

	now := time.Now().Unix()
	if err := r.cleanupExpired(ctx, now); err != nil {
		return nil, err
	}

	requests, err := r.client.ZRevRange(ctx, countsKey, 0, int64(n-1)).Result()
	if err != nil {
		return nil, err
	}
	return requests, nil
}

func (r *RedisRepository) AddQuery(ctx context.Context, query string, at time.Time) error {
	ctx = ensureContext(ctx)
	if query == "" {
		return errors.New("query is empty")
	}

	now := time.Now()
	if at.IsZero() {
		at = now
	}

	nowUnix := now.Unix()
	if at.Unix() < nowUnix-windowSeconds {
		return nil
	}

	if err := r.cleanupExpired(ctx, nowUnix); err != nil {
		return err
	}

	bucketUnix := at.Unix()
	expirationAt := time.Unix(bucketUnix+bucketTTLSeconds, 0)
	ttl := time.Until(expirationAt)
	if ttl <= 0 {
		return nil
	}

	bucketKey := fmt.Sprintf("%s%d", bucketPrefix, bucketUnix)
	pipe := r.client.Pipeline()
	pipe.HIncrBy(ctx, bucketKey, query, 1)
	pipe.Expire(ctx, bucketKey, ttl)
	pipe.ZIncrBy(ctx, countsKey, 1, query)
	pipe.ZAdd(ctx, bucketsKey, &redis.Z{Score: float64(bucketUnix), Member: bucketKey})
	_, err := pipe.Exec(ctx)
	return err
}

func (r *RedisRepository) cleanupExpired(ctx context.Context, now int64) error {
	r.cleanupMu.Lock()
	defer r.cleanupMu.Unlock()

	if now-r.lastCleanup < cleanupIntervalSeconds {
		return nil
	}

	cutoff := now - windowSeconds
	buckets, err := r.client.ZRangeByScore(ctx, bucketsKey, &redis.ZRangeBy{
		Min: "-inf",
		Max: strconv.FormatInt(cutoff, 10),
	}).Result()
	if err != nil {
		return err
	}

	for _, bucketKey := range buckets {
		if _, err := cleanupScript.Run(ctx, r.client, []string{countsKey, bucketsKey, bucketKey}).Result(); err != nil {
			return err
		}
	}

	r.lastCleanup = now
	return nil
}

func ensureContext(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx
}
