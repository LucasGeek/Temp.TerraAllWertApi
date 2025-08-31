package handler

import (
	"api/data/services"
	"api/infra/config"

	"github.com/gofiber/fiber/v2"
)

type OrderHandler struct {
	orderService   *services.OrderService
	userService    *services.UserService
	productService *services.ProductService
	cacheService   *services.CacheService
	config         *config.Config
}

func NewOrderHandler(
	orderService *services.OrderService,
	userService *services.UserService,
	productService *services.ProductService,
	cacheService *services.CacheService,
	cfg *config.Config,
) *OrderHandler {
	return &OrderHandler{
		orderService:   orderService,
		userService:    userService,
		productService: productService,
		cacheService:   cacheService,
		config:         cfg,
	}
}

func (h *OrderHandler) GetOrders(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"message": "Orders endpoint - placeholder",
	})
}

func (h *OrderHandler) CreateOrder(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"message": "Create order endpoint - placeholder",
	})
}