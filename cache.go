package cache

import (
	"context"
	"fmt"
	"reflect"
	"time"

	client "github.com/lvwxx/go-cache-fetch/cache_client"
	"github.com/vmihailenco/msgpack"
	"golang.org/x/sync/singleflight"
)

type Cache struct {
	g              singleflight.Group
	cacheKeyPrefix string
	cacheClient    client.CacheClient
}

func NewCache(cacheClient client.CacheClient, cacheKeyPrefix string) *Cache {
	if cacheKeyPrefix == "" {
		cacheKeyPrefix = "go-cache-fetch"
	}

	return &Cache{
		cacheKeyPrefix: cacheKeyPrefix,
		cacheClient:    cacheClient,
	}
}

func (c *Cache) Fetch(ctx context.Context, key string, result interface{}, ex time.Duration, fn func() (rawResult interface{}, err error)) (ok bool, err error) {
	returnValue := reflect.ValueOf(result).Elem()

	exist, err := c.Get(ctx, key, &result)
	if err != nil {
		return false, err
	}
	if exist {
		return true, nil
	}

	// 防止缓存穿透
	res, err, _ := c.g.Do(key, func() (interface{}, error) {
		data, err := fn()
		if err != nil {
			return data, err
		}
		err = c.Set(ctx, key, data, ex)
		return data, err
	})
	if err != nil {
		return
	}

	returnValue.Set(reflect.ValueOf(res))

	return false, nil
}

func (c *Cache) FetchIgnoreZero(ctx context.Context, key string, result interface{}, ex time.Duration, fn func() (rawResult interface{}, err error)) (ok bool, err error) {
	returnValue := reflect.ValueOf(result).Elem()

	exist, err := c.Get(ctx, key, &result)
	if err != nil {
		return false, err
	}
	if exist {
		if returnValue.IsZero() {
			// 删除老的缓存
			c.Delete(ctx, key)
		} else {
			return true, nil
		}
	}

	// 防止缓存穿透
	res, err, _ := c.g.Do(key, func() (interface{}, error) {
		data, err := fn()
		if err != nil {
			return data, err
		}
		err = c.Set(ctx, key, data, ex)
		return data, err
	})
	if err != nil {
		return
	}

	returnValue.Set(reflect.ValueOf(res))

	return false, nil
}

func (c *Cache) Get(ctx context.Context, key string, returnValue interface{}) (exist bool, err error) {
	result := c.cacheClient.Get(ctx, c.cacheKey(key))
	if result == "" {
		return false, nil
	}

	err = msgpack.Unmarshal([]byte(result), returnValue)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (c *Cache) Set(ctx context.Context, key string, value interface{}, ex time.Duration) (err error) {
	bytes, err := msgpack.Marshal(value)
	if err != nil {
		return
	}

	return c.cacheClient.Set(ctx, c.cacheKey(key), bytes, ex)
}

func (c *Cache) Delete(ctx context.Context, key string) (err error) {
	return c.cacheClient.Del(ctx, c.cacheKey(key))
}

func (c *Cache) cacheKey(key string) string {
	return fmt.Sprintf("%s:%s", c.cacheKeyPrefix, key)
}
