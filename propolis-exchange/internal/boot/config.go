package boot

import (
	"context"
	"fmt"

	"github.com/sethvargo/go-envconfig"
)

type Config struct {
	Env     string `env:"ENV,default=dev"`
	BaseURL string `env:"BASE_URL,required"`
	DataDir string `env:"DATA_DIR"`
	Server  struct {
		Port    string `env:"PORT,default=8080"`
		Origins string `env:"ALLOWED_ORIGINS,required"`
	}
	Postgres struct {
		DatabaseURL string `env:"DATABASE_URL,required"`
	}
}

func Load() (*Config, error) {
	config := &Config{}
	if err := envconfig.Process(context.Background(), config); err != nil {
		return nil, fmt.Errorf("parsing env vars: %w", err)
	}
	return config, nil
}

func (c *Config) IsProduction() bool {
	return c.Env == "prod"
}

func (c *Config) IsDevelopment() bool {
	return c.Env == "dev"
}

func (c *Config) DataDirectory() string {
	return c.DataDir
}
