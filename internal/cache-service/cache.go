package cacheservice

import (
	"fmt"
	"malomopa/internal/config"
	"time"

	"github.com/karlseguin/ccache/v3"
)

const LRUCacheName = "lru"
const DefaultCacheKey = "x"

type Cache interface {
	Get(key string) (any, bool)
	Set(key string, value any)
}

type LRUCache struct {
	cache *ccache.Cache[any]
	ttl   time.Duration
}

func NewCache(cfg *config.CacheConfig) (Cache, error) {
	if cfg == nil {
		return nil, nil
	}
	// if need arises, more elegant and scalable dispatch mechanism
	// can be implemented, but since we only have 1 cache type supported
	// this is okay for now
	if cfg.Name != LRUCacheName {
		return nil, fmt.Errorf("unsupported cache with name %q", cfg.Name)
	}
	return NewLRUCache(cfg), nil
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
		cache: ccache.New(
			ccache.Configure[any]().
				MaxSize(cfg.MaxSize).
				GetsPerPromote(1). // Not optimal, but convinient for testing. TODO: add to config
				ItemsToPrune(1),
		),
		ttl: cfg.TTL,
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
