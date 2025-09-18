package config

import (
	"os"
	"time"

	"admira-etl/internal/constants"
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
		Port:        getEnv("PORT", constants.DefaultPort),
		LogLevel:    getEnv("LOG_LEVEL", constants.DefaultLogLevel),
		HTTPTimeout: constants.DefaultHTTPTimeout * time.Second,
		MaxRetries:  constants.DefaultMaxRetries,
		RetryDelay:  constants.DefaultRetryDelay * time.Second,
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

