package cacheservice

import (
	"malomopa/internal/config"
	"time"

	"github.com/karlseguin/ccache/v3"
)

const DefaultCacheKey = "x"

type Cache interface {
	Get(key string) (any, bool)
	Set(key string, value any)
}

type LRUCache struct {
	cache *ccache.Cache[any]
	ttl   time.Duration
}

func GetFromCacheOrCompute(cache Cache, key string, compute func() (any, error)) (any, error) {
	if cache == nil {
		return compute()
	}
	if value, ok := cache.Get(key); ok {
		return value, nil
	}

	value, err := compute()
	if err != nil {
		return nil, err
	}

	cache.Set(key, value)

	return value, nil
}

func NewLRUCache(cfg *config.CacheConfig) *LRUCache {
	return &LRUCache{
		cache: ccache.New(ccache.Configure[any]().MaxSize(cfg.MaxSize)),
		ttl:   cfg.TTL,
	}
}

func (c *LRUCache) Get(key string) (any, bool) {
	res := c.cache.Get(key)
	if res != nil && !res.Expired() {
		return res.Value(), true
	}
	return nil, false
}

func (c *LRUCache) Set(key string, value any) {
	c.cache.Set(key, value, c.ttl)
}
