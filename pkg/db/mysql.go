package db

import (
	"fmt"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
)

// MySQLConfig captures the shared MySQL connection settings.
type MySQLConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
	Params   map[string]string
}

// LoadMySQLConfig reads shared MySQL settings from environment variables.
func LoadMySQLConfig() MySQLConfig {
	return MySQLConfig{
		Host:     stringValueOrDefault(os.Getenv("MYSQL_HOST"), "localhost"),
		Port:     intValueOrDefault(os.Getenv("MYSQL_PORT"), 3306),
		User:     stringValueOrDefault(os.Getenv("MYSQL_USER"), "root"),
		Password: os.Getenv("MYSQL_PASSWORD"),
		Database: stringValueOrDefault(os.Getenv("MYSQL_DATABASE"), "social_backend"),
		Params: map[string]string{
			"parseTime": "true",
			"loc":       "UTC",
		},
	}
}

// DSN returns a go-sql-driver/mysql compatible DSN.
func (c MySQLConfig) DSN() string {
	query := url.Values{}
	keys := make([]string, 0, len(c.Params))
	for key := range c.Params {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		query.Set(key, c.Params[key])
	}

	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?%s", c.User, c.Password, c.Host, c.Port, c.Database, query.Encode())
}

func intValueOrDefault(raw string, fallback int) int {
	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

func stringValueOrDefault(raw string, fallback string) string {
	if strings.TrimSpace(raw) == "" {
		return fallback
	}
	return raw
}
