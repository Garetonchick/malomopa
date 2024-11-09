package config

import (
	"encoding/json"
	"io"
	"os"
)

type OrderAssignerConfig struct {
	HTTPServer   *HTTPServerConfig   `json:"http_server"`
	Scylla       *ScyllaConfig       `json:"scylla"`
	CacheService *CacheServiceConfig `json:"cache_service"`
}

func LoadConfig(path string) (*OrderAssignerConfig, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var cfg OrderAssignerConfig
	err = json.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
