package config

import (
	"os"
	"time"
)

type Config struct {
	AdsAPIURL   string
	CRMAPIURL   string
	SinkURL     string
	SinkSecret  string
	Port        string
	LogLevel    string
	HTTPTimeout time.Duration
	MaxRetries  int
	RetryDelay  time.Duration
}

func Load() *Config {
	return &Config{
		AdsAPIURL:   getEnv("ADS_API_URL", ""),
		CRMAPIURL:   getEnv("CRM_API_URL", ""),
		SinkURL:     getEnv("SINK_URL", ""),
		SinkSecret:  getEnv("SINK_SECRET", ""),
		Port:        getEnv("PORT", "8080"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
		HTTPTimeout: 30 * time.Second,
		MaxRetries:  3,
		RetryDelay:  time.Second,
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

