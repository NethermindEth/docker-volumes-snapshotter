package config

import (
	"errors"
	"io"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Prefix     string   `yaml:"prefix"`
	BackupFile string   `yaml:"backup_file"`
	Volumes    []string `yaml:"volumes"`
}

func LoadConfig(path string) (*Config, error) {
	configFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer configFile.Close()

	var config Config
	configData, err := io.ReadAll(configFile)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(configData, &config)
	if err != nil {
		return nil, err
	}
	for _, v := range config.Volumes {
		if !filepath.IsAbs(v) {
			return nil, errors.New("volume path must be absolute")
		}
		if _, err := os.Stat(v); err != nil {
			return nil, err
		}
	}
	return &config, nil
}
