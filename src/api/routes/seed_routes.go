package routes

import (
	"github.com/gofiber/fiber/v2"
	"terra-allwert/api/handlers"
	"terra-allwert/infra/config"
	"terra-allwert/infra/middleware"
)

func SetupSeedRoutes(app *fiber.App, cfg *config.Config, authMiddleware *middleware.AuthMiddleware) {
	seedHandler := handlers.NewSeedHandler(cfg)
	
	api := app.Group("/api/v1")
	seeds := api.Group("/seeds", authMiddleware.RequireAuth())

	// Seed endpoints
	seeds.Post("/run", seedHandler.RunSeeds)
	seeds.Post("/enterprises", seedHandler.RunEnterpriseSeeds)
	seeds.Post("/users", seedHandler.RunUserSeeds)
}