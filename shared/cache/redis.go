package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	client *redis.Client
}

// Connect establishes a connection to Redis
func Connect(url string) (*RedisClient, error) {
	opts, err := redis.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("invalid redis url: %w", err)
	}

	client := redis.NewClient(opts)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("could not connect to redis: %w", err)
	}

	return &RedisClient{
		client: client,
	}, nil
}

func (r *RedisClient) Close() error {
	return r.client.Close()
}

// Set stores a value in Redis with an expiration time. The value is JSON marshaled.
func (r *RedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	bytes, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("could not marshal value: %w", err)
	}
	return r.client.Set(ctx, key, bytes, expiration).Err()
}

// Get retrieves a value from Redis and unmarshals it into the provided dest interface.
func (r *RedisClient) Get(ctx context.Context, key string, dest interface{}) error {
	val, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return err // Cache miss
		}
		return fmt.Errorf("could not get value: %w", err)
	}
	if err := json.Unmarshal(val, dest); err != nil {
		return fmt.Errorf("could not unmarshal value: %w", err)
	}
	return nil
}

// Delete removes a key from Redis.
func (r *RedisClient) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

// DeletePattern removes all keys matching the pattern (useful for invalidation).
func (r *RedisClient) DeletePattern(ctx context.Context, pattern string) error {
	iter := r.client.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		key := iter.Val()
		if err := r.client.Del(ctx, key).Err(); err != nil {
			log.Printf("Failed to delete key %s: %v", key, err)
		}
	}
	if err := iter.Err(); err != nil {
		return fmt.Errorf("error during scan: %w", err)
	}
	return nil
}

// IsCacheMiss checks if an error returned by Get is a cache miss.
func IsCacheMiss(err error) bool {
	return err == redis.Nil
}
