package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	ServerPort              string
	CacheTTL                time.Duration
	RateLimitMaxTokens      int
	RateLimitRefillInterval time.Duration
	ProviderTimeout         time.Duration
}

func LoadConfig() (*Config, error) {
	cfg := &Config{
		ServerPort:              getEnv("SERVER_PORT", "8080"),
		CacheTTL:                30 * time.Second,
		RateLimitMaxTokens:      10,
		RateLimitRefillInterval: 1 * time.Minute,
		ProviderTimeout:         2 * time.Second,
	}

	if ttlStr := os.Getenv("CACHE_TTL_SECONDS"); ttlStr != "" {
		ttlSeconds, err := strconv.Atoi(ttlStr)
		if err != nil {
			return nil, fmt.Errorf("invalid CACHE_TTL_SECONDS: %w", err)
		}
		if ttlSeconds <= 0 {
			return nil, fmt.Errorf("CACHE_TTL_SECONDS must be positive, got: %d", ttlSeconds)
		}
		cfg.CacheTTL = time.Duration(ttlSeconds) * time.Second
	}

	if maxReqStr := os.Getenv("RATE_LIMIT_MAX_REQUESTS"); maxReqStr != "" {
		maxReq, err := strconv.Atoi(maxReqStr)
		if err != nil {
			return nil, fmt.Errorf("invalid RATE_LIMIT_MAX_REQUESTS: %w", err)
		}
		if maxReq <= 0 {
			return nil, fmt.Errorf("RATE_LIMIT_MAX_REQUESTS must be positive, got: %d", maxReq)
		}
		cfg.RateLimitMaxTokens = maxReq
	}

	if windowStr := os.Getenv("RATE_LIMIT_WINDOW_MINUTES"); windowStr != "" {
		windowMinutes, err := strconv.Atoi(windowStr)
		if err != nil {
			return nil, fmt.Errorf("invalid RATE_LIMIT_WINDOW_MINUTES: %w", err)
		}
		if windowMinutes <= 0 {
			return nil, fmt.Errorf("RATE_LIMIT_WINDOW_MINUTES must be positive, got: %d", windowMinutes)
		}
		cfg.RateLimitRefillInterval = time.Duration(windowMinutes) * time.Minute
	}

	if timeoutStr := os.Getenv("PROVIDER_TIMEOUT_SECONDS"); timeoutStr != "" {
		timeoutSeconds, err := strconv.Atoi(timeoutStr)
		if err != nil {
			return nil, fmt.Errorf("invalid PROVIDER_TIMEOUT_SECONDS: %w", err)
		}
		if timeoutSeconds <= 0 {
			return nil, fmt.Errorf("PROVIDER_TIMEOUT_SECONDS must be positive, got: %d", timeoutSeconds)
		}
		cfg.ProviderTimeout = time.Duration(timeoutSeconds) * time.Second
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
