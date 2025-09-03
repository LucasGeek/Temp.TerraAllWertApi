package routes

import (
	"github.com/gofiber/fiber/v2"
	"terra-allwert/api/handlers"
	"terra-allwert/domain/interfaces"
	"terra-allwert/infra/auth"
	"terra-allwert/infra/middleware"
)

// SetupAuthRoutes configures authentication routes
func SetupAuthRoutes(
	router fiber.Router,
	userRepo interfaces.UserRepository,
	jwtService *auth.JWTService,
	authMiddleware *middleware.AuthMiddleware,
) {
	// Create auth handler
	authHandler := handlers.NewAuthHandler(userRepo, jwtService)

	// Use the existing /api/v1 group passed from main.go
	auth := router.Group("/auth")

	// Public auth routes
	auth.Post("/login", authHandler.Login)
	auth.Post("/register", authHandler.Register)
	auth.Post("/refresh", authHandler.RefreshToken)
	auth.Post("/forgot-password", authHandler.ForgotPassword)
	auth.Post("/reset-password", authHandler.ResetPassword)

	// Protected auth routes
	authProtected := auth.Use(authMiddleware.RequireAuth())
	authProtected.Post("/logout", authHandler.Logout)
	authProtected.Get("/profile", authHandler.GetProfile)
	authProtected.Put("/profile", authHandler.UpdateProfile)
	authProtected.Post("/change-password", authHandler.ChangePassword)
}