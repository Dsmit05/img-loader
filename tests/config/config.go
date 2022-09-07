//go:build integration
// +build integration

package config

import "github.com/kelseyhightower/envconfig"

const envPrefix = "QA"

type Config struct {
	Host string `split_words:"true" default:"localhost:8191"`
}

func FromEnv() (*Config, error) {
	cfg := &Config{}
	err := envconfig.Process(envPrefix, cfg)
	return cfg, err
}
