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

func (r *Redis) GetTopNRequests(n int) ([]string, error) {
	if n <= 0 {
		return nil, errors.New("n must be positive")
	}

	now := time.Now().Unix()
	if err := r.cleanupExpired(now); err != nil {
		return nil, err
	}

	ctx := context.Background()
	requests, err := r.client.ZRevRange(ctx, countsKey, 0, int64(n-1)).Result()
	if err != nil {
		return nil, err
	}
	return requests, nil
}

func (r *Redis) AddRequest(request string) error {
	if request == "" {
		return errors.New("request is empty")
	}

	now := time.Now().Unix()
	if err := r.cleanupExpired(now); err != nil {
		return err
	}

	bucketKey := fmt.Sprintf("%s%d", bucketPrefix, now)
	ctx := context.Background()
	pipe := r.client.Pipeline()
	pipe.HIncrBy(ctx, bucketKey, request, 1)
	pipe.Expire(ctx, bucketKey, time.Second*time.Duration(bucketTTLSeconds))
	pipe.ZIncrBy(ctx, countsKey, 1, request)
	pipe.ZAdd(ctx, bucketsKey, &redis.Z{Score: float64(now), Member: bucketKey})
	_, err := pipe.Exec(ctx)
	return err
}

func (r *Redis) cleanupExpired(now int64) error {
	r.cleanupMu.Lock()
	defer r.cleanupMu.Unlock()

	if now-r.lastCleanup < cleanupIntervalSeconds {
		return nil
	}

	cutoff := now - windowSeconds
	ctx := context.Background()
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
