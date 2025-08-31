package handlers

import (
	"api/infra/config"

	"github.com/gofiber/fiber/v2"
)

type AdminHandler struct {
	config *config.Config
}

func NewAdminHandler(cfg *config.Config) *AdminHandler {
	return &AdminHandler{
		config: cfg,
	}
}

func (h *AdminHandler) GetStats(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"message": "Admin stats endpoint - placeholder",
	})
}

func (h *AdminHandler) FlushCache(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"message": "Cache flushed - placeholder",
	})
}
