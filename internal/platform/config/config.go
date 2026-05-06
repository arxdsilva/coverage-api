package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	ServerAddr                    string
	DatabaseURL                   string
	MigrationsDir                 string
	APIKeyHeader                  string
	APIKeySecret                  string
	GitHubToken                   string
	GitHubOrgs                    []string
	GitHubAPIBaseURL              string
	GitHubInsightsCacheTTLSeconds int
	GitHubInsightsMaxRepos        int
	GitHubInsightsWindowDays      []int
	GitHubInsightsSyncInterval    time.Duration
	ShutdownTimeout               time.Duration
}

func Load() (Config, error) {
	cfg := Config{
		ServerAddr:                    getEnv("SERVER_ADDR", ":8080"),
		DatabaseURL:                   os.Getenv("DATABASE_URL"),
		MigrationsDir:                 getEnv("MIGRATIONS_DIR", "./migrations"),
		APIKeyHeader:                  getEnv("API_KEY_HEADER", "X-API-Key"),
		APIKeySecret:                  os.Getenv("API_KEY_SECRET"),
		GitHubToken:                   strings.TrimSpace(os.Getenv("GITHUB_TOKEN")),
		GitHubOrgs:                    splitListEnv("GITHUB_ORGS"),
		GitHubAPIBaseURL:              getEnv("GITHUB_API_BASE_URL", "https://api.github.com"),
		GitHubInsightsCacheTTLSeconds: getEnvInt("GITHUB_INSIGHTS_CACHE_TTL_SECONDS", 60),
		GitHubInsightsMaxRepos:        getEnvInt("GITHUB_INSIGHTS_MAX_REPOS", 200),
		GitHubInsightsWindowDays:      parseWindowDays(strings.TrimSpace(os.Getenv("GITHUB_INSIGHTS_WINDOW_DAYS"))),
		GitHubInsightsSyncInterval:    getEnvDuration("GITHUB_INSIGHTS_SYNC_INTERVAL_SECONDS", 3600),
		ShutdownTimeout:               getEnvDuration("SHUTDOWN_TIMEOUT_SECONDS", 10),
	}

	if len(cfg.GitHubInsightsWindowDays) == 0 {
		cfg.GitHubInsightsWindowDays = []int{30, 60, 90}
	}

	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}

	return cfg, nil
}

func getEnvInt(key string, defaultValue int) int {
	v := os.Getenv(key)
	if v == "" {
		return defaultValue
	}
	parsed, err := strconv.Atoi(v)
	if err != nil {
		return defaultValue
	}
	return parsed
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
	if c.MigrationsDir == "" {
		return fmt.Errorf("migrations dir cannot be empty")
	}
	if c.APIKeySecret == "" {
		return fmt.Errorf("api key secret cannot be empty")
	}
	if c.GitHubToken == "" {
		return fmt.Errorf("github token cannot be empty")
	}
	if len(c.GitHubOrgs) == 0 {
		return fmt.Errorf("github orgs cannot be empty")
	}
	if c.GitHubInsightsCacheTTLSeconds < 0 {
		return fmt.Errorf("github insights cache ttl seconds must be >= 0")
	}
	if c.GitHubInsightsMaxRepos < 1 {
		return fmt.Errorf("github insights max repos must be >= 1")
	}
	if len(c.GitHubInsightsWindowDays) == 0 {
		return fmt.Errorf("github insights window days cannot be empty")
	}
	if c.GitHubInsightsSyncInterval <= 0 {
		return fmt.Errorf("github insights sync interval must be > 0")
	}
	return nil
}

func splitListEnv(key string) []string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	seen := map[string]struct{}{}
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		out = append(out, trimmed)
	}
	return out
}

func parseWindowDays(raw string) []int {
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]int, 0, len(parts))
	seen := map[int]struct{}{}
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		value, err := strconv.Atoi(trimmed)
		if err != nil || !isSupportedWindowDays(value) {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

func isSupportedWindowDays(value int) bool {
	return value == 30 || value == 60 || value == 90
}
