package config

import (
	"log"
	"os"
)

func New() *Config {
	return &Config{}
}

type Config struct{}

func (c *Config) RedisUrlOrDie() string {
	urlStr := os.Getenv("REDIS_URL")
	if urlStr == "" {
		log.Fatalf("REDIS_URL is not set")
	}
	return urlStr
}
