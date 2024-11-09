package config

type OrderExecutorConfig struct {
	HTTPServer *HTTPServerConfig `json:"http-server"`
	Scylla     *ScyllaConfig     `json:"scylla"`
}
