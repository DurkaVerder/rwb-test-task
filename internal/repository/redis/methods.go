package redis

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode"

	customErr "github.com/DurkaVerder/rwb-test-task/internal/errors"

	redis "github.com/go-redis/redis/v8"
)

const (
	countsKey              = "requests:counts"
	bucketsKey             = "requests:buckets"
	bucketPrefix           = "requests:bucket:"
	stopListKey            = "stopList"
	windowSeconds          = int64(300)
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

	stopWords, err := r.client.SMembers(ctx, stopListKey).Result()
	if err != nil {
		return nil, err
	}

	if len(stopWords) == 0 {
		requests, err := r.client.ZRevRange(ctx, countsKey, 0, int64(n-1)).Result()
		if err != nil {
			return nil, err
		}
		return requests, nil
	}

	stopSet := make(map[string]struct{}, len(stopWords))
	for _, word := range stopWords {
		if word != "" {
			stopSet[strings.ToLower(strings.TrimSpace(word))] = struct{}{}
		}
	}

	results := make([]string, 0, n)
	start := int64(0)
	page := int64(n * 2)
	if page < 50 {
		page = 50
	}

	for len(results) < n {
		batch, err := r.client.ZRevRange(ctx, countsKey, start, start+page-1).Result()
		if err != nil {
			return nil, err
		}
		if len(batch) == 0 {
			break
		}
		for _, query := range batch {
			if len(results) >= n {
				break
			}
			if containsStopWordTokens(query, stopSet) {
				continue
			}
			results = append(results, query)
		}
		if len(batch) < int(page) {
			break
		}
		start += page
	}
	return results, nil
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

	tokens := uniqueTokens(tokenize(query))
	if len(tokens) == 0 {
		return nil
	}
	if blocked, err := r.containsStopWordTokens(ctx, tokens); err != nil {
		return err
	} else if blocked {
		return nil
	}

	bucketUnix := at.Unix()
	bucketKey := fmt.Sprintf("%s%d", bucketPrefix, bucketUnix)
	pipe := r.client.Pipeline()
	pipe.HIncrBy(ctx, bucketKey, query, 1)
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

func containsStopWordTokens(query string, stopSet map[string]struct{}) bool {
	if len(stopSet) == 0 {
		return false
	}
	for _, token := range tokenize(query) {
		if _, blocked := stopSet[token]; blocked {
			return true
		}
	}
	return false
}

func uniqueTokens(tokens []string) []string {
	unique := make([]string, 0, len(tokens))
	seen := make(map[string]struct{}, len(tokens))
	for _, token := range tokens {
		token = strings.TrimSpace(strings.ToLower(token))
		if token == "" {
			continue
		}
		if _, exists := seen[token]; exists {
			continue
		}
		seen[token] = struct{}{}
		unique = append(unique, token)
	}
	return unique
}

func tokenize(text string) []string {
	return strings.FieldsFunc(strings.ToLower(text), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})
}

func (r *RedisRepository) containsStopWordTokens(ctx context.Context, tokens []string) (bool, error) {
	if len(tokens) == 0 {
		return false, nil
	}

	members := make([]interface{}, len(tokens))
	for i, token := range tokens {
		members[i] = token
	}

	present, err := r.client.SMIsMember(ctx, stopListKey, members...).Result()
	if err != nil {
		return false, err
	}
	for _, found := range present {
		if found {
			return true, nil
		}
	}
	return false, nil
}

func (r *RedisRepository) AddStopWord(ctx context.Context, word string) error {
	ctx = ensureContext(ctx)
	word = strings.TrimSpace(word)
	if word == "" {
		return customErr.InvalidWordError
	}
	word = strings.ToLower(word)

	if _, err := r.client.SAdd(ctx, stopListKey, word).Result(); err != nil {
		return err
	}
	return nil
}

func (r *RedisRepository) RemoveStopWord(ctx context.Context, word string) error {
	ctx = ensureContext(ctx)
	word = strings.TrimSpace(word)
	if word == "" {
		return customErr.InvalidWordError
	}
	word = strings.ToLower(word)

	removed, err := r.client.SRem(ctx, stopListKey, word).Result()
	if err != nil {
		return err
	}
	if removed == 0 {
		return customErr.WordNotFoundError
	}
	return nil
}

func (r *RedisRepository) GetAllStopWords(ctx context.Context) ([]string, error) {
	ctx = ensureContext(ctx)
	var stopWords []string
	stopWords, err := r.client.SMembers(ctx, stopListKey).Result()
	if err != nil {
		return nil, err
	}
	return stopWords, nil
}

func (r *RedisRepository) GetStopWord(ctx context.Context, word string) (string, error) {
	ctx = ensureContext(ctx)
	word = strings.TrimSpace(word)
	if word == "" {
		return "", customErr.InvalidWordError
	}
	word = strings.ToLower(word)

	isMember, err := r.client.SIsMember(ctx, stopListKey, word).Result()
	if err != nil {
		return "", err
	}
	if isMember {
		return word, nil
	}
	return "", customErr.WordNotFoundError
}
