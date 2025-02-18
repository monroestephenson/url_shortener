package cache

import (
	"context"
	"encoding/json"
	"time"

	"url_shortener/internal/models"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client *redis.Client
	ctx    context.Context
}

func NewRedisCache(addr, password string, db int) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	ctx := context.Background()

	// Test connection
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &RedisCache{
		client: client,
		ctx:    ctx,
	}, nil
}

// GetShortURL retrieves a short URL from cache
func (c *RedisCache) GetShortURL(shortCode string) (*models.ShortURL, error) {
	val, err := c.client.Get(c.ctx, "url:"+shortCode).Result()
	if err == redis.Nil {
		return nil, nil // Cache miss
	}
	if err != nil {
		return nil, err
	}

	var shortURL models.ShortURL
	if err := json.Unmarshal([]byte(val), &shortURL); err != nil {
		return nil, err
	}

	return &shortURL, nil
}

// SetShortURL stores a short URL in cache with expiration
func (c *RedisCache) SetShortURL(shortURL *models.ShortURL, expiration time.Duration) error {
	data, err := json.Marshal(shortURL)
	if err != nil {
		return err
	}

	return c.client.Set(c.ctx, "url:"+shortURL.ShortCode, data, expiration).Err()
}

// DeleteShortURL removes a short URL from cache
func (c *RedisCache) DeleteShortURL(shortCode string) error {
	return c.client.Del(c.ctx, "url:"+shortCode).Err()
}

// IncrementAccessCount atomically increments the access count
func (c *RedisCache) IncrementAccessCount(shortCode string) error {
	return c.client.Incr(c.ctx, "count:"+shortCode).Err()
}

// GetAccessCount gets the current access count from cache
func (c *RedisCache) GetAccessCount(shortCode string) (int64, error) {
	return c.client.Get(c.ctx, "count:"+shortCode).Int64()
}

// Close closes the Redis connection
func (c *RedisCache) Close() error {
	return c.client.Close()
}
