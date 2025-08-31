package handlers

import (
	"api/infra/config"

	"github.com/gofiber/fiber/v2"
)

type UserHandler struct {
	config *config.Config
}

func NewUserHandler(cfg *config.Config) *UserHandler {
	return &UserHandler{
		config: cfg,
	}
}

func (h *UserHandler) GetUsers(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"message": "Users endpoint - placeholder",
	})
}

func (h *UserHandler) CreateUser(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"message": "Create user endpoint - placeholder",
	})
}