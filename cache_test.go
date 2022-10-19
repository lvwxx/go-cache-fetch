package cache

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type cacheTest struct {
	n string
}

func (ct *cacheTest) Get(ctx context.Context, key string) string {
	return ct.n
}

func (ct *cacheTest) Set(ctx context.Context, key string, value interface{}, ex time.Duration) error {
	ct.n = "123"
	return nil
}

func (ct *cacheTest) Del(ctx context.Context, key string) error {
	ct.n = ""
	return nil
}

func getData() (data int64, err error) {
	time.Sleep(time.Second)
	return time.Now().UnixNano(), nil
}

func TestFetch(t *testing.T) {
	client := &cacheTest{}
	ctx := context.TODO()

	cache := NewCache(client, "test")
	// 模拟 10 个并发
	var results []int64
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		time.Sleep(time.Microsecond)
		go func() {
			defer wg.Done()
			var result int64
			ok, err := cache.Fetch(ctx, "test", &result, time.Minute, func() (rawResult interface{}, err error) {
				return getData()
			})
			assert.NoError(t, err)
			assert.Equal(t, true, ok)
			results = append(results, result)
		}()
	}

	wg.Wait()
	assert.Equal(t, results[0], results[9])

	result := cache.cacheClient.Get(ctx, "test")
	assert.Equal(t, "123", result)
}
