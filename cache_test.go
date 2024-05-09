package cache

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/vmihailenco/msgpack"
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

func getZeroData() (data string, err error) {
	return "", nil
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
			assert.Equal(t, false, ok)
			results = append(results, result)
		}()
	}

	wg.Wait()
	assert.Equal(t, results[0], results[9])

	result := cache.cacheClient.Get(ctx, "test")
	assert.Equal(t, "123", result)
}

type cacheIgnoreZero struct {
	n []byte
}

func (ct *cacheIgnoreZero) Get(ctx context.Context, key string) string {
	return string(ct.n)
}

func (ct *cacheIgnoreZero) Set(ctx context.Context, key string, value interface{}, ex time.Duration) error {
	ct.n, _ = msgpack.Marshal("")
	return nil
}

func (ct *cacheIgnoreZero) Del(ctx context.Context, key string) error {
	ct.n = []byte("")
	return nil
}

func TestFetchIgnoreZero(t *testing.T) {
	client := &cacheIgnoreZero{}
	ctx := context.TODO()

	cache := NewCache(client, "test")

	var result string
	ok, err := cache.FetchIgnoreZero(ctx, "test", &result, time.Minute, func() (rawResult interface{}, err error) {
		return getZeroData()
	})

	assert.NoError(t, err)
	assert.Equal(t, false, ok)

	ok, err = cache.FetchIgnoreZero(ctx, "test", &result, time.Minute, func() (rawResult interface{}, err error) {
		return getZeroData()
	})
	assert.NoError(t, err)
	assert.Equal(t, false, ok)
}
