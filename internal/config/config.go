package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Remote RemoteConfig `yaml:"remote"`
	Sync   SyncConfig   `yaml:"sync"`
}

type RemoteConfig struct {
	Host string `yaml:"host"`
	Path string `yaml:"path"`
	Port int    `yaml:"port"`
}

type SyncConfig struct {
	Source  string   `yaml:"source"`
	Exclude []string `yaml:"exclude"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if config.Remote.Port == 0 {
		config.Remote.Port = 8080
	}

	return &config, nil
}