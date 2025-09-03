package routes

import (
	"github.com/gofiber/fiber/v2"
	"terra-allwert/api/handlers"
	"terra-allwert/infra/middleware"
)

func SetupPinsRoutes(app *fiber.App, handler *handlers.PinsHandler, authMiddleware *middleware.AuthMiddleware) {
	api := app.Group("/api/v1")

	// Menu Pins routes (all protected)
	menuPins := api.Group("/menu-pins", authMiddleware.RequireAuth())
	menuPins.Post("/", handler.CreateMenuPins)
	menuPins.Get("/:id", handler.GetMenuPinsByID)
	menuPins.Put("/:id", handler.UpdateMenuPins)
	menuPins.Delete("/:id", handler.DeleteMenuPins)
	menuPins.Get("/:menuPinId/markers", handler.GetPinMarkersByMenuPin)
	menuPins.Get("/:menuPinId/markers/search", handler.GetPinMarkersByPosition)

	// Pin Marker routes (all protected)
	pinMarkers := api.Group("/pin-markers", authMiddleware.RequireAuth())
	pinMarkers.Post("/", handler.CreatePinMarker)
	pinMarkers.Put("/:id", handler.UpdatePinMarker)
	pinMarkers.Delete("/:id", handler.DeletePinMarker)
	pinMarkers.Get("/:markerId/images", handler.GetPinMarkerImagesByMarker)

	// Pin Marker Image routes (all protected)
	pinMarkerImages := api.Group("/pin-marker-images", authMiddleware.RequireAuth())
	pinMarkerImages.Post("/", handler.CreatePinMarkerImage)
	pinMarkerImages.Put("/:id", handler.UpdatePinMarkerImage)
	pinMarkerImages.Patch("/:id/position", handler.UpdatePinMarkerImagePosition)
	pinMarkerImages.Delete("/:id", handler.DeletePinMarkerImage)

	// Menu-specific pins routes (all protected)
	menus := api.Group("/menus", authMiddleware.RequireAuth())
	menus.Get("/:menuId/pins", handler.GetMenuPinsByMenuID)
}