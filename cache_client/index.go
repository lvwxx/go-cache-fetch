package client

import (
	"context"
	"time"
)

type CacheClient interface {
	Get(ctx context.Context, key string) string
	Set(ctx context.Context, key string, value interface{}, ex time.Duration) error
	Del(ctx context.Context, key string) error
}
