// Package config handles application configuration loading from environment variables
package config

import (
	"context"
	"log/slog"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

// Config holds application configuration
type Config struct {
	BasePath          string `env:"BASE_PATH"`
	ConversionRetries int    `env:"CONVERSION_RETRIES" envDefault:"3"`
	MaxFileSize       int    `env:"MAX_FILE_SIZE" envDefault:"134217728"`
	MaxWorkers        int    `env:"MAX_WORKERS" envDefault:"8"`
	Port              int    `env:"PORT" envDefault:"3000"`
	PrettyLogs        bool   `env:"PRETTY_LOGS" envDefault:"0"`
	RequestIDHeader   string `env:"REQUEST_ID_HEADER"`
	RequestIDLogLabel string `env:"REQUEST_ID_LOG_LABEL" envDefault:"reqId"`
}

// Load parses environment variables into a Config struct
func Load(ctx context.Context) (*Config, error) {
	if err := godotenv.Load(); err != nil {
		slog.DebugContext(ctx, "No .env file found, relying on system environment variables")
	}

	cfg := &Config{}
	return cfg, env.Parse(cfg)
}
