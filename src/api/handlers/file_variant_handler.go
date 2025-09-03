package handlers

import (
	"fmt"
	"strconv"
	"time"

	"terra-allwert/domain/entities"
	"terra-allwert/domain/interfaces"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type FileVariantHandler struct {
	fileRepo        interfaces.FileRepository
	fileVariantRepo interfaces.FileVariantRepository
	storageService  interfaces.StorageService
}

type CreateVariantRequest struct {
	VariantName string `json:"variant_name" validate:"required"`
	Width       int    `json:"width" validate:"required,min=1"`
	Height      int    `json:"height" validate:"required,min=1"`
}

func NewFileVariantHandler(
	fileRepo interfaces.FileRepository,
	fileVariantRepo interfaces.FileVariantRepository,
	storageService interfaces.StorageService,
) *FileVariantHandler {
	return &FileVariantHandler{
		fileRepo:        fileRepo,
		fileVariantRepo: fileVariantRepo,
		storageService:  storageService,
	}
}

// CreateFileVariant creates a new file variant
// @Summary Create a new file variant
// @Description Create a new file variant (thumbnail, resized version, etc.)
// @Tags file-variants
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param variant body entities.FileVariant true "File variant data"
// @Success 201 {object} entities.FileVariant
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /file-variants [post]
func (h *FileVariantHandler) CreateFileVariant(c *fiber.Ctx) error {
	var variant entities.FileVariant

	if err := c.BodyParser(&variant); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if err := h.fileVariantRepo.Create(c.Context(), &variant); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create file variant",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(variant)
}

// CreateVariantForFile creates a variant for a specific original file
// @Summary Create variant for file
// @Description Create a new variant for a specific original file
// @Tags file-variants
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param fileId path string true "Original File ID"
// @Param request body CreateVariantRequest true "Variant creation request"
// @Success 201 {object} entities.FileVariant
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /files/{fileId}/variants [post]
func (h *FileVariantHandler) CreateVariantForFile(c *fiber.Ctx) error {
	fileIDParam := c.Params("fileId")
	fileID, err := uuid.Parse(fileIDParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid file ID",
		})
	}

	var req CreateVariantRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Check if original file exists
	originalFile, err := h.fileRepo.GetByID(c.Context(), fileID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Original file not found",
		})
	}

	// Generate storage path for variant
	variantStoragePath := generateVariantStoragePath(originalFile.StoragePath, req.VariantName)

	// Create file variant
	variant := &entities.FileVariant{
		ID:             uuid.New(),
		OriginalFileID: fileID,
		VariantName:    req.VariantName,
		StoragePath:    variantStoragePath,
		Width:          req.Width,
		Height:         req.Height,
		FileSizeBytes:  0, // Will be updated after upload
	}

	if err := h.fileVariantRepo.Create(c.Context(), variant); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create file variant",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(variant)
}

// GetFileVariantByID gets a file variant by ID
// @Summary Get file variant by ID
// @Description Get a single file variant by its ID
// @Tags file-variants
// @Produce json
// @Security BearerAuth
// @Param id path string true "File Variant ID"
// @Success 200 {object} entities.FileVariant
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /file-variants/{id} [get]
func (h *FileVariantHandler) GetFileVariantByID(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid file variant ID",
		})
	}

	variant, err := h.fileVariantRepo.GetByID(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "File variant not found",
		})
	}

	return c.JSON(variant)
}

// GetFileVariantsByOriginalFile gets all variants for a specific original file
// @Summary Get variants by original file
// @Description Get all variants for a specific original file
// @Tags file-variants
// @Produce json
// @Security BearerAuth
// @Param fileId path string true "Original File ID"
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Success 200 {array} entities.FileVariant
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /files/{fileId}/variants [get]
func (h *FileVariantHandler) GetFileVariantsByOriginalFile(c *fiber.Ctx) error {
	fileIDParam := c.Params("fileId")
	fileID, err := uuid.Parse(fileIDParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid file ID",
		})
	}

	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	variants, err := h.fileVariantRepo.GetByOriginalFileID(c.Context(), fileID, limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch file variants",
		})
	}

	return c.JSON(variants)
}

// GetFileVariantByName gets a specific variant by original file ID and variant name
// @Summary Get variant by name
// @Description Get a specific variant by original file ID and variant name
// @Tags file-variants
// @Produce json
// @Security BearerAuth
// @Param fileId path string true "Original File ID"
// @Param variantName path string true "Variant Name"
// @Success 200 {object} entities.FileVariant
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /files/{fileId}/variants/{variantName} [get]
func (h *FileVariantHandler) GetFileVariantByName(c *fiber.Ctx) error {
	fileIDParam := c.Params("fileId")
	fileID, err := uuid.Parse(fileIDParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid file ID",
		})
	}

	variantName := c.Params("variantName")
	if variantName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Variant name is required",
		})
	}

	variant, err := h.fileVariantRepo.GetByVariantName(c.Context(), fileID, variantName)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "File variant not found",
		})
	}

	return c.JSON(variant)
}

// GetAllFileVariants gets all file variants with pagination
// @Summary Get all file variants
// @Description Get all file variants with optional pagination
// @Tags file-variants
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Param width query int false "Filter by width"
// @Param height query int false "Filter by height"
// @Success 200 {array} entities.FileVariant
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /file-variants [get]
func (h *FileVariantHandler) GetAllFileVariants(c *fiber.Ctx) error {
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	var variants []*entities.FileVariant
	var err error

	// Check for dimension filters
	if widthStr := c.Query("width"); widthStr != "" {
		if heightStr := c.Query("height"); heightStr != "" {
			width, err1 := strconv.Atoi(widthStr)
			height, err2 := strconv.Atoi(heightStr)

			if err1 == nil && err2 == nil {
				variants, err = h.fileVariantRepo.GetByDimensions(c.Context(), width, height, limit, offset)
			} else {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "Invalid width or height parameters",
				})
			}
		} else {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Height parameter is required when width is specified",
			})
		}
	} else {
		variants, err = h.fileVariantRepo.GetAll(c.Context(), limit, offset)
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch file variants",
		})
	}

	return c.JSON(variants)
}

// UpdateFileVariant updates an existing file variant
// @Summary Update file variant
// @Description Update an existing file variant
// @Tags file-variants
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "File Variant ID"
// @Param variant body entities.FileVariant true "File variant data"
// @Success 200 {object} entities.FileVariant
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /file-variants/{id} [put]
func (h *FileVariantHandler) UpdateFileVariant(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid file variant ID",
		})
	}

	var variant entities.FileVariant
	if err := c.BodyParser(&variant); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	variant.ID = id
	if err := h.fileVariantRepo.Update(c.Context(), &variant); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update file variant",
		})
	}

	return c.JSON(variant)
}

// DeleteFileVariant deletes a file variant
// @Summary Delete file variant
// @Description Delete a file variant by ID (also removes from storage)
// @Tags file-variants
// @Security BearerAuth
// @Param id path string true "File Variant ID"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /file-variants/{id} [delete]
func (h *FileVariantHandler) DeleteFileVariant(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid file variant ID",
		})
	}

	// Get variant to get storage path
	variant, err := h.fileVariantRepo.GetByID(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "File variant not found",
		})
	}

	// Delete from storage
	if err := h.storageService.DeleteFile(c.Context(), variant.StoragePath); err != nil {
		// Log error but don't fail the request
		fmt.Printf("Warning: Failed to delete file variant from storage: %v\n", err)
	}

	// Delete from database
	if err := h.fileVariantRepo.Delete(c.Context(), id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete file variant",
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// DeleteAllVariantsForFile deletes all variants for a specific original file
// @Summary Delete all variants for file
// @Description Delete all variants for a specific original file
// @Tags file-variants
// @Security BearerAuth
// @Param fileId path string true "Original File ID"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /files/{fileId}/variants [delete]
func (h *FileVariantHandler) DeleteAllVariantsForFile(c *fiber.Ctx) error {
	fileIDParam := c.Params("fileId")
	fileID, err := uuid.Parse(fileIDParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid file ID",
		})
	}

	// Get all variants for the file
	variants, err := h.fileVariantRepo.GetByOriginalFileID(c.Context(), fileID, 1000, 0)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch file variants",
		})
	}

	// Delete each variant from storage
	for _, variant := range variants {
		if err := h.storageService.DeleteFile(c.Context(), variant.StoragePath); err != nil {
			// Log error but continue
			fmt.Printf("Warning: Failed to delete file variant from storage: %v\n", err)
		}
	}

	// Delete all variants from database
	if err := h.fileVariantRepo.DeleteByOriginalFile(c.Context(), fileID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete file variants",
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// GetPresignedVariantUploadURL generates a presigned URL for variant upload
// @Summary Get presigned variant upload URL
// @Description Generate a presigned URL for uploading a file variant
// @Tags file-variants
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "File Variant ID"
// @Param expires_in query int false "Expiration in seconds" default(3600)
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /file-variants/{id}/upload-url [get]
func (h *FileVariantHandler) GetPresignedVariantUploadURL(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid file variant ID",
		})
	}

	variant, err := h.fileVariantRepo.GetByID(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "File variant not found",
		})
	}

	// Parse expiration (default 1 hour)
	expiresIn, _ := strconv.Atoi(c.Query("expires_in", "3600"))
	expiration := time.Duration(expiresIn) * time.Second

	// Assume image content type for variants (could be made configurable)
	contentType := "image/jpeg"

	uploadURL, err := h.storageService.GeneratePresignedUploadURL(
		c.Context(),
		variant.StoragePath,
		expiration,
		contentType,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate upload URL",
		})
	}

	return c.JSON(fiber.Map{
		"upload_url": uploadURL.String(),
		"expires_at": time.Now().Add(expiration).UTC().Format(time.RFC3339),
	})
}

// GetPresignedVariantDownloadURL generates a presigned URL for variant download
// @Summary Get presigned variant download URL
// @Description Generate a presigned URL for secure variant download
// @Tags file-variants
// @Produce json
// @Security BearerAuth
// @Param id path string true "File Variant ID"
// @Param expires_in query int false "Expiration in seconds" default(3600)
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /file-variants/{id}/download-url [get]
func (h *FileVariantHandler) GetPresignedVariantDownloadURL(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid file variant ID",
		})
	}

	variant, err := h.fileVariantRepo.GetByID(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "File variant not found",
		})
	}

	// Parse expiration (default 1 hour)
	expiresIn, _ := strconv.Atoi(c.Query("expires_in", "3600"))
	expiration := time.Duration(expiresIn) * time.Second

	downloadURL, err := h.storageService.GeneratePresignedDownloadURL(
		c.Context(),
		variant.StoragePath,
		expiration,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate download URL",
		})
	}

	return c.JSON(fiber.Map{
		"download_url": downloadURL.String(),
		"expires_at":   time.Now().Add(expiration).UTC().Format(time.RFC3339),
	})
}

// ============== UTILITY FUNCTIONS ==============

func generateVariantStoragePath(originalPath, variantName string) string {
	// Extract file parts from original path
	// Example: files/2023/12/01/file-id.jpg -> files/2023/12/01/file-id-thumb.jpg

	// Find the last dot for extension
	lastDot := -1
	for i := len(originalPath) - 1; i >= 0; i-- {
		if originalPath[i] == '.' {
			lastDot = i
			break
		}
	}

	if lastDot == -1 {
		// No extension found
		return fmt.Sprintf("%s-%s", originalPath, variantName)
	}

	basePath := originalPath[:lastDot]
	extension := originalPath[lastDot:]

	return fmt.Sprintf("%s-%s%s", basePath, variantName, extension)
}
