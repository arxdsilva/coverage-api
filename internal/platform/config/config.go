package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	ServerAddr      string
	DatabaseURL     string
	APIKeyHeader    string
	APIKeySecret    string
	ShutdownTimeout time.Duration
}

func Load() (Config, error) {
	cfg := Config{
		ServerAddr:      getEnv("SERVER_ADDR", ":8080"),
		DatabaseURL:     os.Getenv("DATABASE_URL"),
		APIKeyHeader:    getEnv("API_KEY_HEADER", "X-API-Key"),
		APIKeySecret:    os.Getenv("API_KEY_SECRET"),
		ShutdownTimeout: getEnvDuration("SHUTDOWN_TIMEOUT_SECONDS", 10),
	}

	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}

func getEnvDuration(key string, defaultSeconds int) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return time.Duration(defaultSeconds) * time.Second
	}
	seconds, err := strconv.Atoi(v)
	if err != nil || seconds <= 0 {
		return time.Duration(defaultSeconds) * time.Second
	}
	return time.Duration(seconds) * time.Second
}

func (c Config) Validate() error {
	if c.ServerAddr == "" {
		return fmt.Errorf("server address cannot be empty")
	}
	if c.DatabaseURL == "" {
		return fmt.Errorf("database url cannot be empty")
	}
	if c.APIKeySecret == "" {
		return fmt.Errorf("api key secret cannot be empty")
	}
	return nil
}
