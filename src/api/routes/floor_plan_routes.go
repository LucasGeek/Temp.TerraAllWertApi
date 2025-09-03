package routes

import (
	"github.com/gofiber/fiber/v2"
	"terra-allwert/api/handlers"
	"terra-allwert/infra/middleware"
)

func SetupFloorPlanRoutes(app *fiber.App, towerHandler *handlers.TowerHandler, floorHandler *handlers.FloorHandler, authMiddleware *middleware.AuthMiddleware) {
	api := app.Group("/api/v1")

	// Tower routes (all protected)
	towers := api.Group("/towers", authMiddleware.RequireAuth())
	towers.Post("/", towerHandler.CreateTower)
	towers.Get("/", towerHandler.GetTowers)
	towers.Get("/:id", towerHandler.GetTowerByID)
	towers.Put("/:id", towerHandler.UpdateTower)
	towers.Patch("/:id/position", towerHandler.UpdateTowerPosition)
	towers.Delete("/:id", towerHandler.DeleteTower)

	// Floor routes (all protected)
	floors := api.Group("/floors", authMiddleware.RequireAuth())
	floors.Post("/", floorHandler.CreateFloor)
	floors.Get("/", floorHandler.GetFloors)
	floors.Get("/:id", floorHandler.GetFloorByID)
	floors.Put("/:id", floorHandler.UpdateFloor)
	floors.Delete("/:id", floorHandler.DeleteFloor)

	// Menu Floor Plan specific routes (all protected)
	menuFloorPlans := api.Group("/menu-floor-plans", authMiddleware.RequireAuth())
	menuFloorPlans.Get("/:menuFloorPlanId/towers", towerHandler.GetTowersByMenuFloorPlan)

	// Tower-specific floor routes
	towers.Get("/:towerId/floors", floorHandler.GetFloorsByTower)
	towers.Get("/:towerId/floors/:floorNumber", floorHandler.GetFloorByNumber)
}