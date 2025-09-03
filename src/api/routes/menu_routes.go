package routes

import (
	"github.com/gofiber/fiber/v2"
	"terra-allwert/api/handlers"
	"terra-allwert/infra/middleware"
)

func SetupMenuRoutes(app *fiber.App, handler *handlers.MenuHandler, authMiddleware *middleware.AuthMiddleware) {
	api := app.Group("/api/v1")
	
	// Menu routes (all protected)
	menus := api.Group("/menus", authMiddleware.RequireAuth())
	menus.Post("/", handler.CreateMenu)
	menus.Get("/", handler.GetMenus)
	menus.Get("/:id", handler.GetMenuByID)
	menus.Put("/:id", handler.UpdateMenu)
	menus.Patch("/:id/position", handler.UpdateMenuPosition)
	menus.Delete("/:id", handler.DeleteMenu)
	menus.Get("/:parentId/children", handler.GetChildMenus)

	// Enterprise-specific menu routes (all protected)
	enterprises := api.Group("/enterprises", authMiddleware.RequireAuth())
	enterprises.Get("/:enterpriseId/menus", handler.GetMenusByEnterprise)
	enterprises.Get("/:enterpriseId/menus/hierarchy", handler.GetMenuHierarchy)
}