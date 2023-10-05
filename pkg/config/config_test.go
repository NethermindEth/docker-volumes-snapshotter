package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	tempDir := t.TempDir()
	configFile, err := os.CreateTemp(tempDir, ConfigFileName)
	require.NoError(t, err)

	ConfigFilePath = configFile.Name()

	tc := []struct {
		name  string
		setup func(t *testing.T) ([]byte, *Config)
		err   error
	}{
		{
			name: "valid config",
			setup: func(t *testing.T) ([]byte, *Config) {
				volume1, err := os.MkdirTemp(tempDir, "volume1")
				require.NoError(t, err)
				volume2, err := os.CreateTemp(tempDir, "volume2-*.txt")
				require.NoError(t, err)
				configData := []byte(fmt.Sprintf(`prefix: prefix/path
volumes:
- %s
- %s
`, volume1, volume2.Name()))
				config := &Config{
					Prefix:  "prefix/path",
					Volumes: []string{volume1, volume2.Name()},
				}
				return configData, config
			},
			err: nil,
		},
		{
			name: "invalid config, path is not absolute",
			setup: func(t *testing.T) ([]byte, *Config) {
				volume1, err := os.MkdirTemp(tempDir, "volume1")
				require.NoError(t, err)
				volume2, err := os.CreateTemp(tempDir, "volume2-*.txt")
				require.NoError(t, err)
				configData := []byte(fmt.Sprintf(`prefix: prefix/path
volumes:
- %s
- %s
`, volume1, volume2.Name()))
				config := &Config{
					Prefix:  "prefix/path",
					Volumes: []string{volume1, filepath.Dir(volume2.Name())},
				}
				return configData, config
			},
			err: errors.New("volume path must be absolute"),
		},
		{
			name: "invalid config, path does not exist",
			setup: func(t *testing.T) ([]byte, *Config) {
				volume1, err := os.MkdirTemp(tempDir, "volume1")
				require.NoError(t, err)
				volume2, err := os.CreateTemp(tempDir, "volume2-*.txt")
				require.NoError(t, err)
				configData := []byte(fmt.Sprintf(`prefix: prefix/path
volumes:
- %s
- %s
`, volume1, volume2.Name()))
				config := &Config{
					Prefix:  "prefix/path",
					Volumes: []string{volume1 + "-suffix", volume2.Name()},
				}
				return configData, config
			},
			err: errors.New("volume path must be absolute"),
		},
	}

	for _, tt := range tc {
		t.Run(tt.name, func(t *testing.T) {
			configData, wantConfig := tt.setup(t)
			err = configFile.Truncate(0)
			require.NoError(t, err)
			_, err = configFile.Write(configData)
			require.NoError(t, err)

			config, err := LoadConfig()
			if tt.err != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, wantConfig, config)
			}
		})
	}
}

func TestSaveConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configFilePath := filepath.Join(tmpDir, ConfigFileName)

	// Create a config to save
	config := &Config{
		Prefix:  "prefix/path",
		Volumes: []string{"/path/to/volume1", "/path/to/volume2"},
	}

	// Test saving a valid config
	err := config.Save(configFilePath)
	assert.NoError(t, err)

	// Read the saved config file and compare to the original config
	savedConfigData, err := os.ReadFile(configFilePath)
	assert.NoError(t, err)
	assert.Equal(t, []byte(`prefix: prefix/path
volumes:
- /path/to/volume1
- /path/to/volume2
`), savedConfigData)
}
