package storage

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"time"

	"api/domain/entities"
	"api/domain/interfaces"
	"api/infra/config"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type minioService struct {
	client              *minio.Client
	bucketName          string
	baseURL             string
	bulkDownloadService *BulkDownloadService
}

func NewMinioService(cfg *config.MinioConfig) (interfaces.StorageService, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	service := &minioService{
		client:     client,
		bucketName: cfg.BucketName,
		baseURL:    cfg.BaseURL,
	}

	if err := service.ensureBucket(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ensure bucket exists: %w", err)
	}

	return service, nil
}

func (s *minioService) ensureBucket(ctx context.Context) error {
	exists, err := s.client.BucketExists(ctx, s.bucketName)
	if err != nil {
		return err
	}

	if !exists {
		return s.client.MakeBucket(ctx, s.bucketName, minio.MakeBucketOptions{})
	}
	return nil
}

func (s *minioService) GenerateSignedUploadURL(ctx context.Context, fileName, contentType, folder string) (*entities.SignedUploadURL, error) {
	objectPath := s.buildObjectPath(folder, fileName)
	
	reqParams := make(url.Values)
	reqParams.Set("response-content-disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileName))
	if contentType != "" {
		reqParams.Set("response-content-type", contentType)
	}

	uploadURL, err := s.client.PresignedPutObject(ctx, s.bucketName, objectPath, 15*time.Minute)
	if err != nil {
		return nil, fmt.Errorf("failed to generate signed upload URL: %w", err)
	}

	accessURL := s.buildAccessURL(objectPath)

	return &entities.SignedUploadURL{
		UploadURL: uploadURL.String(),
		AccessURL: accessURL,
		ExpiresIn: 15 * 60,
		Fields:    make(map[string]interface{}),
	}, nil
}

func (s *minioService) GenerateSignedDownloadURL(ctx context.Context, objectPath string, expiry time.Duration) (string, error) {
	downloadURL, err := s.client.PresignedGetObject(ctx, s.bucketName, objectPath, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate signed download URL: %w", err)
	}
	return downloadURL.String(), nil
}

func (s *minioService) UploadFile(ctx context.Context, objectPath string, reader io.Reader, size int64, contentType string) error {
	options := minio.PutObjectOptions{
		ContentType: contentType,
	}
	if size > 0 {
		options.PartSize = uint64(size)
	}

	_, err := s.client.PutObject(ctx, s.bucketName, objectPath, reader, size, options)
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}
	return nil
}

func (s *minioService) DeleteFile(ctx context.Context, objectPath string) error {
	err := s.client.RemoveObject(ctx, s.bucketName, objectPath, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

func (s *minioService) FileExists(ctx context.Context, objectPath string) (bool, error) {
	_, err := s.client.StatObject(ctx, s.bucketName, objectPath, minio.StatObjectOptions{})
	if err != nil {
		if minio.ToErrorResponse(err).Code == "NoSuchKey" {
			return false, nil
		}
		return false, fmt.Errorf("failed to check if file exists: %w", err)
	}
	return true, nil
}

func (s *minioService) GetFileMetadata(ctx context.Context, objectPath string) (*entities.FileMetadata, error) {
	objInfo, err := s.client.StatObject(ctx, s.bucketName, objectPath, minio.StatObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get file metadata: %w", err)
	}

	return &entities.FileMetadata{
		FileName:    extractFileName(objectPath),
		FileSize:    objInfo.Size,
		ContentType: objInfo.ContentType,
		UploadedAt:  objInfo.LastModified,
		Checksum:    &objInfo.ETag,
	}, nil
}

func (s *minioService) CreateBulkDownload(ctx context.Context, towerID string) (*entities.BulkDownload, error) {
	if s.bulkDownloadService == nil {
		return nil, fmt.Errorf("bulk download service not configured")
	}
	
	result, err := s.bulkDownloadService.GenerateTowerDownload(ctx, towerID)
	if err != nil {
		return nil, err
	}
	
	return &entities.BulkDownload{
		DownloadURL: result.DownloadURL,
		FileName: result.FileName,
		FileSize: result.FileSize,
		ExpiresIn: result.ExpiresIn,
		CreatedAt: result.CreatedAt,
	}, nil
}

func (s *minioService) GetBulkDownloadStatus(ctx context.Context, downloadID string) (*entities.BulkDownloadStatus, error) {
	return &entities.BulkDownloadStatus{
		ID: downloadID,
		Status: "COMPLETED",
		Progress: 100,
	}, nil
}

func (s *minioService) buildObjectPath(folder, fileName string) string {
	if folder == "" {
		return fileName
	}
	return fmt.Sprintf("%s/%s", folder, fileName)
}

func (s *minioService) buildAccessURL(objectPath string) string {
	return fmt.Sprintf("%s/%s/%s", s.baseURL, s.bucketName, objectPath)
}

func extractFileName(objectPath string) string {
	for i := len(objectPath) - 1; i >= 0; i-- {
		if objectPath[i] == '/' {
			return objectPath[i+1:]
		}
	}
	return objectPath
}