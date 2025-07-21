package config

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/brianvoe/gofakeit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestServiceConfig struct {
	StaticBaseUrl string `koanf:"static_base_url"`
	StaticPath    string `koanf:"static_path"`
}

func TestConfig_Load(t *testing.T) {
	staticPath := gofakeit.Word()
	staticPathNotValid := gofakeit.Word()
	staticBaseUrl := gofakeit.URL()

	// env vars
	t.Setenv("APP_SERVICE_STATIC_PATH", staticPath)

	// config file
	tempDir := t.TempDir()
	configFilename := filepath.Join(tempDir, "config.yaml")
	configData := fmt.Sprintf(
		"service:\n  static_base_url: %s\n  static_path: %s",
		staticBaseUrl,
		staticPathNotValid,
	)
	err := os.WriteFile(configFilename, []byte(configData), os.ModePerm)
	require.NoError(t, err)

	// load config
	config, err := Load[TestServiceConfig]("APP_", configFilename)
	require.NoError(t, err)

	t.Run("yaml", func(t *testing.T) {
		assert.Equal(t, staticBaseUrl, config.Service.StaticBaseUrl)
		assert.NotEqual(t, staticPath, staticPathNotValid) // env var has higher priority
	})

	t.Run("env", func(t *testing.T) {
		assert.Equal(t, staticPath, config.Service.StaticPath)
	})
}
