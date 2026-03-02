package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	Auth      AuthConfig      `mapstructure:"auth"`
	Upload    UploadConfig    `mapstructure:"upload"`
	Image     ImageConfig     `mapstructure:"image"`
	RateLimit RateLimitConfig `mapstructure:"rate_limit"`
	Cache     CacheConfig     `mapstructure:"cache"`
}

type ServerConfig struct {
	Port string `mapstructure:"port"`
	Host string `mapstructure:"host"`
}

type AuthConfig struct {
	Token string `mapstructure:"token"`
}

type UploadConfig struct {
	MaxSize      int64    `mapstructure:"max_size"`
	AllowedTypes []string `mapstructure:"allowed_types"`
	StoragePath  string   `mapstructure:"storage_path"`
}

type ImageConfig struct {
	MaxDimension  int `mapstructure:"max_dimension"`
	Quality       int `mapstructure:"quality"`
	ThumbnailSize int `mapstructure:"thumbnail_size"`
}

type RateLimitConfig struct {
	RequestsPerMinute int `mapstructure:"requests_per_minute"`
	Burst             int `mapstructure:"burst"`
}

type CacheConfig struct {
	RecentUploadsSize int `mapstructure:"recent_uploads_size"`
}

func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./")

	// Environment variable overrides
	viper.AutomaticEnv()
	viper.BindEnv("auth.token", "AUTH_TOKEN")

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
