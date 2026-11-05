package redis_test

import (
	"context"
	"os"
	"testing"
	"time"

	storageRedis "github.com/Cristoff-dev/keyforge/internal/storage/redis"
	"github.com/redis/go-redis/v9"
)

func TestCache_SetAndGet(t *testing.T) {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "localhost:6379"
	}

	ctx := context.Background()
	rdb := redis.NewClient(&redis.Options{
		Addr: redisURL,
	})

	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Skipf("Skipping Redis integration test (Redis unavailable): %v", err)
	}
	defer rdb.Close()

	cache := storageRedis.NewCache(rdb)

	t.Run("Set and Get Value", func(t *testing.T) {
		key := "test_key_set_get"
		val := "test_value_123"

		err := cache.Set(ctx, key, val, 1*time.Minute)
		if err != nil {
			t.Fatalf("unexpected error setting key: %v", err)
		}

		got, err := cache.Get(ctx, key)
		if err != nil {
			t.Fatalf("unexpected error getting key: %v", err)
		}

		if got != val {
			t.Errorf("expected value %s, got %s", val, got)
		}

		// Clean up
		_ = rdb.Del(ctx, key).Err()
	})

	t.Run("Get Non-Existent Key", func(t *testing.T) {
		_, err := cache.Get(ctx, "non_existent_key_999")
		if err == nil {
			t.Error("expected error for non-existent key, got nil")
		}
	})
}

func TestCache_AllowRequest(t *testing.T) {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "localhost:6379"
	}

	ctx := context.Background()
	rdb := redis.NewClient(&redis.Options{
		Addr: redisURL,
	})

	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Skipf("Skipping Redis integration test (Redis unavailable): %v", err)
	}
	defer rdb.Close()

	cache := storageRedis.NewCache(rdb)

	t.Run("Rate Limit Threshold Enforcement", func(t *testing.T) {
		key := "test_ratelimit_key"
		limit := 3
		window := 10 * time.Second

		_ = rdb.Del(ctx, key).Err()

		for i := 1; i <= limit; i++ {
			allowed, err := cache.AllowRequest(ctx, key, limit, window)
			if err != nil {
				t.Fatalf("unexpected error on request %d: %v", i, err)
			}
			if !allowed {
				t.Errorf("expected request %d to be allowed", i)
			}
		}

		allowed, err := cache.AllowRequest(ctx, key, limit, window)
		if err != nil {
			t.Fatalf("unexpected error on 4th request: %v", err)
		}
		if allowed {
			t.Error("expected 4th request to be blocked by rate limiter")
		}

		_ = rdb.Del(ctx, key).Err()
	})
}
