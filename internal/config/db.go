package config

type ScyllaConfig struct {
	Nodes       []string `json:"nodes"`
	Port        int      `json:"port"`
	Keyspace    string   `json:"keyspace"`
	Consistency string   `json:"consistency"`
	NumRetries  int      `json:"num_retries"`
}
