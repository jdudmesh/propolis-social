package boot

import (
	"context"
	"fmt"

	"github.com/sethvargo/go-envconfig"
)

type Config struct {
	Env           string `env:"ENV,default=dev"`
	DataDirectory string `env:"DATA_DIR"`
}

func Load() (Config, error) {
	config := Config{}
	if err := envconfig.Process(context.Background(), &config); err != nil {
		return Config{}, fmt.Errorf("parsing env vars: %w", err)
	}
	return config, nil
}
