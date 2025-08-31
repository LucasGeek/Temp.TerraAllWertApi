package handler

import (
	"api/data/repositories"
	"api/data/services"
	"api/infra/client"
	"api/infra/config"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type AdminHandler struct {
	userRepo       *repositories.UserRepository
	userService    *services.UserService
	cacheService   *services.CacheService
	externalClient *client.ExternalClient
	config         *config.Config
	db             *gorm.DB
}

func NewAdminHandler(
	userRepo *repositories.UserRepository,
	userService *services.UserService,
	cacheService *services.CacheService,
	externalClient *client.ExternalClient,
	cfg *config.Config,
	db *gorm.DB,
) *AdminHandler {
	return &AdminHandler{
		userRepo:       userRepo,
		userService:    userService,
		cacheService:   cacheService,
		externalClient: externalClient,
		config:         cfg,
		db:             db,
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
