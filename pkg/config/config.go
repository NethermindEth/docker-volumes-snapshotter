package config

import (
	"errors"
	"io"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

const ConfigFileName = "config.yml"

var ConfigFilePath = "/" + ConfigFileName

// Config is the configuration for the backup/restore process.
type Config struct {
	Prefix  string   `yaml:"prefix"`
	Volumes []string `yaml:"volumes"`
}

// LoadConfig loads the configuration from the ConfigFilePath file.
func LoadConfig() (*Config, error) {
	configFile, err := os.Open(ConfigFilePath)
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

// Save saves the configuration to the given path in YAML format.
func (c *Config) Save(path string) error {
	configFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer configFile.Close()

	configData, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	_, err = configFile.Write(configData)
	return err
}
