package unit

import (
	"net/http/httptest"
	"testing"

	"api/api/handler"
	"api/infra/config"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func TestHealthHandler_Health(t *testing.T) {
	// Create a new Fiber app
	app := fiber.New()

	// Create config and handler
	cfg := &config.Config{
		App: config.AppConfig{
			Name:        "Test App",
			Version:     "1.0.0",
			Environment: "test",
		},
	}

	info := map[string]interface{}{
		"name":    "Test App",
		"version": "1.0.0",
	}

	healthHandler := handler.NewHealthHandler(nil, nil, info, cfg)

	// Setup route
	app.Get("/health", healthHandler.Health)

	// Create test request
	req := httptest.NewRequest("GET", "/health", nil)
	resp, err := app.Test(req)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}