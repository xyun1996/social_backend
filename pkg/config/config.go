package config

import (
	"fmt"
	"os"
	"strings"
)

// ServiceConfig captures the minimum shared process configuration.
type ServiceConfig struct {
	Name string
	Env  string
	Addr string
}

// LoadServiceConfig reads process-level configuration from environment variables.
func LoadServiceConfig(serviceName string, defaultAddr string) ServiceConfig {
	prefix := strings.ToUpper(strings.ReplaceAll(serviceName, "-", "_"))

	return ServiceConfig{
		Name: serviceName,
		Env:  valueOrDefault(os.Getenv("APP_ENV"), "local"),
		Addr: valueOrDefault(os.Getenv(fmt.Sprintf("%s_ADDR", prefix)), defaultAddr),
	}
}

func valueOrDefault(value string, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}

	return value
}
