package upload

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"go.uber.org/zap"

	"api/domain/entities"
	"api/infra/logger"
)

// DirectUploadService handles direct uploads to MinIO with signed URLs
type DirectUploadService struct {
	minioClient *minio.Client
	bucketName  string
	baseURL     string
}

// UploadConfig contains upload configuration
type UploadConfig struct {
	MaxFileSize    int64         // Maximum file size in bytes
	AllowedTypes   []string      // Allowed MIME types
	ExpiryDuration time.Duration // URL expiry duration
}

// SignedUploadRequest represents a request for signed upload URL
type SignedUploadRequest struct {
	FileName    string `json:"fileName" validate:"required"`
	ContentType string `json:"contentType" validate:"required"`
	Folder      string `json:"folder" validate:"required"`
	FileSize    *int64 `json:"fileSize,omitempty"`
}

// SignedUploadResponse contains the signed upload URL and metadata
type SignedUploadResponse struct {
	UploadURL    string            `json:"uploadUrl"`
	FileKey      string            `json:"fileKey"`
	Fields       map[string]string `json:"fields"`
	ExpiresIn    int               `json:"expiresIn"`
	MaxFileSize  int64             `json:"maxFileSize"`
	AllowedTypes []string          `json:"allowedTypes"`
}

// UploadSession represents an upload session for tracking
type UploadSession struct {
	ID          string            `json:"id"`
	FileKey     string            `json:"fileKey"`
	FileName    string            `json:"fileName"`
	ContentType string            `json:"contentType"`
	FileSize    int64             `json:"fileSize"`
	Folder      string            `json:"folder"`
	Status      UploadStatus      `json:"status"`
	Metadata    map[string]string `json:"metadata"`
	CreatedAt   time.Time         `json:"createdAt"`
	CompletedAt *time.Time        `json:"completedAt,omitempty"`
	ExpiresAt   time.Time         `json:"expiresAt"`
}

type UploadStatus string

const (
	UploadStatusPending   UploadStatus = "PENDING"
	UploadStatusCompleted UploadStatus = "COMPLETED"
	UploadStatusFailed    UploadStatus = "FAILED"
	UploadStatusExpired   UploadStatus = "EXPIRED"
)

var (
	// Default upload configurations for different file types
	DefaultImageConfig = UploadConfig{
		MaxFileSize:    10 * 1024 * 1024, // 10MB
		AllowedTypes:   []string{"image/jpeg", "image/png", "image/webp", "image/gif"},
		ExpiryDuration: 15 * time.Minute,
	}
	
	DefaultVideoConfig = UploadConfig{
		MaxFileSize:    100 * 1024 * 1024, // 100MB
		AllowedTypes:   []string{"video/mp4", "video/webm", "video/mov", "video/avi"},
		ExpiryDuration: 30 * time.Minute,
	}
	
	DefaultDocumentConfig = UploadConfig{
		MaxFileSize:    50 * 1024 * 1024, // 50MB
		AllowedTypes:   []string{"application/pdf", "application/msword", "application/vnd.openxmlformats-officedocument.wordprocessingml.document"},
		ExpiryDuration: 15 * time.Minute,
	}
)

func NewDirectUploadService(minioClient *minio.Client, bucketName, baseURL string) *DirectUploadService {
	return &DirectUploadService{
		minioClient: minioClient,
		bucketName:  bucketName,
		baseURL:     baseURL,
	}
}

// GenerateSignedUploadURL generates a signed URL for direct upload to MinIO
func (s *DirectUploadService) GenerateSignedUploadURL(ctx context.Context, request SignedUploadRequest) (*entities.SignedUploadURL, error) {
	// Validate request
	if err := s.validateUploadRequest(request); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Get upload config based on content type
	config := s.getUploadConfig(request.ContentType)

	// Generate unique file key
	fileKey := s.generateFileKey(request.Folder, request.FileName)

	// Create presigned URL for PUT operation
	expiry := config.ExpiryDuration
	presignedURL, err := s.minioClient.PresignedPutObject(
		ctx,
		s.bucketName,
		fileKey,
		expiry,
	)
	if err != nil {
		logger.Error(ctx, "Failed to generate presigned URL", err,
			zap.String("bucket", s.bucketName),
			zap.String("key", fileKey),
		)
		return nil, fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	// Create upload session for tracking
	session := &UploadSession{
		ID:          uuid.New().String(),
		FileKey:     fileKey,
		FileName:    request.FileName,
		ContentType: request.ContentType,
		Folder:      request.Folder,
		Status:      UploadStatusPending,
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(expiry),
		Metadata:    make(map[string]string),
	}

	if request.FileSize != nil {
		session.FileSize = *request.FileSize
	}

	// Store session (could be in Redis/database for persistence)
	if err := s.storeUploadSession(ctx, session); err != nil {
		logger.Warn(ctx, "Failed to store upload session", zap.Error(err))
	}

	// Store session fields for tracking
	session.Metadata["content_type"] = request.ContentType
	session.Metadata["key"] = fileKey

	logger.Info(ctx, "Signed upload URL generated",
		zap.String("session_id", session.ID),
		zap.String("file_key", fileKey),
		zap.String("content_type", request.ContentType),
		zap.Duration("expires_in", expiry),
	)

	return &entities.SignedUploadURL{
		UploadURL: presignedURL.String(),
		AccessURL: presignedURL.String(),
		Fields:    make(map[string]interface{}),
		ExpiresIn: int(expiry.Seconds()),
	}, nil
}

// ConfirmUpload confirms that an upload was completed successfully
func (s *DirectUploadService) ConfirmUpload(ctx context.Context, sessionID, fileKey string) error {
	// Verify file exists in MinIO
	_, err := s.minioClient.StatObject(ctx, s.bucketName, fileKey, minio.StatObjectOptions{})
	if err != nil {
		return fmt.Errorf("upload verification failed: %w", err)
	}

	// Update session status
	session, err := s.getUploadSession(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to get upload session: %w", err)
	}

	now := time.Now()
	session.Status = UploadStatusCompleted
	session.CompletedAt = &now

	if err := s.updateUploadSession(ctx, session); err != nil {
		logger.Warn(ctx, "Failed to update upload session", zap.Error(err))
	}

	logger.Info(ctx, "Upload confirmed successfully",
		zap.String("session_id", sessionID),
		zap.String("file_key", fileKey),
	)

	return nil
}

// GetUploadStatus returns the status of an upload session
func (s *DirectUploadService) GetUploadStatus(ctx context.Context, sessionID string) (*UploadSession, error) {
	session, err := s.getUploadSession(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get upload session: %w", err)
	}

	// Check if session has expired
	if time.Now().After(session.ExpiresAt) && session.Status == UploadStatusPending {
		session.Status = UploadStatusExpired
		s.updateUploadSession(ctx, session)
	}

	return session, nil
}

// GenerateDownloadURL generates a signed URL for downloading a file
func (s *DirectUploadService) GenerateDownloadURL(ctx context.Context, fileKey string, expiry time.Duration) (string, error) {
	presignedURL, err := s.minioClient.PresignedGetObject(
		ctx,
		s.bucketName,
		fileKey,
		expiry,
		nil,
	)
	if err != nil {
		return "", fmt.Errorf("failed to generate download URL: %w", err)
	}

	logger.Debug(ctx, "Download URL generated",
		zap.String("file_key", fileKey),
		zap.Duration("expiry", expiry),
	)

	return presignedURL.String(), nil
}

// DeleteFile deletes a file from MinIO
func (s *DirectUploadService) DeleteFile(ctx context.Context, fileKey string) error {
	err := s.minioClient.RemoveObject(ctx, s.bucketName, fileKey, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	logger.Info(ctx, "File deleted successfully", zap.String("file_key", fileKey))
	return nil
}

// CleanupExpiredUploads removes expired upload sessions and associated files
func (s *DirectUploadService) CleanupExpiredUploads(ctx context.Context) error {
	// This would typically query a database or Redis for expired sessions
	// For now, implement basic MinIO cleanup for old uploads
	
	// List objects with upload prefix that are older than 24 hours
	objectsCh := s.minioClient.ListObjects(ctx, s.bucketName, minio.ListObjectsOptions{
		Prefix: "uploads/",
	})

	var deletedCount int
	for object := range objectsCh {
		if object.Err != nil {
			logger.Error(ctx, "Error listing objects for cleanup", object.Err)
			continue
		}

		// Check if object is older than 24 hours
		if time.Since(object.LastModified) > 24*time.Hour {
			err := s.minioClient.RemoveObject(ctx, s.bucketName, object.Key, minio.RemoveObjectOptions{})
			if err != nil {
				logger.Error(ctx, "Failed to remove expired upload", err, zap.String("key", object.Key))
				continue
			}
			deletedCount++
		}
	}

	logger.Info(ctx, "Cleanup completed", zap.Int("deleted_files", deletedCount))
	return nil
}

// validateUploadRequest validates the upload request
func (s *DirectUploadService) validateUploadRequest(request SignedUploadRequest) error {
	if request.FileName == "" {
		return fmt.Errorf("file name is required")
	}

	if request.ContentType == "" {
		return fmt.Errorf("content type is required")
	}

	if request.Folder == "" {
		return fmt.Errorf("folder is required")
	}

	// Validate file extension matches content type
	ext := strings.ToLower(filepath.Ext(request.FileName))
	if !s.isValidExtension(ext, request.ContentType) {
		return fmt.Errorf("file extension %s does not match content type %s", ext, request.ContentType)
	}

	// Check if content type is allowed
	config := s.getUploadConfig(request.ContentType)
	if !s.isAllowedContentType(request.ContentType, config.AllowedTypes) {
		return fmt.Errorf("content type %s is not allowed", request.ContentType)
	}

	// Validate file size if provided
	if request.FileSize != nil && *request.FileSize > config.MaxFileSize {
		return fmt.Errorf("file size %d exceeds maximum allowed size %d", *request.FileSize, config.MaxFileSize)
	}

	return nil
}

// generateFileKey generates a unique key for the file
func (s *DirectUploadService) generateFileKey(folder, fileName string) string {
	ext := filepath.Ext(fileName)
	name := strings.TrimSuffix(fileName, ext)
	
	// Clean folder name
	folder = strings.Trim(folder, "/")
	
	// Generate unique identifier
	uniqueID := uuid.New().String()
	
	// Create key: folder/originalname_uuid.ext
	return fmt.Sprintf("%s/%s_%s%s", folder, name, uniqueID, ext)
}

// getUploadConfig returns the appropriate upload configuration based on content type
func (s *DirectUploadService) getUploadConfig(contentType string) UploadConfig {
	switch {
	case strings.HasPrefix(contentType, "image/"):
		return DefaultImageConfig
	case strings.HasPrefix(contentType, "video/"):
		return DefaultVideoConfig
	case strings.HasPrefix(contentType, "application/"):
		return DefaultDocumentConfig
	default:
		return DefaultImageConfig // Default fallback
	}
}

// isAllowedContentType checks if content type is in allowed list
func (s *DirectUploadService) isAllowedContentType(contentType string, allowedTypes []string) bool {
	for _, allowed := range allowedTypes {
		if contentType == allowed {
			return true
		}
	}
	return false
}

// isValidExtension validates that file extension matches content type
func (s *DirectUploadService) isValidExtension(ext, contentType string) bool {
	validExtensions := map[string][]string{
		"image/jpeg": {".jpg", ".jpeg"},
		"image/png":  {".png"},
		"image/webp": {".webp"},
		"image/gif":  {".gif"},
		"video/mp4":  {".mp4"},
		"video/webm": {".webm"},
		"video/mov":  {".mov"},
		"video/avi":  {".avi"},
		"application/pdf": {".pdf"},
	}

	extensions, exists := validExtensions[contentType]
	if !exists {
		return true // Allow unknown types
	}

	for _, validExt := range extensions {
		if ext == validExt {
			return true
		}
	}

	return false
}

// Session storage methods (simplified - should use Redis/DB in production)
var sessionStore = make(map[string]*UploadSession)

func (s *DirectUploadService) storeUploadSession(ctx context.Context, session *UploadSession) error {
	sessionStore[session.ID] = session
	return nil
}

func (s *DirectUploadService) getUploadSession(ctx context.Context, sessionID string) (*UploadSession, error) {
	session, exists := sessionStore[sessionID]
	if !exists {
		return nil, fmt.Errorf("upload session not found")
	}
	return session, nil
}

func (s *DirectUploadService) updateUploadSession(ctx context.Context, session *UploadSession) error {
	sessionStore[session.ID] = session
	return nil
}