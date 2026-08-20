// Package config handles application configuration loading from environment variables
package config

import (
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

// Config holds application configuration
type Config struct {
	BasePath          string     `env:"BASE_PATH"`
	ConversionRetries int        `env:"CONVERSION_RETRIES" envDefault:"3"`
	LogLevel          slog.Level `env:"LOG_LEVEL" envDefault:"info"`
	// MaxFileSize is the maximum uploaded file size in bytes, 0 means unlimited
	MaxFileSize       int64  `env:"MAX_FILE_SIZE" envDefault:"0"`
	MaxWorkers        int    `env:"MAX_WORKERS" envDefault:"8"`
	Port              int    `env:"PORT" envDefault:"3000"`
	PrettyLogs        bool   `env:"PRETTY_LOGS" envDefault:"0"`
	RequestIDHeader   string `env:"REQUEST_ID_HEADER"`
	RequestIDLogLabel string `env:"REQUEST_ID_LOG_LABEL" envDefault:"reqId"`
}

// Load parses environment variables into a Config struct. A missing .env file is
// fine, a malformed one is an error.
func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("loading .env: %w", err)
	}

	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("parsing environment: %w", err)
	}

	return cfg, nil
}
