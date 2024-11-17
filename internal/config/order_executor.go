package config

import (
	"encoding/json"
	"io"
	"os"
)

type OrderExecutorConfig struct {
	HTTPServer *HTTPServerConfig `json:"http_server"`
	Logger     *LoggerConfig     `json:"logger"`
	Scylla     *ScyllaConfig     `json:"scylla"`
}

func LoadExecutorConfig(path string) (*OrderExecutorConfig, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var cfg OrderExecutorConfig
	err = json.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
