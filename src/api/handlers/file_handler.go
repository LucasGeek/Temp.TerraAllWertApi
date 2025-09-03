package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"terra-allwert/domain/entities"
	"terra-allwert/domain/interfaces"
	"terra-allwert/infra/middleware"
	"terra-allwert/infra/storage"
	"terra-allwert/infra/websocket"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type FileHandler struct {
	fileRepo           interfaces.FileRepository
	fileVariantRepo    interfaces.FileVariantRepository
	storageService     interfaces.StorageService
	uploadStateManager *storage.UploadStateManager
	progressHub        *websocket.ProgressHub
	rateLimiter        *middleware.UploadRateLimiter
	circuitBreaker     *middleware.CircuitBreaker
}

func NewFileHandler(
	fileRepo interfaces.FileRepository,
	fileVariantRepo interfaces.FileVariantRepository,
	storageService interfaces.StorageService,
	uploadStateManager *storage.UploadStateManager,
	progressHub *websocket.ProgressHub,
	rateLimiter *middleware.UploadRateLimiter,
	circuitBreaker *middleware.CircuitBreaker,
) *FileHandler {
	return &FileHandler{
		fileRepo:           fileRepo,
		fileVariantRepo:    fileVariantRepo,
		storageService:     storageService,
		uploadStateManager: uploadStateManager,
		progressHub:        progressHub,
		rateLimiter:        rateLimiter,
		circuitBreaker:     circuitBreaker,
	}
}

type PresignedUploadRequest struct {
	FileName    string `json:"file_name" validate:"required"`
	ContentType string `json:"content_type" validate:"required"`
	FileSize    int64  `json:"file_size" validate:"required,min=1"`
}

type PresignedUploadResponse struct {
	UploadURL   string `json:"upload_url"`
	FileID      string `json:"file_id"`
	StoragePath string `json:"storage_path"`
	ExpiresAt   string `json:"expires_at"`
}

type MultipartUploadRequest struct {
	FileName    string `json:"file_name" validate:"required"`
	ContentType string `json:"content_type" validate:"required"`
	FileSize    int64  `json:"file_size" validate:"required,min=1"`
	PartSize    int64  `json:"part_size,omitempty"` // Optional, defaults to 5MB
}

type MultipartUploadResponse struct {
	UploadID    string    `json:"upload_id"`
	FileID      string    `json:"file_id"`
	StoragePath string    `json:"storage_path"`
	PartURLs    []PartURL `json:"part_urls"`
}

type PartURL struct {
	PartNumber int    `json:"part_number"`
	UploadURL  string `json:"upload_url"`
	ExpiresAt  string `json:"expires_at"`
}

type CompleteMultipartRequest struct {
	UploadID string                    `json:"upload_id" validate:"required"`
	Parts    []interfaces.CompletePart `json:"parts" validate:"required"`
}

// ============== PRESIGNED URL ENDPOINTS ==============

// RequestPresignedUploadURL generates a presigned URL for direct file upload
// @Summary Request presigned upload URL
// @Description Generate a presigned URL for direct file upload to storage
// @Tags files
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body PresignedUploadRequest true "Upload request data"
// @Success 200 {object} PresignedUploadResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /files/presigned-upload [post]
func (h *FileHandler) RequestPresignedUploadURL(c *fiber.Ctx) error {
	var req PresignedUploadRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Generate unique file ID and storage path
	fileID := uuid.New()
	ext := filepath.Ext(req.FileName)
	storagePath := generateStoragePath(fileID, ext)

	// Generate presigned URL (24 hours expiration)
	expiration := 24 * time.Hour
	presignedURL, err := h.storageService.GeneratePresignedUploadURL(
		c.Context(),
		storagePath,
		expiration,
		req.ContentType,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate presigned URL",
		})
	}

	// Pre-register file in database
	file := &entities.File{
		ID:            fileID,
		FileType:      determineFileType(req.ContentType),
		MimeType:      req.ContentType,
		Extension:     strings.TrimPrefix(ext, "."),
		OriginalName:  req.FileName,
		StoragePath:   storagePath,
		FileSizeBytes: req.FileSize,
	}

	if err := h.fileRepo.Create(c.Context(), file); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to register file",
		})
	}

	response := PresignedUploadResponse{
		UploadURL:   presignedURL.String(),
		FileID:      fileID.String(),
		StoragePath: storagePath,
		ExpiresAt:   time.Now().Add(expiration).UTC().Format(time.RFC3339),
	}

	return c.JSON(response)
}

// RequestMultipartUpload initiates a multipart upload for large files
// @Summary Request multipart upload
// @Description Initiate a multipart upload for large files
// @Tags files
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body MultipartUploadRequest true "Multipart upload request"
// @Success 200 {object} MultipartUploadResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /files/multipart-upload [post]
func (h *FileHandler) RequestMultipartUpload(c *fiber.Ctx) error {
	var req MultipartUploadRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Set default part size to 5MB if not specified
	if req.PartSize == 0 {
		req.PartSize = 5 * 1024 * 1024 // 5MB
	}

	fileID := uuid.New()
	ext := filepath.Ext(req.FileName)
	storagePath := generateStoragePath(fileID, ext)

	// Initiate multipart upload
	uploadID, err := h.storageService.InitiateMultipartUpload(
		c.Context(),
		storagePath,
		req.ContentType,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to initiate multipart upload",
		})
	}

	// Calculate number of parts
	totalParts := int((req.FileSize + req.PartSize - 1) / req.PartSize)

	// Generate presigned URLs for each part
	expiration := 24 * time.Hour
	partURLs := make([]PartURL, totalParts)

	for i := 1; i <= totalParts; i++ {
		partURL, err := h.storageService.GeneratePresignedPartURL(
			c.Context(),
			storagePath,
			uploadID,
			i,
			expiration,
		)
		if err != nil {
			// Abort multipart upload on error
			h.storageService.AbortMultipartUpload(c.Context(), storagePath, uploadID)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to generate part URLs",
			})
		}

		partURLs[i-1] = PartURL{
			PartNumber: i,
			UploadURL:  partURL.String(),
			ExpiresAt:  time.Now().Add(expiration).UTC().Format(time.RFC3339),
		}
	}

	// Pre-register file in database
	file := &entities.File{
		ID:            fileID,
		FileType:      determineFileType(req.ContentType),
		MimeType:      req.ContentType,
		Extension:     strings.TrimPrefix(ext, "."),
		OriginalName:  req.FileName,
		StoragePath:   storagePath,
		FileSizeBytes: req.FileSize,
	}

	if err := h.fileRepo.Create(c.Context(), file); err != nil {
		// Abort multipart upload on error
		h.storageService.AbortMultipartUpload(c.Context(), storagePath, uploadID)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to register file",
		})
	}

	response := MultipartUploadResponse{
		UploadID:    uploadID,
		FileID:      fileID.String(),
		StoragePath: storagePath,
		PartURLs:    partURLs,
	}

	return c.JSON(response)
}

// CompleteMultipartUpload completes a multipart upload
// @Summary Complete multipart upload
// @Description Complete a multipart upload after all parts are uploaded
// @Tags files
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param fileId path string true "File ID"
// @Param request body CompleteMultipartRequest true "Complete request"
// @Success 200 {object} entities.File
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /files/{fileId}/complete-multipart [post]
func (h *FileHandler) CompleteMultipartUpload(c *fiber.Ctx) error {
	fileIDParam := c.Params("fileId")
	fileID, err := uuid.Parse(fileIDParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid file ID",
		})
	}

	var req CompleteMultipartRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Get file from database
	file, err := h.fileRepo.GetByID(c.Context(), fileID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "File not found",
		})
	}

	// Complete multipart upload
	err = h.storageService.CompleteMultipartUpload(
		c.Context(),
		file.StoragePath,
		req.UploadID,
		req.Parts,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to complete multipart upload",
		})
	}

	// Get file info from storage to update metadata
	fileInfo, err := h.storageService.GetFileInfo(c.Context(), file.StoragePath)
	if err == nil {
		// Update file with actual storage info
		file.FileSizeBytes = fileInfo.Size
		file.FileHash = &fileInfo.ETag

		// Set CDN URL if available
		cdnURL := h.storageService.GetFileURL(c.Context(), file.StoragePath)
		file.CdnURL = &cdnURL

		// Update file in database
		h.fileRepo.Update(c.Context(), file)
	}

	return c.JSON(file)
}

// ============== FILE CRUD ENDPOINTS ==============

// CreateFile creates a new file record
// @Summary Create a new file
// @Description Create a new file record with the provided data
// @Tags files
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param file body entities.File true "File data"
// @Success 201 {object} entities.File
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /files [post]
func (h *FileHandler) CreateFile(c *fiber.Ctx) error {
	var file entities.File

	if err := c.BodyParser(&file); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if err := h.fileRepo.Create(c.Context(), &file); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create file",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(file)
}

// GetFileByID gets a file by ID
// @Summary Get file by ID
// @Description Get a single file by its ID
// @Tags files
// @Produce json
// @Security BearerAuth
// @Param id path string true "File ID"
// @Success 200 {object} entities.File
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /files/{id} [get]
func (h *FileHandler) GetFileByID(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid file ID",
		})
	}

	file, err := h.fileRepo.GetByID(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "File not found",
		})
	}

	return c.JSON(file)
}

// GetFiles gets all files with pagination and filters
// @Summary Get all files
// @Description Get all files with optional pagination and filters
// @Tags files
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Param file_type query string false "File type filter"
// @Param mime_type query string false "MIME type filter"
// @Param uploader query string false "Uploader ID filter"
// @Success 200 {array} entities.File
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /files [get]
func (h *FileHandler) GetFiles(c *fiber.Ctx) error {
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	// Build filters
	filters := interfaces.FileSearchFilters{}

	if fileType := c.Query("file_type"); fileType != "" {
		ft := entities.FileType(fileType)
		filters.FileType = &ft
	}

	if mimeType := c.Query("mime_type"); mimeType != "" {
		filters.MimeType = &mimeType
	}

	if uploader := c.Query("uploader"); uploader != "" {
		if uploaderID, err := uuid.Parse(uploader); err == nil {
			filters.UploadedBy = &uploaderID
		}
	}

	if minSize := c.Query("min_size"); minSize != "" {
		if size, err := strconv.ParseInt(minSize, 10, 64); err == nil {
			filters.MinSize = &size
		}
	}

	if maxSize := c.Query("max_size"); maxSize != "" {
		if size, err := strconv.ParseInt(maxSize, 10, 64); err == nil {
			filters.MaxSize = &size
		}
	}

	var files []*entities.File
	var err error

	// Use search if filters are provided, otherwise get all
	if filters != (interfaces.FileSearchFilters{}) {
		files, err = h.fileRepo.Search(c.Context(), filters, limit, offset)
	} else {
		files, err = h.fileRepo.GetAll(c.Context(), limit, offset)
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch files",
		})
	}

	return c.JSON(files)
}

// UpdateFile updates an existing file
// @Summary Update file
// @Description Update an existing file
// @Tags files
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "File ID"
// @Param file body entities.File true "File data"
// @Success 200 {object} entities.File
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /files/{id} [put]
func (h *FileHandler) UpdateFile(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid file ID",
		})
	}

	var file entities.File
	if err := c.BodyParser(&file); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	file.ID = id
	if err := h.fileRepo.Update(c.Context(), &file); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update file",
		})
	}

	return c.JSON(file)
}

// DeleteFile deletes a file
// @Summary Delete file
// @Description Delete a file by ID (also removes from storage)
// @Tags files
// @Security BearerAuth
// @Param id path string true "File ID"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /files/{id} [delete]
func (h *FileHandler) DeleteFile(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid file ID",
		})
	}

	// Get file to get storage path
	file, err := h.fileRepo.GetByID(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "File not found",
		})
	}

	// Delete from storage
	if err := h.storageService.DeleteFile(c.Context(), file.StoragePath); err != nil {
		// Log error but don't fail the request
		fmt.Printf("Warning: Failed to delete file from storage: %v\n", err)
	}

	// Delete file variants from storage and database
	variants, _ := h.fileVariantRepo.GetByOriginalFileID(c.Context(), id, 1000, 0)
	for _, variant := range variants {
		h.storageService.DeleteFile(c.Context(), variant.StoragePath)
	}
	h.fileVariantRepo.DeleteByOriginalFile(c.Context(), id)

	// Delete from database
	if err := h.fileRepo.Delete(c.Context(), id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete file",
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// GetPresignedDownloadURL generates a presigned URL for file download
// @Summary Get presigned download URL
// @Description Generate a presigned URL for secure file download
// @Tags files
// @Produce json
// @Security BearerAuth
// @Param id path string true "File ID"
// @Param expires_in query int false "Expiration in seconds" default(3600)
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /files/{id}/download-url [get]
func (h *FileHandler) GetPresignedDownloadURL(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid file ID",
		})
	}

	file, err := h.fileRepo.GetByID(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "File not found",
		})
	}

	// Parse expiration (default 1 hour)
	expiresIn, _ := strconv.Atoi(c.Query("expires_in", "3600"))
	expiration := time.Duration(expiresIn) * time.Second

	downloadURL, err := h.storageService.GeneratePresignedDownloadURL(
		c.Context(),
		file.StoragePath,
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

func generateStoragePath(fileID uuid.UUID, extension string) string {
	// Generate storage path: files/YYYY/MM/DD/file-id.ext
	now := time.Now()
	return fmt.Sprintf("files/%d/%02d/%02d/%s%s",
		now.Year(), now.Month(), now.Day(),
		fileID.String(), extension)
}

func determineFileType(contentType string) entities.FileType {
	if strings.HasPrefix(contentType, "image/") {
		return entities.FileTypeImage
	}
	if strings.HasPrefix(contentType, "video/") {
		return entities.FileTypeVideo
	}
	return entities.FileTypeDocument
}

func generateFileHash(content []byte) string {
	hash := sha256.Sum256(content)
	return hex.EncodeToString(hash[:])
}
