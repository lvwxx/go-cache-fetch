package client

import (
	"context"
	"time"

	redis "github.com/go-redis/redis/v8"
)

type RedisClient struct {
	*redis.Client
}

func NewRedis(client *redis.Client) *RedisClient {
	return &RedisClient{
		Client: client,
	}
}

func (r *RedisClient) Get(ctx context.Context, key string) string {
	result, err := r.Client.Get(ctx, key).Result()
	if err != nil {
		return ""
	}

	return result
}

func (r *RedisClient) Set(ctx context.Context, key string, value interface{}, ex time.Duration) error {
	return r.Client.SetEX(ctx, key, value, ex).Err()
}

func (r *RedisClient) Del(ctx context.Context, key string) error {
	return r.Client.Del(ctx, key).Err()
}
