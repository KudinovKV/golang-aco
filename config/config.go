package config

import (
	"github.com/caarlos0/env/v6"
)

type config struct {
	InPath     string `env:"IN_PATH" envDefault:"in\\a280.tsp"`
	OutputPath string `env:"OUT_PATH" envDefault:"out.json"`
	Mode       string `env:"MODE" envDefault:"generate"`
}

func Config() (*config, error) {
	cfg := &config{}

	if err := env.Parse(cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}
