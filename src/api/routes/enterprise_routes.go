package routes

import (
	"github.com/gofiber/fiber/v2"
	"terra-allwert/api/handlers"
	"terra-allwert/infra/middleware"
)

func SetupEnterpriseRoutes(app *fiber.App, handler *handlers.EnterpriseHandler, authMiddleware *middleware.AuthMiddleware) {
	api := app.Group("/api/v1")
	enterprises := api.Group("/enterprises", authMiddleware.RequireAuth())

	// Enterprise routes (all protected)
	enterprises.Post("/", handler.CreateEnterprise)
	enterprises.Get("/", handler.GetEnterprises)
	enterprises.Get("/search", handler.SearchEnterprises)
	enterprises.Get("/slug/:slug", handler.GetEnterpriseBySlug)
	enterprises.Get("/:id", handler.GetEnterpriseByID)
	enterprises.Put("/:id", handler.UpdateEnterprise)
	enterprises.Delete("/:id", handler.DeleteEnterprise)
}