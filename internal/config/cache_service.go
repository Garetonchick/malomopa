package config

import "malomopa/internal/common"

type CacheServiceConfig struct {
	DataSources    []*DataSourceConfig `json:"data_sources"`
	GlobalTimeout  common.Duration     `json:"global_timeout"`
	MaxParallelism int                 `json:"max_parallelism"`
}

type DataSourceConfig struct {
	Name     string           `json:"name"`
	Endpoint string           `json:"endpoint"`
	Deps     []string         `json:"deps,omitempty"`
	Cache    *CacheConfig     `json:"cache,omitempty"`
	Timeout  *common.Duration `json:"timeout,omitempty"`
}

type CacheConfig struct {
	Name    string          `json:"name"`
	TTL     common.Duration `json:"ttl"`
	MaxSize int64           `json:"max_size"`
}
