package routes

import (
	"github.com/gofiber/fiber/v2"
	"terra-allwert/api/handlers"
	"terra-allwert/infra/middleware"
)

func SetupFileRoutes(app *fiber.App, fileHandler *handlers.FileHandler, variantHandler *handlers.FileVariantHandler, authMiddleware *middleware.AuthMiddleware) {
	api := app.Group("/api/v1")

	// ============== FILE ROUTES ==============
	
	// File CRUD routes (all protected)
	files := api.Group("/files", authMiddleware.RequireAuth())
	files.Post("/", fileHandler.CreateFile)
	files.Get("/", fileHandler.GetFiles)
	files.Get("/:id", fileHandler.GetFileByID)
	files.Put("/:id", fileHandler.UpdateFile)
	files.Delete("/:id", fileHandler.DeleteFile)

	// Presigned URL routes for files
	files.Post("/presigned-upload", fileHandler.RequestPresignedUploadURL)
	files.Post("/multipart-upload", fileHandler.RequestMultipartUpload)
	files.Post("/:fileId/complete-multipart", fileHandler.CompleteMultipartUpload)
	files.Get("/:id/download-url", fileHandler.GetPresignedDownloadURL)

	// File variant routes nested under files
	files.Post("/:fileId/variants", variantHandler.CreateVariantForFile)
	files.Get("/:fileId/variants", variantHandler.GetFileVariantsByOriginalFile)
	files.Get("/:fileId/variants/:variantName", variantHandler.GetFileVariantByName)
	files.Delete("/:fileId/variants", variantHandler.DeleteAllVariantsForFile)

	// ============== FILE VARIANT ROUTES ==============
	
	// Standalone file variant routes (all protected)
	fileVariants := api.Group("/file-variants", authMiddleware.RequireAuth())
	fileVariants.Post("/", variantHandler.CreateFileVariant)
	fileVariants.Get("/", variantHandler.GetAllFileVariants)
	fileVariants.Get("/:id", variantHandler.GetFileVariantByID)
	fileVariants.Put("/:id", variantHandler.UpdateFileVariant)
	fileVariants.Delete("/:id", variantHandler.DeleteFileVariant)

	// Presigned URL routes for file variants
	fileVariants.Get("/:id/upload-url", variantHandler.GetPresignedVariantUploadURL)
	fileVariants.Get("/:id/download-url", variantHandler.GetPresignedVariantDownloadURL)
}