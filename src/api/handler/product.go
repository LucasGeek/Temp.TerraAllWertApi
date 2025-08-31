package handler

import (
	"api/data/services"
	"api/infra/config"

	"github.com/gofiber/fiber/v2"
)

type ProductHandler struct {
	productService *services.ProductService
	cacheService   *services.CacheService
	config         *config.Config
}

func NewProductHandler(productService *services.ProductService, cacheService *services.CacheService, cfg *config.Config) *ProductHandler {
	return &ProductHandler{
		productService: productService,
		cacheService:   cacheService,
		config:         cfg,
	}
}

func (h *ProductHandler) GetProducts(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"message": "Products endpoint - placeholder",
	})
}

func (h *ProductHandler) CreateProduct(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"message": "Create product endpoint - placeholder",
	})
}