package handler

import (
	"api/data/services"
	"api/infra/config"

	"github.com/gofiber/fiber/v2"
)

type UserHandler struct {
	userService  *services.UserService
	cacheService *services.CacheService
	config       *config.Config
}

func NewUserHandler(userService *services.UserService, cacheService *services.CacheService, cfg *config.Config) *UserHandler {
	return &UserHandler{
		userService:  userService,
		cacheService: cacheService,
		config:       cfg,
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