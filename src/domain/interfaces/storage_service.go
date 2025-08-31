package interfaces

import (
	"context"
	"io"
	"time"

	"api/domain/entities"
)

type StorageService interface {
	// Direct upload support
	GenerateSignedUploadURL(ctx context.Context, fileName, contentType, folder string) (*entities.SignedUploadURL, error)
	GenerateSignedDownloadURL(ctx context.Context, objectPath string, expiry time.Duration) (string, error)
	
	// File operations
	UploadFile(ctx context.Context, objectPath string, reader io.Reader, size int64, contentType string) error
	DeleteFile(ctx context.Context, objectPath string) error
	FileExists(ctx context.Context, objectPath string) (bool, error)
	GetFileMetadata(ctx context.Context, objectPath string) (*entities.FileMetadata, error)
	
	// Bulk operations
	CreateBulkDownload(ctx context.Context, towerID string) (*entities.BulkDownload, error)
	GetBulkDownloadStatus(ctx context.Context, downloadID string) (*entities.BulkDownloadStatus, error)
}