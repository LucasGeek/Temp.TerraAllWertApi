package middleware

import (
	"api/infra/config"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"time"
)

func ErrorHandler(cfg *config.Config) fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		code := fiber.StatusInternalServerError
		if e, ok := err.(*fiber.Error); ok {
			code = e.Code
		}

		return c.Status(code).JSON(fiber.Map{
			"error":   true,
			"message": err.Error(),
		})
	}
}

func RequestID() fiber.Handler {
	return requestid.New()
}

func Logger(cfg *config.Config) fiber.Handler {
	return logger.New(logger.Config{
		Format: "[${time}] ${status} - ${method} ${path} (${latency})\n",
	})
}

func CORS(cfg *config.Config) fiber.Handler {
	return cors.New(cors.Config{
		AllowOrigins:     "http://localhost:3000,http://localhost:8080",
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Content-Type,Authorization",
		AllowCredentials: true,
	})
}

func Security(cfg *config.Config) fiber.Handler {
	return helmet.New()
}

func RateLimit(cfg *config.Config) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        100,
		Expiration: 1 * time.Minute,
	})
}

func Monitoring(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Add monitoring logic here if needed
		return c.Next()
	}
}