package config

import (
	"os"
)

func New() *Config {
	return &Config{}
}

type Config struct{}

func (c *Config) MustGetRedisURL() string {
	urlStr := os.Getenv("REDIS_URL")
	if urlStr == "" {
		panic("REDIS_URL is not set")
	}
	return urlStr
}
