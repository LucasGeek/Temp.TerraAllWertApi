package unit

import (
	"os"
	"testing"
	"time"

	"api/infra/config"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfig_Defaults(t *testing.T) {
	// Clear environment variables to test defaults
	os.Clearenv()

	cfg, err := config.LoadConfig()

	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "Terra-Allwert-API", cfg.App.Name)
	assert.Equal(t, "1.0.0", cfg.App.Version)
	assert.Equal(t, "development", cfg.App.Environment)
	assert.Equal(t, "3000", cfg.App.Port)
	assert.Equal(t, true, cfg.App.Debug)
	assert.Equal(t, 30*time.Second, cfg.App.GracefulTimeout)
}

func TestLoadConfig_CustomValues(t *testing.T) {
	// Set custom environment variables
	os.Setenv("APP_NAME", "Custom API")
	os.Setenv("APP_VERSION", "2.0.0")
	os.Setenv("APP_ENV", "production")
	os.Setenv("APP_PORT", "8080")
	os.Setenv("APP_DEBUG", "false")

	defer func() {
		os.Clearenv()
	}()

	cfg, err := config.LoadConfig()

	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "Custom API", cfg.App.Name)
	assert.Equal(t, "2.0.0", cfg.App.Version)
	assert.Equal(t, "production", cfg.App.Environment)
	assert.Equal(t, "8080", cfg.App.Port)
	assert.Equal(t, false, cfg.App.Debug)
}

func TestConfig_IsDev(t *testing.T) {
	cfg := &config.Config{
		App: config.AppConfig{
			Environment: "development",
		},
	}

	assert.True(t, cfg.IsDev())
	assert.False(t, cfg.IsPrd())
}

func TestConfig_IsPrd(t *testing.T) {
	cfg := &config.Config{
		App: config.AppConfig{
			Environment: "production",
		},
	}

	assert.False(t, cfg.IsDev())
	assert.True(t, cfg.IsPrd())
}