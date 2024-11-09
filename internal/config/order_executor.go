package config

type OrderExecutorConfig struct {
	HTTPServer *HTTPServerConfig `json:"http_server"`
	Scylla     *ScyllaConfig     `json:"scylla"`
}
