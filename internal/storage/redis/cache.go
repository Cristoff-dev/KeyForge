package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const tokenBucketScript = `
local tokens_key = KEYS[1]
local timestamp_key = KEYS[2]
local capacity = tonumber(ARGV[1])
local refill_rate = tonumber(ARGV[2]) 
local now = tonumber(ARGV[3])
local window = tonumber(ARGV[4])

local tokens = tonumber(redis.call("GET", tokens_key))
if tokens == nil then
	tokens = capacity
end

local last_refreshed = tonumber(redis.call("GET", timestamp_key))
if last_refreshed == nil then
	last_refreshed = now
end

local delta = math.max(0, now - last_refreshed)
local filled_tokens = math.min(capacity, tokens + (delta * refill_rate))
local allowed = 0

if filled_tokens >= 1 then
	allowed = 1
	filled_tokens = filled_tokens - 1
	redis.call("SET", tokens_key, filled_tokens, "EX", window)
	redis.call("SET", timestamp_key, now, "EX", window)
end

return allowed
`

var ErrCacheMiss = errors.New("key not found in cache")

type Cache struct {
	client *redis.Client
}

func NewCache(client *redis.Client) *Cache {
	return &Cache{
		client: client,
	}
}

func (c *Cache) SetHashValidation(ctx context.Context, plainKey string, hash string, ttl time.Duration) error {
	key := fmt.Sprintf("auth:hash:%s", plainKey)
	return c.client.Set(ctx, key, hash, ttl).Err()
}

func (c *Cache) GetHashValidation(ctx context.Context, plainKey string) (string, error) {
	key := fmt.Sprintf("auth:hash:%s", plainKey)
	val, err := c.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return "", ErrCacheMiss
	}
	return val, err
}

func (c *Cache) InvalidateHash(ctx context.Context, plainKey string) error {
	key := fmt.Sprintf("auth:hash:%s", plainKey)
	return c.client.Del(ctx, key).Err()
}

func (c *Cache) IncrementRateLimit(ctx context.Context, clientID string, limit int, window time.Duration) (bool, error) {
	tokensKey := fmt.Sprintf("rate:%s:tokens", clientID)
	timestampKey := fmt.Sprintf("rate:%s:ts", clientID)

	windowSeconds := int64(window.Seconds())
	if windowSeconds == 0 {
		windowSeconds = 1
	}

	refillRate := float64(limit) / float64(windowSeconds)
	now := time.Now().Unix()

	result, err := c.client.Eval(ctx, tokenBucketScript, []string{tokensKey, timestampKey}, limit, refillRate, now, windowSeconds).Result()
	if err != nil {
		return false, err
	}

	allowed, ok := result.(int64)
	if !ok {
		return false, errors.New("invalid script response type")
	}

	return allowed == 1, nil
}

func (c *Cache) Get(ctx context.Context, key string) (string, error) {
	return c.client.Get(ctx, key).Result()
}

func (c *Cache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return c.client.Set(ctx, key, value, expiration).Err()
}

func (c *Cache) AllowRequest(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	count, err := c.client.Incr(ctx, key).Result()
	if err != nil {
		return false, err
	}

	if count == 1 {
		c.client.Expire(ctx, key, window)
	}

	return count <= int64(limit), nil
}
