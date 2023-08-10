# go-cache-fetch

并发下获取数据防止缓存穿透的 golang 库

## 使用

### 下载

```bash
go get github.com/lvwxx/go-cache-fetch
```

cache 客户端以 redis 为例

```golang
import (
  cf "github.com/lvwxx/go-cache-fetch"
  cc "github.com/lvwxx/go-cache-fetch/cache_client"
  redis "github.com/go-redis/redis/v8"
)

cacheCli := cc.NewRedis(redis.Client)
cache := cf.NewCache(cacheCli, "prefix") 


// 优先从缓存中获取数据，如果缓存中没有，从 db 中获取数据并存入缓存，返回获取的值

ok, err := cache.Fetch(ctx, "cacheKey", &result, "time.Minute", func() (interface{}, error) {
  return getDataFromDB()
})

// reslut 就是最后获取到的值
```

如果使用其他缓存客户端，只要实现 CacheClient 的接口即可。

```golang
type CacheClient interface {
  Get(ctx context.Context, key string) string
  Set(ctx context.Context, key string, value interface{}, ex time.Duration) error
  Del(ctx context.Context, key string) error
}
```

## 缓存穿透

使用 singleflight.Group 来防止缓存穿透，在缓存失效时，同一时间高并发下只会有 1 个 DB 的请求，从而减小 DB 的瞬时压力。

