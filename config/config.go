package config

import (
	"fmt"
	"os"
	"reflect"
	"time"

	yaml "gopkg.in/yaml.v2"
)

type Config struct {
	Migrations                string       `yaml:"migrations"`
	RPCs                      []*RPC       `yaml:"rpcs"`
	MasterDB                  DBConfig     `yaml:"master_db"`
	SlaveDB                   DBConfig     `yaml:"slave_db"`
	SlaveDbEnable             bool         `yaml:"slave_db_enable"`
	ApiCacheEnable            bool         `yaml:"api_cache_enable"`
	CacheConfig               CacheConfig  `yaml:"cache_config"`
	RpcServer                 ServerConfig `yaml:"rpc_server"`
	MetricsServer             ServerConfig `yaml:"metrics_server"`
	HttpServer                ServerConfig `yaml:"http_server"`
	WebsocketServer           ServerConfig `yaml:"websocket_server"`
	EmailConfig               EmailConfig  `yaml:"email_config"`
	SMSConfig                 SMSConfig    `yaml:"sms_config"`
	MinioConfig               MinioConfig  `yaml:"minio_config"`
	KodoConfig                KodoConfig   `yaml:"kodo_config"`
	S3Config                  S3Config     `yaml:"s3_config"`
	ElasticsearchConfig       ESConfig     `yaml:"elasticsearch_config"`
	RedisConfig               RedisConfig  `yaml:"redis"`
	CORSAllowedOrigins        string       `yaml:"cors_allowed_origins"`
	JWTSecret                 string       `yaml:"jwt_secret"`
	Domain                    string       `yaml:"domain"`
	PrivateKey                string       `yaml:"private_key"`
	NumConfirmations          uint64       `yaml:"num_confirmations"`
	SafeAbortNonceTooLowCount uint64       `yaml:"safe_abort_nonce_too_low_count"`
	CallerAddress             string       `yaml:"caller_address"`
}

type ChainScannerConfig struct {
	RpcUrl        string `yaml:"rpc_url"`
	ConsumerToken string `yaml:"consumer_token"`
}

type WalletSignConfig struct {
	RpcUrl        string `yaml:"rpc_url"`
	ConsumerToken string `yaml:"consumer_token"`
	ChainName     string `yaml:"chain_name"`
}

type DBConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Name     string `yaml:"name"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}

type EmailConfig struct {
	SMTPHost     string `yaml:"smtp_host"`     // SMTP服务器地址
	SMTPPort     int    `yaml:"smtp_port"`     // SMTP端口
	SMTPUser     string `yaml:"smtp_user"`     // SMTP用户名
	SMTPPassword string `yaml:"smtp_password"` // SMTP密码
	FromName     string `yaml:"from_name"`     // 发件人名称
	FromEmail    string `yaml:"from_email"`    // 发件人邮箱
	UseSSL       bool   `yaml:"use_ssl"`       // 是否使用SSL/TLS
}

type SMSConfig struct {
	AccessKeyId     string `yaml:"access_key_id"`     // 阿里云AccessKeyId
	AccessKeySecret string `yaml:"access_key_secret"` // 阿里云AccessKeySecret
	SignName        string `yaml:"sign_name"`         // 短信签名
	TemplateCode    string `yaml:"template_code"`     // 短信模板代码
	Endpoint        string `yaml:"endpoint"`          // 短信服务端点
}

type MinioConfig struct {
	Endpoint        string `yaml:"endpoint"`
	AccessKeyID     string `yaml:"access_key_id"`
	SecretAccessKey string `yaml:"secret_access_key"`
	UseSSL          bool   `yaml:"use_ssl"`
	BucketName      string `yaml:"bucket_name"`
	BaseURL         string `yaml:"base_url"`
}

type CacheConfig struct {
	ListSize         int           `yaml:"list_size"`
	DetailSize       int           `yaml:"detail_size"`
	ListExpireTime   time.Duration `yaml:"list_expire_time"`
	DetailExpireTime time.Duration `yaml:"detail_expire_time"`
}

type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type KodoConfig struct {
	AccessKey     string `yaml:"access_key"`
	SecretKey     string `yaml:"secret_key"`
	Bucket        string `yaml:"bucket"`
	Domain        string `yaml:"domain"`
	Zone          string `yaml:"zone"`
	UseHTTPS      bool   `yaml:"use_https"`
	UseCdnDomains bool   `yaml:"use_cdn_domains"`
}

type S3Config struct {
	AccessKey    string `yaml:"access_key"`
	SecretKey    string `yaml:"secret_key"`
	Bucket       string `yaml:"bucket"`
	Region       string `yaml:"region"`
	Endpoint     string `yaml:"endpoint"`
	CDNDomain    string `yaml:"cdn_domain"`
	UsePathStyle bool   `yaml:"use_path_style"`
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

type RedisConfig struct {
	Enabled      bool          `yaml:"enable"`        // 是否启用Redis
	Addr         string        `yaml:"addr"`          // Redis地址，格式: host:port
	Password     string        `yaml:"password"`      // Redis密码（可选）
	DB           int           `yaml:"db"`            // Redis数据库索引
	PoolSize     int           `yaml:"pool_size"`     // 连接池大小
	MinIdle      int           `yaml:"min_idle"`      // 最小空闲连接数
	DialTimeout  time.Duration `yaml:"dial_timeout"`  // 连接超时时间
	ReadTimeout  time.Duration `yaml:"read_timeout"`  // 读取超时时间
	WriteTimeout time.Duration `yaml:"write_timeout"` // 写入超时时间
}

type RPC struct {
	RpcUrl    string   `yaml:"rpc_url"`
	ChainId   uint64   `yaml:"chain_id"`
	Contracts Contract `yaml:"contracts"`
}

type Contract struct {
	ReferralRewardManager string `yaml:"referral_reward_manager"`
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
