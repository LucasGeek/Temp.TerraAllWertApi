package storage

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"api/domain/entities"
	"api/domain/interfaces"
	"api/infra/logger"

	"github.com/minio/minio-go/v7"
	"go.uber.org/zap"
)

type BulkDownloadService struct {
	minioClient    *minio.Client
	towerRepo      interfaces.TowerRepository
	apartmentRepo  interfaces.ApartmentRepository
	galleryRepo    interfaces.GalleryRepository
	bucketName     string
	tempBucketName string
}

type BulkDownloadResult struct {
	DownloadURL string    `json:"downloadUrl"`
	FileName    string    `json:"fileName"`
	FileSize    int64     `json:"fileSize"`
	ExpiresIn   int       `json:"expiresIn"`
	CreatedAt   time.Time `json:"createdAt"`
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
		towerRepo:      towerRepo,
		apartmentRepo:  apartmentRepo,
		galleryRepo:    galleryRepo,
		bucketName:     bucketName,
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

func (s *BulkDownloadService) createTowerZip(ctx context.Context, writer *io.PipeWriter, towerName string, apartments []*entities.Apartment) {
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

	// Process each apartment with complete file inclusion
	for _, apartment := range apartments {
		if apartment == nil {
			continue
		}

		apartmentFolder := fmt.Sprintf("%s/apartment_%s/", towerName, apartment.Number)

		// Create apartment info file with detailed information
		aptFile, err := zipWriter.Create(apartmentFolder + "info.txt")
		if err != nil {
			logger.Error(ctx, "Failed to create apartment info file", err, zap.String("apartment", apartment.Number))
			continue
		}

		aptInfo := s.generateApartmentInfo(apartment)
		aptFile.Write([]byte(aptInfo))

		// Include apartment images
		if err := s.includeApartmentImages(ctx, zipWriter, apartment, apartmentFolder+"images/"); err != nil {
			logger.Warn(ctx, "Failed to include apartment images", zap.Error(err), zap.String("apartment", apartment.Number))
		}

		// Include floor plan if available
		if apartment.FloorPlanURL != nil {
			if err := s.includeFloorPlan(ctx, zipWriter, apartment, apartmentFolder); err != nil {
				logger.Warn(ctx, "Failed to include floor plan", zap.Error(err), zap.String("apartment", apartment.Number))
			}
		}

		// Include main image if available
		if apartment.MainImageURL != nil {
			if err := s.includeMainImage(ctx, zipWriter, apartment, apartmentFolder); err != nil {
				logger.Warn(ctx, "Failed to include main image", zap.Error(err), zap.String("apartment", apartment.Number))
			}
		}
	}

	logger.Info(ctx, "Zip creation completed", zap.String("tower", towerName))
}

// generateApartmentInfo creates detailed apartment information
func (s *BulkDownloadService) generateApartmentInfo(apartment *entities.Apartment) string {
	info := fmt.Sprintf("Apartment Information\n")
	info += fmt.Sprintf("=====================\n")
	info += fmt.Sprintf("Number: %s\n", apartment.Number)
	info += fmt.Sprintf("ID: %s\n", apartment.ID)
	info += fmt.Sprintf("Floor ID: %s\n", apartment.FloorID)
	info += fmt.Sprintf("Status: %s\n", apartment.Status)
	info += fmt.Sprintf("Available: %t\n", apartment.Available)
	
	if apartment.Area != nil {
		info += fmt.Sprintf("Area: %s\n", *apartment.Area)
	}
	if apartment.Bedrooms != nil {
		info += fmt.Sprintf("Bedrooms: %d\n", *apartment.Bedrooms)
	}
	if apartment.Suites != nil {
		info += fmt.Sprintf("Suites: %d\n", *apartment.Suites)
	}
	if apartment.ParkingSpots != nil {
		info += fmt.Sprintf("Parking Spots: %d\n", *apartment.ParkingSpots)
	}
	if apartment.Price != nil {
		info += fmt.Sprintf("Price: %.2f\n", *apartment.Price)
	}
	if apartment.SolarPosition != nil {
		info += fmt.Sprintf("Solar Position: %s\n", *apartment.SolarPosition)
	}
	
	info += fmt.Sprintf("Created At: %s\n", apartment.CreatedAt.Format("2006-01-02 15:04:05"))
	info += fmt.Sprintf("Updated At: %s\n", apartment.UpdatedAt.Format("2006-01-02 15:04:05"))
	info += fmt.Sprintf("\nGenerated: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	
	return info
}

// includeApartmentImages downloads and includes all apartment images
func (s *BulkDownloadService) includeApartmentImages(ctx context.Context, zipWriter *zip.Writer, apartment *entities.Apartment, folder string) error {
	// Include images from apartment.Images slice
	for i, image := range apartment.Images {
		imageName := fmt.Sprintf("image_%d_%s", i+1, filepath.Base(image.ImageURL))
		imagePath := folder + imageName
		
		if err := s.downloadFileToZip(ctx, zipWriter, image.ImageURL, imagePath); err != nil {
			logger.Warn(ctx, "Failed to include apartment image", 
				zap.Error(err), 
				zap.String("image_url", image.ImageURL),
				zap.String("apartment", apartment.Number))
			continue
		}
		
		// Create image info file
		infoPath := folder + fmt.Sprintf("image_%d_info.txt", i+1)
		infoFile, err := zipWriter.Create(infoPath)
		if err != nil {
			continue
		}
		
		imageInfo := fmt.Sprintf("Image Information\n")
		imageInfo += fmt.Sprintf("================\n")
		imageInfo += fmt.Sprintf("Original URL: %s\n", image.ImageURL)
		imageInfo += fmt.Sprintf("Order: %d\n", image.Order)
		if image.Description != nil {
			imageInfo += fmt.Sprintf("Description: %s\n", *image.Description)
		}
		imageInfo += fmt.Sprintf("Created At: %s\n", image.CreatedAt.Format("2006-01-02 15:04:05"))
		
		infoFile.Write([]byte(imageInfo))
	}
	
	return nil
}

// includeFloorPlan downloads and includes the apartment floor plan
func (s *BulkDownloadService) includeFloorPlan(ctx context.Context, zipWriter *zip.Writer, apartment *entities.Apartment, folder string) error {
	if apartment.FloorPlanURL == nil {
		return nil
	}
	
	planName := "floor_plan" + filepath.Ext(*apartment.FloorPlanURL)
	planPath := folder + planName
	
	return s.downloadFileToZip(ctx, zipWriter, *apartment.FloorPlanURL, planPath)
}

// includeMainImage downloads and includes the apartment main image
func (s *BulkDownloadService) includeMainImage(ctx context.Context, zipWriter *zip.Writer, apartment *entities.Apartment, folder string) error {
	if apartment.MainImageURL == nil {
		return nil
	}
	
	imageName := "main_image" + filepath.Ext(*apartment.MainImageURL)
	imagePath := folder + imageName
	
	return s.downloadFileToZip(ctx, zipWriter, *apartment.MainImageURL, imagePath)
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
