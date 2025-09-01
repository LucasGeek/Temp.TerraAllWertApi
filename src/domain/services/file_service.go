package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"time"

	"api/domain/entities"
	"api/infra/storage"
	"github.com/google/uuid"
)

// FileService gerencia operações de arquivo
type FileService struct {
	storage *storage.MinIOStorage
}

// NewFileService cria uma nova instância do serviço de arquivos
func NewFileService(storage *storage.MinIOStorage) *FileService {
	return &FileService{
		storage: storage,
	}
}

// GetSignedUploadURL gera uma URL assinada para upload
func (s *FileService) GetSignedUploadURL(ctx context.Context, input *entities.SignedUploadURLInput) (*entities.SignedUploadURLResponse, error) {
	fileID := uuid.New().String()
	
	// Construir caminho no MinIO
	minioPath := fmt.Sprintf("%s/%s/%s-%s", 
		input.RouteID, 
		input.FileType,
		fileID,
		input.FileName,
	)
	
	// Gerar URL assinada para upload (válida por 1 hora)
	uploadURL, err := s.storage.GetPresignedUploadURL(ctx, minioPath, input.ContentType, 3600)
	if err != nil {
		return nil, fmt.Errorf("erro ao gerar URL de upload: %w", err)
	}
	
	expiresAt := time.Now().Add(1 * time.Hour)
	
	return &entities.SignedUploadURLResponse{
		UploadURL: uploadURL,
		MinioPath: minioPath,
		ExpiresAt: expiresAt,
		FileID:    fileID,
	}, nil
}

// ConfirmFileUpload confirma que o upload foi realizado
func (s *FileService) ConfirmFileUpload(ctx context.Context, input *entities.ConfirmFileUploadInput) (*entities.ConfirmFileUploadResponse, error) {
	// Verificar se o arquivo existe no MinIO
	exists, err := s.storage.FileExists(ctx, input.MinioPath)
	if err != nil {
		return nil, fmt.Errorf("erro ao verificar arquivo: %w", err)
	}
	
	if !exists {
		return &entities.ConfirmFileUploadResponse{
			Success: false,
		}, nil
	}
	
	// Obter URL pública do arquivo
	publicURL := s.storage.GetPublicURL(input.MinioPath)
	
	// Gerar URL de download assinada (válida por 24 horas)
	downloadURL, err := s.storage.GetPresignedDownloadURL(ctx, input.MinioPath, 86400)
	if err != nil {
		return nil, fmt.Errorf("erro ao gerar URL de download: %w", err)
	}
	
	// Gerar thumbnail se for imagem
	var thumbnailURL string
	if input.FileType == "image" {
		thumbnailPath := fmt.Sprintf("thumbnails/%s", input.MinioPath)
		// TODO: Implementar geração de thumbnail
		thumbnailURL = s.storage.GetPublicURL(thumbnailPath)
	}
	
	metadata := &entities.FileMetadataExtended{
		ID:           input.FileID,
		URL:          publicURL,
		DownloadURL:  downloadURL,
		ThumbnailURL: thumbnailURL,
		Metadata: map[string]interface{}{
			"originalFileName": input.OriginalFileName,
			"fileSize":         input.FileSize,
			"checksum":         input.Checksum,
			"uploadedAt":       time.Now(),
		},
	}
	
	return &entities.ConfirmFileUploadResponse{
		Success:      true,
		FileMetadata: metadata,
	}, nil
}

// GetSignedDownloadURLs obtém URLs de download para múltiplos arquivos
func (s *FileService) GetSignedDownloadURLs(ctx context.Context, input *entities.SignedDownloadURLsInput) (*entities.SignedDownloadURLsResponse, error) {
	urls := make([]*entities.FileDownloadURL, 0, len(input.FileIDs))
	
	expirationSeconds := int64(input.ExpirationMinutes * 60)
	if expirationSeconds == 0 {
		expirationSeconds = 3600 // 1 hora por padrão
	}
	
	for _, fileID := range input.FileIDs {
		// Buscar caminho do arquivo baseado no fileID
		// TODO: Implementar busca em banco de dados
		minioPath := fmt.Sprintf("%s/%s", input.RouteID, fileID)
		
		downloadURL, err := s.storage.GetPresignedDownloadURL(ctx, minioPath, expirationSeconds)
		if err != nil {
			continue // Pular arquivos com erro
		}
		
		urls = append(urls, &entities.FileDownloadURL{
			FileID:      fileID,
			DownloadURL: downloadURL,
			ExpiresAt:   time.Now().Add(time.Duration(expirationSeconds) * time.Second),
		})
	}
	
	return &entities.SignedDownloadURLsResponse{
		URLs: urls,
	}, nil
}

// CalculateChecksum calcula o SHA256 de um arquivo
func (s *FileService) CalculateChecksum(reader io.Reader) (string, error) {
	hash := sha256.New()
	if _, err := io.Copy(hash, reader); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}