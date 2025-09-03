package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"terra-allwert/api/handlers"
	"terra-allwert/domain/interfaces"
	"terra-allwert/infra/middleware"
	"terra-allwert/infra/storage"
	"terra-allwert/infra/websocket"
)

// SetupOptimizedUploadRoutes configures the optimized upload routes
func SetupOptimizedUploadRoutes(
	app *fiber.App,
	fileRepo interfaces.FileRepository,
	storageService interfaces.StorageService,
	redisClient *redis.Client,
	progressHub *websocket.ProgressHub,
	rateLimiter *middleware.UploadRateLimiter,
	circuitBreaker *middleware.CircuitBreaker,
	authMiddleware *middleware.AuthMiddleware,
) {
	// Initialize upload state manager
	uploadStateManager := storage.NewUploadStateManager(redisClient)

	// Initialize optimized upload handler
	optimizedHandler := handlers.NewOptimizedUploadHandler(
		fileRepo,
		storageService,
		uploadStateManager,
		progressHub,
		rateLimiter,
		circuitBreaker,
	)

	api := app.Group("/api/v1")
	files := api.Group("/files", authMiddleware.RequireAuth())

	// Optimized upload endpoints
	files.Post("/optimized-upload", optimizedHandler.InitiateOptimizedUpload)
	files.Post("/optimized-upload/:uploadId/complete", optimizedHandler.CompleteOptimizedUpload)
	files.Get("/upload-progress/:uploadId", optimizedHandler.GetUploadProgress)
	files.Delete("/optimized-upload/:uploadId/abort", optimizedHandler.AbortOptimizedUpload)
}