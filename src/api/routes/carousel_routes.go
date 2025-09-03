package routes

import (
	"github.com/gofiber/fiber/v2"
	"terra-allwert/api/handlers"
	"terra-allwert/infra/middleware"
)

func SetupCarouselRoutes(app *fiber.App, handler *handlers.CarouselHandler, authMiddleware *middleware.AuthMiddleware) {
	api := app.Group("/api/v1")

	// Menu Carousel routes (all protected)
	menuCarousels := api.Group("/menu-carousels", authMiddleware.RequireAuth())
	menuCarousels.Post("/", handler.CreateMenuCarousel)
	menuCarousels.Get("/:id", handler.GetMenuCarouselByID)
	menuCarousels.Put("/:id", handler.UpdateMenuCarousel)
	menuCarousels.Delete("/:id", handler.DeleteMenuCarousel)
	menuCarousels.Get("/:carouselId/items", handler.GetCarouselItemsByCarousel)

	// Carousel Item routes (all protected)
	carouselItems := api.Group("/carousel-items", authMiddleware.RequireAuth())
	carouselItems.Post("/", handler.CreateCarouselItem)
	carouselItems.Put("/:id", handler.UpdateCarouselItem)
	carouselItems.Patch("/:id/position", handler.UpdateCarouselItemPosition)
	carouselItems.Delete("/:id", handler.DeleteCarouselItem)
	carouselItems.Get("/:itemId/text-overlays", handler.GetTextOverlaysByItem)

	// Text Overlay routes (all protected)
	textOverlays := api.Group("/text-overlays", authMiddleware.RequireAuth())
	textOverlays.Post("/", handler.CreateTextOverlay)
	textOverlays.Put("/:id", handler.UpdateTextOverlay)
	textOverlays.Delete("/:id", handler.DeleteTextOverlay)

	// Menu-specific carousel routes (all protected)
	menus := api.Group("/menus", authMiddleware.RequireAuth())
	menus.Get("/:menuId/carousel", handler.GetMenuCarouselByMenuID)
}