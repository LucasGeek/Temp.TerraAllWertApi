package storage

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"time"

	"api/domain/interfaces"
	"api/infra/logger"

	"github.com/minio/minio-go/v7"
	"go.uber.org/zap"
)

type BulkDownloadService struct {
	minioClient      *minio.Client
	towerRepo        interfaces.TowerRepository
	apartmentRepo    interfaces.ApartmentRepository
	galleryRepo      interfaces.GalleryRepository
	bucketName       string
	tempBucketName   string
}

type BulkDownloadResult struct {
	DownloadURL  string    `json:"downloadUrl"`
	FileName     string    `json:"fileName"`
	FileSize     int64     `json:"fileSize"`
	ExpiresIn    int       `json:"expiresIn"`
	CreatedAt    time.Time `json:"createdAt"`
}

func NewBulkDownloadService(
	minioClient *minio.Client,
	towerRepo interfaces.TowerRepository,
	apartmentRepo interfaces.ApartmentRepository,
	galleryRepo interfaces.GalleryRepository,
	bucketName, tempBucketName string,
) *BulkDownloadService {
	return &BulkDownloadService{
		minioClient:    minioClient,
		towerRepo:     towerRepo,
		apartmentRepo: apartmentRepo,
		galleryRepo:   galleryRepo,
		bucketName:    bucketName,
		tempBucketName: tempBucketName,
	}
}

func (s *BulkDownloadService) GenerateTowerDownload(ctx context.Context, towerID string) (*BulkDownloadResult, error) {
	// Get tower and all related data
	tower, err := s.towerRepo.GetByID(ctx, towerID)
	if err != nil {
		logger.Error(ctx, "Failed to get tower", err, zap.String("tower_id", towerID))
		return nil, err
	}

	// Get all apartments for the tower
	apartments, err := s.apartmentRepo.GetByTowerID(ctx, towerID)
	if err != nil {
		logger.Error(ctx, "Failed to get apartments", err, zap.String("tower_id", towerID))
		return nil, err
	}

	// Create temporary zip file in MinIO
	zipFileName := fmt.Sprintf("tower_%s_%d.zip", towerID, time.Now().Unix())
	zipObjectName := fmt.Sprintf("downloads/%s", zipFileName)

	// Create a pipe for streaming zip data to MinIO
	reader, writer := io.Pipe()

	// Start zip creation in a goroutine
	go s.createTowerZip(ctx, writer, tower.Name, apartments)

	// Upload zip to MinIO
	uploadInfo, err := s.minioClient.PutObject(
		ctx,
		s.tempBucketName,
		zipObjectName,
		reader,
		-1, // unknown size
		minio.PutObjectOptions{
			ContentType: "application/zip",
		},
	)
	if err != nil {
		logger.Error(ctx, "Failed to upload zip", err, zap.String("object", zipObjectName))
		return nil, err
	}

	// Generate presigned download URL (expires in 1 hour)
	expiry := time.Hour
	downloadURL, err := s.minioClient.PresignedGetObject(
		ctx,
		s.tempBucketName,
		zipObjectName,
		expiry,
		nil,
	)
	if err != nil {
		logger.Error(ctx, "Failed to generate download URL", err)
		return nil, err
	}

	logger.Info(ctx, "Bulk download created", 
		zap.String("tower_id", towerID),
		zap.String("file_name", zipFileName),
		zap.Int64("file_size", uploadInfo.Size),
	)

	return &BulkDownloadResult{
		DownloadURL: downloadURL.String(),
		FileName:    zipFileName,
		FileSize:    uploadInfo.Size,
		ExpiresIn:   int(expiry.Seconds()),
		CreatedAt:   time.Now(),
	}, nil
}

func (s *BulkDownloadService) createTowerZip(ctx context.Context, writer *io.PipeWriter, towerName string, apartments []interface{}) {
	defer writer.Close()

	zipWriter := zip.NewWriter(writer)
	defer zipWriter.Close()

	// Create tower folder structure
	towerFolder := fmt.Sprintf("%s/", towerName)
	
	// Add tower info file
	infoFile, err := zipWriter.Create(towerFolder + "tower_info.txt")
	if err != nil {
		logger.Error(ctx, "Failed to create info file", err)
		return
	}

	towerInfo := fmt.Sprintf("Tower: %s\nGenerated: %s\nTotal Apartments: %d\n",
		towerName, time.Now().Format("2006-01-02 15:04:05"), len(apartments))
	infoFile.Write([]byte(towerInfo))

	// Process each apartment (this is a simplified example)
	for i, apt := range apartments {
		apartmentFolder := fmt.Sprintf("%s/apartment_%d/", towerName, i+1)
		
		// Create apartment folder
		aptFile, err := zipWriter.Create(apartmentFolder + "info.txt")
		if err != nil {
			continue
		}
		
		aptInfo := fmt.Sprintf("Apartment #%d\nProcessed: %s\n", i+1, time.Now().Format("2006-01-02 15:04:05"))
		aptFile.Write([]byte(aptInfo))

		// Here you would add logic to download and include:
		// - Apartment images
		// - Floor plans
		// - Documents
		// This would require additional service methods
	}

	logger.Info(ctx, "Zip creation completed", zap.String("tower", towerName))
}

func (s *BulkDownloadService) downloadFileToZip(ctx context.Context, zipWriter *zip.Writer, objectName, zipPath string) error {
	// Get object from MinIO
	object, err := s.minioClient.GetObject(ctx, s.bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return err
	}
	defer object.Close()

	// Create file in zip
	zipFile, err := zipWriter.Create(zipPath)
	if err != nil {
		return err
	}

	// Copy file content to zip
	_, err = io.Copy(zipFile, object)
	return err
}

// CleanupExpiredDownloads removes old download files
func (s *BulkDownloadService) CleanupExpiredDownloads(ctx context.Context) error {
	// List objects in downloads folder
	objectsCh := s.minioClient.ListObjects(ctx, s.tempBucketName, minio.ListObjectsOptions{
		Prefix: "downloads/",
	})

	for object := range objectsCh {
		if object.Err != nil {
			logger.Error(ctx, "Error listing objects", object.Err)
			continue
		}

		// Check if file is older than 24 hours
		if time.Since(object.LastModified) > 24*time.Hour {
			err := s.minioClient.RemoveObject(ctx, s.tempBucketName, object.Key, minio.RemoveObjectOptions{})
			if err != nil {
				logger.Error(ctx, "Failed to remove expired file", err, zap.String("object", object.Key))
			} else {
				logger.Info(ctx, "Removed expired download", zap.String("object", object.Key))
			}
		}
	}

	return nil
}