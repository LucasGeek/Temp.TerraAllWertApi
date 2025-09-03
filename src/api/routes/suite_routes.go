package routes

import (
	"github.com/gofiber/fiber/v2"
	"terra-allwert/api/handlers"
	"terra-allwert/infra/middleware"
)

func SetupSuiteRoutes(app *fiber.App, handler *handlers.SuiteHandler, authMiddleware *middleware.AuthMiddleware) {
	api := app.Group("/api/v1")
	
	// Suite routes (all protected)
	suites := api.Group("/suites", authMiddleware.RequireAuth())
	suites.Post("/", handler.CreateSuite)
	suites.Get("/", handler.GetSuites)
	suites.Get("/search", handler.SearchSuites)
	suites.Get("/:id", handler.GetSuiteByID)
	suites.Put("/:id", handler.UpdateSuite)
	suites.Patch("/:id/status", handler.UpdateSuiteStatus)
	suites.Delete("/:id", handler.DeleteSuite)

	// Floor-based suite routes (all protected)
	floors := api.Group("/floors", authMiddleware.RequireAuth())
	floors.Get("/:floorId/suites", handler.GetSuitesByFloor)
}