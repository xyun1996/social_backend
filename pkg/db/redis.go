package db

import (
	"fmt"
	"os"
)

// RedisConfig captures the shared Redis connection settings.
type RedisConfig struct {
	Addr     string
	Username string
	Password string
	DB       int
}

// LoadRedisConfig reads shared Redis settings from environment variables.
func LoadRedisConfig() RedisConfig {
	return RedisConfig{
		Addr:     stringValueOrDefault(os.Getenv("REDIS_ADDR"), "localhost:6379"),
		Username: os.Getenv("REDIS_USERNAME"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       intValueOrDefault(os.Getenv("REDIS_DB"), 0),
	}
}

// URL returns a redis:// style connection string for documentation and debugging.
func (c RedisConfig) URL() string {
	auth := ""
	if c.Username != "" || c.Password != "" {
		auth = c.Username
		if c.Password != "" {
			auth += ":" + c.Password
		}
		auth += "@"
	}
	return fmt.Sprintf("redis://%s%s/%d", auth, c.Addr, c.DB)
}
