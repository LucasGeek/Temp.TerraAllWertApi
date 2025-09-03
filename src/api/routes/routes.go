package routes

import (
	"github.com/gofiber/fiber/v2"
	"terra-allwert/api/handlers"
	"terra-allwert/infra/middleware"
)

// SetupAllRoutes configures all API routes
func SetupAllRoutes(app *fiber.App, handlers *Handlers, authMiddleware *middleware.AuthMiddleware) {
	// Setup individual route groups
	SetupEnterpriseRoutes(app, handlers.EnterpriseHandler, authMiddleware)
	SetupMenuRoutes(app, handlers.MenuHandler, authMiddleware)
	SetupFloorPlanRoutes(app, handlers.TowerHandler, handlers.FloorHandler, authMiddleware)
	SetupSuiteRoutes(app, handlers.SuiteHandler, authMiddleware)
	SetupCarouselRoutes(app, handlers.CarouselHandler, authMiddleware)
	SetupPinsRoutes(app, handlers.PinsHandler, authMiddleware)
	SetupFileRoutes(app, handlers.FileHandler, handlers.FileVariantHandler, authMiddleware)
}

// Handlers holds all handler instances
type Handlers struct {
	EnterpriseHandler   *handlers.EnterpriseHandler
	MenuHandler         *handlers.MenuHandler
	TowerHandler        *handlers.TowerHandler
	FloorHandler        *handlers.FloorHandler
	SuiteHandler        *handlers.SuiteHandler
	CarouselHandler     *handlers.CarouselHandler
	PinsHandler         *handlers.PinsHandler
	FileHandler         *handlers.FileHandler
	FileVariantHandler  *handlers.FileVariantHandler
}