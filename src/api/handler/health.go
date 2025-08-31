package handler

import (
	"api/infra/config"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type HealthHandler struct {
	db     *gorm.DB
	redis  *redis.Client
	info   interface{}
	config *config.Config
}

func NewHealthHandler(db *gorm.DB, redis *redis.Client, info interface{}, cfg *config.Config) *HealthHandler {
	return &HealthHandler{
		db:     db,
		redis:  redis,
		info:   info,
		config: cfg,
	}
}

func (h *HealthHandler) Health(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status": "ok",
		"info":   h.info,
	})
}