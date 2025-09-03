package interfaces

import (
	"context"
	"io"
	"net/url"
	"time"
)

type StorageService interface {
	// Presigned URLs for direct upload/download
	GeneratePresignedUploadURL(ctx context.Context, objectKey string, expiration time.Duration, contentType string) (*url.URL, error)
	GeneratePresignedDownloadURL(ctx context.Context, objectKey string, expiration time.Duration) (*url.URL, error)
	
	// File operations
	UploadFile(ctx context.Context, objectKey string, reader io.Reader, size int64, contentType string) error
	DownloadFile(ctx context.Context, objectKey string) (io.ReadCloser, error)
	DeleteFile(ctx context.Context, objectKey string) error
	CopyFile(ctx context.Context, sourceKey, destKey string) error
	MoveFile(ctx context.Context, sourceKey, destKey string) error
	
	// File information
	FileExists(ctx context.Context, objectKey string) (bool, error)
	GetFileInfo(ctx context.Context, objectKey string) (*FileInfo, error)
	GetFileURL(ctx context.Context, objectKey string) string
	
	// Bucket operations
	EnsureBucket(ctx context.Context, bucketName string) error
	ListFiles(ctx context.Context, prefix string, limit int) ([]FileInfo, error)
	
	// Multipart upload for large files
	InitiateMultipartUpload(ctx context.Context, objectKey string, contentType string) (string, error)
	GeneratePresignedPartURL(ctx context.Context, objectKey, uploadID string, partNumber int, expiration time.Duration) (*url.URL, error)
	CompleteMultipartUpload(ctx context.Context, objectKey, uploadID string, parts []CompletePart) error
	AbortMultipartUpload(ctx context.Context, objectKey, uploadID string) error
}

type FileInfo struct {
	Key          string    `json:"key"`
	Size         int64     `json:"size"`
	LastModified time.Time `json:"last_modified"`
	ContentType  string    `json:"content_type"`
	ETag         string    `json:"etag"`
}

type CompletePart struct {
	PartNumber int    `json:"part_number"`
	ETag       string `json:"etag"`
}