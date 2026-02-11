package config

import (
	"fmt"
	yaml "gopkg.in/yaml.v2"
	"os"
	"reflect"
)

type Config struct {
	Migrations          string           `yaml:"migrations"`
	MasterDB            DBConfig         `yaml:"master_db"`
	SlaveDB             DBConfig         `yaml:"slave_db"`
	RpcServer           ServerConfig     `yaml:"rpc_server"`
	SlaveDbEnable       bool             `yaml:"slave_db_enable"`
	MetricsServer       ServerConfig     `yaml:"metrics_server"`
	HttpServer          ServerConfig     `yaml:"http_server"`
	WebsocketServer     ServerConfig     `yaml:"websocket_server"`
	ElasticsearchConfig ESConfig         `yaml:"elasticsearch_config"`
	CORSAllowedOrigins  string           `yaml:"cors_allowed_origins"`
	Sportradar          SportradarConfig `yaml:"sportradar"`
}

type DBConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Name     string `yaml:"name"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}

type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type SportradarConfig struct {
	AccessLevel string `yaml:"access_level"`
	ApiKey      string `yaml:"api_key"`
}

type ESConfig struct {
	Addresses    []string `yaml:"addresses"`     // Elasticsearch节点地址列表，如 ["http://localhost:9200"]
	Username     string   `yaml:"username"`      // 用户名（可选）
	Password     string   `yaml:"password"`      // 密码（可选）
	CloudID      string   `yaml:"cloud_id"`      // Elastic Cloud ID（可选）
	APIKey       string   `yaml:"api_key"`       // API Key（可选）
	ServiceToken string   `yaml:"service_token"` // Service Token（可选）
	Enable       bool     `yaml:"enable"`        // 是否启用Elasticsearch
}

func New(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", path, err)
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal yaml: %w", err)
	}

	if reflect.DeepEqual(cfg, &Config{}) {
		return nil, fmt.Errorf("config file %s is empty or invalid", path)
	}

	return cfg, nil
}
