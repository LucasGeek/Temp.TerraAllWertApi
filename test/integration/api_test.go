package integration

import (
	"net/http/httptest"
	"testing"

	"api/api/handler"
	"api/api/router"
	"api/data/repositories"
	"api/data/services"
	"api/infra/config"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func TestAPIRoutes(t *testing.T) {
	// Create a new Fiber app
	app := fiber.New()

	// Create mock config
	cfg := &config.Config{
		App: config.AppConfig{
			Name:        "Test API",
			Version:     "1.0.0",
			Environment: "test",
		},
	}

	// Create mock repositories (with nil db for now)
	userRepo := repositories.NewUserRepository(nil)
	productRepo := repositories.NewProductRepository(nil)
	orderRepo := repositories.NewOrderRepository(nil)

	// Create services
	cacheService := services.NewCacheService(nil)
	userService := services.NewUserService(userRepo)
	productService := services.NewProductService(productRepo)
	orderService := services.NewOrderService(orderRepo, productRepo)

	// Create handlers
	userHandler := handler.NewUserHandler(userService, cacheService, cfg)
	productHandler := handler.NewProductHandler(productService, cacheService, cfg)
	orderHandler := handler.NewOrderHandler(orderService, userService, productService, cacheService, cfg)
	healthHandler := handler.NewHealthHandler(nil, nil, map[string]string{"status": "test"}, cfg)
	adminHandler := handler.NewAdminHandler(userRepo, productRepo, orderRepo, userService, productService, orderService, cacheService, nil, cfg, nil)

	// Setup routes
	router.SetupRoutes(app, userHandler, productHandler, orderHandler, healthHandler, adminHandler, cfg)

	tests := []struct {
		name     string
		method   string
		path     string
		expected int
	}{
		{"Health Check", "GET", "/health", 200},
		{"Get Users", "GET", "/api/v1/users", 200},
		{"Get Products", "GET", "/api/v1/products", 200},
		{"Get Orders", "GET", "/api/v1/orders", 200},
		{"API Docs", "GET", "/api/v1/docs", 200},
		{"Admin Stats", "GET", "/api/v1/admin/stats", 200},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			resp, err := app.Test(req)

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, resp.StatusCode)
		})
	}
}