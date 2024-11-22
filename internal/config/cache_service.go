package config

import "time"

type CacheServiceConfig struct {
	DataSources []*DataSourceConfig `json:"data_sources"`
}

type DataSourceConfig struct {
	Name     string       `json:"name"`
	Endpoint string       `json:"endpoint"`
	Deps     []string     `json:"deps,omitempty"`
	Cache    *CacheConfig `json:"cache,omitempty"`
}

type CacheConfig struct {
	Name    string        `json:"name"`
	TTL     time.Duration `json:"ttl"`
	MaxSize int64         `json:"max_size"`
}
