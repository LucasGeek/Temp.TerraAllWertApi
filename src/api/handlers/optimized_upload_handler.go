package handlers

import (
	"context"
	"fmt"
	"path/filepath"
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

// OptimizedUploadHandler handles large file uploads with best practices
type OptimizedUploadHandler struct {
	fileRepo           interfaces.FileRepository
	storageService     interfaces.StorageService
	uploadStateManager *storage.UploadStateManager
	progressHub        *websocket.ProgressHub
	rateLimiter        *middleware.UploadRateLimiter
	circuitBreaker     *middleware.CircuitBreaker
}

func NewOptimizedUploadHandler(
	fileRepo interfaces.FileRepository,
	storageService interfaces.StorageService,
	uploadStateManager *storage.UploadStateManager,
	progressHub *websocket.ProgressHub,
	rateLimiter *middleware.UploadRateLimiter,
	circuitBreaker *middleware.CircuitBreaker,
) *OptimizedUploadHandler {
	return &OptimizedUploadHandler{
		fileRepo:           fileRepo,
		storageService:     storageService,
		uploadStateManager: uploadStateManager,
		progressHub:        progressHub,
		rateLimiter:        rateLimiter,
		circuitBreaker:     circuitBreaker,
	}
}

// OptimizedUploadRequest represents a request for optimized file upload
type OptimizedUploadRequest struct {
	FileName    string `json:"file_name" validate:"required"`
	ContentType string `json:"content_type" validate:"required"`
	FileSize    int64  `json:"file_size" validate:"required,min=1"`
	UserID      string `json:"user_id" validate:"required"`
}

// OptimizedUploadResponse represents the response for optimized upload
type OptimizedUploadResponse struct {
	FileID    string                 `json:"file_id"`
	UploadID  string                 `json:"upload_id,omitempty"`
	UploadURL string                 `json:"upload_url,omitempty"`
	PartURLs  []PartUploadInfo       `json:"part_urls,omitempty"`
	Method    string                 `json:"method"` // "direct", "presigned", "multipart"
	ExpiresAt string                 `json:"expires_at"`
	Metadata  map[string]interface{} `json:"metadata"`
}

type PartUploadInfo struct {
	PartNumber int    `json:"part_number"`
	UploadURL  string `json:"upload_url"`
	Size       int64  `json:"size"`
}

// InitiateOptimizedUpload handles large file uploads using best practices
// @Summary Initiate optimized file upload
// @Description Automatically chooses the best upload method based on file size and system load
// @Tags files
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param upload body OptimizedUploadRequest true "Upload request"
// @Success 200 {object} OptimizedUploadResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 429 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /files/optimized-upload [post]
func (h *OptimizedUploadHandler) InitiateOptimizedUpload(c *fiber.Ctx) error {
	var req OptimizedUploadRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	ctx := c.Context()

	// Check circuit breaker
	if !h.circuitBreaker.IsAllowed() {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error":       "Service temporarily unavailable",
			"retry_after": "30",
		})
	}

	// Rate limiting check
	if !h.rateLimiter.Allow(req.UserID, req.FileSize) {
		category := middleware.GetFileSizeCategory(req.FileSize)
		return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
			"error":         "Upload rate limit exceeded",
			"file_size":     req.FileSize,
			"size_category": string(category),
			"retry_after":   "60",
		})
	}

	// Generate file ID
	fileID := uuid.New()

	// Choose upload method based on file size and best practices
	uploadMethod := h.chooseUploadMethod(req.FileSize)

	var response OptimizedUploadResponse
	var err error

	switch uploadMethod {
	case "direct":
		response, err = h.handleDirectUpload(ctx, fileID, req)
	case "presigned":
		response, err = h.handlePresignedUpload(ctx, fileID, req)
	case "multipart":
		response, err = h.handleMultipartUpload(ctx, fileID, req)
	default:
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Invalid upload method determined",
		})
	}

	if err != nil {
		h.circuitBreaker.RecordFailure()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to initiate upload: %v", err),
		})
	}

	h.circuitBreaker.RecordSuccess()

	// Determine file type and extension
	extension := strings.ToLower(filepath.Ext(req.FileName))
	var fileType entities.FileType
	if strings.HasPrefix(req.ContentType, "image/") {
		fileType = entities.FileTypeImage
	} else if strings.HasPrefix(req.ContentType, "video/") {
		fileType = entities.FileTypeVideo
	} else {
		fileType = entities.FileTypeDocument
	}

	// Create file record
	now := time.Now()
	file := &entities.File{
		ID:            fileID,
		OriginalName:  req.FileName,
		MimeType:      req.ContentType,
		Extension:     extension,
		FileType:      fileType,
		FileSizeBytes: req.FileSize,
		StoragePath:   "uploading", // Will be updated when upload completes
		CreatedAt:     now,
		UpdatedAt:     &now,
	}

	if err := h.fileRepo.Create(ctx, file); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create file record",
		})
	}

	// Send initial progress update
	h.progressHub.BroadcastProgress(
		req.UserID,
		response.FileID,
		"file_upload",
		0,
		"initiated",
		fmt.Sprintf("Upload initiated for %s (%s method)", req.FileName, uploadMethod),
		map[string]interface{}{
			"file_size": req.FileSize,
			"method":    uploadMethod,
		},
	)

	return c.JSON(response)
}

// chooseUploadMethod selects the optimal upload method based on file size
func (h *OptimizedUploadHandler) chooseUploadMethod(fileSize int64) string {
	switch {
	case fileSize < 16*1024*1024: // < 16MB - direct upload through API
		return "direct"
	case fileSize < 100*1024*1024: // 16MB - 100MB - presigned URL
		return "presigned"
	default: // > 100MB - multipart upload
		return "multipart"
	}
}

// handleDirectUpload handles small files directly through the API
func (h *OptimizedUploadHandler) handleDirectUpload(ctx context.Context, fileID uuid.UUID, req OptimizedUploadRequest) (OptimizedUploadResponse, error) {
	return OptimizedUploadResponse{
		FileID:    fileID.String(),
		Method:    "direct",
		ExpiresAt: time.Now().Add(15 * time.Minute).Format(time.RFC3339),
		Metadata: map[string]interface{}{
			"upload_endpoint": "/api/v1/files/direct-upload",
			"max_size":        16 * 1024 * 1024,
		},
	}, nil
}

// handlePresignedUpload generates presigned URLs for medium-sized files
func (h *OptimizedUploadHandler) handlePresignedUpload(ctx context.Context, fileID uuid.UUID, req OptimizedUploadRequest) (OptimizedUploadResponse, error) {
	storageKey := fmt.Sprintf("files/%s/%s", fileID.String(), req.FileName)
	expiration := 60 * time.Minute

	uploadURL, err := h.storageService.GeneratePresignedUploadURL(ctx, storageKey, expiration, req.ContentType)
	if err != nil {
		return OptimizedUploadResponse{}, err
	}

	return OptimizedUploadResponse{
		FileID:    fileID.String(),
		UploadURL: uploadURL.String(),
		Method:    "presigned",
		ExpiresAt: time.Now().Add(expiration).Format(time.RFC3339),
		Metadata: map[string]interface{}{
			"storage_key": storageKey,
			"max_size":    100 * 1024 * 1024,
		},
	}, nil
}

// handleMultipartUpload initiates multipart upload for large files
func (h *OptimizedUploadHandler) handleMultipartUpload(ctx context.Context, fileID uuid.UUID, req OptimizedUploadRequest) (OptimizedUploadResponse, error) {
	storageKey := fmt.Sprintf("files/%s/%s", fileID.String(), req.FileName)

	uploadID, err := h.storageService.InitiateMultipartUpload(ctx, storageKey, req.ContentType)
	if err != nil {
		return OptimizedUploadResponse{}, err
	}

	// Calculate optimal part size based on file size
	partSize := h.calculateOptimalPartSize(req.FileSize)
	numParts := (req.FileSize + partSize - 1) / partSize

	// Generate presigned URLs for each part
	var partURLs []PartUploadInfo
	expiration := 2 * time.Hour // Longer expiration for large files

	for i := int64(1); i <= numParts; i++ {
		partURL, err := h.storageService.GeneratePresignedPartURL(ctx, storageKey, uploadID, int(i), expiration)
		if err != nil {
			return OptimizedUploadResponse{}, fmt.Errorf("failed to generate part URL %d: %w", i, err)
		}

		// Calculate size for this part
		currentPartSize := partSize
		if i == numParts {
			currentPartSize = req.FileSize - (i-1)*partSize
		}

		partURLs = append(partURLs, PartUploadInfo{
			PartNumber: int(i),
			UploadURL:  partURL.String(),
			Size:       currentPartSize,
		})
	}

	// Save upload state
	uploadState := &storage.UploadState{
		UploadID:     uploadID,
		ObjectKey:    storageKey,
		TotalSize:    req.FileSize,
		UploadedSize: 0,
		Parts:        make(map[int]string),
		ContentType:  req.ContentType,
		UserID:       req.UserID,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		Status:       storage.UploadStatusInitiated,
	}

	if err := h.uploadStateManager.SaveUploadState(ctx, uploadState); err != nil {
		return OptimizedUploadResponse{}, fmt.Errorf("failed to save upload state: %w", err)
	}

	return OptimizedUploadResponse{
		FileID:    fileID.String(),
		UploadID:  uploadID,
		PartURLs:  partURLs,
		Method:    "multipart",
		ExpiresAt: time.Now().Add(expiration).Format(time.RFC3339),
		Metadata: map[string]interface{}{
			"storage_key": storageKey,
			"part_size":   partSize,
			"num_parts":   numParts,
		},
	}, nil
}

// calculateOptimalPartSize determines the best part size for multipart upload
func (h *OptimizedUploadHandler) calculateOptimalPartSize(fileSize int64) int64 {
	switch {
	case fileSize > 5*1024*1024*1024: // > 5GB
		return 128 * 1024 * 1024 // 128MB parts
	case fileSize > 1024*1024*1024: // > 1GB
		return 64 * 1024 * 1024 // 64MB parts
	case fileSize > 100*1024*1024: // > 100MB
		return 32 * 1024 * 1024 // 32MB parts
	default:
		return 16 * 1024 * 1024 // 16MB parts
	}
}

// CompleteOptimizedUpload completes a multipart upload
// @Summary Complete optimized upload
// @Description Complete a multipart upload with part information
// @Tags files
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param uploadId path string true "Upload ID"
// @Param completion body CompleteMultipartRequest true "Completion data"
// @Success 200 {object} entities.File
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /files/optimized-upload/{uploadId}/complete [post]
func (h *OptimizedUploadHandler) CompleteOptimizedUpload(c *fiber.Ctx) error {
	uploadID := c.Params("uploadId")

	var req CompleteMultipartRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	ctx := c.Context()

	// Get upload state
	uploadState, err := h.uploadStateManager.GetUploadState(ctx, uploadID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Upload not found",
		})
	}

	// Complete the multipart upload
	if err := h.storageService.CompleteMultipartUpload(ctx, uploadState.ObjectKey, uploadID, req.Parts); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to complete upload",
		})
	}

	// Update file status
	fileID, _ := uuid.Parse(uploadState.ObjectKey[6:42]) // Extract UUID from "files/{uuid}/filename"
	file, err := h.fileRepo.GetByID(ctx, fileID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "File record not found",
		})
	}

	file.StoragePath = uploadState.ObjectKey
	now := time.Now()
	file.UpdatedAt = &now

	if err := h.fileRepo.Update(ctx, file); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update file record",
		})
	}

	// Update upload state
	uploadState.Status = storage.UploadStatusCompleted
	uploadState.UploadedSize = uploadState.TotalSize
	h.uploadStateManager.SaveUploadState(ctx, uploadState)

	// Send completion progress update
	h.progressHub.BroadcastProgress(
		uploadState.UserID,
		file.ID.String(),
		"file_upload",
		100,
		"completed",
		fmt.Sprintf("Upload completed successfully for %s", file.OriginalName),
		map[string]interface{}{
			"file_size":    file.FileSizeBytes,
			"storage_path": file.StoragePath,
		},
	)

	// Clean up upload state
	h.uploadStateManager.DeleteUploadState(ctx, uploadID)

	return c.JSON(file)
}

// GetUploadProgress returns the current progress of an upload
// @Summary Get upload progress
// @Description Get the current progress of a file upload
// @Tags files
// @Produce json
// @Security BearerAuth
// @Param uploadId path string true "Upload ID"
// @Success 200 {object} storage.UploadState
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /files/upload-progress/{uploadId} [get]
func (h *OptimizedUploadHandler) GetUploadProgress(c *fiber.Ctx) error {
	uploadID := c.Params("uploadId")

	ctx := c.Context()
	uploadState, err := h.uploadStateManager.GetUploadState(ctx, uploadID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Upload not found",
		})
	}

	// Calculate progress percentage
	progress := uploadState.CalculateProgress()

	return c.JSON(fiber.Map{
		"upload_id":     uploadState.UploadID,
		"object_key":    uploadState.ObjectKey,
		"total_size":    uploadState.TotalSize,
		"uploaded_size": uploadState.UploadedSize,
		"progress":      progress,
		"status":        uploadState.Status,
		"created_at":    uploadState.CreatedAt,
		"updated_at":    uploadState.UpdatedAt,
		"parts":         len(uploadState.Parts),
	})
}

// AbortOptimizedUpload cancels an ongoing upload
// @Summary Abort upload
// @Description Cancel an ongoing multipart upload
// @Tags files
// @Security BearerAuth
// @Param uploadId path string true "Upload ID"
// @Success 204 "No Content"
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /files/optimized-upload/{uploadId}/abort [delete]
func (h *OptimizedUploadHandler) AbortOptimizedUpload(c *fiber.Ctx) error {
	uploadID := c.Params("uploadId")

	ctx := c.Context()
	uploadState, err := h.uploadStateManager.GetUploadState(ctx, uploadID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Upload not found",
		})
	}

	// Abort the multipart upload in storage
	if err := h.storageService.AbortMultipartUpload(ctx, uploadState.ObjectKey, uploadID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to abort upload",
		})
	}

	// Update upload state
	uploadState.Status = storage.UploadStatusAborted
	h.uploadStateManager.SaveUploadState(ctx, uploadState)

	// Send abort progress update
	h.progressHub.BroadcastProgress(
		uploadState.UserID,
		uploadState.ObjectKey,
		"file_upload",
		0,
		"aborted",
		"Upload was aborted by user",
		map[string]interface{}{
			"reason": "user_requested",
		},
	)

	// Clean up upload state
	h.uploadStateManager.DeleteUploadState(ctx, uploadID)

	return c.SendStatus(fiber.StatusNoContent)
}
