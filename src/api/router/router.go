package router

import (
	"api/api/handler"
	"api/infra/config"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(
	app *fiber.App,
	userHandler *handler.UserHandler,
	productHandler *handler.ProductHandler,
	orderHandler *handler.OrderHandler,
	healthHandler *handler.HealthHandler,
	adminHandler *handler.AdminHandler,
	cfg *config.Config,
) {
	// Health check
	app.Get("/health", healthHandler.Health)

	// API v1 routes
	v1 := app.Group("/api/v1")

	// User routes
	users := v1.Group("/users")
	users.Get("/", userHandler.GetUsers)
	users.Post("/", userHandler.CreateUser)

	// Product routes
	products := v1.Group("/products")
	products.Get("/", productHandler.GetProducts)
	products.Post("/", productHandler.CreateProduct)

	// Order routes
	orders := v1.Group("/orders")
	orders.Get("/", orderHandler.GetOrders)
	orders.Post("/", orderHandler.CreateOrder)

	// Admin routes
	admin := v1.Group("/admin")
	admin.Get("/stats", adminHandler.GetStats)
	admin.Post("/cache/flush", adminHandler.FlushCache)

	// Documentation endpoint placeholder
	v1.Get("/docs", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "API Documentation - placeholder",
			"version": "v1",
		})
	})
}